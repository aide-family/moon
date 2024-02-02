package dashboardservice

import (
	"github.com/google/wire"
)

// ProviderSetDashboardService 注入AuthService
var ProviderSetDashboardService = wire.NewSet(NewDashboardService, NewChartService)
