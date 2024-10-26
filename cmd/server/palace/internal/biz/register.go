package biz

import "github.com/google/wire"

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(
	NewUserBiz,
	NewCaptchaBiz,
	NewAuthorizationBiz,
	NewResourceBiz,
	NewTeamBiz,
	NewTeamRoleBiz,
	NewMenuBiz,
	NewDatasourceBiz,
	NewStrategyBiz,
	NewStrategyGroupBiz,
	NewStrategyCountBiz,
	NewMetricBiz,
	NewDictBiz,
	NewTemplateBiz,
	NewAlarmBiz,
	NewDashboardBiz,
	NewAlarmGroupBiz,
	NewAlarmPageBiz,
	NewSubscriptionStrategyBiz,
	NewAlarmHookBiz,
	NewAlarmHistoryBiz,
	NewInviteBiz,
	NewUserMessageBiz,
	NewServerRegisterBiz,
)
