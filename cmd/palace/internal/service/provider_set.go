package service

import (
	"github.com/google/wire"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewAuthService,
	NewHealthService,
	NewServerService,
	NewResourceService,
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
)
