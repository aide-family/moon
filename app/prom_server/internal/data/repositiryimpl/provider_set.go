package repositiryimpl

import (
	"github.com/google/wire"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/dashboard"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/syslog"

	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/alarmhistory"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/alarmintervene"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/alarmpage"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/alarmrealtime"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/alarmsuppress"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/alarmupgrade"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/api"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/cache"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/captcha"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/chatgroup"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/dataimpl"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/endpoint"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/msg"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/notify"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/ping"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/promdict"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/role"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/strategy"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/strategygroup"
	"github.com/aide-family/moon/app/prom_server/internal/data/repositiryimpl/user"
)

// ProviderSetRepository 注入repository依赖
var ProviderSetRepository = wire.NewSet(
	ping.NewPingRepo,
	promdict.NewPromDictRepo,
	strategy.NewStrategyRepo,
	alarmpage.NewAlarmPageRepo,
	alarmhistory.NewAlarmHistoryRepo,
	strategygroup.NewStrategyGroupRepo,
	role.NewRoleRepo,
	user.NewUserRepo,
	cache.NewCacheRepo,
	endpoint.NewEndpointRepo,
	api.NewApiRepo,
	captcha.NewCaptchaRepo,
	chatgroup.NewChatGroupRepo,
	notify.NewNotifyRepo,
	dataimpl.NewDataRepo,
	alarmrealtime.NewAlarmRealtime,
	alarmintervene.NewAlarmIntervene,
	alarmsuppress.NewAlarmSuppress,
	alarmupgrade.NewAlarmUpgrade,
	msg.NewMsgRepo,
	dashboard.NewDashboardRepo,
	dashboard.NewChartRepo,
	syslog.NewSysLogRepo,
	notify.NewNotifyTemplateRepo,
)
