//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"

	"prometheus-manager/pkg/helper/plog"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/app/prom_server/internal/data/repositiryimple"
	"prometheus-manager/app/prom_server/internal/server"
	"prometheus-manager/app/prom_server/internal/service"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
	"prometheus-manager/app/prom_server/internal/service/authservice"
	"prometheus-manager/app/prom_server/internal/service/dictservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"
	"prometheus-manager/app/prom_server/internal/service/systemservice"
)

// wireApp init kratos application.
func wireApp(*string) (*kratos.App, func(), error) {
	panic(wire.Build(
		ProviderSetCore,
		server.ProviderSetServer,
		data.ProviderSetData,
		service.ProviderSetService,
		conf.ProviderSetConf,
		plog.ProviderSetPLog,
		dictservice.ProviderSetDictService,
		promservice.ProviderSetProm,
		alarmservice.ProviderSetAlarm,
		authservice.ProviderSetAuthService,
		systemservice.ProviderSetSystem,
		biz.ProviderSetBiz,
		repositiryimple.ProviderSetRepository,
		newApp,
	))
}
