package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/interflows"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/pkg/util/interflow"
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
	serverUrl := s.interflowConf.GetServer()
	msg := &interflow.HookMsg{
		Topic: req.Topic,
		Value: req.Value,
		Key:   []byte(serverUrl),
	}
	sendCh <- msg
	return &pb.ReceiveResponse{}, nil
}
