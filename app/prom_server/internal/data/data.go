package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"prometheus-manager/app/prom_server/internal/conf"
)

// ProviderSetData is data providers.
var ProviderSetData = wire.NewSet(NewData)

// Data .
type Data struct {
	db     *gorm.DB
	client *redis.Client

	log *log.Helper
}

// DB gorm DB对象
func (d *Data) DB() *gorm.DB {
	return d.db
}

// Client redis client
func (d *Data) Client() *redis.Client {
	return d.client
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
