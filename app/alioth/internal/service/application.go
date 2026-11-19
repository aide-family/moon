package service

import (
	"context"
	"encoding/json"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/application"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// ApplicationService 把应用与发布包版本请求转成用例入参。
type ApplicationService struct {
	v1.UnimplementedApplicationServiceServer
	catalog *application.Catalog
}

func NewApplicationService(catalog *application.Catalog) *ApplicationService {
	return &ApplicationService{catalog: catalog}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, req *v1.CreateApplicationRequest) (*v1.CreateApplicationResponse, error) {
	item, err := s.catalog.CreateApplication(ctx, bo.Application{
		Name: req.GetName(), DisplayName: req.GetDisplayName(), Category: req.GetCategory(),
		Description: req.GetDescription(), Readme: req.GetReadme(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.CreateApplicationResponse{Application: applicationItem(*item)}, nil
}

func (s *ApplicationService) UpdateApplication(ctx context.Context, req *v1.UpdateApplicationRequest) (*v1.UpdateApplicationResponse, error) {
	item, err := s.catalog.UpdateApplication(ctx, bo.Application{
		UID: req.GetUid(), DisplayName: req.GetDisplayName(), Category: req.GetCategory(),
		Description: req.GetDescription(), Readme: req.GetReadme(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateApplicationResponse{Application: applicationItem(*item)}, nil
}

func (s *ApplicationService) DeleteApplication(ctx context.Context, req *v1.DeleteApplicationRequest) (*v1.DeleteApplicationResponse, error) {
	if err := s.catalog.DeleteApplication(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteApplicationResponse{}, nil
}

func (s *ApplicationService) GetApplication(ctx context.Context, req *v1.GetApplicationRequest) (*v1.GetApplicationResponse, error) {
	item, err := s.catalog.GetApplication(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetApplicationResponse{Application: applicationItem(*item)}, nil
}

func (s *ApplicationService) ListApplications(ctx context.Context, req *v1.ListApplicationsRequest) (*v1.ListApplicationsResponse, error) {
	items, total, page, pageSize, err := s.catalog.ListApplications(ctx, req.GetPage(), req.GetPageSize(), bo.ApplicationListFilter{
		Name: req.GetName(), Category: req.GetCategory(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ApplicationItem, 0, len(items))
	for _, item := range items {
		out = append(out, applicationItem(item))
	}
	return &v1.ListApplicationsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *ApplicationService) CreateApplicationVersion(ctx context.Context, req *v1.CreateApplicationVersionRequest) (*v1.CreateApplicationVersionResponse, error) {
	version := bo.ApplicationVersion{
		Version: req.GetVersion(), Changelog: req.GetChangelog(),
		Resources: versionResourceInputs(req.GetResources()),
	}
	item, err := s.catalog.CreateVersion(ctx, req.GetApplicationUid(), version)
	if err != nil {
		return nil, err
	}
	return &v1.CreateApplicationVersionResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) UpdateApplicationVersion(ctx context.Context, req *v1.UpdateApplicationVersionRequest) (*v1.UpdateApplicationVersionResponse, error) {
	version := bo.ApplicationVersion{
		UID: req.GetUid(), Changelog: req.GetChangelog(),
		Resources: versionResourceInputs(req.GetResources()),
	}
	item, err := s.catalog.UpdateVersion(ctx, req.GetApplicationUid(), version)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateApplicationVersionResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) DeleteApplicationVersion(ctx context.Context, req *v1.DeleteApplicationVersionRequest) (*v1.DeleteApplicationVersionResponse, error) {
	if err := s.catalog.DeleteVersion(ctx, req.GetApplicationUid(), req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteApplicationVersionResponse{}, nil
}

func (s *ApplicationService) ValidateApplicationVersion(ctx context.Context, req *v1.ValidateApplicationVersionRequest) (*v1.ValidateApplicationVersionResponse, error) {
	version, errs, warnings, err := s.catalog.Validate(ctx, req.GetApplicationUid(), req.GetUid())
	if err != nil {
		return nil, err
	}
	resp := &v1.ValidateApplicationVersionResponse{Errors: errs, Warnings: warnings}
	if version != nil {
		resp.Version = applicationVersionItem(*version)
	}
	return resp, nil
}

func (s *ApplicationService) PublishApplicationVersion(ctx context.Context, req *v1.PublishApplicationVersionRequest) (*v1.PublishApplicationVersionResponse, error) {
	item, err := s.catalog.Publish(ctx, req.GetApplicationUid(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.PublishApplicationVersionResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) DeprecateApplicationVersion(ctx context.Context, req *v1.DeprecateApplicationVersionRequest) (*v1.DeprecateApplicationVersionResponse, error) {
	item, err := s.catalog.Deprecate(ctx, req.GetApplicationUid(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.DeprecateApplicationVersionResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) ListApplicationVersions(ctx context.Context, req *v1.ListApplicationVersionsRequest) (*v1.ListApplicationVersionsResponse, error) {
	items, total, page, pageSize, err := s.catalog.ListVersions(ctx, req.GetApplicationUid(), req.GetPage(), req.GetPageSize(), bo.ApplicationVersionFilter{Status: req.GetStatus()})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ApplicationVersionItem, 0, len(items))
	for _, item := range items {
		out = append(out, applicationVersionItem(item))
	}
	return &v1.ListApplicationVersionsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *ApplicationService) GetApplicationVersion(ctx context.Context, req *v1.GetApplicationVersionRequest) (*v1.GetApplicationVersionResponse, error) {
	item, err := s.catalog.GetVersion(ctx, req.GetApplicationUid(), req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetApplicationVersionResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) DuplicateApplicationVersion(ctx context.Context, req *v1.DuplicateApplicationVersionRequest) (*v1.DuplicateApplicationVersionResponse, error) {
	item, err := s.catalog.DuplicateVersion(ctx, req.GetApplicationUid(), req.GetUid(), req.GetVersion(), req.GetChangelog())
	if err != nil {
		return nil, err
	}
	return &v1.DuplicateApplicationVersionResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) CreateApplicationVersionResource(ctx context.Context, req *v1.CreateApplicationVersionResourceRequest) (*v1.CreateApplicationVersionResourceResponse, error) {
	input := bo.ApplicationResource{
		Item: req.GetResource().GetItem(), Component: req.GetResource().GetComponent(),
		Manifest: req.GetResource().GetManifest(),
	}
	item, err := s.catalog.CreateVersionResource(ctx, req.GetApplicationUid(), req.GetUid(), input, req.GetExpectedRevision())
	if err != nil {
		return nil, err
	}
	return &v1.CreateApplicationVersionResourceResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) GetApplicationVersionResource(ctx context.Context, req *v1.GetApplicationVersionResourceRequest) (*v1.GetApplicationVersionResourceResponse, error) {
	resource, err := s.catalog.GetVersionResource(ctx, req.GetApplicationUid(), req.GetUid(), req.GetResourceUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetApplicationVersionResourceResponse{Resource: applicationResourceItem(*resource)}, nil
}

func (s *ApplicationService) SaveApplicationVersionResource(ctx context.Context, req *v1.SaveApplicationVersionResourceRequest) (*v1.SaveApplicationVersionResourceResponse, error) {
	item, err := s.catalog.SaveVersionResource(ctx, req.GetApplicationUid(), req.GetUid(), req.GetResourceUid(),
		req.GetComponent(), req.GetManifest(), req.GetSecretValues(), req.GetClearSecretKeys(), req.GetExpectedRevision())
	if err != nil {
		return nil, err
	}
	return &v1.SaveApplicationVersionResourceResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) DeleteApplicationVersionResource(ctx context.Context, req *v1.DeleteApplicationVersionResourceRequest) (*v1.DeleteApplicationVersionResourceResponse, error) {
	item, err := s.catalog.DeleteVersionResource(ctx, req.GetApplicationUid(), req.GetUid(), req.GetResourceUid(), req.GetExpectedRevision())
	if err != nil {
		return nil, err
	}
	return &v1.DeleteApplicationVersionResourceResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) SaveApplicationVersionComponent(ctx context.Context, req *v1.SaveApplicationVersionComponentRequest) (*v1.SaveApplicationVersionComponentResponse, error) {
	item, err := s.catalog.SaveVersionComponent(ctx, req.GetApplicationUid(), req.GetUid(),
		req.GetComponentUid(), req.GetName(), req.GetDescription(), req.GetExpectedRevision())
	if err != nil {
		return nil, err
	}
	return &v1.SaveApplicationVersionComponentResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) DeleteApplicationVersionComponent(ctx context.Context, req *v1.DeleteApplicationVersionComponentRequest) (*v1.DeleteApplicationVersionComponentResponse, error) {
	item, err := s.catalog.DeleteVersionComponent(ctx, req.GetApplicationUid(), req.GetUid(),
		req.GetComponentUid(), req.GetExpectedRevision())
	if err != nil {
		return nil, err
	}
	return &v1.DeleteApplicationVersionComponentResponse{Version: applicationVersionItem(*item)}, nil
}

func (s *ApplicationService) ValidateApplicationVersionResource(ctx context.Context, req *v1.ValidateApplicationVersionResourceRequest) (*v1.ValidateApplicationVersionResourceResponse, error) {
	errs, warnings, err := s.catalog.ValidateVersionResource(ctx, req.GetApplicationUid(), req.GetUid(), req.GetComponent(), req.GetManifest())
	if err != nil {
		return nil, err
	}
	return &v1.ValidateApplicationVersionResourceResponse{Errors: errs, Warnings: warnings}, nil
}

// ApplicationDeploymentService 把部署组请求转成用例入参。
type ApplicationDeploymentService struct {
	v1.UnimplementedApplicationDeploymentServiceServer
	deployer *application.Deployer
}

func NewApplicationDeploymentService(deployer *application.Deployer) *ApplicationDeploymentService {
	return &ApplicationDeploymentService{deployer: deployer}
}

func (s *ApplicationDeploymentService) CreateApplicationDeployment(ctx context.Context, req *v1.CreateApplicationDeploymentRequest) (*v1.CreateApplicationDeploymentResponse, error) {
	item, err := s.deployer.Create(ctx, req.GetApplicationUid(), req.GetVersionUid(), req.GetName(), req.GetTargetNamespace(), req.GetClusterUids(), deployValues(req.GetValues()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateApplicationDeploymentResponse{Deployment: applicationDeploymentItem(*item)}, nil
}

func (s *ApplicationDeploymentService) ListApplicationDeployments(ctx context.Context, req *v1.ListApplicationDeploymentsRequest) (*v1.ListApplicationDeploymentsResponse, error) {
	items, total, page, pageSize, err := s.deployer.List(ctx, req.GetPage(), req.GetPageSize(), bo.ApplicationReleaseFilter{
		Name: req.GetName(), Status: req.GetStatus(), ApplicationUID: req.GetApplicationUid(), ClusterUID: req.GetClusterUid(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ApplicationDeploymentItem, 0, len(items))
	for _, item := range items {
		out = append(out, applicationDeploymentItem(item))
	}
	return &v1.ListApplicationDeploymentsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *ApplicationDeploymentService) GetApplicationDeployment(ctx context.Context, req *v1.GetApplicationDeploymentRequest) (*v1.GetApplicationDeploymentResponse, error) {
	item, err := s.deployer.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetApplicationDeploymentResponse{Deployment: applicationDeploymentItem(*item)}, nil
}

func (s *ApplicationDeploymentService) UpgradeApplicationDeployment(ctx context.Context, req *v1.UpgradeApplicationDeploymentRequest) (*v1.UpgradeApplicationDeploymentResponse, error) {
	item, err := s.deployer.Upgrade(ctx, req.GetUid(), req.GetVersionUid(), req.GetClusterUids(), deployValues(req.GetValues()))
	if err != nil {
		return nil, err
	}
	return &v1.UpgradeApplicationDeploymentResponse{Deployment: applicationDeploymentItem(*item)}, nil
}

func (s *ApplicationDeploymentService) RollbackApplicationDeployment(ctx context.Context, req *v1.RollbackApplicationDeploymentRequest) (*v1.RollbackApplicationDeploymentResponse, error) {
	item, err := s.deployer.Rollback(ctx, req.GetUid(), req.GetTargetVersionUid(), req.GetClusterUids())
	if err != nil {
		return nil, err
	}
	return &v1.RollbackApplicationDeploymentResponse{Deployment: applicationDeploymentItem(*item)}, nil
}

func (s *ApplicationDeploymentService) RedeployApplicationDeployment(ctx context.Context, req *v1.RedeployApplicationDeploymentRequest) (*v1.RedeployApplicationDeploymentResponse, error) {
	item, err := s.deployer.Redeploy(ctx, req.GetUid(), req.GetClusterUids())
	if err != nil {
		return nil, err
	}
	return &v1.RedeployApplicationDeploymentResponse{Deployment: applicationDeploymentItem(*item)}, nil
}

func (s *ApplicationDeploymentService) UninstallApplicationDeployment(ctx context.Context, req *v1.UninstallApplicationDeploymentRequest) (*v1.UninstallApplicationDeploymentResponse, error) {
	item, err := s.deployer.Uninstall(ctx, req.GetUid(), req.GetClusterUids(), req.GetKeepPersistentVolumeClaims(), req.GetDeleteClusterScopedResources())
	if err != nil {
		return nil, err
	}
	return &v1.UninstallApplicationDeploymentResponse{Deployment: applicationDeploymentItem(*item)}, nil
}

func (s *ApplicationDeploymentService) ListApplicationReleaseOperations(ctx context.Context, req *v1.ListApplicationReleaseOperationsRequest) (*v1.ListApplicationReleaseOperationsResponse, error) {
	items, total, page, pageSize, err := s.deployer.ListOperations(ctx, req.GetUid(), req.GetReleaseUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ApplicationOperationItem, 0, len(items))
	for _, item := range items {
		out = append(out, applicationOperationItem(item))
	}
	return &v1.ListApplicationReleaseOperationsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

// ApplicationReleaseService 读取单个集群实例与操作记录。
type ApplicationReleaseService struct {
	v1.UnimplementedApplicationReleaseServiceServer
	deployer *application.Deployer
}

func NewApplicationReleaseService(deployer *application.Deployer) *ApplicationReleaseService {
	return &ApplicationReleaseService{deployer: deployer}
}

func (s *ApplicationReleaseService) GetApplicationRelease(ctx context.Context, req *v1.GetApplicationReleaseRequest) (*v1.GetApplicationReleaseResponse, error) {
	item, err := s.deployer.GetRelease(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetApplicationReleaseResponse{Release: applicationReleaseItem(*item)}, nil
}

func (s *ApplicationReleaseService) ListApplicationReleases(ctx context.Context, req *v1.ListApplicationReleasesRequest) (*v1.ListApplicationReleasesResponse, error) {
	items, total, page, pageSize, err := s.deployer.ListReleases(ctx, req.GetPage(), req.GetPageSize(), bo.ApplicationReleaseFilter{
		ApplicationUID: req.GetApplicationUid(), ClusterUID: req.GetClusterUid(), Status: req.GetStatus(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ApplicationReleaseItem, 0, len(items))
	for _, item := range items {
		out = append(out, applicationReleaseItem(item))
	}
	return &v1.ListApplicationReleasesResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *ApplicationReleaseService) ListApplicationReleaseOperations(ctx context.Context, req *v1.ListReleaseOperationsRequest) (*v1.ListReleaseOperationsResponse, error) {
	items, total, page, pageSize, err := s.deployer.ListOperations(ctx, "", req.GetUid(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ApplicationOperationItem, 0, len(items))
	for _, item := range items {
		out = append(out, applicationOperationItem(item))
	}
	return &v1.ListReleaseOperationsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *ApplicationReleaseService) GetApplicationOperation(ctx context.Context, req *v1.GetApplicationOperationRequest) (*v1.GetApplicationOperationResponse, error) {
	item, err := s.deployer.GetOperation(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetApplicationOperationResponse{Operation: applicationOperationItem(*item)}, nil
}

func applicationItem(item bo.Application) *v1.ApplicationItem {
	return &v1.ApplicationItem{
		Uid: item.UID, Name: item.Name, DisplayName: item.DisplayName, Category: item.Category,
		Description: item.Description, Readme: item.Readme, VersionCount: item.VersionCount,
		LatestVersion: item.LatestVersion, CreatedAt: item.CreatedAt.Unix(),
	}
}

func applicationVersionItem(item bo.ApplicationVersion) *v1.ApplicationVersionItem {
	out := &v1.ApplicationVersionItem{
		Uid: item.UID, ApplicationUid: item.ApplicationUID, Version: item.Version, Status: item.Status,
		Changelog: item.Changelog, CreatedAt: item.CreatedAt.Unix(), Revision: item.Revision,
		ComponentCount: item.ComponentCount, ResourceCount: item.ResourceCount,
		AllowedActions: item.AllowedActions, CopiedFromVersionUid: item.CopiedFromVersionUID,
	}
	for _, resource := range item.Resources {
		out.Resources = append(out.Resources, applicationResourceItem(resource))
	}
	for _, component := range item.Components {
		mapped := &v1.ApplicationVersionComponent{
			Uid: component.UID, Name: component.Name, DisplayName: component.DisplayName,
			SortOrder: component.SortOrder, ResourceCount: component.ResourceCount,
			Description: component.Description,
		}
		for _, resource := range component.Resources {
			mapped.Resources = append(mapped.Resources, applicationResourceItem(resource))
		}
		out.Components = append(out.Components, mapped)
	}
	for _, image := range item.Images {
		out.Images = append(out.Images, &v1.ApplicationVersionImageItem{
			ResourceItem: image.ResourceItem, Container: image.Container, Image: image.Image,
			RegistryUid: image.RegistryUID, RegistryName: image.RegistryName,
			Repository: image.Repository, Tag: image.Tag, Digest: image.Digest,
			Platforms: image.Platforms, External: image.External, Component: image.Component,
			ResourceUid: image.ResourceUID, ContainerPath: image.ContainerPath, Project: image.Project,
			Source: image.Source, ResolvedImage: image.ResolvedImage,
		})
	}
	if item.PublishedAt != nil {
		out.PublishedAt = item.PublishedAt.Unix()
	}
	if item.StatusChangedAt != nil {
		out.StatusChangedAt = item.StatusChangedAt.Unix()
	}
	if item.DeprecatedAt != nil {
		out.DeprecatedAt = item.DeprecatedAt.Unix()
	}
	if raw, err := json.Marshal(item.Validation); err == nil {
		out.Validation = string(raw)
	}
	return out
}

func applicationResourceItem(resource bo.ApplicationResource) *v1.ApplicationResource {
	return &v1.ApplicationResource{
		Item: resource.Item, Manifest: resource.Manifest, ApiVersion: resource.APIVersion,
		Kind: resource.Kind, Name: resource.Name, Namespace: resource.Namespace, Scope: resource.Scope,
		ClusterScoped: resource.ClusterScoped, Images: resource.Images, SecretKeys: resource.SecretKeys,
		Uid: resource.UID, Component: resource.Component, ContentHash: resource.ContentHash,
	}
}

func versionResourceInputs(items []*v1.ApplicationResourceInput) []bo.ApplicationResource {
	out := make([]bo.ApplicationResource, 0, len(items))
	for _, item := range items {
		out = append(out, bo.ApplicationResource{
			Item: item.GetItem(), Component: item.GetComponent(), Manifest: item.GetManifest(),
		})
	}
	return out
}

func applicationDeploymentItem(item bo.ApplicationDeployment) *v1.ApplicationDeploymentItem {
	out := &v1.ApplicationDeploymentItem{
		Uid: item.UID, Name: item.Name, ApplicationUid: item.ApplicationUID,
		ApplicationName: item.ApplicationName, VersionUid: item.VersionUID, Version: item.Version,
		TargetNamespace: item.TargetNamespace, Status: item.Status, Message: item.Message,
		CreatedAt: item.CreatedAt.Unix(),
	}
	if item.DeployedAt != nil {
		out.DeployedAt = item.DeployedAt.Unix()
	}
	for _, cluster := range item.Clusters {
		out.Clusters = append(out.Clusters, &v1.DeploymentCluster{
			Uid: cluster.UID, Name: cluster.Name, Status: cluster.Status,
			ReleaseUid: cluster.ReleaseUID, Version: cluster.Version, Message: cluster.Message,
			CurrentVersion: cluster.CurrentVersion, CurrentVersionUid: cluster.CurrentVersionUID,
			InSync: cluster.InSync,
		})
	}
	for _, release := range item.Releases {
		out.Releases = append(out.Releases, applicationReleaseItem(release))
	}
	return out
}

func applicationReleaseItem(item bo.ApplicationRelease) *v1.ApplicationReleaseItem {
	out := &v1.ApplicationReleaseItem{
		Uid: item.UID, DeploymentUid: item.DeploymentUID, ClusterUid: item.ClusterUID,
		ClusterName: item.ClusterName, TargetNamespace: item.TargetNamespace, Name: item.Name,
		Status: item.Status, VersionUid: item.VersionUID, Version: item.Version, Message: item.Message,
		LastOperationUid: item.LastOperationUID, ApplicationUid: item.ApplicationUID,
		ApplicationName: item.ApplicationName,
	}
	if item.DeployedAt != nil {
		out.DeployedAt = item.DeployedAt.Unix()
	}
	return out
}

func applicationOperationItem(item bo.ApplicationOperation) *v1.ApplicationOperationItem {
	out := &v1.ApplicationOperationItem{
		Uid: item.UID, DeploymentUid: item.DeploymentUID, ReleaseUid: item.ReleaseUID,
		ClusterUid: item.ClusterUID, ClusterName: item.ClusterName, Action: item.Action, Status: item.Status,
		FromVersionUid: item.FromVersionUID, ToVersionUid: item.ToVersionUID, Message: item.Message,
		OperatorName: item.OperatorName, CreatedAt: item.CreatedAt.Unix(),
	}
	if item.StartedAt != nil {
		out.StartedAt = item.StartedAt.Unix()
	}
	if item.FinishedAt != nil {
		out.FinishedAt = item.FinishedAt.Unix()
	}
	for _, resource := range item.Resources {
		out.Resources = append(out.Resources, &v1.ResourceRef{
			ApiVersion: resource.APIVersion, Kind: resource.Kind, Namespace: resource.Namespace, Name: resource.Name,
		})
	}
	return out
}

func deployValues(values *v1.ApplicationDeployValues) bo.ApplicationDeployValues {
	if values == nil {
		return bo.ApplicationDeployValues{}
	}
	return bo.ApplicationDeployValues{
		Replicas: values.GetReplicas(), ImageTags: values.GetImageTags(),
		IngressHosts: values.GetIngressHosts(), SecretValues: values.GetSecretValues(),
	}
}
