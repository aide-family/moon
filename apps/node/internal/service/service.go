package service

import (
	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewPullService,
	NewLoadService,
	NewPushService,
	NewPingService,
)

const (
	loadModuleName = "service/load"
	pingModuleName = "service/ping"
	pullModuleName = "service/pull"
	pushModuleName = "service/push"
)
