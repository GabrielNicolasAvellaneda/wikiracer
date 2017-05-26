package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/darkonie/wikiracer/control"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/pprof"
)

// BasePath represents a base path for wikiracing.
const BasePath = "/api/v1"

type key int

var jobPoolMgrKey key = 1

func jpManagerFromContext(ctx context.Context) (*control.JobPoolManager, bool) {
	jPool, ok := ctx.Value(jobPoolMgrKey).(*control.JobPoolManager)
	return jPool, ok
}

// NewRouter returns a configured gorilla mux router with loaded paths and handlers.
func NewRouter(jpManager *control.JobPoolManager) *mux.Router {
	router := mux.NewRouter()

	route := router.PathPrefix(BasePath).Subrouter()

	// get job info.
	route.Path("/job").Handler(jobMiddleware(jobInfoHandler, jpManager)).Methods("GET")
	route.Path("/job/{id}").Handler(jobMiddleware(jobInfoHandler, jpManager)).Methods("GET")

	// use server sent events to track the job status in real time.
	route.Path("/job/{id}/sse").Handler(jobMiddleware(jobInfoSSEHandler, jpManager)).Methods("GET")

	// kick off a new job.
	route.Path("/job").Handler(jobMiddleware(jobStartHandler, jpManager)).Methods("POST")

	// cancel an active job.
	route.Path("/job/{id}/cancel").Handler(jobMiddleware(jobCancelHandler, jpManager)).Methods("POST")

	// add debug endpoints
	debug := router.PathPrefix("/debug").Subrouter()
	debug.Path("/pprof").HandlerFunc(pprof.Index).Methods("GET")
	debug.Path("/cmdline").HandlerFunc(pprof.Cmdline).Methods("GET")
	debug.Path("/profile").HandlerFunc(pprof.Profile).Methods("GET")
	debug.Path("/symbol").HandlerFunc(pprof.Symbol).Methods("GET")
	debug.Path("/trace").HandlerFunc(pprof.Trace).Methods("GET")
	debug.Path("/{profile}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		profile := mux.Vars(req)["profile"]
		pprof.Handler(profile).ServeHTTP(w, req)
	}).Methods("GET")

	return router
}

func jobsFromRequest(r *http.Request) (map[string]*control.Job, error) {
	jobID, useID := mux.Vars(r)["id"]
	jpManager, ok := jpManagerFromContext(r.Context())
	if !ok {
		return nil, errors.New("unable to get a job pool from context")
	}

	if useID {
		j, found := jpManager.GetJob(jobID)
		if !found {
			return nil, fmt.Errorf("job %s not found in job pool", jobID)
		}
		return map[string]*control.Job{jobID: j}, nil
	}

	return jpManager.Pool, nil
}
