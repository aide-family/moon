package convert

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/safety"

	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

func NamespaceToBo(namespace *do.Namespace) *bo.NamespaceItemBo {
	if pointer.IsNil(namespace.Metadata) {
		namespace.Metadata = safety.NewMap(make(map[string]string))
	}
	return &bo.NamespaceItemBo{
		UID:       namespace.UID,
		Name:      namespace.Name,
		CreatedAt: namespace.CreatedAt,
		UpdatedAt: namespace.UpdatedAt,
		Metadata:  namespace.Metadata.Map(),
		Status:    namespace.Status,
	}
}

func NamespaceToDo(ctx context.Context, bo *bo.CreateNamespaceBo) *do.Namespace {
	namespaceModel := &do.Namespace{
		Name:     bo.Name,
		Metadata: safety.NewMap(bo.Metadata),
		Status:   bo.Status,
	}
	namespaceModel.WithCreator(contextx.GetUserUID(ctx))
	return namespaceModel
}

func NamespaceToSelectBo(namespace *do.Namespace) *bo.NamespaceItemSelectBo {
	return &bo.NamespaceItemSelectBo{
		UID:      namespace.UID,
		Name:     namespace.Name,
		Status:   namespace.Status,
		Disabled: namespace.Status != enum.GlobalStatus_ENABLED || namespace.DeletedAt.Valid,
		Tooltip:  namespace.Name,
	}
}
