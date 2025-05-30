package service

import (
	"github.com/google/wire"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewAuthService,
	NewHealthService,
	NewServerService,
	NewMenuService,
	NewUserService,
	NewCallbackService,
	NewTeamDashboardService,
	NewTeamDatasourceService,
	NewTeamDictService,
	NewTeamNoticeService,
	NewTeamStrategyService,
	NewTeamService,
	NewSystemService,
	NewLoadService,
	NewTeamLogService,
	NewAlertService,
	NewTimeEngineService,
	NewTeamStrategyMetricService,
)
