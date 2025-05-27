//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl"
	"github.com/aide-family/moon/cmd/palace/internal/data/rpc"
	"github.com/aide-family/moon/cmd/palace/internal/server"
	"github.com/aide-family/moon/cmd/palace/internal/service"
	portal_service "github.com/aide-family/moon/cmd/palace/internal/service/portal"
)

// wireApp init kratos application.
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		data.ProviderSetData,
		impl.ProviderSetImpl,
		rpc.ProviderSetRPC,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		portal_service.ProviderSetPortalService,
		server.ProviderSetServer,
		newApp,
	))
}
