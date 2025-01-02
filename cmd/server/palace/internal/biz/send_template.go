package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

type (
	// SendTemplateBiz 告警发送模板
	SendTemplateBiz struct {
		sendTemplateRepo repository.SendTemplateRepo
		teamTemplateRepo repository.TeamSendTemplate
	}
)

// NewSendTemplateBiz 创建告警发送模板
func NewSendTemplateBiz(sendTemplateRepo repository.SendTemplateRepo, teamTemplateRepo repository.TeamSendTemplate) *SendTemplateBiz {
	return &SendTemplateBiz{
		sendTemplateRepo: sendTemplateRepo,
		teamTemplateRepo: teamTemplateRepo,
	}
}

func (b *SendTemplateBiz) getSendTemplateRepo(ctx context.Context) repository.SendTemplateRepo {
	if middleware.GetSourceType(ctx).IsSystem() {
		return b.sendTemplateRepo
	}
	return b.teamTemplateRepo
}

// CreateSendTemplate 创建告警发送模板
func (b *SendTemplateBiz) CreateSendTemplate(ctx context.Context, param *bo.CreateSendTemplate) error {
	// 校验名称是否存在
	isExist := b.templateNameIsExist(ctx, param.Name)
	if isExist {
		return merr.ErrorI18nToastSendTemplateNameExist(ctx, param.Name)
	}
	err := b.getSendTemplateRepo(ctx).Create(ctx, param)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateSendTemplate 获取告警发送模板
func (b *SendTemplateBiz) UpdateSendTemplate(ctx context.Context, param *bo.UpdateSendTemplate) error {
	err := b.getSendTemplateRepo(ctx).UpdateByID(ctx, param)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateSendTemplateStatus 更新告警发送模板状态
func (b *SendTemplateBiz) UpdateSendTemplateStatus(ctx context.Context, param *bo.UpdateSendTemplateStatusParams) error {
	err := b.getSendTemplateRepo(ctx).UpdateStatusByIds(ctx, param)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// GetSendTemplateDetail 获取告警发送模板
func (b *SendTemplateBiz) GetSendTemplateDetail(ctx context.Context, ID uint32) (imodel.ISendTemplate, error) {
	sendTemplate, err := b.getSendTemplateRepo(ctx).GetByID(ctx, ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastSendTemplateTypeNotExist(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return sendTemplate, nil
}

// SendTemplateList 获取告警发送模板列表
func (b *SendTemplateBiz) SendTemplateList(ctx context.Context, param *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error) {
	templates, err := b.getSendTemplateRepo(ctx).FindByPage(ctx, param)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return templates, nil
}

// DeleteSendTemplate 删除告警发送模板
func (b *SendTemplateBiz) DeleteSendTemplate(ctx context.Context, ID uint32) error {
	err := b.getSendTemplateRepo(ctx).DeleteByID(ctx, ID)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// templateNameIsExist 模板名称是否存在
func (b *SendTemplateBiz) templateNameIsExist(ctx context.Context, name string) bool {
	template, err := b.getSendTemplateRepo(ctx).GetTemplateInfoByName(ctx, name)
	if err != nil {
		return false
	}
	if !types.IsNil(template) {
		return true
	}
	return false
}
