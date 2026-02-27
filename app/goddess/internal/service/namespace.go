package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	magicboxv1 "github.com/aide-family/magicbox/api/v1"
)

func NewNamespaceService(namespaceBiz *biz.Namespace) *NamespaceService {
	return &NamespaceService{
		namespaceBiz: namespaceBiz,
	}
}

type NamespaceService struct {
	magicboxv1.UnimplementedNamespaceServer

	namespaceBiz *biz.Namespace
}

func (s *NamespaceService) CreateNamespace(ctx context.Context, req *magicboxv1.CreateNamespaceRequest) (*magicboxv1.CreateNamespaceReply, error) {
	createNamespaceBo := bo.NewCreateNamespaceBo(req)
	if err := s.namespaceBiz.CreateNamespace(ctx, createNamespaceBo); err != nil {
		return nil, err
	}
	return &magicboxv1.CreateNamespaceReply{}, nil
}

func (s *NamespaceService) UpdateNamespace(ctx context.Context, req *magicboxv1.UpdateNamespaceRequest) (*magicboxv1.UpdateNamespaceReply, error) {
	updateNamespaceBo := bo.NewUpdateNamespaceBo(req)
	if err := s.namespaceBiz.UpdateNamespace(ctx, updateNamespaceBo); err != nil {
		return nil, err
	}
	return &magicboxv1.UpdateNamespaceReply{}, nil
}

func (s *NamespaceService) UpdateNamespaceStatus(ctx context.Context, req *magicboxv1.UpdateNamespaceStatusRequest) (*magicboxv1.UpdateNamespaceStatusReply, error) {
	updateNamespaceStatusBo := bo.NewUpdateNamespaceStatusBo(req)
	if err := s.namespaceBiz.UpdateNamespaceStatus(ctx, updateNamespaceStatusBo); err != nil {
		return nil, err
	}
	return &magicboxv1.UpdateNamespaceStatusReply{}, nil
}

func (s *NamespaceService) DeleteNamespace(ctx context.Context, req *magicboxv1.DeleteNamespaceRequest) (*magicboxv1.DeleteNamespaceReply, error) {
	if err := s.namespaceBiz.DeleteNamespace(ctx, snowflake.ParseInt64(req.Uid)); err != nil {
		return nil, err
	}
	return &magicboxv1.DeleteNamespaceReply{}, nil
}

func (s *NamespaceService) GetNamespace(ctx context.Context, req *magicboxv1.GetNamespaceRequest) (*magicboxv1.NamespaceItem, error) {
	namespaceItemBo, err := s.namespaceBiz.GetNamespace(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return namespaceItemBo.ToAPIV1NamespaceItem(), nil
}

func (s *NamespaceService) ListNamespace(ctx context.Context, req *magicboxv1.ListNamespaceRequest) (*magicboxv1.ListNamespaceReply, error) {
	listNamespaceBo := bo.NewListNamespaceBo(req)
	listNamespacePageResponseBo, err := s.namespaceBiz.ListNamespace(ctx, listNamespaceBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListNamespaceReply(listNamespacePageResponseBo), nil
}

func (s *NamespaceService) SelectNamespace(ctx context.Context, req *magicboxv1.SelectNamespaceRequest) (*magicboxv1.SelectNamespaceReply, error) {
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
