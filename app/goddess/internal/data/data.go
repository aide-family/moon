// Package data is the data package for the Goddess service.
package data

import (
	"context"

	_ "github.com/aide-family/magicbox/connect/orm/mysql"
	_ "github.com/aide-family/magicbox/connect/orm/sqlite"
	_ "github.com/aide-family/magicbox/oauth/feishu"
	_ "github.com/aide-family/magicbox/oauth/gitee"
	_ "github.com/aide-family/magicbox/oauth/github"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/connect"
	"github.com/aide-family/magicbox/plugin/cache"
	"github.com/aide-family/magicbox/plugin/cache/mem"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/gorm"

	"github.com/aide-family/goddess/internal/conf"
)

// ProviderSetData is a set of data providers.
var ProviderSetData = wire.NewSet(New)

// New a data and returns.
func New(c *conf.Bootstrap, helper *klog.Helper) (*Data, func(), error) {
	d := &Data{
		helper:      helper,
		c:           c,
		closes:      safety.NewSyncMap(make(map[string]func() error)),
		reloadFuncs: safety.NewSyncMap(make(map[string]func())),
	}

	if err := d.initRegistry(); err != nil {
		return nil, d.close, err
	}

	cacheDriver := mem.CacheDriver()
	cache, err := cache.New(context.Background(), cacheDriver)
	if err != nil {
		return nil, d.close, err
	}
	d.cache = cache
	d.closes.Set("cache", func() error { return cache.Close() })
	db, closer, err := connect.NewDB(c.GetDatabase())
	if err != nil {
		return nil, d.close, err
	}
	d.db = db
	d.closes.Set("db", closer)

	return d, d.close, nil
}

type Data struct {
	helper      *klog.Helper
	c           *conf.Bootstrap
	registry    connect.Report
	cache       cache.Interface
	db          *gorm.DB
	closes      *safety.SyncMap[string, func() error] // 使用SyncMap保证并发安全
	reloadFuncs *safety.SyncMap[string, func()]
}

func (d *Data) AppendClose(name string, close func() error) {
	d.closes.Set(name, close)
}

func (d *Data) close() {
	d.closes.Range(func(name string, close func() error) bool {
		if err := close(); err != nil {
			d.helper.Errorw("msg", "close db failed", "name", name, "error", err)
			return true // 继续遍历
		}
		d.helper.Debugw("msg", "close success", "name", name)
		return true // 继续遍历
	})
}

func (d *Data) Registry() connect.Report {
	return d.registry
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

func (d *Data) DB() *gorm.DB {
	return d.db
}

func (d *Data) Cache() cache.Interface {
	return d.cache
}
