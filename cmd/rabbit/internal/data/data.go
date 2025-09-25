// Package data is a data package for kratos.
package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/rabbit/internal/conf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/email"
	"github.com/aide-family/moon/pkg/plugin/hook"
	"github.com/aide-family/moon/pkg/plugin/sms"
	"github.com/aide-family/moon/pkg/util/safety"
)

// ProviderSetData is a set of data providers.
var ProviderSetData = wire.NewSet(New)

func New(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	var err error
	dataConf := c.GetData()
	data := &Data{
		dataConf: dataConf,
		emails:   safety.NewMap[string, email.Email](),
		smss:     safety.NewMap[string, sms.Sender](),
		hooks:    safety.NewMap[string, hook.Sender](),
		helper:   log.NewHelper(log.With(logger, "module", "data")),
	}
	data.cache, err = cache.NewCache(dataConf.GetCache())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if err = data.cache.Close(); err != nil {
			log.NewHelper(logger).Errorw("method", "close cache", "err", err)
		}
	}
	return data, cleanup, nil
}

type Data struct {
	dataConf *conf.Data
	cache    cache.Cache
	emails   *safety.Map[string, email.Email]
	smss     *safety.Map[string, sms.Sender]
	hooks    *safety.Map[string, hook.Sender]

	helper *log.Helper
}

func (d *Data) GetCache() cache.Cache {
	return d.cache
}

func (d *Data) GetEmail(name string) (email.Email, bool) {
	emailInstance, ok := d.emails.Get(name)
	if !ok {
		return nil, false
	}
	return emailInstance.Copy(), true
}

func (d *Data) SetEmail(name string, email email.Email) {
	d.emails.Set(name, email)
}

func (d *Data) GetSms(name string) (sms.Sender, bool) {
	return d.smss.Get(name)
}

func (d *Data) SetSms(name string, sms sms.Sender) {
	d.smss.Set(name, sms)
}

func (d *Data) GetHook(name string) (hook.Sender, bool) {
	return d.hooks.Get(name)
}

func (d *Data) SetHook(name string, hook hook.Sender) {
	d.hooks.Set(name, hook)
}
