package repoimpl

import (
	"github.com/google/wire"
)

var ProviderSetRepoImpl = wire.NewSet(
	NewUserRepository,
	NewCaptchaRepository,
	NewTeamRepository,
	NewCacheRepository,
	NewResourceRepository,
	NewTeamRoleRepository,
	NewTeamMenuRepository,
	NewMenuRepository,
	NewDatasourceRepository,
	NewDatasourceMetricRepository,
	NewLockRepository,
	NewMetricRepository,
	NewDictRepository,
	NewTemplateRepository,
	NewStrategyRepository,
	NewStrategyGroupRepository,
)
