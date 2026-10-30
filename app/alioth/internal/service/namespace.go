package service

import (
	"context"
	"strings"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/feature"
	"github.com/aide-family/moon/app/alioth/internal/biz/namespace"
	"github.com/aide-family/moon/app/alioth/internal/biz/order"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

type NamespaceService struct {
	v1.UnimplementedNamespaceServiceServer
	tenancy      *namespace.Tenancy
	provisioning *feature.Provisioning
	purchasing   *order.Purchasing
}

func NewNamespaceService(tenancy *namespace.Tenancy, provisioning *feature.Provisioning, purchasing *order.Purchasing) *NamespaceService {
	return &NamespaceService{tenancy: tenancy, provisioning: provisioning, purchasing: purchasing}
}

func (s *NamespaceService) CreateNamespace(ctx context.Context, req *v1.CreateNamespaceRequest) (*v1.CreateNamespaceResponse, error) {
	item, err := s.tenancy.CreateNamespace(ctx, strings.TrimSpace(req.GetName()), strings.TrimSpace(req.GetRemark()), req.GetPackageUid())
	if err != nil {
		return nil, err
	}
	return &v1.CreateNamespaceResponse{Namespace: namespaceItem(*item)}, nil
}

// ListPublicPackages 返回创建租户时可选的功能套餐。登录即可调用，不需要系统权限。
func (s *NamespaceService) ListPublicPackages(ctx context.Context, _ *v1.ListPublicPackagesRequest) (*v1.ListPublicPackagesResponse, error) {
	items, err := s.purchasing.ListPublicPackages(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListPublicPackagesResponse{Items: featurePackageItems(items)}, nil
}

func (s *NamespaceService) CreateNamespaceOrder(ctx context.Context, req *v1.CreateNamespaceOrderRequest) (*v1.CreateNamespaceOrderResponse, error) {
	item, err := s.purchasing.CreateOrder(ctx, req.GetName(), req.GetRemark(), req.GetPackageUid())
	if err != nil {
		return nil, err
	}
	return &v1.CreateNamespaceOrderResponse{Order: namespaceOrderItem(*item)}, nil
}

func (s *NamespaceService) PayNamespaceOrder(ctx context.Context, req *v1.PayNamespaceOrderRequest) (*v1.PayNamespaceOrderResponse, error) {
	item, ns, err := s.purchasing.PayOrder(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.PayNamespaceOrderResponse{Order: namespaceOrderItem(*item), Namespace: namespaceItem(*ns)}, nil
}

func (s *NamespaceService) CancelNamespaceOrder(ctx context.Context, req *v1.CancelNamespaceOrderRequest) (*v1.CancelNamespaceOrderResponse, error) {
	item, err := s.purchasing.CancelOrder(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.CancelNamespaceOrderResponse{Order: namespaceOrderItem(*item)}, nil
}

func (s *NamespaceService) ListNamespaceOrders(ctx context.Context, req *v1.ListNamespaceOrdersRequest) (*v1.ListNamespaceOrdersResponse, error) {
	items, total, page, pageSize, err := s.purchasing.ListOrders(ctx, req.GetPage(), req.GetPageSize(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.ListNamespaceOrdersResponse{Items: namespaceOrderItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *NamespaceService) ListSystemNamespaceOrders(ctx context.Context, req *v1.ListSystemNamespaceOrdersRequest) (*v1.ListSystemNamespaceOrdersResponse, error) {
	items, total, page, pageSize, err := s.purchasing.ListNamespaceOrders(ctx, req.GetUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	return &v1.ListSystemNamespaceOrdersResponse{Items: namespaceOrderItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
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

// UpdateSystemNamespaceStatus 由平台停用或启用租户。停用原因必填，成功后通知负责人。
func (s *NamespaceService) UpdateSystemNamespaceStatus(ctx context.Context, req *v1.UpdateSystemNamespaceStatusRequest) (*v1.UpdateSystemNamespaceStatusResponse, error) {
	item, err := s.tenancy.SetSystemNamespaceStatus(ctx, req.GetUid(), req.GetStatus(), req.GetReason())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateSystemNamespaceStatusResponse{Namespace: namespaceItem(*item)}, nil
}

// ArchiveSystemNamespace 归档租户：软删除租户及其全部业务数据，不可恢复。
func (s *NamespaceService) ArchiveSystemNamespace(ctx context.Context, req *v1.ArchiveSystemNamespaceRequest) (*v1.ArchiveSystemNamespaceResponse, error) {
	if err := s.tenancy.ArchiveNamespace(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.ArchiveSystemNamespaceResponse{}, nil
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
	item, err := s.provisioning.CreatePackage(ctx, bo.FeaturePackage{
		Name:         strings.TrimSpace(req.GetName()),
		Remark:       strings.TrimSpace(req.GetRemark()),
		DurationDays: req.GetDurationDays(),
		Visibility:   req.GetVisibility(),
		MaxMembers:   req.GetMaxMembers(),
		PriceCents:   req.GetPriceCents(),
		Features:     req.GetFeatures(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateFeaturePackageResponse{Package: featurePackageItem(*item)}, nil
}

func (s *NamespaceService) UpdateFeaturePackage(ctx context.Context, req *v1.UpdateFeaturePackageRequest) (*v1.UpdateFeaturePackageResponse, error) {
	item, err := s.provisioning.UpdatePackage(ctx, bo.FeaturePackage{
		UID:          req.GetUid(),
		Name:         strings.TrimSpace(req.GetName()),
		Remark:       strings.TrimSpace(req.GetRemark()),
		DurationDays: req.GetDurationDays(),
		Visibility:   req.GetVisibility(),
		MaxMembers:   req.GetMaxMembers(),
		PriceCents:   req.GetPriceCents(),
		Features:     req.GetFeatures(),
	})
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
	packageUID, packageName, features, packageFeatures, extraFeatures, expiresAt, expired, durationDays := namespaceFeatureFields(*item)
	return &v1.GetNamespaceFeaturesResponse{
		PackageUid: packageUID, PackageName: packageName, Features: features,
		PackageFeatures: packageFeatures, ExtraFeatures: extraFeatures,
		ExpiresAt: expiresAt, Expired: expired, DurationDays: durationDays,
	}, nil
}

func (s *NamespaceService) UpdateNamespaceFeatures(ctx context.Context, req *v1.UpdateNamespaceFeaturesRequest) (*v1.UpdateNamespaceFeaturesResponse, error) {
	item, err := s.provisioning.UpdateNamespaceFeatures(ctx, req.GetUid(), req.GetPackageUid(), req.GetFeatures(), req.GetExpiresAt())
	if err != nil {
		return nil, err
	}
	packageUID, packageName, features, packageFeatures, extraFeatures, expiresAt, expired, durationDays := namespaceFeatureFields(*item)
	return &v1.UpdateNamespaceFeaturesResponse{
		PackageUid: packageUID, PackageName: packageName, Features: features,
		PackageFeatures: packageFeatures, ExtraFeatures: extraFeatures,
		ExpiresAt: expiresAt, Expired: expired, DurationDays: durationDays,
	}, nil
}
