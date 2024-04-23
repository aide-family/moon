package hook

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
)

var _ interflow.Interflow = (*hookGrpcInterflow)(nil)

type (
	hookGrpcInterflow struct {
	}
)

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

func NewHookGrpcInterflow(c GrpcConfig, logger log.Logger) interflow.Interflow {
	return &hookGrpcInterflow{}
}
