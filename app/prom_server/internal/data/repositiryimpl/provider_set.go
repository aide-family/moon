package repositiryimpl

import (
	"github.com/google/wire"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/dashboard"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/syslog"

	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmhistory"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmintervene"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmpage"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmrealtime"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmsuppress"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/alarmupgrade"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/api"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/cache"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/captcha"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/chatgroup"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/dataimpl"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/endpoint"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/msg"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/notify"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/ping"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/promdict"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/role"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/strategy"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/strategygroup"
	"prometheus-manager/app/prom_server/internal/data/repositiryimpl/user"
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
)
