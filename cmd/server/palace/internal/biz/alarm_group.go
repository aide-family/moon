package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewAlarmGroupBiz 创建告警分组业务
func NewAlarmGroupBiz(strategy repository.AlarmGroup) *AlarmGroupBiz {
	return &AlarmGroupBiz{
		strategyRepo: strategy,
	}
}

type (
	// AlarmGroupBiz 告警分组业务
	AlarmGroupBiz struct {
		strategyRepo repository.AlarmGroup
	}
)

// CreateAlarmGroup 创建告警分组
func (s *AlarmGroupBiz) CreateAlarmGroup(ctx context.Context, params *bo.CreateAlarmNoticeGroupParams) (*bizmodel.AlarmNoticeGroup, error) {
	alarmGroup, err := s.strategyRepo.CreateAlarmGroup(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return alarmGroup, nil
}

// UpdateAlarmGroup 更新告警分组
func (s *AlarmGroupBiz) UpdateAlarmGroup(ctx context.Context, params *bo.UpdateAlarmNoticeGroupParams) error {
	if err := s.strategyRepo.UpdateAlarmGroup(ctx, params); !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nAlarmGroupDataNotFoundErr(ctx)
		}
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// GetAlarmGroupDetail 获取告警分组详情
func (s *AlarmGroupBiz) GetAlarmGroupDetail(ctx context.Context, groupID uint32) (*bizmodel.AlarmNoticeGroup, error) {
	alarmGroup, err := s.strategyRepo.GetAlarmGroup(ctx, groupID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nAlarmGroupDataNotFoundErr(ctx)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return alarmGroup, nil
}

// DeleteAlarmGroup 删除告警分组
func (s *AlarmGroupBiz) DeleteAlarmGroup(ctx context.Context, alarmID uint32) error {
	if err := s.strategyRepo.DeleteAlarmGroup(ctx, alarmID); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// UpdateStatus 更新告警分组状态
func (s *AlarmGroupBiz) UpdateStatus(ctx context.Context, params *bo.UpdateAlarmNoticeGroupStatusParams) error {
	if err := s.strategyRepo.UpdateStatus(ctx, params); !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nAlarmGroupDataNotFoundErr(ctx)
		}
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// ListPage 分页查询告警分组
func (s *AlarmGroupBiz) ListPage(ctx context.Context, params *bo.QueryAlarmNoticeGroupListParams) ([]*bizmodel.AlarmNoticeGroup, error) {
	alarmGroups, err := s.strategyRepo.AlarmGroupPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return alarmGroups, nil
}
