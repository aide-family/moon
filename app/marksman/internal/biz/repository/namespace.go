package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/marksman/internal/biz/bo"
)

type Namespace interface {
	GetNamespace(ctx context.Context, uid snowflake.ID) (*bo.NamespaceItemBo, error)
	SelectNamespace(ctx context.Context, req *bo.SelectNamespaceBo) (*bo.SelectNamespaceBoResult, error)
}
