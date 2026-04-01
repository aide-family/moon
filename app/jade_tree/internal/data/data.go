package data

import (
	"context"

	_ "github.com/aide-family/magicbox/connect/orm/mysql"
	_ "github.com/aide-family/magicbox/connect/orm/postgres"
	_ "github.com/aide-family/magicbox/connect/orm/sqlite"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/aide-family/jade_tree/internal/conf"
)

var ProviderSetData = wire.NewSet(New)

func New(c *conf.Bootstrap, helper *klog.Helper) (*Data, func(), error) {
	d := &Data{
		helper: helper,
		c:      c,
		closes: safety.NewSyncMap(make(map[string]func() error)),
	}

	if err := d.initRegistry(); err != nil {
		return nil, d.close, err
	}

	dbConf := c.GetDatabase()
	if pointer.IsNil(dbConf) || dbConf.GetDialector() == config.ORMConfig_TYPE_UNKNOWN {
		return nil, d.close, merr.ErrorInvalidArgument("database configuration is required")
	}

	db, closeDB, err := connect.NewDB(dbConf)
	if err != nil {
		return nil, d.close, err
	}
	d.db = db
	d.closes.Set("db", closeDB)

	node, err := snowflake.NewNode(hello.NodeID())
	if err != nil {
		_ = closeDB()
		return nil, d.close, err
	}
	d.node = node

	return d, d.close, nil
}

type Data struct {
	helper   *klog.Helper
	c        *conf.Bootstrap
	registry connect.Report
	db       *gorm.DB
	node     *snowflake.Node
	closes   *safety.SyncMap[string, func() error]
}

func (d *Data) AppendClose(name string, close func() error) {
	d.closes.Set(name, close)
}

func (d *Data) close() {
	d.closes.Range(func(name string, close func() error) bool {
		if err := close(); err != nil && d.helper != nil {
			d.helper.Errorw("msg", "close resource failed", "name", name, "error", err)
		}
		return true
	})
}

func (d *Data) initRegistry() error {
	report := d.c.GetReport()
	if pointer.IsNil(report) || report.GetReportType() == config.ReportConfig_REPORT_TYPE_UNKNOWN {
		return nil
	}
	reportInstance, closer, err := connect.NewReport(report)
	if err != nil {
		return err
	}
	d.registry = reportInstance
	d.closes.Set("report", closer)
	return nil
}

func (d *Data) Registry() connect.Report { return d.registry }

func (d *Data) Node() *snowflake.Node { return d.node }

func (d *Data) DB() *gorm.DB { return d.db }

func (d *Data) Config() *conf.Bootstrap { return d.c }

// RequireDB returns the database handle or panics if database is not configured.
func (d *Data) RequireDB() *gorm.DB {
	if d.db == nil {
		panic("database is not configured: set bootstrap.database in config")
	}
	return d.db
}

// PingDB verifies connectivity when a database is configured.
func (d *Data) PingDB(ctx context.Context) error {
	if d.db == nil {
		return nil
	}
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
