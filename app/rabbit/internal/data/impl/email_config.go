package impl

import (
	"context"
	"errors"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
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

func NewEmailConfigRepository(d *data.Data) repository.EmailConfig {
	query.SetDefault(d.DB())
	return &emailConfigRepository{Data: d}
}

type emailConfigRepository struct {
	*data.Data
}

// CreateEmailConfig implements [repository.EmailConfig].
func (e *emailConfigRepository) CreateEmailConfig(ctx context.Context, req *bo.CreateEmailConfigBo) (snowflake.ID, error) {
	emailConfigMutation := query.EmailConfig
	emailConfigDo := convert.ToEmailConfigDo(ctx, req)
	if err := emailConfigMutation.WithContext(ctx).Create(emailConfigDo); err != nil {
		return 0, err
	}
	return emailConfigDo.ID, nil
}

// DeleteEmailConfig implements [repository.EmailConfig].
func (e *emailConfigRepository) DeleteEmailConfig(ctx context.Context, uid snowflake.ID) error {
	namespace := contextx.GetNamespace(ctx)
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(namespace.Int64()), emailConfig.ID.Eq(uid.Int64()))
	_, err := wrappers.Delete()
	return err
}

// GetEmailConfig implements [repository.EmailConfig].
func (e *emailConfigRepository) GetEmailConfig(ctx context.Context, uid snowflake.ID) (*bo.EmailConfigItemBo, error) {
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), emailConfig.ID.Eq(uid.Int64()))
	emailConfigDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("email config not found")
		}
		return nil, err
	}
	return convert.ToEmailConfigBO(emailConfigDO), nil
}

// GetEmailConfigByName implements [repository.EmailConfig].
func (e *emailConfigRepository) GetEmailConfigByName(ctx context.Context, name string) (*bo.EmailConfigItemBo, error) {
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), emailConfig.Name.Eq(name))
	emailConfigDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("email config not found")
		}
		return nil, err
	}
	return convert.ToEmailConfigBO(emailConfigDO), nil
}

// ListEmailConfig implements [repository.EmailConfig].
func (e *emailConfigRepository) ListEmailConfig(ctx context.Context, req *bo.ListEmailConfigBo) (*bo.PageResponseBo[*bo.EmailConfigItemBo], error) {
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(emailConfig.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(emailConfig.Status.Eq(int32(req.Status)))
	}
	if pointer.IsNotNil(req.PageRequestBo) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(req.Limit()).Offset(req.Offset())
	}
	emailConfigs, err := wrappers.Order(emailConfig.CreatedAt.Desc()).Find()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pageRequestBo := bo.NewPageRequestBo(req.Page, req.PageSize)
			pageRequestBo.WithTotal(0)
			req.PageRequestBo = pageRequestBo
			return bo.NewPageResponseBo(req.PageRequestBo, []*bo.EmailConfigItemBo{}), nil
		}
		return nil, err
	}
	emailConfigItems := make([]*bo.EmailConfigItemBo, 0, len(emailConfigs))
	for _, emailConfig := range emailConfigs {
		emailConfigItems = append(emailConfigItems, convert.ToEmailConfigBO(emailConfig))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, emailConfigItems), nil
}

// SelectEmailConfig implements [repository.EmailConfig].
func (e *emailConfigRepository) SelectEmailConfig(ctx context.Context, req *bo.SelectEmailConfigBo) (*bo.SelectEmailConfigBoResult, error) {
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()))

	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(emailConfig.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(emailConfig.Status.Eq(int32(req.Status)))
	}

	// 获取总数
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}

	// 游标分页：如果提供了lastUID，则查询UID小于lastUID的记录
	if req.LastUID > 0 {
		wrappers = wrappers.Where(emailConfig.ID.Lt(req.LastUID.Int64()))
	}

	// 限制返回数量
	wrappers = wrappers.Limit(int(req.Limit))

	// 按UID倒序排列（snowflake ID按时间生成，与CreatedAt一致）
	emailConfigs, err := wrappers.Order(emailConfig.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}

	// 获取最后一个UID，用于下次分页
	var lastUID snowflake.ID
	if len(emailConfigs) > 0 {
		lastUID = emailConfigs[len(emailConfigs)-1].ID
	}
	emailConfigItems := make([]*bo.EmailConfigItemSelectBo, 0, len(emailConfigs))
	for _, emailConfig := range emailConfigs {
		emailConfigItems = append(emailConfigItems, convert.ToEmailConfigItemSelectBO(emailConfig))
	}

	return &bo.SelectEmailConfigBoResult{
		Items:   emailConfigItems,
		Total:   total,
		LastUID: lastUID,
	}, nil
}

// UpdateEmailConfig implements [repository.EmailConfig].
func (e *emailConfigRepository) UpdateEmailConfig(ctx context.Context, req *bo.UpdateEmailConfigBo) error {
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), emailConfig.ID.Eq(req.UID.Int64()))
	columns := []field.AssignExpr{
		emailConfig.Name.Value(req.Name),
		emailConfig.Host.Value(req.Host),
		emailConfig.Port.Value(req.Port),
		emailConfig.Username.Value(req.Username),
		emailConfig.Password.Value(strutil.EncryptString(req.Password)),
	}
	_, err := wrappers.UpdateColumnSimple(columns...)
	return err
}

// UpdateEmailConfigStatus implements [repository.EmailConfig].
func (e *emailConfigRepository) UpdateEmailConfigStatus(ctx context.Context, req *bo.UpdateEmailConfigStatusBo) error {
	emailConfig := query.EmailConfig
	wrappers := emailConfig.WithContext(ctx).Where(emailConfig.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), emailConfig.ID.Eq(req.UID.Int64()))
	_, err := wrappers.UpdateColumn(emailConfig.Status, req.Status)
	return err
}
