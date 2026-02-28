package convert

import (
	"time"

	apiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

func ToNamespaceItemBo(namespaceModel *apiv1.NamespaceItem) *bo.NamespaceItemBo {
	if namespaceModel == nil {
		return nil
	}
	createdAt, _ := time.Parse(time.DateTime, namespaceModel.CreatedAt)
	updatedAt, _ := time.Parse(time.DateTime, namespaceModel.UpdatedAt)
	return &bo.NamespaceItemBo{
		UID:       snowflake.ParseInt64(namespaceModel.Uid),
		Name:      namespaceModel.Name,
		Status:    namespaceModel.Status,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func ToNamespaceItemSelectBo(namespaceItemSelect *apiv1.NamespaceItemSelect) *bo.NamespaceItemSelectBo {
	if namespaceItemSelect == nil {
		return nil
	}
	return &bo.NamespaceItemSelectBo{
		Value:    namespaceItemSelect.Value,
		Label:    namespaceItemSelect.Label,
		Disabled: namespaceItemSelect.Disabled,
		Tooltip:  namespaceItemSelect.Tooltip,
	}
}
