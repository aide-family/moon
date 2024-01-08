package alarmservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/vo"

	pb "prometheus-manager/api/alarm/hook"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/times"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
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
	historyBos := make([]*bo.AlarmHistoryBO, 0, len(alertList))
	for _, alert := range alertList {
		startTime := times.ParseAlertTime(alert.GetStartsAt())
		endTime := times.ParseAlertTime(alert.GetEndsAt())
		labels := strategy.Labels(alert.GetLabels())
		annotations := strategy.Annotations(alert.GetAnnotations())
		historyBos = append(historyBos, &bo.AlarmHistoryBO{
			Md5:        alert.GetFingerprint(),
			StrategyId: uint32(labels.StrategyId()),
			LevelId:    uint32(labels.LevelId()),
			Status:     vo.ToAlarmStatus(alert.GetStatus()),
			StartAt:    startTime.Unix(),
			EndAt:      endTime.Unix(),
			Instance:   strategy.MapToLabels(alert.GetLabels()).GetInstance(),
			Duration:   int64(endTime.Sub(startTime).Seconds()),
			Info: &bo.AlertBo{
				Status:       alert.GetStatus(),
				Labels:       &labels,
				Annotations:  &annotations,
				StartsAt:     times.ParseAlertTime(alert.GetStartsAt()).Unix(),
				EndsAt:       times.ParseAlertTime(alert.GetEndsAt()).Unix(),
				GeneratorURL: alert.GetGeneratorURL(),
				Fingerprint:  alert.GetFingerprint(),
			},
		})
	}
	_, err := s.historyBiz.HandleHistory(ctx, historyBos...)
	if err != nil {
		return nil, err
	}
	return &pb.HookV1Reply{
		Msg:  "handle alert info is success",
		Code: 0,
	}, nil
}
