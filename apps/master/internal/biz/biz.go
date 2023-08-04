package biz

import (
	"context"
	"github.com/google/wire"
	promBizV1 "prometheus-manager/apps/master/internal/biz/prom/v1"
	"prometheus-manager/apps/master/internal/service"
	promServiceV1 "prometheus-manager/apps/master/internal/service/prom/v1"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewCrudLogic,
	wire.Bind(new(service.ICrudLogic), new(*CrudLogic)),
	NewPingLogic,
	wire.Bind(new(service.IPingLogic), new(*PingLogic)),
	NewPushLogic,
	wire.Bind(new(service.IPushLogic), new(*PushLogic)),
	promBizV1.NewDirLogic,
	wire.Bind(new(promServiceV1.IDirLogic), new(*promBizV1.DirLogic)),
	promBizV1.NewFileLogic,
	wire.Bind(new(promServiceV1.IFileLogic), new(*promBizV1.FileLogic)),
	promBizV1.NewGroupLogic,
	wire.Bind(new(promServiceV1.IGroupLogic), new(*promBizV1.GroupLogic)),
	promBizV1.NewRuleLogic,
	wire.Bind(new(promServiceV1.IRuleLogic), new(*promBizV1.RuleLogic)),
	promBizV1.NewNodeLogic,
	wire.Bind(new(promServiceV1.INodeLogic), new(*promBizV1.NodeLogic)),
)

type V1Repo interface {
	V1(ctx context.Context) string
}
