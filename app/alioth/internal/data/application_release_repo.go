package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationdeployment"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationrelease"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/applicationreleaseoperation"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

var _ repository.ApplicationReleaseRepo = (*applicationReleaseRepo)(nil)

func NewApplicationReleaseRepo(deps *Dependencies, logger *slog.Logger) repository.ApplicationReleaseRepo {
	return &applicationReleaseRepo{repoBase: newBase(deps, logger)}
}

type applicationReleaseRepo struct{ *repoBase }

type storedDeployValues struct {
	Replicas     map[string]int32  `json:"replicas,omitempty"`
	ImageTags    map[string]string `json:"imageTags,omitempty"`
	IngressHosts map[string]string `json:"ingressHosts,omitempty"`
}

func (r *applicationReleaseRepo) CreateDeployment(ctx context.Context, deployment bo.ApplicationDeployment, releases []bo.ApplicationRelease, operations []bo.ApplicationOperation) (*bo.ApplicationDeployment, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	clusters, _ := json.Marshal(deployment.Clusters)
	values, _ := json.Marshal(nonSecretValues(deployment.Values))
	images, _ := json.Marshal(deployment.Images)
	cipher, err := r.encryptSecrets(deployment.Values.SecretValues)
	if err != nil {
		return nil, err
	}
	deploymentUID := id.New()
	_, err = tx.ApplicationDeployment.Create().
		SetUID(deploymentUID).SetNamespaceUID(deployment.NamespaceUID).SetName(deployment.Name).
		SetApplicationUID(deployment.ApplicationUID).SetApplicationName(deployment.ApplicationName).
		SetVersionUID(deployment.VersionUID).SetVersion(deployment.Version).
		SetTargetNamespace(deployment.TargetNamespace).SetClusters(string(clusters)).
		SetValues(string(values)).SetSecretValuesCipher(cipher).SetImages(string(images)).
		SetStatus(applicationdeployment.StatusPENDING).SetCreatedBy(deployment.CreatedBy).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	releaseUIDs := make(map[string]string, len(releases))
	for index := range releases {
		release := releases[index]
		releaseUID := id.New()
		releaseUIDs[release.ClusterUID] = releaseUID
		releaseValues, _ := json.Marshal(nonSecretValues(release.Values))
		releaseImages, _ := json.Marshal(release.Images)
		releaseCipher, err := r.encryptSecrets(release.Values.SecretValues)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ApplicationRelease.Create().
			SetUID(releaseUID).SetNamespaceUID(release.NamespaceUID).SetDeploymentUID(deploymentUID).
			SetClusterUID(release.ClusterUID).SetClusterName(release.ClusterName).
			SetTargetNamespace(release.TargetNamespace).SetName(release.Name).
			SetApplicationUID(release.ApplicationUID).SetApplicationName(release.ApplicationName).
			SetVersionUID(release.VersionUID).SetVersion(release.Version).
			SetStatus(applicationrelease.StatusPENDING).SetValues(string(releaseValues)).
			SetSecretValuesCipher(releaseCipher).SetImages(string(releaseImages)).Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	for index := range operations {
		operation := operations[index]
		opValues, _ := json.Marshal(nonSecretValues(operation.Values))
		opResources, _ := json.Marshal(operation.Resources)
		opDiff, _ := json.Marshal(operation.Diff)
		opCipher, err := r.encryptSecrets(operation.Values.SecretValues)
		if err != nil {
			return nil, err
		}
		if _, err := tx.ApplicationReleaseOperation.Create().
			SetUID(id.New()).SetNamespaceUID(operation.NamespaceUID).
			SetApplicationUID(operation.ApplicationUID).SetDeploymentUID(deploymentUID).
			SetReleaseUID(releaseUIDs[operation.ClusterUID]).SetClusterUID(operation.ClusterUID).
			SetClusterName(operation.ClusterName).
			SetBatchUID(operation.BatchUID).SetAction(applicationreleaseoperation.Action(operation.Action)).
			SetStatus(applicationreleaseoperation.StatusPENDING).SetToVersionUID(operation.ToVersionUID).
			SetValues(string(opValues)).SetSecretValuesCipher(opCipher).SetResources(string(opResources)).
			SetDiff(string(opDiff)).SetOperatorUID(operation.OperatorUID).SetOperatorName(operation.OperatorName).
			SetKeepPersistentVolumeClaims(operation.KeepPVC).SetDeleteClusterScopedResources(operation.DeleteClusterScoped).Save(ctx); err != nil {
			return nil, r.fail(err)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetDeployment(ctx, deployment.NamespaceUID, deploymentUID)
}

func (r *applicationReleaseRepo) GetDeployment(ctx context.Context, namespaceUID, uid string) (*bo.ApplicationDeployment, error) {
	row, err := r.client.ApplicationDeployment.Query().Where(
		applicationdeployment.UIDEQ(uid), applicationdeployment.NamespaceUIDEQ(namespaceUID),
		applicationdeployment.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toDeployment(row)
	releases, err := r.releases(ctx, namespaceUID, uid)
	if err != nil {
		return nil, err
	}
	item.Releases = releases
	item.Clusters = clusterMatrix(item.Clusters, releases, item.VersionUID)
	return &item, nil
}

func (r *applicationReleaseRepo) ListDeployments(ctx context.Context, namespaceUID string, filter bo.ApplicationReleaseFilter, page, pageSize int) ([]bo.ApplicationDeployment, int, error) {
	query := r.client.ApplicationDeployment.Query().Where(
		applicationdeployment.NamespaceUIDEQ(namespaceUID), applicationdeployment.DeletedAtIsNil(),
	)
	if filter.Name != "" {
		query.Where(applicationdeployment.NameContainsFold(filter.Name))
	}
	if filter.Status != "" {
		query.Where(applicationdeployment.StatusEQ(applicationdeployment.Status(filter.Status)))
	}
	if filter.ApplicationUID != "" {
		query.Where(applicationdeployment.ApplicationUIDEQ(filter.ApplicationUID))
	}
	if filter.ClusterUID != "" {
		uids, err := r.client.ApplicationRelease.Query().Where(
			applicationrelease.NamespaceUIDEQ(namespaceUID),
			applicationrelease.ClusterUIDEQ(filter.ClusterUID),
			applicationrelease.DeletedAtIsNil(),
		).Select(applicationrelease.FieldDeploymentUID).Strings(ctx)
		if err != nil {
			return nil, 0, r.fail(err)
		}
		if len(uids) == 0 {
			return []bo.ApplicationDeployment{}, 0, nil
		}
		query.Where(applicationdeployment.UIDIn(uids...))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(applicationdeployment.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.ApplicationDeployment, 0, len(rows))
	for _, row := range rows {
		item := toDeployment(row)
		releases, err := r.releases(ctx, namespaceUID, row.UID)
		if err != nil {
			return nil, 0, err
		}
		item.Releases = releases
		item.Clusters = clusterMatrix(item.Clusters, releases, item.VersionUID)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *applicationReleaseRepo) UpdateDeployment(ctx context.Context, in bo.ApplicationDeployment) error {
	clusters, _ := json.Marshal(in.Clusters)
	values, _ := json.Marshal(nonSecretValues(in.Values))
	images, _ := json.Marshal(in.Images)
	update := r.client.ApplicationDeployment.Update().
		Where(applicationdeployment.UIDEQ(in.UID), applicationdeployment.NamespaceUIDEQ(in.NamespaceUID)).
		SetVersionUID(in.VersionUID).SetVersion(in.Version).SetClusters(string(clusters)).
		SetValues(string(values)).SetImages(string(images)).SetStatus(applicationdeployment.Status(in.Status)).
		SetMessage(in.Message)
	if in.DeployedAt != nil {
		update.SetDeployedAt(*in.DeployedAt)
	}
	_, err := update.Save(ctx)
	return r.fail(err)
}

func (r *applicationReleaseRepo) GetRelease(ctx context.Context, namespaceUID, uid string) (*bo.ApplicationRelease, error) {
	row, err := r.client.ApplicationRelease.Query().Where(
		applicationrelease.UIDEQ(uid), applicationrelease.NamespaceUIDEQ(namespaceUID),
		applicationrelease.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toRelease(row)
	return &item, nil
}

func (r *applicationReleaseRepo) ListReleases(ctx context.Context, namespaceUID string, filter bo.ApplicationReleaseFilter, page, pageSize int) ([]bo.ApplicationRelease, int, error) {
	query := r.client.ApplicationRelease.Query().Where(
		applicationrelease.NamespaceUIDEQ(namespaceUID), applicationrelease.DeletedAtIsNil(),
	)
	if filter.ApplicationUID != "" {
		query.Where(applicationrelease.ApplicationUIDEQ(filter.ApplicationUID))
	}
	if filter.ClusterUID != "" {
		query.Where(applicationrelease.ClusterUIDEQ(filter.ClusterUID))
	}
	if filter.Status != "" {
		query.Where(applicationrelease.StatusEQ(applicationrelease.Status(filter.Status)))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(applicationrelease.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.ApplicationRelease, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRelease(row))
	}
	return items, total, nil
}

func (r *applicationReleaseRepo) UpdateRelease(ctx context.Context, in bo.ApplicationRelease) error {
	values, _ := json.Marshal(nonSecretValues(in.Values))
	images, _ := json.Marshal(in.Images)
	update := r.client.ApplicationRelease.Update().
		Where(applicationrelease.UIDEQ(in.UID), applicationrelease.NamespaceUIDEQ(in.NamespaceUID)).
		SetVersionUID(in.VersionUID).SetVersion(in.Version).SetStatus(applicationrelease.Status(in.Status)).
		SetValues(string(values)).SetImages(string(images)).SetMessage(in.Message).
		SetLastOperationUID(in.LastOperationUID)
	if in.DeployedAt != nil {
		update.SetDeployedAt(*in.DeployedAt)
	}
	_, err := update.Save(ctx)
	return r.fail(err)
}

func (r *applicationReleaseRepo) CreateOperation(ctx context.Context, in bo.ApplicationOperation) (*bo.ApplicationOperation, error) {
	values, _ := json.Marshal(nonSecretValues(in.Values))
	resources, _ := json.Marshal(in.Resources)
	diff, _ := json.Marshal(in.Diff)
	cipher, err := r.encryptSecrets(in.Values.SecretValues)
	if err != nil {
		return nil, err
	}
	row, err := r.client.ApplicationReleaseOperation.Create().
		SetUID(id.New()).SetNamespaceUID(in.NamespaceUID).SetApplicationUID(in.ApplicationUID).
		SetDeploymentUID(in.DeploymentUID).SetReleaseUID(in.ReleaseUID).
		SetClusterUID(in.ClusterUID).SetClusterName(in.ClusterName).SetBatchUID(in.BatchUID).
		SetAction(applicationreleaseoperation.Action(in.Action)).
		SetStatus(applicationreleaseoperation.Status(in.Status)).
		SetFromVersionUID(in.FromVersionUID).SetToVersionUID(in.ToVersionUID).
		SetValues(string(values)).SetSecretValuesCipher(cipher).SetResources(string(resources)).
		SetDiff(string(diff)).SetMessage(in.Message).SetOperatorUID(in.OperatorUID).
		SetOperatorName(in.OperatorName).SetKeepPersistentVolumeClaims(in.KeepPVC).
		SetDeleteClusterScopedResources(in.DeleteClusterScoped).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	return r.GetOperation(ctx, in.NamespaceUID, row.UID)
}

func (r *applicationReleaseRepo) GetOperation(ctx context.Context, namespaceUID, uid string) (*bo.ApplicationOperation, error) {
	row, err := r.client.ApplicationReleaseOperation.Query().Where(
		applicationreleaseoperation.UIDEQ(uid), applicationreleaseoperation.NamespaceUIDEQ(namespaceUID),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toOperation(row)
	return &item, nil
}

func (r *applicationReleaseRepo) ListOperations(ctx context.Context, namespaceUID, deploymentUID, releaseUID string, page, pageSize int) ([]bo.ApplicationOperation, int, error) {
	query := r.client.ApplicationReleaseOperation.Query().Where(applicationreleaseoperation.NamespaceUIDEQ(namespaceUID))
	if deploymentUID != "" {
		query.Where(applicationreleaseoperation.DeploymentUIDEQ(deploymentUID))
	}
	if releaseUID != "" {
		query.Where(applicationreleaseoperation.ReleaseUIDEQ(releaseUID))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(applicationreleaseoperation.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.ApplicationOperation, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOperation(row))
	}
	return items, total, nil
}

func (r *applicationReleaseRepo) ClaimOperation(ctx context.Context, owner string, now, leaseUntil time.Time) (*bo.ApplicationOperation, error) {
	const claimSQL = `
UPDATE application_release_operations
SET status = 'RUNNING', started_at = $1, lease_owner = $2, lease_until = $3, updated_at = $1
WHERE uid = (
  SELECT uid FROM application_release_operations
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
	return r.GetOperation(ctx, namespaceUID, uid)
}

func (r *applicationReleaseRepo) UpdateOperation(ctx context.Context, in bo.ApplicationOperation) error {
	resources, _ := json.Marshal(in.Resources)
	diff, _ := json.Marshal(in.Diff)
	update := r.client.ApplicationReleaseOperation.Update().
		Where(applicationreleaseoperation.UIDEQ(in.UID), applicationreleaseoperation.NamespaceUIDEQ(in.NamespaceUID)).
		SetStatus(applicationreleaseoperation.Status(in.Status)).SetMessage(in.Message).
		SetResources(string(resources)).SetDiff(string(diff)).SetLeaseOwner("")
	if in.StartedAt != nil {
		update.SetStartedAt(*in.StartedAt)
	}
	if in.FinishedAt != nil {
		update.SetFinishedAt(*in.FinishedAt)
	}
	_, err := update.Save(ctx)
	return r.fail(err)
}

func (r *applicationReleaseRepo) HasActiveOperation(ctx context.Context, releaseUID string) (bool, error) {
	ok, err := r.client.ApplicationReleaseOperation.Query().Where(
		applicationreleaseoperation.ReleaseUIDEQ(releaseUID),
		applicationreleaseoperation.StatusIn(applicationreleaseoperation.StatusPENDING, applicationreleaseoperation.StatusRUNNING),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *applicationReleaseRepo) LastSuccessfulOperation(ctx context.Context, releaseUID string) (*bo.ApplicationOperation, error) {
	row, err := r.client.ApplicationReleaseOperation.Query().Where(
		applicationreleaseoperation.ReleaseUIDEQ(releaseUID),
		applicationreleaseoperation.StatusEQ(applicationreleaseoperation.StatusSUCCEEDED),
	).Order(ent.Desc(applicationreleaseoperation.FieldFinishedAt)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, r.fail(err)
	}
	item := toOperation(row)
	return &item, nil
}

func (r *applicationReleaseRepo) releases(ctx context.Context, namespaceUID, deploymentUID string) ([]bo.ApplicationRelease, error) {
	rows, err := r.client.ApplicationRelease.Query().Where(
		applicationrelease.NamespaceUIDEQ(namespaceUID), applicationrelease.DeploymentUIDEQ(deploymentUID),
		applicationrelease.DeletedAtIsNil(),
	).Order(ent.Asc(applicationrelease.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.ApplicationRelease, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRelease(row))
	}
	return items, nil
}

func (r *applicationReleaseRepo) encryptSecrets(values map[string]string) (string, error) {
	if len(values) == 0 {
		return "", nil
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return "", merr.Server("request failed")
	}
	cipher, err := r.box.encrypt(string(raw))
	if err != nil {
		return "", err
	}
	return cipher, nil
}

func nonSecretValues(values bo.ApplicationDeployValues) storedDeployValues {
	return storedDeployValues{Replicas: values.Replicas, ImageTags: values.ImageTags, IngressHosts: values.IngressHosts}
}

func readValues(raw string) bo.ApplicationDeployValues {
	var stored storedDeployValues
	_ = json.Unmarshal([]byte(raw), &stored)
	return bo.ApplicationDeployValues{Replicas: stored.Replicas, ImageTags: stored.ImageTags, IngressHosts: stored.IngressHosts}
}

func clusterMatrix(clusters []bo.DeploymentCluster, releases []bo.ApplicationRelease, desiredVersionUID string) []bo.DeploymentCluster {
	index := make(map[string]bo.ApplicationRelease, len(releases))
	for _, release := range releases {
		index[release.ClusterUID] = release
	}
	for position := range clusters {
		if release, ok := index[clusters[position].UID]; ok {
			clusters[position].Status = release.Status
			clusters[position].ReleaseUID = release.UID
			clusters[position].Version = release.Version
			clusters[position].Message = release.Message
			// 只有出现过成功操作后，release 上的版本才是集群实际运行的版本。
			if release.LastOperationUID != "" {
				clusters[position].CurrentVersionUID = release.VersionUID
				clusters[position].CurrentVersion = release.Version
			}
			clusters[position].InSync = clusters[position].CurrentVersionUID != "" &&
				clusters[position].CurrentVersionUID == desiredVersionUID
		}
	}
	return clusters
}

func toDeployment(row *ent.ApplicationDeployment) bo.ApplicationDeployment {
	item := bo.ApplicationDeployment{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name,
		ApplicationUID: row.ApplicationUID, ApplicationName: row.ApplicationName,
		VersionUID: row.VersionUID, Version: row.Version, TargetNamespace: row.TargetNamespace,
		Status: string(row.Status), Message: row.Message, CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt, DeployedAt: row.DeployedAt, Values: readValues(row.Values),
	}
	_ = json.Unmarshal([]byte(row.Clusters), &item.Clusters)
	_ = json.Unmarshal([]byte(row.Images), &item.Images)
	return item
}

func toRelease(row *ent.ApplicationRelease) bo.ApplicationRelease {
	item := bo.ApplicationRelease{
		UID: row.UID, NamespaceUID: row.NamespaceUID, DeploymentUID: row.DeploymentUID,
		ClusterUID: row.ClusterUID, ClusterName: row.ClusterName, TargetNamespace: row.TargetNamespace,
		Name: row.Name, ApplicationUID: row.ApplicationUID, ApplicationName: row.ApplicationName,
		VersionUID: row.VersionUID, Version: row.Version, Status: string(row.Status),
		Message: row.Message, LastOperationUID: row.LastOperationUID, DeployedAt: row.DeployedAt,
		CreatedAt: row.CreatedAt, Values: readValues(row.Values),
	}
	_ = json.Unmarshal([]byte(row.Images), &item.Images)
	return item
}

func toOperation(row *ent.ApplicationReleaseOperation) bo.ApplicationOperation {
	item := bo.ApplicationOperation{
		UID: row.UID, NamespaceUID: row.NamespaceUID, ApplicationUID: row.ApplicationUID,
		DeploymentUID: row.DeploymentUID, ReleaseUID: row.ReleaseUID,
		ClusterUID: row.ClusterUID, ClusterName: row.ClusterName, BatchUID: row.BatchUID,
		Action: string(row.Action), Status: string(row.Status),
		FromVersionUID: row.FromVersionUID, ToVersionUID: row.ToVersionUID,
		Message: row.Message, OperatorUID: row.OperatorUID, OperatorName: row.OperatorName,
		KeepPVC: row.KeepPersistentVolumeClaims, DeleteClusterScoped: row.DeleteClusterScopedResources,
		LeaseOwner: row.LeaseOwner, LeaseUntil: row.LeaseUntil, StartedAt: row.StartedAt,
		FinishedAt: row.FinishedAt, CreatedAt: row.CreatedAt, Values: readValues(row.Values),
	}
	_ = json.Unmarshal([]byte(row.Resources), &item.Resources)
	_ = json.Unmarshal([]byte(row.Diff), &item.Diff)
	return item
}

func (r *applicationReleaseRepo) HasActiveDeploymentsForApplication(ctx context.Context, namespaceUID, applicationUID string) (bool, error) {
	ok, err := r.client.ApplicationDeployment.Query().Where(
		applicationdeployment.NamespaceUIDEQ(namespaceUID),
		applicationdeployment.ApplicationUIDEQ(applicationUID),
		applicationdeployment.DeletedAtIsNil(),
		applicationdeployment.StatusNEQ(applicationdeployment.StatusUNINSTALLED),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *applicationReleaseRepo) ReleaseUsedVersion(ctx context.Context, namespaceUID, releaseUID, versionUID string) (bool, error) {
	release, err := r.client.ApplicationRelease.Query().Where(
		applicationrelease.UIDEQ(releaseUID), applicationrelease.NamespaceUIDEQ(namespaceUID),
		applicationrelease.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	if release.VersionUID == versionUID {
		return true, nil
	}
	ok, err := r.client.ApplicationReleaseOperation.Query().Where(
		applicationreleaseoperation.NamespaceUIDEQ(namespaceUID),
		applicationreleaseoperation.ReleaseUIDEQ(releaseUID),
		applicationreleaseoperation.Or(
			applicationreleaseoperation.FromVersionUIDEQ(versionUID),
			applicationreleaseoperation.ToVersionUIDEQ(versionUID),
		),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}

func (r *applicationReleaseRepo) HasDeploymentsForVersion(ctx context.Context, namespaceUID, versionUID string) (bool, error) {
	ok, err := r.client.ApplicationRelease.Query().Where(
		applicationrelease.NamespaceUIDEQ(namespaceUID),
		applicationrelease.VersionUIDEQ(versionUID),
		applicationrelease.DeletedAtIsNil(),
		applicationrelease.StatusNEQ(applicationrelease.StatusUNINSTALLED),
	).Exist(ctx)
	if err != nil {
		return false, r.fail(err)
	}
	return ok, nil
}
