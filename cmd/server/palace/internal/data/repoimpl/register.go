package repoimpl

import (
	"github.com/google/wire"
)

// ProviderSetRepoImpl wire Set
var ProviderSetRepoImpl = wire.NewSet(
	NewUserRepository,
	NewCaptchaRepository,
	NewTeamRepository,
	NewCacheRepository,
	NewResourceRepository,
	NewTeamResourceRepository,
	NewTeamRoleRepository,
	NewTeamMenuRepository,
	NewMenuRepository,
	NewDatasourceRepository,
	NewDatasourceMetricRepository,
	NewLockRepository,
	NewMetricRepository,
	NewDictRepository,
	NewTeamDictRepository,
	NewTemplateRepository,
	NewStrategyRepository,
	NewStrategyGroupRepository,
	NewStrategyCountRepository,
	NewRealtimeAlarmRepository,
	NewDashboardRepository,
	NewAlarmGroupRepository,
	NewAlarmPageRepository,
	NewSubscriberStrategyRepository,
	NewAlarmHookRepository,
	NewGithubUserRepository,
	NewInviteRepository,
	NewUserMessageRepository,
)
