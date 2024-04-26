package alarmservice

import (
	"context"

	pb "github.com/aide-family/moon/api/alarm/hook"
	"github.com/aide-family/moon/app/prom_server/internal/biz"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/aide-family/moon/pkg/util/times"
	"github.com/go-kratos/kratos/v2/log"
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
		endsAt := endTime.Unix()
		duration := int64(endTime.Sub(startTime).Seconds())
		if endsAt <= 0 {
			endsAt = 0
			duration = 0
		}
		historyBos = append(historyBos, &bo.AlarmHistoryBO{
			Md5:        alert.GetFingerprint(),
			StrategyId: labels.StrategyId(),
			LevelId:    labels.LevelId(),
			Status:     vobj.ToAlarmStatus(alert.GetStatus()),
			StartsAt:   startTime.Unix(),
			EndsAt:     endsAt,
			Instance:   strategy.MapToLabels(alert.GetLabels()).GetInstance(),
			Duration:   duration,
			Info: &bo.AlertBo{
				Status:       alert.GetStatus(),
				Labels:       &labels,
				Annotations:  &annotations,
				StartsAt:     alert.GetStartsAt(),
				EndsAt:       alert.GetEndsAt(),
				GeneratorURL: alert.GetGeneratorURL(),
				Fingerprint:  hash.MD5(alert.GetFingerprint()),
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
