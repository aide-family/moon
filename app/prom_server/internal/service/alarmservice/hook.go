package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/hook"
	"prometheus-manager/app/prom_server/internal/biz"
)

type HookService struct {
	pb.UnimplementedHookServer

	log *log.Helper

	historyBiz *biz.HistoryBiz
}

func NewHookService(historyBiz *biz.HistoryBiz, logger log.Logger) *HookService {
	return &HookService{
		log:        log.NewHelper(log.With(logger, "module", "service.alarm.hook")),
		historyBiz: historyBiz,
	}
}

func (s *HookService) V1(ctx context.Context, req *pb.HookV1Request) (*pb.HookV1Reply, error) {
	return &pb.HookV1Reply{}, nil
}
