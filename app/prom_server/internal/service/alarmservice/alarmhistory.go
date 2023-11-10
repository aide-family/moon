package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/history"
	"prometheus-manager/app/prom_server/internal/biz/alarmbiz"
)

type HistoryService struct {
	pb.UnimplementedHistoryServer

	log *log.Helper

	historyBiz *alarmbiz.HistoryBiz
}

func NewHistoryService(historyBiz *alarmbiz.HistoryBiz, logger log.Logger) *HistoryService {
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

	alarmHistoryInfo := &api.AlarmHistoryV1{
		Id:          historyDetailBO.Id,
		AlarmId:     0,
		AlarmName:   historyDetailBO.Md5,
		AlarmLevel:  "historyDetailBO.LevelId",
		AlarmStatus: historyDetailBO.Status.String(),
	}

	return &pb.GetHistoryReply{AlarmHistory: alarmHistoryInfo}, nil
}

func (s *HistoryService) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) (*pb.ListHistoryReply, error) {
	historyList, pgInfo, err := s.historyBiz.ListHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	list := make([]*api.AlarmHistoryV1, 0, len(historyList))
	for _, v := range historyList {
		list = append(list, &api.AlarmHistoryV1{
			Id:          v.Id,
			AlarmId:     0,
			AlarmName:   v.Md5,
			AlarmLevel:  "v.LevelId",
			AlarmStatus: v.Status.String(),
		})
	}

	return &pb.ListHistoryReply{
		Page: &api.PageReply{
			Curr:  int32(pgInfo.GetCurr()),
			Size:  int32(pgInfo.GetSize()),
			Total: pgInfo.GetTotal(),
		},
		List: list,
	}, nil
}
