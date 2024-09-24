package biz

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

// NewAlarmBiz 创建告警相关业务逻辑
func NewAlarmBiz(alarmRepository repository.Alarm, alarmRawRepository repository.AlarmRaw, historyRepository repository.HistoryRepository) *AlarmBiz {
	return &AlarmBiz{
		alarmRepository:    alarmRepository,
		alarmRawRepository: alarmRawRepository,
		historyRepository:  historyRepository,
	}
}

// AlarmBiz 告警相关业务逻辑
type AlarmBiz struct {
	alarmRepository    repository.Alarm
	alarmRawRepository repository.AlarmRaw
	historyRepository  repository.HistoryRepository
}

// GetRealTimeAlarm 获取实时告警明细
func (b *AlarmBiz) GetRealTimeAlarm(ctx context.Context, params *bo.GetRealTimeAlarmParams) (*alarmmodel.RealtimeAlarm, error) {
	return b.alarmRepository.GetRealTimeAlarm(ctx, params)
}

// ListRealTimeAlarms 获取实时告警列表
func (b *AlarmBiz) ListRealTimeAlarms(ctx context.Context, params *bo.GetRealTimeAlarmsParams) ([]*alarmmodel.RealtimeAlarm, error) {
	return b.alarmRepository.GetRealTimeAlarms(ctx, params)
}

// SaveAlertQueue 保存告警队列
func (b *AlarmBiz) SaveAlertQueue(param *bo.CreateAlarmHookRawParams) error {
	return b.alarmRepository.SaveAlertQueue(param)
}

// CreateAlarmRawInfo 创建告警原始信息
func (b *AlarmBiz) CreateAlarmRawInfo(ctx context.Context, param *bo.CreateAlarmHookRawParams) (*alarmmodel.AlarmRaw, error) {
	alarmRawJson, err := json.Marshal(param)
	if !types.IsNil(err) {
		return nil, err
	}

	alarmRaw := &bo.CreateAlarmRawParams{
		RawInfo:     string(alarmRawJson),
		Fingerprint: param.Fingerprint,
	}
	return b.alarmRawRepository.CreateAlarmRaw(ctx, alarmRaw)
}

func (b *AlarmBiz) CreateAlarmInfo(ctx context.Context, params *bo.CreateAlarmHookRawParams) error {
	rawInfo, err := b.CreateAlarmRawInfo(ctx, params)
	if !types.IsNil(err) {
		return err
	}
	for _, alert := range params.Alerts {
		alert.RawID = rawInfo.ID
		// 保存告警历史
		err := b.historyRepository.CreateAlarmHistory(ctx, alert)
		if err != nil {
			log.Error(ctx, "create alarm history error", err)
		}
		// 保存实时告警
		if err := b.alarmRepository.CreateRealTimeAlarm(ctx, alert); err != nil {
			return err
		}
	}
	return nil
}
