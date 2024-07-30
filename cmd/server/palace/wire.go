//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package palace

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/microserver"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/microserver/microserverrepoimpl"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/repoimpl"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/server"
	"github.com/aide-family/moon/cmd/server/palace/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*palaceconf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSetServer,
		data.ProviderSetData,
		repoimpl.ProviderSetRepoImpl,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		microserverrepoimpl.ProviderSetRPCRepoImpl,
		microserver.ProviderSetRPCConn,
		runtimecache.ProviderSetRuntimeCache,
		newApp,
	))
}
