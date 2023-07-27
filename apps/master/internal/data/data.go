package data

import (
	"prometheus-manager/apps/master/internal/biz"
	"prometheus-manager/apps/master/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewPingRepo,
	wire.Bind(new(biz.IPingRepo), new(*PingRepo)),
	NewCrudRepo,
	wire.Bind(new(biz.ICrudRepo), new(*CrudRepo)),
)

// Data .
type Data struct {
	// TODO wrapped database client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
