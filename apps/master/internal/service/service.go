package service

import (
	"github.com/google/wire"
	promV1 "prometheus-manager/apps/master/internal/service/prom/v1"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(
	NewCrudService,
	NewPingService,
	promV1.NewDirService,
	promV1.NewFileService,
	promV1.NewGroupService,
	promV1.NewRuleService,
	promV1.NewNodeService,
)
