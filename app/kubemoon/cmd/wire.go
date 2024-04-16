//go:build wireinject
// +build wireinject

package main

import (
	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/aide-family/moon/app/kubemoon/internal/data"
	"github.com/aide-family/moon/app/kubemoon/internal/server"
	"github.com/aide-family/moon/app/kubemoon/internal/service"
	"github.com/aide-family/moon/pkg/helper/plog"
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*string) (*kratos.App, func(), error) {
	panic(wire.Build(
		data.ProviderSetData,
		service.ProviderSetService,
		server.ProviderSetServer,
		conf.ProviderSetConf,
		plog.ProviderSetPLog,
		ProviderSetCore,
		newApp,
	))
}
