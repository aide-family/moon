//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package moon

import (
	"github.com/aide-cloud/moon/cmd/moon/internal/data/repoimpl"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-cloud/moon/cmd/moon/internal/biz"
	"github.com/aide-cloud/moon/cmd/moon/internal/conf"
	"github.com/aide-cloud/moon/cmd/moon/internal/data"
	"github.com/aide-cloud/moon/cmd/moon/internal/server"
	"github.com/aide-cloud/moon/cmd/moon/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSetServer,
		data.ProviderSetData,
		repoimpl.ProviderSetRepoImpl,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		newApp,
	))
}
