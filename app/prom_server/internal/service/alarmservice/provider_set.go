package alarmservice

import (
	"github.com/google/wire"
)

// ProviderSetAlarm 注入prometheus相关服务
var ProviderSetAlarm = wire.NewSet(
	NewHistoryService,
	NewHookService,
	NewAlarmPageService,
	NewRealtimeService,
)
