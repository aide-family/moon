package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/imagesync"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

var _ repository.ImageSyncRepo = (*imageSyncRepo)(nil)

func NewImageSyncRepo(deps *Dependencies, logger *slog.Logger) repository.ImageSyncRepo {
	return &imageSyncRepo{repoBase: newBase(deps, logger)}
}

type imageSyncRepo struct{ *repoBase }

func (r *imageSyncRepo) CreateImageSync(ctx context.Context, in bo.ImageSync, sourcePassword string) (*bo.ImageSync, error) {
	cipher := ""
	if in.SourceKind == "EXTERNAL" && sourcePassword != "" {
		encoded, err := r.box.encrypt(sourcePassword)
		if err != nil {
			return nil, err
		}
		cipher = encoded
	}
	platforms, _ := json.Marshal(in.SourcePlatforms)
	row, err := r.client.ImageSync.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetSourceKind(imagesync.SourceKind(in.SourceKind)).
		SetSourceRegistryUID(in.SourceRegistryUID).
		SetSourceRegistryName(in.SourceRegistryName).
		SetSourceEndpoint(in.SourceEndpoint).
		SetSourceImage(in.SourceImage).
		SetSourceUsername(in.SourceUsername).
		SetSourceRepository(in.SourceRepository).
		SetSourceTag(in.SourceTag).
		SetSourceDigest(in.SourceDigest).
		SetSourcePlatforms(string(platforms)).
		SetSourceCredentialCipher(cipher).
		SetSourceAllowInsecureTLS(in.SourceAllowInsecureTLS).
		SetTargetRegistryUID(in.TargetRegistryUID).
		SetTargetRegistryName(in.TargetRegistryName).
		SetTargetEndpoint(in.TargetEndpoint).
		SetTargetRepository(in.TargetRepository).
		SetTargetTag(in.TargetTag).
		SetMode(imagesync.Mode(in.Mode)).
		SetOverwrite(in.Overwrite).
		SetStatus(imagesync.StatusPENDING).
		SetOperatorUID(in.OperatorUID).
		SetOperatorName(in.OperatorName).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetImageSync(ctx, row.NamespaceUID, row.UID)
}

func (r *imageSyncRepo) GetImageSync(ctx context.Context, namespaceUID, uid string) (*bo.ImageSync, error) {
	row, err := r.client.ImageSync.Query().Where(
		imagesync.UIDEQ(uid), imagesync.NamespaceUIDEQ(namespaceUID),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("image sync not found")
		}
		return nil, r.fail(err)
	}
	item := toImageSync(row)
	return &item, nil
}

func (r *imageSyncRepo) ListImageSyncs(ctx context.Context, namespaceUID string, filter bo.ImageSyncFilter, page, pageSize int) ([]bo.ImageSync, int, error) {
	query := r.client.ImageSync.Query().Where(imagesync.NamespaceUIDEQ(namespaceUID))
	if filter.Status != "" {
		query.Where(imagesync.StatusEQ(imagesync.Status(filter.Status)))
	}
	if filter.SourceKind != "" {
		query.Where(imagesync.SourceKindEQ(imagesync.SourceKind(filter.SourceKind)))
	}
	if filter.RegistryUID != "" {
		query.Where(imagesync.Or(
			imagesync.SourceRegistryUIDEQ(filter.RegistryUID),
			imagesync.TargetRegistryUIDEQ(filter.RegistryUID),
		))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(imagesync.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.ImageSync, 0, len(rows))
	for _, row := range rows {
		items = append(items, toImageSync(row))
	}
	return items, total, nil
}

func (r *imageSyncRepo) RetryImageSync(ctx context.Context, namespaceUID, uid string) (*bo.ImageSync, error) {
	n, err := r.client.ImageSync.Update().
		Where(imagesync.UIDEQ(uid), imagesync.NamespaceUIDEQ(namespaceUID),
			imagesync.StatusIn(imagesync.StatusFAILED, imagesync.StatusCANCELED)).
		SetStatus(imagesync.StatusPENDING).
		SetStage("").
		SetMessage("").
		SetCopiedBlobs(0).
		SetTotalBlobs(0).
		SetTargetDigest("").
		SetLeaseOwner("").
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.Conflict("image sync is not retryable")
	}
	return r.GetImageSync(ctx, namespaceUID, uid)
}

func (r *imageSyncRepo) CancelImageSync(ctx context.Context, namespaceUID, uid string) (*bo.ImageSync, error) {
	n, err := r.client.ImageSync.Update().
		Where(imagesync.UIDEQ(uid), imagesync.NamespaceUIDEQ(namespaceUID),
			imagesync.StatusIn(imagesync.StatusPENDING, imagesync.StatusRUNNING)).
		SetStatus(imagesync.StatusCANCELED).
		SetMessage("canceled by user").
		SetLeaseOwner("").
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.Conflict("image sync cannot be canceled")
	}
	return r.GetImageSync(ctx, namespaceUID, uid)
}

func (r *imageSyncRepo) ClaimImageSync(ctx context.Context, owner string, now, leaseUntil time.Time) (*bo.ImageSync, error) {
	const claimSQL = `
UPDATE image_syncs
SET status = 'RUNNING', started_at = $1, lease_owner = $2, lease_until = $3, updated_at = $1, attempt = attempt + 1
WHERE uid = (
  SELECT uid FROM image_syncs
  WHERE status = 'PENDING'
  ORDER BY created_at ASC
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
RETURNING uid, namespace_uid`
	var uid, namespaceUID string
	err := r.db.QueryRowContext(ctx, claimSQL, now, owner, leaseUntil).Scan(&uid, &namespaceUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, r.fail(err)
	}
	return r.GetImageSync(ctx, namespaceUID, uid)
}

func (r *imageSyncRepo) UpdateImageSync(ctx context.Context, in bo.ImageSync) error {
	update := r.client.ImageSync.Update().
		Where(imagesync.UIDEQ(in.UID), imagesync.NamespaceUIDEQ(in.NamespaceUID), imagesync.StatusEQ(imagesync.StatusRUNNING)).
		SetStatus(imagesync.Status(in.Status)).
		SetStage(in.Stage).
		SetCopiedBlobs(int(in.CopiedBlobs)).
		SetTotalBlobs(int(in.TotalBlobs)).
		SetMessage(in.Message).
		SetTargetDigest(in.TargetDigest).
		SetSourceDigest(in.SourceDigest).
		SetLeaseOwner("")
	if in.StartedAt != nil {
		update.SetStartedAt(*in.StartedAt)
	}
	if in.FinishedAt != nil {
		update.SetFinishedAt(*in.FinishedAt)
	}
	_, err := update.Save(ctx)
	return r.fail(err)
}

func (r *imageSyncRepo) ClearExpiredCredential(ctx context.Context, before time.Time) (int, error) {
	n, err := r.client.ImageSync.Update().
		Where(
			imagesync.SourceKindEQ(imagesync.SourceKindEXTERNAL),
			imagesync.SourceCredentialCipherNEQ(""),
			imagesync.StatusIn(imagesync.StatusSUCCEEDED, imagesync.StatusFAILED, imagesync.StatusCANCELED),
			imagesync.FinishedAtNotNil(),
			imagesync.FinishedAtLT(before),
		).
		SetSourceCredentialCipher("").
		Save(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return n, nil
}

func toImageSync(row *ent.ImageSync) bo.ImageSync {
	var platforms []string
	_ = json.Unmarshal([]byte(row.SourcePlatforms), &platforms)
	return bo.ImageSync{
		UID: row.UID, NamespaceUID: row.NamespaceUID,
		SourceKind: string(row.SourceKind), SourceRegistryUID: row.SourceRegistryUID,
		SourceRegistryName: row.SourceRegistryName, SourceEndpoint: row.SourceEndpoint,
		SourceImage: row.SourceImage, SourceUsername: row.SourceUsername, SourceRepository: row.SourceRepository,
		SourceTag: row.SourceTag, SourceDigest: row.SourceDigest,
		SourcePlatforms: platforms, SourceCredentialCipher: row.SourceCredentialCipher,
		SourceAllowInsecureTLS: row.SourceAllowInsecureTLS,
		TargetRegistryUID:      row.TargetRegistryUID, TargetRegistryName: row.TargetRegistryName,
		TargetEndpoint: row.TargetEndpoint, TargetRepository: row.TargetRepository,
		TargetTag: row.TargetTag, TargetDigest: row.TargetDigest,
		Mode: string(row.Mode), Overwrite: row.Overwrite, Status: string(row.Status),
		Stage: row.Stage, CopiedBlobs: int32(row.CopiedBlobs), TotalBlobs: int32(row.TotalBlobs),
		Message: row.Message, Attempt: int32(row.Attempt),
		OperatorUID: row.OperatorUID, OperatorName: row.OperatorName,
		LeaseOwner: row.LeaseOwner, LeaseUntil: row.LeaseUntil,
		StartedAt: row.StartedAt, FinishedAt: row.FinishedAt, CreatedAt: row.CreatedAt,
	}
}

// sourceCredential 读取并解密外部源的临时密码。
func (r *imageSyncRepo) sourceCredential(ctx context.Context, uid string) (string, error) {
	row, err := r.client.ImageSync.Query().Where(imagesync.UIDEQ(uid)).Only(ctx)
	if err != nil {
		return "", r.fail(err)
	}
	if row.SourceCredentialCipher == "" {
		return "", nil
	}
	plain, err := r.box.decrypt(row.SourceCredentialCipher)
	if err != nil {
		r.log.Error(fmt.Sprintf("image sync credential: %v", err))
		return "", merr.Server("request failed")
	}
	return plain, nil
}
