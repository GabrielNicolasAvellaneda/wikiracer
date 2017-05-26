package supervisor

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/darkonie/wikiracer/api"
	"github.com/darkonie/wikiracer/control"
	"github.com/sirupsen/logrus"
)

var defaultPort = 8081

// Start start HTTP server.
func Start() error {

	port := defaultPort
	if p := os.Getenv("WIKI_PORT"); p != "" {
		portFromEnvVar, err := strconv.Atoi(p)
		if err != nil {
			logrus.Errorf("unable to parse port, using default %d", defaultPort)
		} else {
			port = portFromEnvVar
		}
	}

	jpManager := control.NewJobPoolManager()
	logrus.Infof("Start server on :%d", port)
	logrus.Infof("Use http://127.0.0.1:%d/api/v1/ for more help", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), api.NewRouter(jpManager))
}
