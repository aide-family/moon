//go:build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/biz"
	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/app/alkaid/internal/data"
	"github.com/aide-family/Alioth/app/alkaid/internal/server"
	"github.com/aide-family/Alioth/app/alkaid/internal/service"
)

func wireApp(*conf.Server, *conf.Alioth, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
