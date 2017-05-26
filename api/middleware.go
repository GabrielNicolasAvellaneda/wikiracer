package api

import (
	"context"
	"net/http"

	"github.com/darkonie/wikiracer/control"
)

func jobMiddleware(fn http.HandlerFunc, jpManager *control.JobPoolManager) http.Handler {
	next := http.HandlerFunc(fn)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), jobPoolMgrKey, jpManager)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
