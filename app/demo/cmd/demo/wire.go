//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"github.com/aide-family/moon/app/demo/internal/biz"
	"github.com/aide-family/moon/app/demo/internal/conf"
	"github.com/aide-family/moon/app/demo/internal/data"
	"github.com/aide-family/moon/app/demo/internal/server"
	"github.com/aide-family/moon/app/demo/internal/service"
	"github.com/aide-family/moon/pkg/helper/plog"
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
		newApp,
	))
}
