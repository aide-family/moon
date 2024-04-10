//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"github.com/aide-family/moon/app/prom_server/internal/service/dashboardservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/interflowservice"

	"github.com/aide-family/moon/pkg/helper/plog"

	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/conf"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl"
	"github.com/aide-family/moon/app/prom_server/internal/server"
	"github.com/aide-family/moon/app/prom_server/internal/service"
	"github.com/aide-family/moon/app/prom_server/internal/service/alarmservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/authservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/promservice"
	"github.com/aide-family/moon/app/prom_server/internal/service/systemservice"
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
		interflowservice.ProviderSetInterflowService,
		promservice.ProviderSetProm,
		alarmservice.ProviderSetAlarm,
		authservice.ProviderSetAuthService,
		systemservice.ProviderSetSystem,
		dashboardservice.ProviderSetDashboardService,
		biz.ProviderSetBiz,
		repositiryimpl.ProviderSetRepository,
		newApp,
	))
}
