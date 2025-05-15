//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/moon-monitor/moon/cmd/rabbit/internal/biz"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/conf"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/data"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/data/impl"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/data/rpc"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/server"
	"github.com/moon-monitor/moon/cmd/rabbit/internal/service"
)

// wireApp init wired
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		data.ProviderSetData,
		impl.ProviderSetImpl,
		rpc.ProviderSetRpc,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		server.ProviderSetServer,
		newApp,
	))
}
