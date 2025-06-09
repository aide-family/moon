package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewConfig(configRepo repository.Config, logger log.Logger) *Config {
	return &Config{
		helper:     log.NewHelper(log.With(logger, "module", "biz.config")),
		configRepo: configRepo,
	}
}

type Config struct {
	helper     *log.Helper
	configRepo repository.Config
}

func (c *Config) GetEmailConfig(ctx context.Context, params *bo.GetEmailConfigParams) bo.EmailConfig {
	if validate.IsNil(params.Name) || *params.Name == "" {
		return params.DefaultEmailConfig
	}
	emailConfig, ok := c.configRepo.GetEmailConfig(ctx, params.TeamID, *params.Name)
	if !ok || !emailConfig.GetEnable() {
		return params.DefaultEmailConfig
	}
	return emailConfig
}

func (c *Config) SetEmailConfig(ctx context.Context, params *bo.SetEmailConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetEmailConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) GetSMSConfig(ctx context.Context, params *bo.GetSMSConfigParams) bo.SMSConfig {
	if validate.IsNil(params.Name) || *params.Name == "" {
		return params.DefaultSMSConfig
	}
	smsConfig, ok := c.configRepo.GetSMSConfig(ctx, params.TeamID, *params.Name)
	if !ok || !smsConfig.GetEnable() {
		return params.DefaultSMSConfig
	}
	return smsConfig
}

func (c *Config) SetSMSConfig(ctx context.Context, params *bo.SetSMSConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetSMSConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) GetHookConfig(ctx context.Context, params *bo.GetHookConfigParams) bo.HookConfig {
	if validate.IsNil(params.Name) || *params.Name == "" {
		return params.DefaultHookConfig
	}
	hookConfig, ok := c.configRepo.GetHookConfig(ctx, params.TeamID, *params.Name)
	if !ok || !hookConfig.GetEnable() {
		return params.DefaultHookConfig
	}
	return hookConfig
}

func (c *Config) SetHookConfig(ctx context.Context, params *bo.SetHookConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetHookConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) GetNoticeGroupConfig(ctx context.Context, params *bo.GetNoticeGroupConfigParams) bo.NoticeGroup {
	if validate.IsNil(params.Name) || *params.Name == "" {
		return params.DefaultNoticeGroup
	}
	noticeGroupConfig, ok := c.configRepo.GetNoticeGroupConfig(ctx, params.TeamID, *params.Name)
	if !ok {
		return params.DefaultNoticeGroup
	}
	return noticeGroupConfig
}

func (c *Config) SetNoticeGroupConfig(ctx context.Context, params *bo.SetNoticeGroupConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetNoticeGroupConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) RemoveConfig(ctx context.Context, params *bo.RemoveConfigParams) error {
	return c.configRepo.RemoveConfig(ctx, params)
}
