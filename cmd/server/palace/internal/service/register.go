package service

import (
	"github.com/aide-family/moon/cmd/server/palace/internal/service/alarm"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/datasource"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/dict"
	fileservice "github.com/aide-family/moon/cmd/server/palace/internal/service/file"
	historyservice "github.com/aide-family/moon/cmd/server/palace/internal/service/history"
	hookservice "github.com/aide-family/moon/cmd/server/palace/internal/service/hook"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/invite"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/menu"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/realtime"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/subscriber"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/system"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/team"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/template"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/user"

	"github.com/google/wire"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	user.NewUserService,
	user.NewMessageService,
	authorization.NewAuthorizationService,
	resource.NewResourceService,
	menu.NewMenuService,
	team.NewTeamService,
	team.NewRoleService,
	datasource.NewDatasourceService,
	datasource.NewMetricService,
	strategy.NewStrategyService,
	strategy.NewTemplateService,
	dict.NewDictService,
	realtime.NewDashboardService,
	realtime.NewAlarmService,
	alarm.NewAlarmService,
	alarm.NewSendService,
	alarm.NewTimeEngineRuleService,
	realtime.NewAlarmPageSelfService,
	realtime.NewStatisticsService,
	subscriber.NewSubscriberService,
	hookservice.NewHookService,
	NewAlertService,
	invite.NewInviteService,
	historyservice.NewHistoryService,
	fileservice.NewFileService,
	NewServerService,
	system.NewSystemService,
	template.NewSendTemplateService,
)
