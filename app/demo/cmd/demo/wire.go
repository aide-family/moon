//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"prometheus-manager/app/demo/internal/biz"
	"prometheus-manager/app/demo/internal/conf"
	"prometheus-manager/app/demo/internal/data"
	"prometheus-manager/app/demo/internal/server"
	"prometheus-manager/app/demo/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSetServer, data.ProviderSetData, biz.ProviderSetBiz, service.ProviderSetService, newApp))
}
