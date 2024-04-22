package data

import (
	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(
	NewData,
)

// Data .
type Data struct {
	log *log.Helper
}

// NewData .
func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	d := &Data{
		log: log.NewHelper(log.With(logger, "module", "data")),
	}
	cleanup := func() {
		d.log.Info("closing the data resources")
	}

	return d, cleanup, nil
}
