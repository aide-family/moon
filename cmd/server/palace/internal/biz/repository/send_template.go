package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// SendTemplateRepo 告警发送模板仓库接口
	SendTemplateRepo interface {
		// Create 创建告警发送模板
		Create(ctx context.Context, params *bo.CreateSendTemplate) error
		// UpdateByID 更新告警发送模板
		UpdateByID(ctx context.Context, params *bo.UpdateSendTemplate) error
		// DeleteByID 删除告警发送模板
		DeleteByID(ctx context.Context, id uint32) error
		// FindByPage 分页查询告警发送模板
		FindByPage(ctx context.Context, params *bo.QuerySendTemplateListParams) ([]imodel.ISendTemplate, error)
		// UpdateStatusByIds 批量更新告警发送模板状态
		UpdateStatusByIds(ctx context.Context, status *bo.UpdateSendTemplateStatusParams) error
		// GetByID 根据ID查询告警发送模板
		GetByID(ctx context.Context, id uint32) (imodel.ISendTemplate, error)
		// GetTemplateInfoByName 根据名称查询告警发送模板
		GetTemplateInfoByName(ctx context.Context, name string, sendType vobj.AlarmSendType) (imodel.ISendTemplate, error)
	}
)
