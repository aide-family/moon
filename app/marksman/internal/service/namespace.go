package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"

	namespacev1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/marksman/internal/biz"
	"github.com/aide-family/marksman/internal/biz/bo"
)

func NewNamespaceService(namespaceBiz *biz.Namespace) *NamespaceService {
	return &NamespaceService{
		namespaceBiz: namespaceBiz,
	}
}

type NamespaceService struct {
	namespacev1.UnimplementedNamespaceServer

	namespaceBiz *biz.Namespace
}

func (s *NamespaceService) GetNamespace(ctx context.Context, req *namespacev1.GetNamespaceRequest) (*namespacev1.NamespaceItem, error) {
	namespaceItemBo, err := s.namespaceBiz.GetNamespace(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return namespaceItemBo.ToAPIV1NamespaceItem(), nil
}

func (s *NamespaceService) SelectNamespace(ctx context.Context, req *namespacev1.SelectNamespaceRequest) (*namespacev1.SelectNamespaceReply, error) {
	selectBo := bo.NewSelectNamespaceBo(req)
	result, err := s.namespaceBiz.SelectNamespace(ctx, selectBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1SelectNamespaceReply(result), nil
}

func (s *NamespaceService) HasNamespace(ctx context.Context) (snowflake.ID, error) {
	namespace := contextx.GetNamespace(ctx)
	if namespace <= 0 {
		return 0, merr.ErrorForbidden("namespace is required, please set the namespace in the request header or metadata, Example: %s: default", cnst.HTTPHeaderXNamespace)
	}
	namespaceItemBo, err := s.namespaceBiz.GetNamespace(ctx, namespace)
	if err != nil {
		if merr.IsNotFound(err) {
			return 0, merr.ErrorForbidden("namespace %s not allowed", namespace)
		}
		return 0, err
	}
	if namespaceItemBo.Status != enum.GlobalStatus_ENABLED {
		return 0, merr.ErrorForbidden("namespace %s is not allowed", namespace)
	}
	return namespaceItemBo.UID, nil
}
