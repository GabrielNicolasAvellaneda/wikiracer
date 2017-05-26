package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/darkonie/wikiracer/control"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// request is a structure for user to submit requests
type request struct {
	Timeout         string `json:"timeout"`
	StartPage       string `json:"start_page"`
	DestinationPage string `json:"destination_page"`
	Comment         string `json:"comment"`
	Workers         int    `json:"workers"`
	CrawlMethod     string `json:"crawl_method"`
}

// response is structure used to send back user status.
type response struct {
	ID  string `json:"id"`
	Msg string `json:"msg"`
}

func jobInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jobs, err := jobsFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(jobs)
	if err != nil {
		http.Error(w, "unable to encode job info", http.StatusInternalServerError)
		return
	}
}

// TODO: implement this
func jobInfoSSEHandler(w http.ResponseWriter, r *http.Request) {
	jpManager, ok := jpManagerFromContext(r.Context())
	if !ok {
		http.Error(w, "unable to get a job from context", http.StatusInternalServerError)
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "unable to get id from request", http.StatusInternalServerError)
		return
	}

	job, ok := jpManager.GetJob(id)
	if !ok {
		http.Error(w, "job not found "+id, http.StatusBadRequest)
		return
	}

	// Set response headers.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")
	f := w.(http.Flusher)
	notify := w.(http.CloseNotifier).CloseNotify()
	f.Flush()

	for {
		select {
		case <-notify:
			return
		case <-time.After(time.Second):
			body, err := json.Marshal(job.Duration)
			if err != nil {
				logrus.Errorf("error marshalling sse response: %s", err)
				continue
			}
			io.Copy(w, bytes.NewReader(body))
			f.Flush()
		}
	}
}

func jobStartHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jpManager, ok := jpManagerFromContext(r.Context())
	if !ok {
		http.Error(w, "unable to get a job from context", http.StatusInternalServerError)
		return
	}

	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.StartPage == "" || req.DestinationPage == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		logrus.Errorf("error parsing timeout %s. Using default timeout 1 min", req.Timeout)
		timeout = time.Duration(time.Minute)
	}

	id, err := jpManager.AddJob(req.StartPage, req.DestinationPage, req.Comment, req.CrawlMethod, timeout, req.Workers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	err = jpManager.StartJob(ctx, cancel, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&response{
		ID:  id,
		Msg: "successfully started a new job",
	}); err != nil {
		logrus.Errorf("error encoding response: %s", err)
	}

}

func jobCancelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jpManager, ok := jpManagerFromContext(r.Context())
	if !ok {
		http.Error(w, "unable to get a job manager from context", http.StatusInternalServerError)
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		http.Error(w, "unable to get id from request", http.StatusInternalServerError)
		return
	}

	job, ok := jpManager.GetJob(id)
	if !ok {
		http.Error(w, "job not found "+id, http.StatusBadRequest)
		return
	}

	err := job.Stop(control.Cancelled)
	if err != nil {
		logrus.Errorf("error cancelling a job %s: %s", id, err)
	}
}
