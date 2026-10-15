package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/namespace"
)

type NamespaceService struct {
	v1.UnimplementedNamespaceServiceServer
	namespaces *namespace.Usecase
}

func NewNamespaceService(uc *namespace.Usecase) *NamespaceService {
	return &NamespaceService{namespaces: uc}
}

func (s *NamespaceService) CreateNamespace(ctx context.Context, req *v1.CreateNamespaceRequest) (*v1.CreateNamespaceResponse, error) {
	item, err := s.namespaces.CreateNamespace(ctx, req.GetName(), req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.CreateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) UpdateNamespace(ctx context.Context, req *v1.UpdateNamespaceRequest) (*v1.UpdateNamespaceResponse, error) {
	item, err := s.namespaces.UpdateNamespace(ctx, req.GetUid(), req.GetName(), req.GetRemark(), req.GetLogo())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) UpdateNamespaceStatus(ctx context.Context, req *v1.UpdateNamespaceStatusRequest) (*v1.UpdateNamespaceStatusResponse, error) {
	item, err := s.namespaces.UpdateNamespaceStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNamespaceStatusResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) DeleteNamespace(ctx context.Context, req *v1.DeleteNamespaceRequest) (*v1.DeleteNamespaceResponse, error) {
	if err := s.namespaces.DeleteNamespace(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNamespaceResponse{}, nil
}

func (s *NamespaceService) GetNamespace(ctx context.Context, req *v1.GetNamespaceRequest) (*v1.GetNamespaceResponse, error) {
	item, err := s.namespaces.GetNamespace(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) ListNamespaces(ctx context.Context, req *v1.ListNamespacesRequest) (*v1.ListNamespacesResponse, error) {
	items, total, page, pageSize, err := s.namespaces.ListNamespaces(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNamespacesResponse{Items: namespaceItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NamespaceService) SelectNamespaces(ctx context.Context, req *v1.SelectNamespacesRequest) (*v1.SelectNamespacesResponse, error) {
	items, err := s.namespaces.SelectNamespaces(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNamespacesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *NamespaceService) ListSimpleNamespaces(ctx context.Context, req *v1.ListSimpleNamespacesRequest) (*v1.ListSimpleNamespacesResponse, error) {
	items, total, _, _, err := s.namespaces.ListNamespaces(ctx, req.GetPage(), req.GetPageSize(), "")
	if err != nil {
		return nil, err
	}
	return &v1.ListSimpleNamespacesResponse{Items: namespaceItems(items), Total: int32(total)}, nil
}

func (s *NamespaceService) TransferNamespaceOwner(ctx context.Context, req *v1.TransferNamespaceOwnerRequest) (*v1.TransferNamespaceOwnerResponse, error) {
	item, err := s.namespaces.TransferOwner(ctx, req.GetUid(), req.GetUserUid())
	if err != nil {
		return nil, err
	}
	return &v1.TransferNamespaceOwnerResponse{Namespace: namespaceItem(*item)}, nil
}
