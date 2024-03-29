//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"prometheus-manager/app/prom_agent/internal/biz"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/data"
	"prometheus-manager/app/prom_agent/internal/data/repositiryimpl"
	"prometheus-manager/app/prom_agent/internal/server"
	"prometheus-manager/app/prom_agent/internal/service"
	"prometheus-manager/pkg/helper/plog"
)

// wireApp init kratos application.
func wireApp(*string) (*kratos.App, func(), error) {
	panic(wire.Build(
		ProviderSetCore,
		server.ProviderSetServer,
		data.ProviderSetData,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		conf.ProviderSetConf,
		plog.ProviderSetPLog,
		repositiryimpl.ProviderSetRepoImpl,
		newApp,
	))
}
