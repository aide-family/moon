package data

import (
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData, NewPingRepo)

// Data .
type Data struct {
	db *gorm.DB

	log *log.Helper
}

// DB gorm DB对象
func (d *Data) DB() *gorm.DB {
	return d.db
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	d := &Data{
		log: log.NewHelper(log.With(logger, "module", "data")),
	}
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return d, cleanup, nil
}
