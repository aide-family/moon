package service

import (
	"github.com/google/wire"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/datasource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/resource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/team"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/user"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewGreeterService,
	user.NewUserService,
	authorization.NewAuthorizationService,
	resource.NewResourceService,
	resource.NewMenuService,
	team.NewTeamService,
	team.NewRoleService,
	datasource.NewDatasourceService,
)
