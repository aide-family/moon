package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
)

type (

	// AlarmSendRepository is a AlarmSend repo.
	AlarmSendRepository interface {
		// GetAlarmSendHistory 获取告警发送历史
		GetAlarmSendHistory(ctx context.Context, param *bo.GetAlarmSendHistoryParams) (*alarmmodel.AlarmSendHistory, error)
		// AlarmSendHistoryList 获取告警发送历史列表
		AlarmSendHistoryList(ctx context.Context, param *bo.QueryAlarmSendHistoryListParams) ([]*alarmmodel.AlarmSendHistory, error)
		// SaveAlarmSendHistory 保存告警发送记录
		SaveAlarmSendHistory(ctx context.Context, param *bo.CreateAlarmSendParams) error
		// RetryAlarmSend 重试告警发送
		RetryAlarmSend(ctx context.Context, param *bo.RetryAlarmSendParams) error
		// GetRetryNumberByRequestID 获取重试次数
		GetRetryNumberByRequestID(ctx context.Context, requestID string, teamID uint32) (int, error)
	}
)
