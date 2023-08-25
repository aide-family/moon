package service

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewPingService,
	NewPushService,
	NewPromService,
	NewAlarmPageService,
	NewDictService,
	NewWatchService,
)

const (
	dictModuleName      = "service/dict"
	alarmPageModuleName = "service/alarmPage"
	promModuleName      = "service/prom"
	pushModuleName      = "service/push"
	pingModuleName      = "service/ping"
	watchModuleName     = "service/watch"
)
