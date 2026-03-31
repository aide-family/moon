//go:build wireinject
// +build wireinject

package all

import (
	"github.com/go-kratos/kratos/v2"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/jade_tree/cmd/run"
	"github.com/aide-family/jade_tree/internal/biz"
	"github.com/aide-family/jade_tree/internal/conf"
	"github.com/aide-family/jade_tree/internal/data"
	"github.com/aide-family/jade_tree/internal/data/impl"
	"github.com/aide-family/jade_tree/internal/server"
	"github.com/aide-family/jade_tree/internal/service"
)

func WireApp(serviceName string, bc *conf.Bootstrap, helper *klog.Helper) ([]*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSetServerAll, service.ProviderSetService, biz.ProviderSetBiz, impl.ProviderSetImpl, data.ProviderSetData, run.NewApp))
}
