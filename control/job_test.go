package control

import (
	"context"
	"fmt"

	"github.com/darkonie/wikiracer/worker"
	"testing"
	"time"
	"strings"
)

type fakeCrawler struct {

}


func (f fakeCrawler) Fetch(ctx context.Context, link string) (*worker.Page, error) {
	switch string(link) {
	case "Mike Tyson":
		return &worker.Page{
			Name: "Mike Tyson",
			Links: map[string]bool{"AAA": true},
		}, nil
	case "AAA":
		return &worker.Page{
			Name: "AAA",
			Links: map[string]bool{"BBB": true},
		}, nil
	case "BBB":
		return &worker.Page{
			Name: "BBB",
			Links: map[string]bool{"Ukraine": true},
		}, nil

	default:
	}

	return nil, fmt.Errorf("%s not found", link)
}

func TestNewJob(t *testing.T) {
	job := NewJob("Mike Tyson", "Ukraine", "My comment", "123", "",
		time.Second, 10, nil)

	// dirty hack
	job.newWorker = func() worker.WikiCrawler {
		return &fakeCrawler{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	job.Start(ctx, cancel)
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for context")
	}

	expected := "Mike Tyson_AAA_BBB_Ukraine"
	result := strings.Join(job.Path, "_")
	if result != expected {
		t.Fatalf("expect %s. Got %s", expected, result)
	}
}