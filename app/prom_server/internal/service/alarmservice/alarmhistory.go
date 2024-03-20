package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/server/alarm/history"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type HistoryService struct {
	pb.UnimplementedHistoryServer

	log *log.Helper

	historyBiz *biz.HistoryBiz
}

func NewHistoryService(historyBiz *biz.HistoryBiz, logger log.Logger) *HistoryService {
	return &HistoryService{
		log:        log.NewHelper(log.With(logger, "module", "service.alarm.history")),
		historyBiz: historyBiz,
	}
}

func (s *HistoryService) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryReply, error) {
	historyDetailBO, err := s.historyBiz.GetHistoryDetail(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	alarmHistoryInfo := historyDetailBO.ToApiV1()

	return &pb.GetHistoryReply{AlarmHistory: alarmHistoryInfo}, nil
}

func (s *HistoryService) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) (*pb.ListHistoryReply, error) {
	pgInfo := bo.NewPage(req.GetPage().GetCurr(), req.GetPage().GetSize())
	listReq := &bo.ListHistoryRequest{
		Page:            pgInfo,
		Keyword:         req.GetKeyword(),
		FiringStartAt:   req.GetFiringStartAt(),
		FiringEndAt:     req.GetFiringEndAt(),
		ResolvedStartAt: req.GetResolvedStartAt(),
		ResolvedEndAt:   req.GetResolvedEndAt(),
		Status:          vo.AlarmStatus(req.GetStatus()),
		AlarmPageIds:    req.GetAlarmPages(),
		StrategyIds:     req.GetStrategyIds(),
		AlarmLevelIds:   req.GetAlarmLevelIds(),
		Duration:        req.GetDuration(),
	}
	historyList, err := s.historyBiz.ListHistory(ctx, listReq)
	if err != nil {
		return nil, err
	}

	list := make([]*api.AlarmHistoryV1, 0, len(historyList))
	for _, v := range historyList {
		list = append(list, v.ToApiV1())
	}

	return &pb.ListHistoryReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetRespCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}
