package metric

import (
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(srv *http.Server) {
	srv.Handle("/metrics", promhttp.Handler())
}
