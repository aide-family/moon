//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"prometheus-manager/apps/node/internal/biz"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/data"
	"prometheus-manager/apps/node/internal/server"
	"prometheus-manager/apps/node/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
