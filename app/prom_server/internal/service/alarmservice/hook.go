package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/hook"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/valueobj"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/times"
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
	alertList := req.GetAlerts()
	historyBos := make([]*dobo.AlarmHistoryBO, 0, len(alertList))
	for _, alert := range alertList {
		startTime := times.ParseAlertTime(alert.GetStartsAt())
		endTime := times.ParseAlertTime(alert.GetEndsAt())
		labels := strategy.Labels(alert.GetLabels())
		historyBos = append(historyBos, &dobo.AlarmHistoryBO{
			Md5:        alert.GetFingerprint(),
			StrategyId: uint32(labels.StrategyId()),
			LevelId:    uint32(labels.LevelId()),
			Status:     valueobj.ToAlarmStatus(alert.GetStatus()),
			StartAt:    startTime.Unix(),
			EndAt:      endTime.Unix(),
			Instance:   strategy.Labels(alert.GetLabels()).Get("instance"),
			Duration:   int64(endTime.Sub(startTime).Seconds()),
			Info: &dobo.AlertBo{
				Status:       alert.GetStatus(),
				Labels:       alert.GetLabels(),
				Annotations:  alert.GetAnnotations(),
				StartsAt:     times.ParseAlertTime(alert.GetStartsAt()).Unix(),
				EndsAt:       times.ParseAlertTime(alert.GetEndsAt()).Unix(),
				GeneratorURL: alert.GetGeneratorURL(),
				Fingerprint:  alert.GetFingerprint(),
			},
		})
	}
	_, err := s.historyBiz.CreateHistory(ctx, historyBos...)
	if err != nil {
		return nil, err
	}
	return &pb.HookV1Reply{
		Msg:  "create history success",
		Code: 0,
	}, nil
}
