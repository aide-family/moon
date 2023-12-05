package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/history"

	"prometheus-manager/app/prom_server/internal/biz"
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
	historyList, pgInfo, err := s.historyBiz.ListHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	list := make([]*api.AlarmHistoryV1, 0, len(historyList))
	for _, v := range historyList {
		list = append(list, v.ToApiV1())
	}

	return &pb.ListHistoryReply{
		Page: &api.PageReply{
			Curr:  pgInfo.GetCurr(),
			Size:  pgInfo.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}
