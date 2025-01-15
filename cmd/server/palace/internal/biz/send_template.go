package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

type (
	// SendTemplateBiz 告警发送模板
	SendTemplateBiz struct {
		sysTemplateRepo  repository.SendTemplateRepo
		teamTemplateRepo repository.TeamSendTemplate
	}
)

// NewSendTemplateBiz 创建告警发送模板
func NewSendTemplateBiz(sendTemplateRepo repository.SendTemplateRepo, teamTemplateRepo repository.TeamSendTemplate) *SendTemplateBiz {
	return &SendTemplateBiz{
		sysTemplateRepo:  sendTemplateRepo,
		teamTemplateRepo: teamTemplateRepo,
	}
}

func (b *SendTemplateBiz) getSendTemplateRepo(ctx context.Context) repository.SendTemplateRepo {
	if middleware.GetSourceType(ctx).IsSystem() {
		return b.sysTemplateRepo
	}
	return b.teamTemplateRepo
}

// CreateSendTemplate 创建告警发送模板
func (b *SendTemplateBiz) CreateSendTemplate(ctx context.Context, param *bo.CreateSendTemplate) error {
	// 校验名称是否存在
	if err := b.templateNameIsExist(ctx, param.Name, param.SendType); types.IsNotNil(err) {
		return err
	}
	if err := b.getSendTemplateRepo(ctx).Create(ctx, param); types.IsNotNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateSendTemplate 获取告警发送模板
func (b *SendTemplateBiz) UpdateSendTemplate(ctx context.Context, param *bo.UpdateSendTemplate) error {
	temp := param.UpdateParam
	// 校验名称是否存在
	if err := b.templateNameIsExist(ctx, temp.Name, temp.SendType, param.ID); types.IsNotNil(err) {
		return err
	}
	if err := b.getSendTemplateRepo(ctx).UpdateByID(ctx, param); types.IsNotNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateSendTemplateStatus 更新告警发送模板状态
func (b *SendTemplateBiz) UpdateSendTemplateStatus(ctx context.Context, param *bo.UpdateSendTemplateStatusParams) error {
	if err := b.getSendTemplateRepo(ctx).UpdateStatusByIds(ctx, param); types.IsNotNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// GetSendTemplateDetail 获取告警发送模板
func (b *SendTemplateBiz) GetSendTemplateDetail(ctx context.Context, id uint32) (imodel.ISendTemplate, error) {
	sendTemplate, err := b.getSendTemplateRepo(ctx).GetByID(ctx, id)
	if types.IsNotNil(err) {
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
	if types.IsNotNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return templates, nil
}

// DeleteSendTemplate 删除告警发送模板
func (b *SendTemplateBiz) DeleteSendTemplate(ctx context.Context, ID uint32) error {
	if err := b.getSendTemplateRepo(ctx).DeleteByID(ctx, ID); types.IsNotNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// templateNameIsExist 模板名称是否存在
func (b *SendTemplateBiz) templateNameIsExist(ctx context.Context, name string, sendType vobj.AlarmSendType, id ...uint32) error {
	template, err := b.getSendTemplateRepo(ctx).GetTemplateInfoByName(ctx, name, sendType)
	if types.IsNotNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if types.IsNotNil(template) && (len(id) == 0 || template.GetID() != id[0]) {
		return merr.ErrorI18nToastSendTemplateNameExist(ctx, name)
	}
	return nil
}
