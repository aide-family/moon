package interflowservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/interflows"
	"prometheus-manager/pkg/util/interflow"
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
	sendCh := interflow.GetSendInterflowCh()
	agentUrl := "http://localhost:8001/api/v1/interflows/receive"
	msg := &interflow.HookMsg{
		Topic: req.Topic,
		Value: req.Value,
		Key:   []byte(agentUrl),
	}
	sendCh <- msg
	return &pb.ReceiveResponse{}, nil
}
