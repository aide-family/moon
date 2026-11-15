package registry

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const syncLease = 5 * time.Minute

// Synchronizer 创建并执行镜像同步任务。远端元数据实时读取，
// 任务持久化在 image_syncs 表，由调度器按租约认领。
type Synchronizer struct {
	registries repository.RegistryRepo
	syncs      repository.ImageSyncRepo
	images     repository.ImageRegistryRepo
	access     access.Checker
	log        *slog.Logger
	owner      string
}

func NewSynchronizer(
	registries repository.RegistryRepo,
	syncs repository.ImageSyncRepo,
	images repository.ImageRegistryRepo,
	checker access.Checker,
	logger *slog.Logger,
) *Synchronizer {
	return &Synchronizer{registries: registries, syncs: syncs, images: images, access: checker, log: logger, owner: id.New()}
}

func (sync *Synchronizer) List(ctx context.Context, pageNo, pageSize int32, filter bo.ImageSyncFilter) ([]bo.ImageSync, int, int, int, error) {
	ns, err := sync.access.Allow(ctx, perm.RegistrySyncRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := sync.syncs.ListImageSyncs(ctx, ns, filter, p, s)
	return items, total, p, s, sync.access.Wrap(err)
}

func (sync *Synchronizer) Get(ctx context.Context, uid string) (*bo.ImageSync, error) {
	ns, err := sync.access.Allow(ctx, perm.RegistrySyncRead)
	if err != nil {
		return nil, err
	}
	item, err := sync.syncs.GetImageSync(ctx, ns, uid)
	return item, sync.access.Wrap(err)
}

func (sync *Synchronizer) Preview(ctx context.Context, in bo.ImageSyncInput) (*bo.ImageSyncPreview, error) {
	ns, err := sync.access.Allow(ctx, perm.RegistrySyncWrite)
	if err != nil {
		return nil, err
	}
	if err := validateImageSyncInput(&in); err != nil {
		return nil, err
	}
	target, err := sync.registries.GetRegistry(ctx, ns, in.TargetRegistryUID)
	if err != nil {
		return nil, sync.access.Wrap(err)
	}
	preview, err := sync.resolveSource(ctx, ns, in)
	if err != nil {
		return nil, err
	}
	exists, digest, err := sync.images.TagExists(ctx, *target, in.TargetRepository, in.TargetTag)
	if err != nil {
		return nil, sync.access.Wrap(err)
	}
	preview.TargetExists = exists
	preview.TargetDigest = digest
	return preview, nil
}

func (sync *Synchronizer) Create(ctx context.Context, inputs []bo.ImageSyncInput) ([]bo.ImageSync, []bo.ImageSyncFailure, error) {
	ns, err := sync.access.Allow(ctx, perm.RegistrySyncWrite)
	if err != nil {
		return nil, nil, err
	}
	if len(inputs) == 0 {
		return nil, nil, merr.Invalid("at least one image is required")
	}
	if len(inputs) > 50 {
		return nil, nil, merr.Invalid("too many images in one request")
	}
	operator, _ := actor.UserID(ctx)
	created := make([]bo.ImageSync, 0, len(inputs))
	failed := make([]bo.ImageSyncFailure, 0)
	for index := range inputs {
		in := inputs[index]
		item, err := sync.createOne(ctx, ns, in, operator)
		if err != nil {
			failed = append(failed, bo.ImageSyncFailure{Index: int32(index), Message: err.Error()})
			continue
		}
		created = append(created, *item)
	}
	return created, failed, nil
}

func (sync *Synchronizer) createOne(ctx context.Context, namespaceUID string, in bo.ImageSyncInput, operator string) (*bo.ImageSync, error) {
	if err := validateImageSyncInput(&in); err != nil {
		return nil, err
	}
	target, err := sync.registries.GetRegistry(ctx, namespaceUID, in.TargetRegistryUID)
	if err != nil {
		return nil, sync.access.Wrap(err)
	}
	preview, err := sync.resolveSource(ctx, namespaceUID, in)
	if err != nil {
		return nil, err
	}
	exists, _, err := sync.images.TagExists(ctx, *target, in.TargetRepository, in.TargetTag)
	if err != nil {
		return nil, sync.access.Wrap(err)
	}
	if exists && !in.Overwrite {
		return nil, merr.Conflict("target tag already exists")
	}
	sourceRepo, sourceTag, sourceDigest := splitImageRef(preview.SourceImage)
	if preview.SourceDigest != "" {
		sourceDigest = preview.SourceDigest
	}
	mode := "ALL_PLATFORMS"
	if len(in.Platforms) > 0 {
		mode = "SELECTED_PLATFORMS"
	}
	task := bo.ImageSync{
		NamespaceUID: namespaceUID, SourceKind: in.SourceKind,
		SourceRegistryUID: in.SourceRegistryUID, SourceImage: preview.SourceImage,
		SourceRepository: sourceRepo, SourceTag: sourceTag, SourceDigest: sourceDigest,
		SourcePlatforms: preview.SourcePlatforms, SourceAllowInsecureTLS: in.SourceAllowInsecureTLS,
		TargetRegistryUID: target.UID, TargetRegistryName: target.Name, TargetEndpoint: target.Endpoint,
		TargetRepository: in.TargetRepository, TargetTag: in.TargetTag,
		Mode: mode, Overwrite: in.Overwrite, Status: "PENDING",
		OperatorUID: operator, OperatorName: operator,
	}
	if in.SourceKind == "CONNECTION" {
		source, err := sync.registries.GetRegistry(ctx, namespaceUID, in.SourceRegistryUID)
		if err != nil {
			return nil, sync.access.Wrap(err)
		}
		task.SourceRegistryName = source.Name
		task.SourceEndpoint = source.Endpoint
	}
	if len(in.Platforms) > 0 {
		task.SourcePlatforms = in.Platforms
	}
	return sync.syncs.CreateImageSync(ctx, task, in.SourcePassword)
}

// Retry 让失败/取消的任务重新排队。临时凭据仍在保留窗口内时可直接重试。
func (sync *Synchronizer) Retry(ctx context.Context, uid string) (*bo.ImageSync, error) {
	ns, err := sync.access.Allow(ctx, perm.RegistrySyncWrite)
	if err != nil {
		return nil, err
	}
	item, err := sync.syncs.RetryImageSync(ctx, ns, uid)
	return item, sync.access.Wrap(err)
}

func (sync *Synchronizer) Cancel(ctx context.Context, uid string) (*bo.ImageSync, error) {
	ns, err := sync.access.Allow(ctx, perm.RegistrySyncWrite)
	if err != nil {
		return nil, err
	}
	item, err := sync.syncs.CancelImageSync(ctx, ns, uid)
	return item, sync.access.Wrap(err)
}

// ExecuteNext 认领一条待执行任务并同步执行。调度循环按 tick 调用它。
func (sync *Synchronizer) ExecuteNext(ctx context.Context) error {
	now := time.Now().UTC()
	item, err := sync.syncs.ClaimImageSync(ctx, sync.owner, now, now.Add(syncLease))
	if err != nil || item == nil {
		return err
	}
	sync.execute(ctx, *item)
	return nil
}

func (sync *Synchronizer) execute(ctx context.Context, item bo.ImageSync) {
	started := time.Now().UTC()
	item.StartedAt = &started
	sync.stage(ctx, item, "PULL_MANIFEST")
	digest, err := sync.images.CopyImage(ctx, item)
	finished := time.Now().UTC()
	item.FinishedAt = &finished
	if err != nil {
		item.Status = "FAILED"
		item.Stage = "FAILED"
		item.Message = err.Error()
		if updateErr := sync.syncs.UpdateImageSync(ctx, item); updateErr != nil {
			sync.log.Error("update image sync: " + updateErr.Error())
		}
		return
	}
	item.Status = "SUCCEEDED"
	item.Stage = "VERIFY"
	item.TargetDigest = digest
	item.Message = ""
	if updateErr := sync.syncs.UpdateImageSync(ctx, item); updateErr != nil {
		sync.log.Error("update image sync: " + updateErr.Error())
	}
}

func (sync *Synchronizer) stage(ctx context.Context, item bo.ImageSync, stage string) {
	item.Stage = stage
	if err := sync.syncs.UpdateImageSync(ctx, item); err != nil {
		sync.log.Error("update image sync stage: " + err.Error())
	}
}

func (sync *Synchronizer) resolveSource(ctx context.Context, namespaceUID string, in bo.ImageSyncInput) (*bo.ImageSyncPreview, error) {
	source := bo.ImageSyncSource{
		Kind: in.SourceKind, RegistryUID: in.SourceRegistryUID, Image: in.SourceImage,
		Username: in.SourceUsername, Password: in.SourcePassword, AllowInsecureTLS: in.SourceAllowInsecureTLS,
	}
	if in.SourceKind == "CONNECTION" {
		if _, err := sync.registries.GetRegistry(ctx, namespaceUID, in.SourceRegistryUID); err != nil {
			return nil, sync.access.Wrap(err)
		}
	}
	preview, err := sync.images.ResolveSource(ctx, source, nil)
	return preview, sync.access.Wrap(err)
}

// Cleanup 清理终态超过 24 小时的外部临时凭据。
func (sync *Synchronizer) Cleanup(ctx context.Context) error {
	_, err := sync.syncs.ClearExpiredCredential(ctx, time.Now().UTC().Add(-24*time.Hour))
	return err
}

func validateImageSyncInput(in *bo.ImageSyncInput) error {
	in.SourceKind = strings.TrimSpace(in.SourceKind)
	in.TargetRegistryUID = strings.TrimSpace(in.TargetRegistryUID)
	in.TargetRepository = strings.Trim(strings.TrimSpace(in.TargetRepository), "/")
	in.TargetTag = strings.TrimSpace(in.TargetTag)
	if in.SourceKind != "CONNECTION" && in.SourceKind != "EXTERNAL" {
		return merr.Invalid("invalid image sync source")
	}
	if in.SourceKind == "CONNECTION" {
		in.SourceRegistryUID = strings.TrimSpace(in.SourceRegistryUID)
		if in.SourceRegistryUID == "" || strings.TrimSpace(in.SourceImage) == "" {
			return merr.Invalid("source registry and image are required")
		}
	} else {
		in.SourceImage = strings.TrimSpace(in.SourceImage)
		if in.SourceImage == "" || len(in.SourceImage) > maxSourceImageBytes {
			return merr.Invalid("external image reference is required")
		}
	}
	if in.TargetRegistryUID == "" || in.TargetRepository == "" || in.TargetTag == "" {
		return merr.Invalid("target registry, repository and tag are required")
	}
	if len([]rune(in.TargetRepository)) > maxRepositoryRunes || len([]rune(in.TargetTag)) > maxTagRunes {
		return merr.Invalid("target repository or tag is too long")
	}
	for _, platform := range in.Platforms {
		parts := strings.Split(platform, "/")
		if len(parts) < 2 || len(parts) > 3 || parts[0] == "" || parts[1] == "" {
			return merr.Invalid("invalid target platform")
		}
	}
	return nil
}

func splitImageRef(image string) (repository, tag, digest string) {
	image = strings.TrimSpace(image)
	if index := strings.LastIndex(image, "@"); index > 0 {
		digest = image[index+1:]
		image = image[:index]
	}
	slash := strings.LastIndex(image, "/")
	if index := strings.LastIndex(image, ":"); index > slash {
		tag = image[index+1:]
		image = image[:index]
	}
	return image, tag, digest
}
