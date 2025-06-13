package biz

import (
	"github.com/google/wire"
)

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(
	NewAuthBiz,
	NewPermissionBiz,
	NewMenuBiz,
	NewUserBiz,
	NewDashboardBiz,
	NewServerBiz,
	NewDictBiz,
	NewTeamBiz,
	NewTeamHookBiz,
	NewMessageBiz,
	NewSystemBiz,
	NewTeamNoticeBiz,
	NewTeamDatasourceBiz,
	NewTeamStrategyBiz,
	NewTeamStrategyGroupBiz,
	NewTeamStrategyMetricBiz,
	NewLogsBiz,
	NewRealtimeBiz,
	NewTeamDatasourceQueryBiz,
	NewTimeEngineBiz,
	NewEventBus,
)
