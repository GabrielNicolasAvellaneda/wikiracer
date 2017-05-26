package control

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/darkonie/wikiracer/primitives"
	"github.com/darkonie/wikiracer/worker"
	"github.com/sirupsen/logrus"
)

// define the job status codes
const (
	// PageFound status is used when the destination page found.
	PageFound = iota

	// Cancelled status is used to indicate that the user cancelled the job.
	Cancelled

	// Timedout status is used when job reached timeout.
	Timedout

	// Unchanged initial job state.
	Unchanged
)

// NewJob returns a new job structure.
func NewJob(startLink, endLink, comment, id, crawlerType string, timeout time.Duration, workers int, client *http.Client) *Job {

	dequeueChan := make(chan interface{})
	j := &Job{
		Status:    Unchanged,
		Comment:   comment,
		StartLink: worker.Link(startLink),
		EndLink:   worker.Link(endLink),
		Timeout:   timeout.String(),
		Workers:   workers,

		g:           primitives.NewGraph(),
		dequeueChan: dequeueChan,
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

	g primitives.Graph
	q primitives.Q

	dequeueChan chan interface{}
	client      *http.Client

	newWorker func() worker.WikiCrawler

	cancel context.CancelFunc
	id     string

	Path      []string    `json:"path,omitempty"`
	IsRunning bool        `json:"is_running"`
	StartLink worker.Link `json:"start_link"`
	EndLink   worker.Link `json:"end_link"`
	Status    int         `json:"status"`
	Comment   string      `json:"comment"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
	Timeout   string      `json:"timeout"`
	Errors    []string    `json:"errors"`
	Workers   int         `json:"workers"`

	// stats
	Duration       *JobDuration `json:"duration"`
	PagesProcessed uint64       `json:"pages_processed"`
	PagesVisited   uint64       `json:"pages_visited"`
}

// Start a new job
func (j *Job) Start(ctx context.Context, cancel context.CancelFunc) error {
	j.Lock()
	defer j.Unlock()
	if j.IsRunning {
		return errors.New("job is already running")
	}

	j.cancel = cancel

	// start an iter queue with context
	j.q = primitives.NewWatchIterQueue(ctx, j.dequeueChan)
	j.IsRunning = true
	j.StartTime = time.Now()

	j.q.Enqueue(&worker.Page{Name: j.StartLink})
	go j.start(ctx)
	return nil
}

// process page takes a page adds NEW nodes to graph, excludes already added nodes and returns a new sanitized page.
func (j *Job) processPage(page *worker.Page) {
	j.Lock()
	defer j.Unlock()

	name := string(page.Name)
	j.g.GAddNode(name)

	for link := range page.Links {
		linkStr := string(link)

		j.g.GAddNode(linkStr)
		err := j.g.GAddEdge(name, linkStr, 1)
		if err != nil {
			logrus.Errorf("unable to add edge between %s and %s. %s", name, linkStr, err)
			continue
		}
	}
}

func (j *Job) addError(err error) {
	j.Lock()
	defer j.Unlock()

	j.Errors = append(j.Errors, err.Error())
}

func (j *Job) start(ctx context.Context) {
	visit := &visitedMap{
		m: make(map[string]bool),
	}

	done := make(chan struct{})


	// default to 100 workers
	workers := 100
	if j.Workers > 1 && j.Workers < 1000 {
		workers = j.Workers
	}

	for i := 0; i < workers; i++ {
		go func() {
			w := j.newWorker()
			for {
				select {
				case <-ctx.Done():
					j.Stop(Timedout)
					return

				case <- done:
					return

				case item := <-j.dequeueChan:
					j.PagesVisited = visit.len()

					page, ok := item.(*worker.Page)
					if !ok {
						logrus.Errorf("received item is not page. Got %+v", item)
						break
					}

					j.processPage(page)
					if page.Has(j.EndLink) {
						j.Stop(PageFound)
						return
					}

					for _, l := range j.g.GBFS(string(page.Name)) {
						if l == "" {
							continue
						}
						if visit.visited(l) {
							continue
						}

						fetchedPage, err := w.Fetch(ctx, worker.Link(l))
						if err != nil {
							if err != context.Canceled {
								logrus.Error(err)
								j.addError(err)
							}
							continue
						}
						atomic.AddUint64(&j.PagesProcessed, 1)
						j.q.Enqueue(fetchedPage)
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
	if reason == PageFound {
		var err error
		j.Path, err = j.g.GFindPath(string(j.StartLink), string(j.EndLink))
		return err
	}
	logrus.Infof("stopping job %s with status %d", j.id, reason)

	return nil
}
