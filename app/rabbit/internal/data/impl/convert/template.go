package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/data/impl/do"
)

func ToTemplateDO(ctx context.Context, req *bo.CreateTemplateBo) *do.Template {
	model := &do.Template{
		NamespaceUID: contextx.GetNamespace(ctx),
		Name:         req.Name,
		MessageType:  req.MessageType,
		JSONData:     []byte(req.JSONData),
		Status:       enum.GlobalStatus_ENABLED,
	}
	model.WithCreator(contextx.GetUserUID(ctx))
	return model
}

func ToTemplateItemBo(templateDO *do.Template) *bo.TemplateItemBo {
	return &bo.TemplateItemBo{
		UID:         templateDO.ID,
		Name:        templateDO.Name,
		MessageType: templateDO.MessageType,
		JSONData:    string(templateDO.JSONData),
		Status:      templateDO.Status,
		CreatedAt:   templateDO.CreatedAt,
		UpdatedAt:   templateDO.UpdatedAt,
	}
}

func ToTemplateItemSelectBo(templateDO *do.Template) *bo.TemplateItemSelectBo {
	return &bo.TemplateItemSelectBo{
		UID:      templateDO.ID,
		Name:     templateDO.Name,
		Status:   templateDO.Status,
		Disabled: templateDO.Status == enum.GlobalStatus_DISABLED || templateDO.DeletedAt.Valid,
		Tooltip:  templateDO.Name,
	}
}
