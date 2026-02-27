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

func NewTemplateRepository(d *data.Data) repository.Template {
	query.SetDefault(d.DB())
	return &templateRepository{Data: d}
}

type templateRepository struct {
	*data.Data
}

// DeleteTemplate implements [repository.Template].
func (t *templateRepository) DeleteTemplate(ctx context.Context, uid snowflake.ID) error {
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), template.ID.Eq(uid.Int64()))
	_, err := wrappers.Delete()
	return err
}

// GetTemplate implements [repository.Template].
func (t *templateRepository) GetTemplate(ctx context.Context, uid snowflake.ID) (*bo.TemplateItemBo, error) {
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), template.ID.Eq(uid.Int64()))
	templateDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("template not found")
		}
		return nil, err
	}
	return convert.ToTemplateItemBo(templateDO), nil
}

// GetTemplateByName implements [repository.Template].
func (t *templateRepository) GetTemplateByName(ctx context.Context, name string) (*bo.TemplateItemBo, error) {
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), template.Name.Eq(name))
	templateDO, err := wrappers.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("template not found")
		}
		return nil, err
	}
	return convert.ToTemplateItemBo(templateDO), nil
}

// ListTemplate implements [repository.Template].
func (t *templateRepository) ListTemplate(ctx context.Context, req *bo.ListTemplateBo) (*bo.PageResponseBo[*bo.TemplateItemBo], error) {
	namespace := contextx.GetNamespace(ctx)
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(namespace.Int64()))
	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(template.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(template.Status.Eq(int32(req.Status)))
	}
	if req.MessageType > enum.MessageType_MessageType_UNKNOWN {
		wrappers = wrappers.Where(template.MessageType.Eq(int32(req.MessageType)))
	}
	if pointer.IsNotNil(req.PageRequestBo) {
		total, err := wrappers.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		wrappers = wrappers.Limit(req.Limit()).Offset(req.Offset())
	}
	templates, err := wrappers.Order(template.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	templateItems := make([]*bo.TemplateItemBo, 0, len(templates))
	for _, template := range templates {
		templateItems = append(templateItems, convert.ToTemplateItemBo(template))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, templateItems), nil
}

// SelectTemplate implements [repository.Template].
func (t *templateRepository) SelectTemplate(ctx context.Context, req *bo.SelectTemplateBo) (*bo.SelectTemplateBoResult, error) {
	namespace := contextx.GetNamespace(ctx)
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(namespace.Int64()))

	if strutil.IsNotEmpty(req.Keyword) {
		wrappers = wrappers.Where(template.Name.Like("%" + req.Keyword + "%"))
	}
	if req.Status > enum.GlobalStatus_GlobalStatus_UNKNOWN {
		wrappers = wrappers.Where(template.Status.Eq(int32(req.Status)))
	}
	if req.MessageType > enum.MessageType_MessageType_UNKNOWN {
		wrappers = wrappers.Where(template.MessageType.Eq(int32(req.MessageType)))
	}

	// 获取总数
	total, err := wrappers.Count()
	if err != nil {
		return nil, err
	}

	// 游标分页：如果提供了lastUID，则查询UID小于lastUID的记录
	if req.LastUID > 0 {
		wrappers = wrappers.Where(template.ID.Lt(req.LastUID.Int64()))
	}

	// 限制返回数量
	wrappers = wrappers.Limit(int(req.Limit))

	// 按UID倒序排列（snowflake ID按时间生成，与CreatedAt一致）
	templates, err := wrappers.Order(template.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}

	// 获取最后一个UID，用于下次分页
	var lastUID snowflake.ID
	if len(templates) > 0 {
		lastUID = templates[len(templates)-1].ID
	}
	templateItems := make([]*bo.TemplateItemSelectBo, 0, len(templates))
	for _, template := range templates {
		templateItems = append(templateItems, convert.ToTemplateItemSelectBo(template))
	}

	return &bo.SelectTemplateBoResult{
		Items:   templateItems,
		Total:   total,
		LastUID: lastUID,
	}, nil
}

// UpdateTemplate implements [repository.Template].
func (t *templateRepository) UpdateTemplate(ctx context.Context, req *bo.UpdateTemplateBo) error {
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), template.ID.Eq(req.UID.Int64()))
	columns := []field.AssignExpr{
		template.Name.Value(req.Name),
		template.MessageType.Value(int32(req.MessageType)),
		template.JSONData.Value([]byte(req.JSONData)),
	}
	_, err := wrappers.UpdateColumnSimple(columns...)
	return err
}

// UpdateTemplateStatus implements [repository.Template].
func (t *templateRepository) UpdateTemplateStatus(ctx context.Context, req *bo.UpdateTemplateStatusBo) error {
	template := query.Template
	wrappers := template.WithContext(ctx).Where(template.NamespaceUID.Eq(contextx.GetNamespace(ctx).Int64()), template.ID.Eq(req.UID.Int64()))
	_, err := wrappers.UpdateColumn(template.Status, req.Status)
	return err
}

// CreateTemplate implements [repository.Template].
func (t *templateRepository) CreateTemplate(ctx context.Context, req *bo.CreateTemplateBo) (snowflake.ID, error) {
	template := query.Template
	templateDO := convert.ToTemplateDO(ctx, req)
	if err := template.WithContext(ctx).Create(templateDO); err != nil {
		return 0, err
	}
	return templateDO.ID, nil
}
