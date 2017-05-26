package control

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/darkonie/wikiracer/primitives"
	"github.com/darkonie/wikiracer/worker"
	"github.com/sirupsen/logrus"
)

// define the job status codes
const (
	// PageFound status is used when the destination page found.
	PageFound = iota

	// Running status is used when the job is started.
	Running
	// Cancelled status is used to indicate that the user cancelled the job.
	Cancelled

	// Unchanged initial job state.
	Unchanged
)

// NewJob returns a new job structure.
func NewJob(startLink, endLink, comment, id, crawlerType string, timeout time.Duration, workers int, client *http.Client) *Job {

	// default to 100 workers
	jobWorkers := 100
	if workers > 0 && workers <= 1000 {
		jobWorkers = workers
	}

	dequeueChan := make(chan interface{})
	j := &Job{
		Status:    Unchanged,
		Comment:   comment,
		StartLink: startLink,
		EndLink:   endLink,
		Timeout:   timeout.String(),
		Workers:   jobWorkers,

		dequeueChan: dequeueChan,
		resultChan:  make(chan *worker.Page),
		client:      client,
		cancel:      func() {},
		id:          id,
	}

	d := &JobDuration{
		t1: &j.StartTime,
		t2: &j.EndTime,
	}

	j.Duration = d

	// crawler can be [api, html]
	if crawlerType == "html" {
		j.newWorker = func() worker.WikiCrawler {
			return worker.NewHTMLWikiCrawler(client)
		}
	} else {
		j.newWorker = func() worker.WikiCrawler {
			return worker.NewAPIWikiCrawler(client)
		}
	}

	return j
}

// Job provides control over wikiracing.
type Job struct {
	sync.Mutex

	q primitives.Q

	dequeueChan chan interface{}
	resultChan chan *worker.Page
	client      *http.Client

	newWorker func() worker.WikiCrawler

	cancel context.CancelFunc
	id     string

	Path      []string    `json:"path"`
	IsRunning bool        `json:"is_running"`
	StartLink string      `json:"start_link"`
	EndLink   string      `json:"end_link"`
	Status    int         `json:"status"`
	Comment   string      `json:"comment"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Timeout   string      `json:"timeout"`
	Errors    []string    `json:"errors"`
	Workers   int         `json:"workers"`

	// stats
	Duration       *JobDuration `json:"duration"`
	PagesVisited   uint64       `json:"pages_visited"`
	Depth          int          `json:"depth"`
}

func (j *Job) updateJobDepth(page *worker.Page) {
	j.Lock()
	defer j.Unlock()
	if page.Depth > j.Depth {
		j.Depth = page.Depth
	}
}

func (j *Job) addError(err error) {
	j.Lock()
	defer j.Unlock()

	j.Errors = append(j.Errors, err.Error())
}

// Start a new job
func (j *Job) Start(ctx context.Context, cancel context.CancelFunc) error {
	j.Lock()
	defer j.Unlock()
	if j.IsRunning {
		return errors.New("job is already running")
	}

	j.cancel = cancel
	j.q = primitives.NewPQueue(ctx, j.dequeueChan)
	j.IsRunning = true
	j.Status = Running
	j.StartTime = time.Now()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case page := <- j.resultChan:
				j.updateJobDepth(page)
				if page.Name == j.EndLink {
					j.updatePath(page)
					j.Stop(PageFound)
					return
				}

				if _, ok := page.Links[j.EndLink]; ok {
					j.updatePath(&worker.Page{Name:j.EndLink, Prev: page})
					j.Stop(PageFound)
					return
				}

				depth := page.Depth + 1
				for link := range page.Links {
					newPage := &worker.Page{Name:link, Prev: page, Depth: depth}
					j.q.Enqueue(newPage, depth)
				}
			}
		}
	}()

	// submit start page.
	j.q.Enqueue(&worker.Page{Name: j.StartLink}, 1)
	go j.start(ctx)
	return nil
}

func (j *Job)updatePath(page *worker.Page) {
	var path []string
	for p := page; p != nil; p = p.Prev {
		path = append(path, string(p.Name))
	}

	// reverse slice of strings in go :)
	reversePath := func(path []string) []string {
		for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
			path[i], path[j] = path[j], path[i]
		}
		return path
	}

	j.Path = reversePath(path)
}

func (j *Job) start(ctx context.Context) {
	visit := &visitedMap{
		m: make(map[string]bool),
	}

	for i := 0; i < j.Workers; i++ {
		go func() {
			w := j.newWorker()
			for {
				select {
				case <-ctx.Done():
					return

				case item := <-j.dequeueChan:
					j.PagesVisited = visit.len()
					pair, ok := item.(*primitives.Pair)
					if !ok {
						logrus.Errorf("received item is not a Pair. Got %+v", item)
						break
					}

					req, ok := pair.Item.(*worker.Page)
					if !ok {
						logrus.Errorf("received item is not a string. Got %+v", item)
						break
					}

					if visit.visited(string(req.Name)) {
						continue
					}

					page, err := w.Fetch(ctx, req.Name)
					if err != nil {
						logrus.Error(err)
						continue
					}

					req.Depth = pair.Priority
					req.Links = page.Links

					// make sure we don't block if
					select {
					case j.resultChan <- req:
					case <-time.After(time.Second):
					}
				}
			}
		}()
	}
}

// Stop a job in progress
func (j *Job) Stop(reason int) error {
	j.Lock()
	defer j.Unlock()
	if !j.IsRunning {
		return errors.New("job is not running")
	}
	j.cancel()
	j.IsRunning = false
	j.EndTime = time.Now()
	j.Status = reason
	return nil
}
