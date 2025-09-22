package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/rabbit/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
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

func (c *Config) GetEmailConfig(ctx context.Context, params *bo.GetEmailConfigParams) (bo.EmailConfig, error) {
	if validate.IsNil(params) || validate.TextIsNull(params.Name) {
		return nil, merr.ErrorParams("No email configuration is available")
	}
	emailConfig, ok := c.configRepo.GetEmailConfig(ctx, params.TeamID, params.Name)
	if !ok || !emailConfig.GetEnable() {
		return nil, merr.ErrorParams("No email configuration is available")
	}
	return emailConfig, nil
}

func (c *Config) SetEmailConfig(ctx context.Context, params *bo.SetEmailConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetEmailConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) GetSMSConfig(ctx context.Context, params *bo.GetSMSConfigParams) (bo.SMSConfig, error) {
	if validate.IsNil(params) || validate.TextIsNull(params.Name) {
		return nil, merr.ErrorParams("No SMS configuration is available")
	}
	smsConfig, ok := c.configRepo.GetSMSConfig(ctx, params.TeamID, params.Name)
	if !ok || !smsConfig.GetEnable() {
		return nil, merr.ErrorParams("No SMS configuration is available")
	}
	return smsConfig, nil
}

func (c *Config) SetSMSConfig(ctx context.Context, params *bo.SetSMSConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetSMSConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) GetHookConfig(ctx context.Context, params *bo.GetHookConfigParams) (bo.HookConfig, error) {
	if validate.IsNil(params) || validate.TextIsNull(params.Name) {
		return nil, merr.ErrorParams("No hook configuration is available")
	}
	hookConfig, ok := c.configRepo.GetHookConfig(ctx, params.TeamID, params.Name)
	if !ok || !hookConfig.GetEnable() {
		return nil, merr.ErrorParams("No hook configuration is available")
	}
	return hookConfig, nil
}

func (c *Config) SetHookConfig(ctx context.Context, params *bo.SetHookConfigParams) error {
	if len(params.Configs) == 0 {
		return nil
	}
	return c.configRepo.SetHookConfig(ctx, params.TeamID, params.Configs...)
}

func (c *Config) GetNoticeGroupConfig(ctx context.Context, params *bo.GetNoticeGroupConfigParams) (bo.NoticeGroup, error) {
	if validate.IsNil(params) || validate.TextIsNull(params.Name) {
		return nil, merr.ErrorParams("No notice group configuration is available")
	}
	noticeGroupConfig, ok := c.configRepo.GetNoticeGroupConfig(ctx, params.TeamID, params.Name)
	if !ok {
		return nil, merr.ErrorParams("No notice group configuration is available")
	}
	return noticeGroupConfig, nil
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
