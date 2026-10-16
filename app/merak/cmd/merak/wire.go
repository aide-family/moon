//go:build wireinject

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/biz"
	"github.com/aide-family/Alioth/app/merak/internal/conf"
	"github.com/aide-family/Alioth/app/merak/internal/data"
	"github.com/aide-family/Alioth/app/merak/internal/server"
	"github.com/aide-family/Alioth/app/merak/internal/service"
)

func wireApp(*conf.Server, *conf.Alioth, *conf.Redis, *conf.Notify, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
