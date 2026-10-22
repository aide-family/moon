package server

import (
	"log/slog"
	"time"

	khttp "github.com/go-kratos/kratos/v3/transport/http"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/megrez/internal/conf"
	"github.com/aide-family/Alioth/app/megrez/internal/service"
	"github.com/aide-family/Alioth/pkg/healthserver"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewHTTPServer)

func NewHTTPServer(c *conf.Server, health *service.HealthService, logger *slog.Logger) (*khttp.Server, error) {
	if c == nil || c.Http == nil {
		return nil, merr.Server("http addr is empty")
	}
	var timeout time.Duration
	if c.Http.Timeout != nil {
		timeout = c.Http.Timeout.AsDuration()
	}
	return healthserver.NewServer(c.Http.Addr, c.Http.Network, timeout, health, health, logger)
}
