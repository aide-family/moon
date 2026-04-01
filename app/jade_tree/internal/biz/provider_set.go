package biz

import "github.com/google/wire"
import "github.com/aide-family/jade_tree/internal/biz/collector"

var ProviderSetBiz = wire.NewSet(NewHealth, NewSSHCommand, NewMachineInfo, NewProbeTask, collector.NewProbeCollector)
