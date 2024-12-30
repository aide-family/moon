package biz

import (
	"context"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/imodel"
)

type (
	SendTemplateBiz struct {
	}
)

// CreateSendTemplate 创建告警发送模板
func (b *SendTemplateBiz) CreateSendTemplate(ctx context.Context, param *bo.CreateAlarmSendParams) error {

	return nil
}

// UpdateSendTemplate 获取告警发送模板
func (b *SendTemplateBiz) UpdateSendTemplate(ctx context.Context, param *bo.UpdateSendTemplate) error {

	return nil
}

// UpdateSendTemplateStatus 更新告警发送模板状态
func (b *SendTemplateBiz) UpdateSendTemplateStatus(ctx context.Context, param *bo.UpdateSendTemplateStatus) error {

	return nil
}

// GetSendTemplateDetail 获取告警发送模板
func (b *SendTemplateBiz) GetSendTemplateDetail(ctx context.Context, ID uint32) (imodel.ISendTemplate, error) {
	return nil, nil
}

// SendTemplateList 获取告警发送模板列表
func (b *SendTemplateBiz) SendTemplateList(ctx context.Context, param *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error) {

	return nil, nil
}

// DeleteSendTemplate 删除告警发送模板
func (b *SendTemplateBiz) DeleteSendTemplate(ctx context.Context, ID uint32) error {

	return nil
}
