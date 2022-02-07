//go:build prometheus
// +build prometheus

package httpd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	http.Handle("/metrics", promhttp.Handler())
}
