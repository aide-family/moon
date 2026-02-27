package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewEmailConfig(
	emailConfigRepo repository.EmailConfig,
	helper *klog.Helper,
) *EmailConfig {
	return &EmailConfig{
		emailConfigRepo: emailConfigRepo,
		helper:          klog.NewHelper(klog.With(helper.Logger(), "biz", "email_config")),
	}
}

type EmailConfig struct {
	helper          *klog.Helper
	emailConfigRepo repository.EmailConfig
}

func (c *EmailConfig) CreateEmailConfig(ctx context.Context, req *bo.CreateEmailConfigBo) (snowflake.ID, error) {
	if emailConfig, err := c.emailConfigRepo.GetEmailConfigByName(ctx, req.Name); err == nil {
		return 0, merr.ErrorParams("email config %s already exists, uid: %s", req.Name, emailConfig.UID)
	} else if !merr.IsNotFound(err) {
		c.helper.Errorw("msg", "check email config exists failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create email config %s failed", req.Name).WithCause(err)
	}
	uid, err := c.emailConfigRepo.CreateEmailConfig(ctx, req)
	if err != nil {
		c.helper.Errorw("msg", "create email config failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("create email config %s failed", req.Name).WithCause(err)
	}
	return uid, nil
}

func (c *EmailConfig) UpdateEmailConfig(ctx context.Context, req *bo.UpdateEmailConfigBo) error {
	existEmailConfig, err := c.emailConfigRepo.GetEmailConfigByName(ctx, req.Name)
	if err != nil && !merr.IsNotFound(err) {
		c.helper.Errorw("msg", "check email config exists failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("update email config %s failed", req.Name).WithCause(err)
	} else if existEmailConfig != nil && existEmailConfig.UID != req.UID {
		return merr.ErrorParams("email config %s already exists", req.Name)
	}
	if err := c.emailConfigRepo.UpdateEmailConfig(ctx, req); err != nil {
		c.helper.Errorw("msg", "update email config failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("update email config %s failed", req.Name).WithCause(err)
	}
	return nil
}

func (c *EmailConfig) UpdateEmailConfigStatus(ctx context.Context, req *bo.UpdateEmailConfigStatusBo) error {
	if err := c.emailConfigRepo.UpdateEmailConfigStatus(ctx, req); err != nil {
		c.helper.Errorw("msg", "update email config status failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update email config status %d failed", req.UID).WithCause(err)
	}
	return nil
}

func (c *EmailConfig) DeleteEmailConfig(ctx context.Context, uid snowflake.ID) error {
	if err := c.emailConfigRepo.DeleteEmailConfig(ctx, uid); err != nil {
		c.helper.Errorw("msg", "delete email config failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete email config %s failed", uid).WithCause(err)
	}
	return nil
}

func (c *EmailConfig) GetEmailConfig(ctx context.Context, uid snowflake.ID) (*bo.EmailConfigItemBo, error) {
	emailConfig, err := c.emailConfigRepo.GetEmailConfig(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, err
		}
		c.helper.Errorw("msg", "get email config failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get email config %s failed", uid).WithCause(err)
	}
	return emailConfig, nil
}

func (c *EmailConfig) ListEmailConfig(ctx context.Context, req *bo.ListEmailConfigBo) (*bo.PageResponseBo[*bo.EmailConfigItemBo], error) {
	pageResponseBo, err := c.emailConfigRepo.ListEmailConfig(ctx, req)
	if err != nil {
		c.helper.Errorw("msg", "list email config failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list email config failed").WithCause(err)
	}

	return pageResponseBo, nil
}

func (c *EmailConfig) SelectEmailConfig(ctx context.Context, req *bo.SelectEmailConfigBo) (*bo.SelectEmailConfigBoResult, error) {
	result, err := c.emailConfigRepo.SelectEmailConfig(ctx, req)
	if err != nil {
		c.helper.Errorw("msg", "select email config failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select email config failed").WithCause(err)
	}
	return result, nil
}
