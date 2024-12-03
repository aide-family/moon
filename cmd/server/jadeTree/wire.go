//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package jadeTree

import (
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/biz"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/data"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/data/microserver"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/data/repoimpl"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/jadetreeconf"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/server"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*jadetreeconf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSetServer,
		data.ProviderSetData,
		microserver.ProviderSetRPCConn,
		repoimpl.ProviderSetRepoImpl,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		newApp,
	))
}
