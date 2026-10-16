//go:build wireinject

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/biz"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data"
	"github.com/aide-family/Alioth/app/alioth/internal/server"
	"github.com/aide-family/Alioth/app/alioth/internal/service"
)

func wireApp(*conf.Server, *conf.Data, *conf.Auth, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, access.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
