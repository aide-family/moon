package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

type Namespace interface {
	AllNamespaces(ctx context.Context) ([]*bo.NamespaceItemBo, error)
	CreateNamespace(ctx context.Context, req *bo.CreateNamespaceBo) error
	UpdateNamespace(ctx context.Context, req *bo.UpdateNamespaceBo) error
	UpdateNamespaceStatus(ctx context.Context, req *bo.UpdateNamespaceStatusBo) error
	DeleteNamespace(ctx context.Context, uid snowflake.ID) error
	GetNamespace(ctx context.Context, uid snowflake.ID) (*bo.NamespaceItemBo, error)
	GetNamespaceByName(ctx context.Context, name string) (*bo.NamespaceItemBo, error)
	ListNamespacesByUIDs(ctx context.Context, uids []snowflake.ID) ([]*bo.NamespaceItemBo, error)
	ListNamespace(ctx context.Context, req *bo.ListNamespaceBo) (*bo.PageResponseBo[*bo.NamespaceItemBo], error)
	SelectNamespace(ctx context.Context, req *bo.SelectNamespaceBo) (*bo.SelectNamespaceBoResult, error)
}
