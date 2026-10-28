package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// NamespaceRepo 读写租户，以及用户加入过的租户。
type NamespaceRepo interface {
	CreateNamespace(ctx context.Context, name, remark, ownerUID string) (*bo.Namespace, error)
	ListUserNamespaces(ctx context.Context, userUID string) ([]bo.Namespace, error)
	ListNamespaces(ctx context.Context, userUID string, page, pageSize int, name string) ([]bo.Namespace, int, error)
	SelectNamespaces(ctx context.Context, userUID string, limit int, lastUID string) ([]bo.SelectItem, error)
	GetNamespace(ctx context.Context, uid string) (*bo.Namespace, error)
	UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*bo.Namespace, error)
	UpdateNamespaceStatus(ctx context.Context, uid, status string) (*bo.Namespace, error)
	DeleteNamespace(ctx context.Context, uid string) error
	TransferOwner(ctx context.Context, namespaceUID, targetUserUID string) (*bo.Namespace, error)
}
