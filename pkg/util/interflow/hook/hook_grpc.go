package hook

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
)

var _ interflow.AgentInterflow = (*hookGrpcInterflow)(nil)
var _ interflow.ServerInterflow = (*hookGrpcInterflow)(nil)

type (
	hookGrpcInterflow struct {
	}
)

func (h *hookGrpcInterflow) SendAgent(ctx context.Context, to string, msg *interflow.HookMsg) error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) ServerOnlineNotify(agentUrls []string) error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) ServerOfflineNotify(agentUrls []string) error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) Send(ctx context.Context, msg *interflow.HookMsg) error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) Receive() error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) SetHandles(handles map[consts.TopicType]interflow.Callback) error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) Close() error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) OnlineNotify() error {
	//TODO implement me
	panic("implement me")
}

func (h *hookGrpcInterflow) OfflineNotify() error {
	//TODO implement me
	panic("implement me")
}

func NewAgentHookGrpcInterflow(c GrpcConfig, logger log.Logger) (interflow.AgentInterflow, error) {
	return &hookGrpcInterflow{}, nil
}

func NewServerHookGrpcInterflow(network interflow.Network, logger log.Logger) interflow.ServerInterflow {
	return &hookGrpcInterflow{}
}
