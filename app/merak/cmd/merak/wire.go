//go:build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/biz"
	"github.com/aide-family/Alioth/app/merak/internal/conf"
	"github.com/aide-family/Alioth/app/merak/internal/data"
	"github.com/aide-family/Alioth/app/merak/internal/server"
	"github.com/aide-family/Alioth/app/merak/internal/service"
)

func wireApp(*conf.Server, *conf.Alioth, *conf.Redis, *conf.Notify, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
