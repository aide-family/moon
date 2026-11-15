package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/registry"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// RegistryService 把镜像仓库连接与远端浏览请求转成用例入参。
type RegistryService struct {
	v1.UnimplementedRegistryServiceServer
	registry *registry.Registry
	explorer *registry.Explorer
}

func NewRegistryService(registry *registry.Registry, explorer *registry.Explorer) *RegistryService {
	return &RegistryService{registry: registry, explorer: explorer}
}

func (s *RegistryService) CreateRegistry(ctx context.Context, req *v1.CreateRegistryRequest) (*v1.CreateRegistryResponse, error) {
	in := bo.Registry{
		Name: req.GetName(), Driver: req.GetDriver(), Endpoint: req.GetEndpoint(),
		Username: req.GetUsername(), AllowInsecureTLS: req.GetAllowInsecureTls(), Remark: req.GetRemark(),
	}
	item, err := s.registry.Create(ctx, in, req.GetPassword())
	if err != nil {
		return nil, err
	}
	return &v1.CreateRegistryResponse{Registry: registryItem(*item)}, nil
}

func (s *RegistryService) UpdateRegistry(ctx context.Context, req *v1.UpdateRegistryRequest) (*v1.UpdateRegistryResponse, error) {
	in := bo.Registry{
		UID: req.GetUid(), Name: req.GetName(), Driver: req.GetDriver(), Endpoint: req.GetEndpoint(),
		Username: req.GetUsername(), AllowInsecureTLS: req.GetAllowInsecureTls(), Remark: req.GetRemark(),
	}
	item, err := s.registry.Update(ctx, in, req.GetPassword(), req.GetClearPassword())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRegistryResponse{Registry: registryItem(*item)}, nil
}

func (s *RegistryService) DeleteRegistry(ctx context.Context, req *v1.DeleteRegistryRequest) (*v1.DeleteRegistryResponse, error) {
	if err := s.registry.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteRegistryResponse{}, nil
}

func (s *RegistryService) GetRegistry(ctx context.Context, req *v1.GetRegistryRequest) (*v1.GetRegistryResponse, error) {
	item, err := s.registry.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetRegistryResponse{Registry: registryItem(*item)}, nil
}

func (s *RegistryService) ListRegistries(ctx context.Context, req *v1.ListRegistriesRequest) (*v1.ListRegistriesResponse, error) {
	items, total, page, pageSize, err := s.registry.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName(), req.GetDriver())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.RegistryItem, 0, len(items))
	for _, item := range items {
		out = append(out, registryItem(item))
	}
	return &v1.ListRegistriesResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *RegistryService) TestRegistry(ctx context.Context, req *v1.TestRegistryRequest) (*v1.TestRegistryResponse, error) {
	probe, item, err := s.registry.Test(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.TestRegistryResponse{Probe: registryProbe(*probe), Registry: registryItem(*item)}, nil
}

func (s *RegistryService) ProbeRegistry(ctx context.Context, req *v1.ProbeRegistryRequest) (*v1.ProbeRegistryResponse, error) {
	probe, err := s.registry.Probe(ctx, req.GetDriver(), req.GetEndpoint(), req.GetUsername(), req.GetPassword(), req.GetAllowInsecureTls())
	if err != nil {
		return nil, err
	}
	return &v1.ProbeRegistryResponse{Probe: registryProbe(*probe)}, nil
}

func (s *RegistryService) ListRegistryProjects(ctx context.Context, req *v1.ListRegistryProjectsRequest) (*v1.ListRegistryProjectsResponse, error) {
	items, err := s.explorer.ListProjects(ctx, req.GetUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.RegistryProjectItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.RegistryProjectItem{Name: item.Name, RepositoryCount: item.RepositoryCount})
	}
	return &v1.ListRegistryProjectsResponse{Items: out}, nil
}

func (s *RegistryService) ListRegistryRepositories(ctx context.Context, req *v1.ListRegistryRepositoriesRequest) (*v1.ListRegistryRepositoriesResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListRepositories(ctx, req.GetUid(), req.GetProject(), req.GetName(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.RegistryRepositoryItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.RegistryRepositoryItem{Name: item.Name, Project: item.Project, TagCount: item.TagCount, UpdatedAt: item.UpdatedAt})
	}
	return &v1.ListRegistryRepositoriesResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *RegistryService) ListRegistryTags(ctx context.Context, req *v1.ListRegistryTagsRequest) (*v1.ListRegistryTagsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListTags(ctx, req.GetUid(), req.GetProject(), req.GetRepository(), req.GetArchitecture(), req.GetName(), req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.RegistryTagItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.RegistryTagItem{
			Tag: item.Tag, Digest: item.Digest, Size: item.Size, Platforms: item.Platforms,
			PushedAt: item.PushedAt, ReferenceCount: item.ReferenceCount,
		})
	}
	return &v1.ListRegistryTagsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *RegistryService) GetRegistryTag(ctx context.Context, req *v1.GetRegistryTagRequest) (*v1.GetRegistryTagResponse, error) {
	detail, err := s.explorer.GetTag(ctx, req.GetUid(), req.GetProject(), req.GetRepository(), req.GetTag())
	if err != nil {
		return nil, err
	}
	return &v1.GetRegistryTagResponse{Tag: registryTagDetail(*detail)}, nil
}

func (s *RegistryService) CopyRegistryTag(ctx context.Context, req *v1.CopyRegistryTagRequest) (*v1.CopyRegistryTagResponse, error) {
	item, err := s.explorer.CopyTag(ctx, req.GetUid(), req.GetProject(), req.GetRepository(), req.GetTag(), req.GetDigest(), req.GetTargetRepository(), req.GetTargetTag())
	if err != nil {
		return nil, err
	}
	return &v1.CopyRegistryTagResponse{Tag: registryTagItem(*item)}, nil
}

func (s *RegistryService) DeleteRegistryTag(ctx context.Context, req *v1.DeleteRegistryTagRequest) (*v1.DeleteRegistryTagResponse, error) {
	if err := s.explorer.DeleteTag(ctx, req.GetUid(), req.GetProject(), req.GetRepository(), req.GetTag(), req.GetDigest()); err != nil {
		return nil, err
	}
	return &v1.DeleteRegistryTagResponse{}, nil
}

// ImageSyncService 把同步任务请求转成用例入参。
type ImageSyncService struct {
	v1.UnimplementedImageSyncServiceServer
	sync *registry.Synchronizer
}

func NewImageSyncService(sync *registry.Synchronizer) *ImageSyncService {
	return &ImageSyncService{sync: sync}
}

func (s *ImageSyncService) ListImageSyncs(ctx context.Context, req *v1.ListImageSyncsRequest) (*v1.ListImageSyncsResponse, error) {
	items, total, page, pageSize, err := s.sync.List(ctx, req.GetPage(), req.GetPageSize(), bo.ImageSyncFilter{
		Status: req.GetStatus(), SourceKind: req.GetSourceKind(), RegistryUID: req.GetRegistryUid(),
	})
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ImageSyncItem, 0, len(items))
	for _, item := range items {
		out = append(out, imageSyncItem(item))
	}
	return &v1.ListImageSyncsResponse{Items: out, Total: int32(total), Page: int32(page), PageSize: int32(pageSize)}, nil
}

func (s *ImageSyncService) PreviewImageSync(ctx context.Context, req *v1.PreviewImageSyncRequest) (*v1.PreviewImageSyncResponse, error) {
	in := imageSyncInput(req.GetSource())
	in.TargetRegistryUID = req.GetTargetRegistryUid()
	in.TargetRepository = req.GetTargetRepository()
	in.TargetTag = req.GetTargetTag()
	preview, err := s.sync.Preview(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.PreviewImageSyncResponse{
		SourceImage: preview.SourceImage, SourceDigest: preview.SourceDigest,
		SourcePlatforms: preview.SourcePlatforms, TargetExists: preview.TargetExists, TargetDigest: preview.TargetDigest,
	}, nil
}

func (s *ImageSyncService) CreateImageSyncs(ctx context.Context, req *v1.CreateImageSyncsRequest) (*v1.CreateImageSyncsResponse, error) {
	inputs := make([]bo.ImageSyncInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		inputs = append(inputs, imageSyncInput(item))
	}
	created, failed, err := s.sync.Create(ctx, inputs)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.ImageSyncItem, 0, len(created))
	for _, item := range created {
		out = append(out, imageSyncItem(item))
	}
	failures := make([]*v1.ImageSyncFailure, 0, len(failed))
	for _, item := range failed {
		failures = append(failures, &v1.ImageSyncFailure{Index: item.Index, Message: item.Message})
	}
	return &v1.CreateImageSyncsResponse{Created: out, Failed: failures}, nil
}

func (s *ImageSyncService) GetImageSync(ctx context.Context, req *v1.GetImageSyncRequest) (*v1.GetImageSyncResponse, error) {
	item, err := s.sync.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetImageSyncResponse{Item: imageSyncItem(*item)}, nil
}

func (s *ImageSyncService) RetryImageSync(ctx context.Context, req *v1.RetryImageSyncRequest) (*v1.RetryImageSyncResponse, error) {
	item, err := s.sync.Retry(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.RetryImageSyncResponse{Item: imageSyncItem(*item)}, nil
}

func (s *ImageSyncService) CancelImageSync(ctx context.Context, req *v1.CancelImageSyncRequest) (*v1.CancelImageSyncResponse, error) {
	item, err := s.sync.Cancel(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.CancelImageSyncResponse{Item: imageSyncItem(*item)}, nil
}

func registryItem(item bo.Registry) *v1.RegistryItem {
	return &v1.RegistryItem{
		Uid: item.UID, Name: item.Name, Driver: item.Driver, Endpoint: item.Endpoint,
		Username: item.Username, AllowInsecureTls: item.AllowInsecureTLS, Remark: item.Remark,
		LastCheckedAt: item.LastCheckedAt, LastError: item.LastError, CreatedAt: item.CreatedAt.Unix(),
	}
}

func registryProbe(item bo.RegistryProbe) *v1.RegistryProbe {
	return &v1.RegistryProbe{Endpoint: item.Endpoint, Driver: item.Driver, RepositoryCount: item.RepositoryCount}
}

func registryTagItem(item bo.RegistryTag) *v1.RegistryTagItem {
	return &v1.RegistryTagItem{
		Tag: item.Tag, Digest: item.Digest, Size: item.Size, Platforms: item.Platforms,
		PushedAt: item.PushedAt, ReferenceCount: item.ReferenceCount,
	}
}

func registryTagDetail(item bo.RegistryTagDetail) *v1.RegistryTagDetail {
	out := &v1.RegistryTagDetail{
		Tag: item.Tag, Digest: item.Digest, Size: item.Size, MediaType: item.MediaType,
		ConfigDigest: item.ConfigDigest, Labels: item.Labels, CreatedAt: item.CreatedAt,
	}
	for _, platform := range item.Platforms {
		out.Platforms = append(out.Platforms, &v1.RegistryPlatform{
			Os: platform.OS, Architecture: platform.Architecture, Variant: platform.Variant,
			Digest: platform.Digest, Size: platform.Size,
		})
	}
	for _, reference := range item.References {
		out.References = append(out.References, &v1.RegistryTagReference{
			ApplicationUid: reference.ApplicationUID, ApplicationName: reference.ApplicationName,
			VersionUid: reference.VersionUID, Version: reference.Version, Status: reference.Status,
		})
	}
	return out
}

func imageSyncItem(item bo.ImageSync) *v1.ImageSyncItem {
	out := &v1.ImageSyncItem{
		Uid: item.UID, SourceKind: item.SourceKind, SourceRegistryUid: item.SourceRegistryUID,
		SourceRegistryName: item.SourceRegistryName, SourceImage: item.SourceImage,
		SourceDigest: item.SourceDigest, SourcePlatforms: item.SourcePlatforms,
		TargetRegistryUid: item.TargetRegistryUID, TargetRegistryName: item.TargetRegistryName,
		TargetRepository: item.TargetRepository, TargetTag: item.TargetTag, TargetDigest: item.TargetDigest,
		Mode: item.Mode, Overwrite: item.Overwrite, Status: item.Status, Stage: item.Stage,
		CopiedBlobs: item.CopiedBlobs, TotalBlobs: item.TotalBlobs, Message: item.Message,
		Attempt: item.Attempt, OperatorName: item.OperatorName, CreatedAt: item.CreatedAt.Unix(),
	}
	if item.StartedAt != nil {
		out.StartedAt = item.StartedAt.Unix()
	}
	if item.FinishedAt != nil {
		out.FinishedAt = item.FinishedAt.Unix()
	}
	return out
}

func imageSyncInput(item *v1.ImageSyncInput) bo.ImageSyncInput {
	if item == nil {
		return bo.ImageSyncInput{}
	}
	return bo.ImageSyncInput{
		SourceKind: item.GetSourceKind(), SourceRegistryUID: item.GetSourceRegistryUid(),
		SourceImage: item.GetSourceImage(), SourceUsername: item.GetSourceUsername(),
		SourcePassword: item.GetSourcePassword(), SourceAllowInsecureTLS: item.GetSourceAllowInsecureTls(),
		TargetRegistryUID: item.GetTargetRegistryUid(), TargetRepository: item.GetTargetRepository(),
		TargetTag: item.GetTargetTag(), Platforms: item.GetPlatforms(), Overwrite: item.GetOverwrite(),
	}
}
