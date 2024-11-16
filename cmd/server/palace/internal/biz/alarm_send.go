package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/alarmmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

type (
	// AlarmSendBiz 告警发送相关业务
	AlarmSendBiz struct {
		alarmSendRepository repository.AlarmSendRepository
	}
)

// NewAlarmSendBiz 创建告警发送相关业务
func NewAlarmSendBiz(alarmSendRepository repository.AlarmSendRepository) *AlarmSendBiz {
	return &AlarmSendBiz{alarmSendRepository: alarmSendRepository}
}

// GetAlarmSendDetail 获取告警发送详情
func (a *AlarmSendBiz) GetAlarmSendDetail(ctx context.Context, param *bo.GetAlarmSendHistoryParams) (*alarmmodel.AlarmSendHistory, error) {
	return a.alarmSendRepository.GetAlarmSendHistory(ctx, param)
}

// ListAlarmSendHistories 获取告警发送历史列表
func (a *AlarmSendBiz) ListAlarmSendHistories(ctx context.Context, param *bo.QueryAlarmSendHistoryListParams) ([]*alarmmodel.AlarmSendHistory, error) {
	return a.alarmSendRepository.AlarmSendHistoryList(ctx, param)
}

// RetryAlarmSend 重试告警发送
func (a *AlarmSendBiz) RetryAlarmSend(ctx context.Context, param *bo.RetryAlarmSendParams) error {
	if types.IsNil(param.RequestID) {
		return merr.ErrorI18nParameterRelatedIdMustNotBeEmpty(ctx)
	}
	return a.alarmSendRepository.RetryAlarmSend(ctx, param)
}
