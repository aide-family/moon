package service

import (
	"context"

	pb "github.com/aide-family/moon/api/interflows"
	"github.com/aide-family/moon/app/prom_agent/internal/conf"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
)

type HookInterflowService struct {
	pb.UnimplementedHookInterflowServer

	log           *log.Helper
	interflowConf *conf.Interflow
}

func NewHookInterflowService(interflowConf *conf.Interflow, logger log.Logger) *HookInterflowService {
	return &HookInterflowService{
		log:           log.NewHelper(log.With(logger, "module", "service.interflow")),
		interflowConf: interflowConf,
	}
}

func (s *HookInterflowService) Receive(ctx context.Context, req *pb.ReceiveRequest) (*pb.ReceiveResponse, error) {
	sendCh := interflow.GetSendInterflowCh()
	msg := &interflow.HookMsg{
		Topic: req.GetTopic(),
		Value: req.GetValue(),
	}
	sendCh <- msg
	return &pb.ReceiveResponse{}, nil
}
