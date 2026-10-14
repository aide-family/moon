//go:build wireinject

package main

import (
	"github.com/aide-family/Alioth/app/megrez/internal/biz"
	"github.com/aide-family/Alioth/app/megrez/internal/conf"
	"github.com/aide-family/Alioth/app/megrez/internal/data"
	"github.com/aide-family/Alioth/app/megrez/internal/server"
	"github.com/aide-family/Alioth/app/megrez/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

func wireApp(*conf.Server, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
