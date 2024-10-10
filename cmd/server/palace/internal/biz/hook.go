package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewAlarmHookBiz new AlarmHookBiz
func NewAlarmHookBiz(alarmHook repository.AlarmHook) *AlarmHookBiz {
	return &AlarmHookBiz{
		alarmHookRepo: alarmHook,
	}
}

type (
	// AlarmHookBiz is a greeter service.
	AlarmHookBiz struct {
		alarmHookRepo repository.AlarmHook
	}
)

// CreateAlarmHook create alarm hook
func (s *AlarmHookBiz) CreateAlarmHook(ctx context.Context, params *bo.CreateAlarmHookParams) (*bizmodel.AlarmHook, error) {
	alarmHook, err := s.alarmHookRepo.CreateAlarmHook(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return alarmHook, nil
}

// UpdateAlarmHook update alarm hook
func (s *AlarmHookBiz) UpdateAlarmHook(ctx context.Context, params *bo.UpdateAlarmHookParams) error {
	if err := s.alarmHookRepo.UpdateAlarmHook(ctx, params); !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastAlarmHookNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// DeleteAlarmHook delete alarm hook
func (s *AlarmHookBiz) DeleteAlarmHook(ctx context.Context, ID uint32) error {
	if err := s.alarmHookRepo.DeleteAlarmHook(ctx, ID); !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastAlarmHookNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// GetAlarmHook get alarm hook
func (s *AlarmHookBiz) GetAlarmHook(ctx context.Context, ID uint32) (*bizmodel.AlarmHook, error) {
	alarmHook, err := s.alarmHookRepo.GetAlarmHook(ctx, ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastAlarmHookNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return alarmHook, nil
}

// ListPage list alarm hook page
func (s *AlarmHookBiz) ListPage(ctx context.Context, params *bo.QueryAlarmHookListParams) ([]*bizmodel.AlarmHook, error) {
	alarmHooks, err := s.alarmHookRepo.ListAlarmHook(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return alarmHooks, nil
}

// UpdateStatus update alarm hook status
func (s *AlarmHookBiz) UpdateStatus(ctx context.Context, params *bo.UpdateAlarmHookStatusParams) error {
	if err := s.alarmHookRepo.UpdateAlarmHookStatus(ctx, params); !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastAlarmHookNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}
