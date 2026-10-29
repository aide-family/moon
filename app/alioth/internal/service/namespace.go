package service

import (
	"context"
	"strings"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/feature"
	"github.com/aide-family/moon/app/alioth/internal/biz/namespace"
)

type NamespaceService struct {
	v1.UnimplementedNamespaceServiceServer
	tenancy      *namespace.Tenancy
	provisioning *feature.Provisioning
}

func NewNamespaceService(tenancy *namespace.Tenancy, provisioning *feature.Provisioning) *NamespaceService {
	return &NamespaceService{tenancy: tenancy, provisioning: provisioning}
}

func (s *NamespaceService) CreateNamespace(ctx context.Context, req *v1.CreateNamespaceRequest) (*v1.CreateNamespaceResponse, error) {
	item, err := s.tenancy.CreateNamespace(ctx, req.GetName(), req.GetRemark())
	if err != nil {
		return nil, err
	}
	return &v1.CreateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) UpdateNamespace(ctx context.Context, req *v1.UpdateNamespaceRequest) (*v1.UpdateNamespaceResponse, error) {
	item, err := s.tenancy.UpdateNamespace(ctx, req.GetUid(), req.GetName(), req.GetRemark(), req.GetLogo())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) UpdateNamespaceStatus(ctx context.Context, req *v1.UpdateNamespaceStatusRequest) (*v1.UpdateNamespaceStatusResponse, error) {
	item, err := s.tenancy.UpdateNamespaceStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateNamespaceStatusResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) DeleteNamespace(ctx context.Context, req *v1.DeleteNamespaceRequest) (*v1.DeleteNamespaceResponse, error) {
	if err := s.tenancy.DeleteNamespace(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteNamespaceResponse{}, nil
}

func (s *NamespaceService) GetNamespace(ctx context.Context, req *v1.GetNamespaceRequest) (*v1.GetNamespaceResponse, error) {
	item, err := s.tenancy.GetNamespace(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) ListNamespaces(ctx context.Context, req *v1.ListNamespacesRequest) (*v1.ListNamespacesResponse, error) {
	items, total, page, pageSize, err := s.tenancy.ListNamespaces(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListNamespacesResponse{Items: namespaceItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NamespaceService) SelectNamespaces(ctx context.Context, req *v1.SelectNamespacesRequest) (*v1.SelectNamespacesResponse, error) {
	items, err := s.tenancy.SelectNamespaces(ctx, req.GetLimit(), req.GetLastUid())
	if err != nil {
		return nil, err
	}
	return &v1.SelectNamespacesResponse{Items: selectItems(items), LastUid: lastUID(items)}, nil
}

func (s *NamespaceService) ListSimpleNamespaces(ctx context.Context, req *v1.ListSimpleNamespacesRequest) (*v1.ListSimpleNamespacesResponse, error) {
	items, total, _, _, err := s.tenancy.ListNamespaces(ctx, req.GetPage(), req.GetPageSize(), "")
	if err != nil {
		return nil, err
	}
	return &v1.ListSimpleNamespacesResponse{Items: namespaceItems(items), Total: int32(total)}, nil
}

func (s *NamespaceService) ListSystemNamespaces(ctx context.Context, req *v1.ListSystemNamespacesRequest) (*v1.ListSystemNamespacesResponse, error) {
	items, total, page, pageSize, err := s.tenancy.ListAllNamespaces(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.ListSystemNamespacesResponse{Items: namespaceItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NamespaceService) TransferNamespaceOwner(ctx context.Context, req *v1.TransferNamespaceOwnerRequest) (*v1.TransferNamespaceOwnerResponse, error) {
	item, err := s.tenancy.TransferOwner(ctx, req.GetUid(), req.GetUserUid())
	if err != nil {
		return nil, err
	}
	return &v1.TransferNamespaceOwnerResponse{Namespace: namespaceItem(*item)}, nil
}

func (s *NamespaceService) ListFeatures(ctx context.Context, _ *v1.ListFeaturesRequest) (*v1.ListFeaturesResponse, error) {
	items, err := s.provisioning.ListFeatures(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListFeaturesResponse{Items: featureItems(items)}, nil
}

func (s *NamespaceService) ListFeaturePackages(ctx context.Context, req *v1.ListFeaturePackagesRequest) (*v1.ListFeaturePackagesResponse, error) {
	items, total, page, pageSize, err := s.provisioning.ListPackages(ctx, req.GetPage(), req.GetPageSize(), strings.TrimSpace(req.GetName()))
	if err != nil {
		return nil, err
	}
	return &v1.ListFeaturePackagesResponse{Items: featurePackageItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NamespaceService) GetFeaturePackage(ctx context.Context, req *v1.GetFeaturePackageRequest) (*v1.GetFeaturePackageResponse, error) {
	item, err := s.provisioning.GetPackage(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetFeaturePackageResponse{Package: featurePackageItem(*item)}, nil
}

func (s *NamespaceService) CreateFeaturePackage(ctx context.Context, req *v1.CreateFeaturePackageRequest) (*v1.CreateFeaturePackageResponse, error) {
	item, err := s.provisioning.CreatePackage(ctx, strings.TrimSpace(req.GetName()), strings.TrimSpace(req.GetRemark()), req.GetFeatures())
	if err != nil {
		return nil, err
	}
	return &v1.CreateFeaturePackageResponse{Package: featurePackageItem(*item)}, nil
}

func (s *NamespaceService) UpdateFeaturePackage(ctx context.Context, req *v1.UpdateFeaturePackageRequest) (*v1.UpdateFeaturePackageResponse, error) {
	item, err := s.provisioning.UpdatePackage(ctx, req.GetUid(), strings.TrimSpace(req.GetName()), strings.TrimSpace(req.GetRemark()), req.GetFeatures())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateFeaturePackageResponse{Package: featurePackageItem(*item)}, nil
}

func (s *NamespaceService) DeleteFeaturePackage(ctx context.Context, req *v1.DeleteFeaturePackageRequest) (*v1.DeleteFeaturePackageResponse, error) {
	if err := s.provisioning.DeletePackage(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteFeaturePackageResponse{}, nil
}

func (s *NamespaceService) GetNamespaceFeatures(ctx context.Context, req *v1.GetNamespaceFeaturesRequest) (*v1.GetNamespaceFeaturesResponse, error) {
	item, err := s.provisioning.GetNamespaceFeatures(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	packageUID, packageName, features, packageFeatures := namespaceFeatureFields(*item)
	return &v1.GetNamespaceFeaturesResponse{PackageUid: packageUID, PackageName: packageName, Features: features, PackageFeatures: packageFeatures}, nil
}

func (s *NamespaceService) UpdateNamespaceFeatures(ctx context.Context, req *v1.UpdateNamespaceFeaturesRequest) (*v1.UpdateNamespaceFeaturesResponse, error) {
	item, err := s.provisioning.UpdateNamespaceFeatures(ctx, req.GetUid(), req.GetPackageUid(), req.GetFeatures())
	if err != nil {
		return nil, err
	}
	packageUID, packageName, features, packageFeatures := namespaceFeatureFields(*item)
	return &v1.UpdateNamespaceFeaturesResponse{PackageUid: packageUID, PackageName: packageName, Features: features, PackageFeatures: packageFeatures}, nil
}
