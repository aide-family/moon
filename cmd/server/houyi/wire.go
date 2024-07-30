//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package houyi

import (
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data/microserver"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data/repoimpl"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/server"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*houyiconf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSetServer,
		data.ProviderSetData,
		repoimpl.ProviderSetRepoImpl,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		microserver.ProviderSetRPCConn,
		newApp,
	))
}
