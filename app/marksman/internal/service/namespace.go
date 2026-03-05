package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/marksman/internal/biz"
)

func NewNamespaceService(namespaceBiz *biz.Namespace) *NamespaceService {
	return &NamespaceService{
		Namespace: namespaceBiz,
	}
}

type NamespaceService struct {
	*biz.Namespace
}

func (s *NamespaceService) HasNamespace(ctx context.Context) (snowflake.ID, error) {
	namespace := contextx.GetNamespace(ctx)
	if namespace <= 0 {
		return 0, merr.ErrorForbidden("namespace is required, please set the namespace in the request header or metadata, Example: %s: default", cnst.HTTPHeaderXNamespace)
	}
	namespaceItemBo, err := s.Namespace.GetNamespace(ctx, &goddessv1.GetNamespaceRequest{
		Uid: namespace.Int64(),
	})
	if err != nil {
		if merr.IsNotFound(err) {
			return 0, merr.ErrorForbidden("namespace %s not allowed", namespace)
		}
		return 0, err
	}
	if namespaceItemBo.Status != enum.GlobalStatus_ENABLED {
		return 0, merr.ErrorForbidden("namespace %s is not allowed", namespace)
	}
	return snowflake.ParseInt64(namespaceItemBo.Uid), nil
}
