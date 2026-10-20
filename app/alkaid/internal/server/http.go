package server

import (
	"log/slog"
	"time"

	khttp "github.com/go-kratos/kratos/v3/transport/http"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/app/alkaid/internal/service"
	"github.com/aide-family/Alioth/pkg/healthserver"
	"github.com/aide-family/Alioth/pkg/merr"
)

func NewHTTPServer(c *conf.Server, health *service.HealthService, logger *slog.Logger) (*khttp.Server, error) {
	if c == nil || c.Http == nil {
		return nil, merr.Server("http addr is empty")
	}
	var timeout time.Duration
	if c.Http.Timeout != nil {
		timeout = c.Http.Timeout.AsDuration()
	}
	return healthserver.New(c.Http.Addr, c.Http.Network, timeout, health, logger)
}
