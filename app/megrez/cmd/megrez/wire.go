//go:build wireinject

package main

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"

	"github.com/aide-family/moon/app/megrez/internal/biz"
	"github.com/aide-family/moon/app/megrez/internal/conf"
	"github.com/aide-family/moon/app/megrez/internal/data"
	"github.com/aide-family/moon/app/megrez/internal/server"
	"github.com/aide-family/moon/app/megrez/internal/service"
)

func wireApp(*conf.Server, *conf.Alioth, *conf.Redis, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
