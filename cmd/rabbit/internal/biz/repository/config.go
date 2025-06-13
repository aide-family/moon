package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
)

type Config interface {
	GetEmailConfig(ctx context.Context, teamID uint32, name string) (bo.EmailConfig, bool)
	GetEmailConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.EmailConfig, error)
	SetEmailConfig(ctx context.Context, teamID uint32, configs ...bo.EmailConfig) error

	GetSMSConfig(ctx context.Context, teamID uint32, name string) (bo.SMSConfig, bool)
	GetSMSConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.SMSConfig, error)
	SetSMSConfig(ctx context.Context, teamID uint32, configs ...bo.SMSConfig) error

	GetHookConfig(ctx context.Context, teamID uint32, name string) (bo.HookConfig, bool)
	GetHookConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.HookConfig, error)
	SetHookConfig(ctx context.Context, teamID uint32, configs ...bo.HookConfig) error

	GetNoticeGroupConfig(ctx context.Context, teamID uint32, name string) (bo.NoticeGroup, bool)
	GetNoticeGroupConfigs(ctx context.Context, teamID uint32, names ...string) ([]bo.NoticeGroup, error)
	SetNoticeGroupConfig(ctx context.Context, teamID uint32, configs ...bo.NoticeGroup) error

	RemoveConfig(ctx context.Context, params *bo.RemoveConfigParams) error
}
