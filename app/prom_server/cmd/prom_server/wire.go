//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/alarmbiz"
	"prometheus-manager/app/prom_server/internal/biz/dictbiz"
	"prometheus-manager/app/prom_server/internal/biz/prombiz"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/app/prom_server/internal/data/repository"
	"prometheus-manager/app/prom_server/internal/server"
	"prometheus-manager/app/prom_server/internal/service"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
	"prometheus-manager/app/prom_server/internal/service/authservice"
	"prometheus-manager/app/prom_server/internal/service/dictservice"
	"prometheus-manager/app/prom_server/internal/service/promservice"
	"prometheus-manager/pkg/plog"
)

// wireApp init kratos application.
func wireApp(*string) (*kratos.App, func(), error) {
	panic(wire.Build(
		ProviderSetCore,
		server.ProviderSetServer,
		data.ProviderSetData,
		biz.ProviderSetBiz,
		service.ProviderSetService,
		conf.ProviderSetConf,
		plog.ProviderSetPlog,
		dictservice.ProviderSetDictService,
		promservice.ProviderSetProm,
		alarmservice.ProviderSetAlarm,
		authservice.ProviderSetAuthService,
		dictbiz.ProviderSetDictBiz,
		prombiz.ProviderSetPromBiz,
		alarmbiz.ProviderSetAlarmBiz,
		repository.ProviderSetRepository,
		newApp,
	))
}
