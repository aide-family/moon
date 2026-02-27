package gormimpl

import (
	"time"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/domain/namespace/v1/gormimpl/model"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"
)

func ConvertNamespaceItemSelect(namespaceDo *model.Namespace) *apiv1.NamespaceItemSelect {
	if namespaceDo == nil {
		return nil
	}
	return &apiv1.NamespaceItemSelect{
		Value:    namespaceDo.ID.Int64(),
		Label:    namespaceDo.Name,
		Disabled: namespaceDo.DeletedAt.Valid || namespaceDo.Status != enum.GlobalStatus_ENABLED,
		Tooltip:  namespaceDo.Remark,
	}
}

func ConvertNamespaceItem(namespaceDo *model.Namespace) *apiv1.NamespaceItem {
	if namespaceDo == nil {
		return nil
	}
	if pointer.IsNil(namespaceDo.Metadata) {
		namespaceDo.Metadata = safety.NewMap(make(map[string]string))
	}
	return &apiv1.NamespaceItem{
		Uid:       namespaceDo.ID.Int64(),
		Name:      namespaceDo.Name,
		Metadata:  namespaceDo.Metadata.Map(),
		Status:    namespaceDo.Status,
		CreatedAt: namespaceDo.CreatedAt.Format(time.DateTime),
		UpdatedAt: namespaceDo.UpdatedAt.Format(time.DateTime),
		Remark:    namespaceDo.Remark,
	}
}
