package control

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NewJobPoolManager creates a new instance of JobPoolManager.
func NewJobPoolManager() *JobPoolManager {
	return &JobPoolManager{
		Pool: make(map[string]*Job),
		client: &http.Client{
			//Timeout: time.Second*10,
			Transport: &http.Transport{
			//TLSHandshakeTimeout: time.Second*10,
			},
		},
	}
}

// JobPoolManager represents a pool of jobs.
type JobPoolManager struct {
	sync.RWMutex

	Pool map[string]*Job `json:"pool"`

	client *http.Client
}

// AddJob adds a new job to a pool.
func (jp *JobPoolManager) AddJob(startLink, endLink, comment, crawlerType string, timeout time.Duration, workers int) (string, error) {
	jp.Lock()
	defer jp.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	// assuming id is unique
	jp.Pool[id.String()] = NewJob(startLink, endLink, comment, id.String(), crawlerType, timeout, workers, jp.client)
	return id.String(), nil
}

// GetJob returns a job from a pool.
func (jp *JobPoolManager) GetJob(id string) (*Job, bool) {
	jp.RLock()
	defer jp.RUnlock()
	j, ok := jp.Pool[id]
	return j, ok
}

// StartJob starts a job from a pool.
func (jp *JobPoolManager) StartJob(ctx context.Context, cancel context.CancelFunc, id string) error {
	job, ok := jp.GetJob(id)
	if !ok {
		return fmt.Errorf("job %s does not exist in the pool", id)
	}

	return job.Start(ctx, cancel)
}
