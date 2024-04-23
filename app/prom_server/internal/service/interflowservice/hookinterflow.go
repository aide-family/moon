package interflowservice

import (
	"context"

	pb "github.com/aide-family/moon/api/interflows"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/aide-family/moon/pkg/util/interflow/hook"
	"github.com/go-kratos/kratos/v2/log"
)

type HookInterflowService struct {
	pb.UnimplementedHookInterflowServer
	log *log.Helper
}

func NewHookInterflowService(logger log.Logger) *HookInterflowService {
	return &HookInterflowService{
		log: log.NewHelper(log.With(logger, "module", "service.interflow")),
	}
}

func (s *HookInterflowService) Receive(ctx context.Context, req *pb.ReceiveRequest) (*pb.ReceiveResponse, error) {
	sendCh := hook.GetSendInterflowCh()
	msg := &interflow.HookMsg{
		Topic: req.GetTopic(),
		Value: req.GetValue(),
		To:    req.GetKey(),
	}
	sendCh <- msg
	return &pb.ReceiveResponse{}, nil
}
