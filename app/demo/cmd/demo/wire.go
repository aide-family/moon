//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"prometheus-manager/app/demo/internal/biz"
	"prometheus-manager/app/demo/internal/conf"
	"prometheus-manager/app/demo/internal/data"
	"prometheus-manager/app/demo/internal/server"
	"prometheus-manager/app/demo/internal/service"
	"prometheus-manager/pkg/plog"
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
		plog.ProviderSetPlog,
		newApp,
	))
}
