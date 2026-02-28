//go:build wireinject
// +build wireinject

// Package grpc is the grpc command for the marksman service
package grpc

import (
	"github.com/go-kratos/kratos/v2"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/aide-family/marksman/cmd/run"
	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/data/impl"
	"github.com/aide-family/marksman/internal/server"
	"github.com/aide-family/marksman/internal/service"
)

func WireApp(serviceName string, bc *conf.Bootstrap, helper *klog.Helper) ([]*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSetServerGRPC,
		service.ProviderSetService,
		biz.ProviderSetBiz,
		impl.ProviderSetImpl,
		data.ProviderSetData,
		run.NewApp,
	))
}
