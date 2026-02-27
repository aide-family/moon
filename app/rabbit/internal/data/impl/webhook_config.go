package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/convert"
	"github.com/aide-family/rabbit/internal/data/impl/query"
)

func NewWebhookConfigRepository(d *data.Data) repository.WebhookConfig {
	query.SetDefault(d.DB())
	return &webhookConfigRepository{Data: d}
}

type webhookConfigRepository struct {
	*data.Data
}

// DeleteWebhookConfig implements [repository.WebhookConfig].
func (w *webhookConfigRepository) DeleteWebhookConfig(ctx context.Context, uid snowflake.ID) error {
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), webhookConfig.ID.Eq(uid.Int64()))
	_, err := wrappers.Delete()
	return err
}

// GetWebhookConfig implements [repository.WebhookConfig].
func (w *webhookConfigRepository) GetWebhookConfig(ctx context.Context, uid snowflake.ID) (*bo.WebhookItemBo, error) {
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), webhookConfig.ID.Eq(uid.Int64()))
	webhookConfigDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("webhook config not found")
		}
		return nil, err
	}
	return convert.ToWebhookConfigItemBo(webhookConfigDO), nil
}

// GetWebhookConfigByName implements [repository.WebhookConfig].
func (w *webhookConfigRepository) GetWebhookConfigByName(ctx context.Context, name string) (*bo.WebhookItemBo, error) {
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), webhookConfig.Name.Eq(name))
	webhookConfigDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("webhook config not found")
		}
		return nil, err
	}
	return convert.ToWebhookConfigItemBo(webhookConfigDO), nil
}

// ListWebhookConfig implements [repository.WebhookConfig].
func (w *webhookConfigRepository) ListWebhookConfig(ctx context.Context, req *bo.ListWebhookBo) (*bo.PageResponseBo[*bo.WebhookItemBo], error) {
	namespace := contextx.GetNamespace(ctx)
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(namespace.Int64()))
	if req.App > enum.WebhookAPP_WebhookAPP_UNKNOWN {
		wrappers = wrappers.Where(webhookConfig.App.Eq(int32(req.App)))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(webhookConfig.Status.Eq(int32(req.Status)))
	}
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(webhookConfig.Name.Like("%" + req.Keyword + "%"))
	}
	if pointer.IsNotNil(req.PageRequestBo) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(req.Limit()).Offset(req.Offset())
	}
	webhookConfigs, err := wrappers.Order(webhookConfig.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	webhookConfigItems := make([]*bo.WebhookItemBo, 0, len(webhookConfigs))
	for _, webhookConfig := range webhookConfigs {
		webhookConfigItems = append(webhookConfigItems, convert.ToWebhookConfigItemBo(webhookConfig))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, webhookConfigItems), nil
}

// SelectWebhookConfig implements [repository.WebhookConfig].
func (w *webhookConfigRepository) SelectWebhookConfig(ctx context.Context, req *bo.SelectWebhookBo) (*bo.SelectWebhookBoResult, error) {
	namespace := contextx.GetNamespace(ctx)
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(namespace.Int64()))

	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(webhookConfig.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(webhookConfig.Status.Eq(int32(req.Status)))
	}
	if req.App > enum.WebhookAPP_WebhookAPP_UNKNOWN {
		wrappers = wrappers.Where(webhookConfig.App.Eq(int32(req.App)))
	}

	// 获取总数
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}

	// 游标分页：如果提供了lastUID，则查询UID小于lastUID的记录
	if req.LastUID > 0 {
		wrappers = wrappers.Where(webhookConfig.ID.Lt(req.LastUID.Int64()))
	}

	// 限制返回数量
	wrappers = wrappers.Limit(int(req.Limit))

	// 按UID倒序排列（snowflake ID按时间生成，与CreatedAt一致）
	webhookConfigs, err := wrappers.Order(webhookConfig.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}

	// 获取最后一个UID，用于下次分页
	var lastUID snowflake.ID
	if len(webhookConfigs) > 0 {
		lastUID = webhookConfigs[len(webhookConfigs)-1].ID
	}
	webhookConfigItems := make([]*bo.WebhookItemSelectBo, 0, len(webhookConfigs))
	for _, webhookConfig := range webhookConfigs {
		webhookConfigItems = append(webhookConfigItems, convert.ToWebhookConfigItemSelectBo(webhookConfig))
	}

	return &bo.SelectWebhookBoResult{
		Items:   webhookConfigItems,
		Total:   total,
		LastUID: lastUID,
	}, nil
}

// UpdateWebhookConfig implements [repository.WebhookConfig].
func (w *webhookConfigRepository) UpdateWebhookConfig(ctx context.Context, req *bo.UpdateWebhookBo) error {
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), webhookConfig.ID.Eq(req.UID.Int64()))
	columns := []field.AssignExpr{
		webhookConfig.Name.Value(req.Name),
		webhookConfig.URL.Value(req.URL),
		webhookConfig.Method.Value(int32(req.Method)),
		webhookConfig.Headers.Value(safety.NewMap(req.Headers)),
		webhookConfig.Secret.Value(strutil.EncryptString(req.Secret)),
	}
	_, err := wrappers.UpdateColumnSimple(columns...)
	return err
}

// UpdateWebhookStatus implements [repository.WebhookConfig].
func (w *webhookConfigRepository) UpdateWebhookStatus(ctx context.Context, req *bo.UpdateWebhookStatusBo) error {
	webhookConfig := query.WebhookConfig
	wrappers := webhookConfig.WithContext(ctx).Where(webhookConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), webhookConfig.ID.Eq(req.UID.Int64()))
	_, err := wrappers.UpdateColumn(webhookConfig.Status, req.Status)
	return err
}

// CreateWebhookConfig implements [repository.WebhookConfig].
func (w *webhookConfigRepository) CreateWebhookConfig(ctx context.Context, req *bo.CreateWebhookBo) (snowflake.ID, error) {
	webhookConfig := query.WebhookConfig
	webhookConfigDO := convert.ToWebhookConfigDO(ctx, req)
	if err := webhookConfig.WithContext(ctx).Create(webhookConfigDO); err != nil {
		return 0, err
	}
	return webhookConfigDO.ID, nil
}
