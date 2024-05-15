//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package demo

import (
	"github.com/aide-cloud/moon/cmd/server/demo/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/demo/internal/data"
	"github.com/aide-cloud/moon/cmd/server/demo/internal/data/repoimpl"
	"github.com/aide-cloud/moon/cmd/server/demo/internal/democonf"
	"github.com/aide-cloud/moon/cmd/server/demo/internal/server"
	"github.com/aide-cloud/moon/cmd/server/demo/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*democonf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSetServer,
		data.ProviderSetData,
		repoimpl.ProviderSetRepoImpl,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		newApp,
	))
}
