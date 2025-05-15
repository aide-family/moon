package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/rabbit/internal/biz/bo"
)

type Config interface {
	GetEmailConfig(ctx context.Context, teamID string, name string) (bo.EmailConfig, bool)
	GetEmailConfigs(ctx context.Context, teamID string, names ...string) ([]bo.EmailConfig, error)
	SetEmailConfig(ctx context.Context, teamID string, configs ...bo.EmailConfig) error

	GetSMSConfig(ctx context.Context, teamID string, name string) (bo.SMSConfig, bool)
	GetSMSConfigs(ctx context.Context, teamID string, names ...string) ([]bo.SMSConfig, error)
	SetSMSConfig(ctx context.Context, teamID string, configs ...bo.SMSConfig) error

	GetHookConfig(ctx context.Context, teamID string, name string) (bo.HookConfig, bool)
	GetHookConfigs(ctx context.Context, teamID string, names ...string) ([]bo.HookConfig, error)
	SetHookConfig(ctx context.Context, teamID string, configs ...bo.HookConfig) error

	GetNoticeGroupConfig(ctx context.Context, teamID string, name string) (bo.NoticeGroup, bool)
	GetNoticeGroupConfigs(ctx context.Context, teamID string, names ...string) ([]bo.NoticeGroup, error)
	SetNoticeGroupConfig(ctx context.Context, teamID string, configs ...bo.NoticeGroup) error

	GetNoticeUserConfig(ctx context.Context, teamID string, name string) (bo.NoticeUser, bool)
	GetNoticeUserConfigs(ctx context.Context, teamID string, names ...string) ([]bo.NoticeUser, error)
	SetNoticeUserConfig(ctx context.Context, teamID string, configs ...bo.NoticeUser) error

	RemoveConfig(ctx context.Context, params *bo.RemoveConfigParams) error
}
