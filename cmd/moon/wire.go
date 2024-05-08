//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package moon

import (
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
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
