package biz

import "github.com/google/wire"

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(
	NewGreeterUsecase,
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
)
