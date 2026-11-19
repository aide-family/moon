package application

import (
	"context"
	"log/slog"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// Deployer 负责部署组预检、创建与升级/回滚/卸载操作编排。
type Deployer struct {
	apps       repository.ApplicationRepo
	releases   repository.ApplicationReleaseRepo
	clusters   repository.ClusterRepo
	registries repository.RegistryRepo
	images     repository.ImageRegistryRepo
	applier    repository.ApplicationApplier
	access     access.Checker
	log        *slog.Logger
}

func NewDeployer(
	apps repository.ApplicationRepo,
	releases repository.ApplicationReleaseRepo,
	clusters repository.ClusterRepo,
	registries repository.RegistryRepo,
	images repository.ImageRegistryRepo,
	applier repository.ApplicationApplier,
	checker access.Checker,
	logger *slog.Logger,
) *Deployer {
	return &Deployer{
		apps: apps, releases: releases, clusters: clusters, registries: registries,
		images: images, applier: applier, access: checker, log: logger,
	}
}

func (deployer *Deployer) List(ctx context.Context, pageNo, pageSize int32, filter bo.ApplicationReleaseFilter) ([]bo.ApplicationDeployment, int, int, int, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := deployer.releases.ListDeployments(ctx, ns, filter, p, s)
	return items, total, p, s, deployer.access.Wrap(err)
}

// ListReleases 按应用/集群/状态分页读取集群实例，供发布记录页使用。
func (deployer *Deployer) ListReleases(ctx context.Context, pageNo, pageSize int32, filter bo.ApplicationReleaseFilter) ([]bo.ApplicationRelease, int, int, int, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := deployer.releases.ListReleases(ctx, ns, filter, p, s)
	return items, total, p, s, deployer.access.Wrap(err)
}

func (deployer *Deployer) Get(ctx context.Context, uid string) (*bo.ApplicationDeployment, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseRead)
	if err != nil {
		return nil, err
	}
	item, err := deployer.releases.GetDeployment(ctx, ns, uid)
	return item, deployer.access.Wrap(err)
}

func (deployer *Deployer) GetRelease(ctx context.Context, uid string) (*bo.ApplicationRelease, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseRead)
	if err != nil {
		return nil, err
	}
	item, err := deployer.releases.GetRelease(ctx, ns, uid)
	return item, deployer.access.Wrap(err)
}

func (deployer *Deployer) ListOperations(ctx context.Context, deploymentUID, releaseUID string, pageNo, pageSize int32) ([]bo.ApplicationOperation, int, int, int, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := deployer.releases.ListOperations(ctx, ns, deploymentUID, releaseUID, p, s)
	return items, total, p, s, deployer.access.Wrap(err)
}

func (deployer *Deployer) GetOperation(ctx context.Context, uid string) (*bo.ApplicationOperation, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseRead)
	if err != nil {
		return nil, err
	}
	item, err := deployer.releases.GetOperation(ctx, ns, uid)
	return item, deployer.access.Wrap(err)
}

func (deployer *Deployer) Create(ctx context.Context, applicationUID, versionUID, name, targetNamespace string, clusterUIDs []string, values bo.ApplicationDeployValues) (*bo.ApplicationDeployment, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseWrite)
	if err != nil {
		return nil, err
	}
	if err := deployer.access.Require(ctx, perm.ClusterRead); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	targetNamespace = strings.TrimSpace(targetNamespace)
	if name == "" || targetNamespace == "" {
		return nil, merr.Invalid("deployment name and namespace are required")
	}
	clusterUIDs = uniqueStrings(clusterUIDs)
	if len(clusterUIDs) == 0 || len(clusterUIDs) > 20 {
		return nil, merr.Invalid("select between 1 and 20 clusters")
	}
	version, err := deployer.apps.GetVersion(ctx, ns, applicationUID, versionUID)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	if version.Status != "PUBLISHED" {
		return nil, merr.Conflict("only a published version can be deployed")
	}
	app, err := deployer.apps.GetApplication(ctx, ns, applicationUID)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	clusters := make([]bo.DeploymentCluster, 0, len(clusterUIDs))
	for _, clusterUID := range clusterUIDs {
		cluster, err := deployer.clusters.GetCluster(ctx, ns, clusterUID)
		if err != nil {
			return nil, deployer.access.Wrap(err)
		}
		clusters = append(clusters, bo.DeploymentCluster{UID: cluster.UID, Name: cluster.Name, Status: "PENDING"})
	}
	if err := validateDeployValues(values, version); err != nil {
		return nil, err
	}
	if err := deployer.preflight(ctx, ns, clusters, targetNamespace, version); err != nil {
		return nil, err
	}
	operator, _ := actor.UserID(ctx)
	batchUID := id.New()
	deployment := bo.ApplicationDeployment{
		NamespaceUID: ns, Name: name, ApplicationUID: app.UID, ApplicationName: app.Name,
		VersionUID: version.UID, Version: version.Version, TargetNamespace: targetNamespace,
		Clusters: clusters, Values: values, Images: version.Images, Status: "PENDING",
		CreatedBy: operator,
	}
	releases := make([]bo.ApplicationRelease, 0, len(clusters))
	operations := make([]bo.ApplicationOperation, 0, len(clusters))
	for _, cluster := range clusters {
		releases = append(releases, bo.ApplicationRelease{
			NamespaceUID: ns, ClusterUID: cluster.UID, ClusterName: cluster.Name,
			TargetNamespace: targetNamespace, Name: name, ApplicationUID: app.UID,
			ApplicationName: app.Name, VersionUID: version.UID, Version: version.Version,
			Status: "PENDING", Values: values, Images: version.Images,
		})
		operations = append(operations, bo.ApplicationOperation{
			NamespaceUID: ns, ApplicationUID: app.UID, BatchUID: batchUID,
			ClusterUID: cluster.UID, ClusterName: cluster.Name,
			Action: "INSTALL", Status: "PENDING", ToVersionUID: version.UID,
			Values: values, OperatorUID: operator, OperatorName: operator, KeepPVC: true,
		})
	}
	item, err := deployer.releases.CreateDeployment(ctx, deployment, releases, operations)
	return item, deployer.access.Wrap(err)
}

func (deployer *Deployer) Upgrade(ctx context.Context, uid, versionUID string, clusterUIDs []string, values bo.ApplicationDeployValues) (*bo.ApplicationDeployment, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseWrite)
	if err != nil {
		return nil, err
	}
	deployment, err := deployer.releases.GetDeployment(ctx, ns, uid)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	version, err := deployer.apps.GetVersion(ctx, ns, deployment.ApplicationUID, versionUID)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	if version.Status != "PUBLISHED" {
		return nil, merr.Conflict("only a published version can be upgraded to")
	}
	if err := validateDeployValues(values, version); err != nil {
		return nil, err
	}
	return deployer.createOperations(ctx, deployment, version, clusterUIDs, values, "UPGRADE")
}

func (deployer *Deployer) Rollback(ctx context.Context, uid, versionUID string, clusterUIDs []string) (*bo.ApplicationDeployment, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseWrite)
	if err != nil {
		return nil, err
	}
	deployment, err := deployer.releases.GetDeployment(ctx, ns, uid)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	version, err := deployer.apps.GetVersion(ctx, ns, deployment.ApplicationUID, versionUID)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	// 回滚目标只允许已发布版本，或当前集群实例历史上运行过的已下架版本。
	if version.Status != "PUBLISHED" {
		if version.Status != "DEPRECATED" {
			return nil, merr.Conflict("only a published or previously deployed version can be rolled back to")
		}
		for _, release := range selectReleases(deployment.Releases, clusterUIDs) {
			used, err := deployer.releases.ReleaseUsedVersion(ctx, ns, release.UID, version.UID)
			if err != nil {
				return nil, deployer.access.Wrap(err)
			}
			if !used {
				return nil, merr.Conflict("deprecated version is not in this release history")
			}
		}
	}
	return deployer.createOperations(ctx, deployment, version, clusterUIDs, deployment.Values, "ROLLBACK")
}

func (deployer *Deployer) Redeploy(ctx context.Context, uid string, clusterUIDs []string) (*bo.ApplicationDeployment, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseWrite)
	if err != nil {
		return nil, err
	}
	deployment, err := deployer.releases.GetDeployment(ctx, ns, uid)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	version, err := deployer.apps.GetVersion(ctx, ns, deployment.ApplicationUID, deployment.VersionUID)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	if version.Status != "PUBLISHED" && version.Status != "DEPRECATED" {
		return nil, merr.Conflict("draft version cannot be redeployed")
	}
	return deployer.createOperations(ctx, deployment, version, clusterUIDs, deployment.Values, "REDEPLOY")
}

func (deployer *Deployer) Uninstall(ctx context.Context, uid string, clusterUIDs []string, keepPVC, deleteClusterScoped bool) (*bo.ApplicationDeployment, error) {
	ns, err := deployer.access.Allow(ctx, perm.ReleaseDelete)
	if err != nil {
		return nil, err
	}
	if deleteClusterScoped {
		if err := deployer.access.Require(ctx, perm.ReleaseCluster); err != nil {
			return nil, merr.Forbidden("deleting cluster-scoped resources requires release:cluster")
		}
	}
	deployment, err := deployer.releases.GetDeployment(ctx, ns, uid)
	if err != nil {
		return nil, deployer.access.Wrap(err)
	}
	selected := selectReleases(deployment.Releases, clusterUIDs)
	if len(selected) == 0 {
		return nil, merr.Invalid("no active cluster release selected")
	}
	batchUID := id.New()
	operator, _ := actor.UserID(ctx)
	for _, release := range selected {
		active, err := deployer.releases.HasActiveOperation(ctx, release.UID)
		if err != nil {
			return nil, deployer.access.Wrap(err)
		}
		if active {
			return nil, merr.Conflict("a cluster release already has an active operation")
		}
		if _, err := deployer.releases.CreateOperation(ctx, bo.ApplicationOperation{
			NamespaceUID: ns, ApplicationUID: deployment.ApplicationUID, DeploymentUID: deployment.UID,
			ReleaseUID: release.UID, ClusterUID: release.ClusterUID,
			BatchUID: batchUID, ClusterName: release.ClusterName, Action: "UNINSTALL",
			Status: "PENDING", FromVersionUID: release.VersionUID, ToVersionUID: release.VersionUID,
			Values: release.Values, OperatorUID: operator, OperatorName: operator,
			KeepPVC: keepPVC, DeleteClusterScoped: deleteClusterScoped,
		}); err != nil {
			return nil, deployer.access.Wrap(err)
		}
	}
	deployment.Status = "UNINSTALLING"
	if err := deployer.releases.UpdateDeployment(ctx, *deployment); err != nil {
		return nil, deployer.access.Wrap(err)
	}
	return deployer.Get(ctx, uid)
}

func (deployer *Deployer) createOperations(ctx context.Context, deployment *bo.ApplicationDeployment, version *bo.ApplicationVersion, clusterUIDs []string, values bo.ApplicationDeployValues, action string) (*bo.ApplicationDeployment, error) {
	selected := selectReleases(deployment.Releases, clusterUIDs)
	if len(selected) == 0 {
		return nil, merr.Invalid("no active cluster release selected")
	}
	clusters := make([]bo.DeploymentCluster, 0, len(selected))
	for _, release := range selected {
		clusters = append(clusters, bo.DeploymentCluster{UID: release.ClusterUID, Name: release.ClusterName})
	}
	if err := deployer.preflight(ctx, deployment.NamespaceUID, clusters, deployment.TargetNamespace, version); err != nil {
		return nil, err
	}
	batchUID := id.New()
	operator, _ := actor.UserID(ctx)
	for _, release := range selected {
		active, err := deployer.releases.HasActiveOperation(ctx, release.UID)
		if err != nil {
			return nil, deployer.access.Wrap(err)
		}
		if active {
			return nil, merr.Conflict("a cluster release already has an active operation")
		}
		if _, err := deployer.releases.CreateOperation(ctx, bo.ApplicationOperation{
			NamespaceUID: deployment.NamespaceUID, ApplicationUID: deployment.ApplicationUID,
			DeploymentUID: deployment.UID, ReleaseUID: release.UID, ClusterUID: release.ClusterUID,
			BatchUID: batchUID, ClusterName: release.ClusterName, Action: action,
			Status: "PENDING", FromVersionUID: release.VersionUID, ToVersionUID: version.UID,
			Values: values, OperatorUID: operator, OperatorName: operator, KeepPVC: true,
		}); err != nil {
			return nil, deployer.access.Wrap(err)
		}
	}
	deployment.VersionUID = version.UID
	deployment.Version = version.Version
	deployment.Status = "DEPLOYING"
	if err := deployer.releases.UpdateDeployment(ctx, *deployment); err != nil {
		return nil, deployer.access.Wrap(err)
	}
	return deployer.Get(ctx, deployment.UID)
}

func (deployer *Deployer) preflight(ctx context.Context, namespaceUID string, clusters []bo.DeploymentCluster, targetNamespace string, version *bo.ApplicationVersion) error {
	clusterScoped := false
	for _, cluster := range clusters {
		refs, err := deployer.applier.Preflight(ctx, cluster.UID, targetNamespace, version.Resources)
		if err != nil {
			return err
		}
		for _, ref := range refs {
			if ref.Namespace == "" {
				clusterScoped = true
			}
		}
		for _, resource := range version.Resources {
			if resource.ClusterScoped {
				clusterScoped = true
			}
		}
	}
	if clusterScoped {
		if err := deployer.access.Require(ctx, perm.ReleaseCluster); err != nil {
			return merr.Forbidden("cluster-scoped resources require release:cluster")
		}
	}
	for _, image := range version.Images {
		if image.External || image.RegistryUID == "" || image.Tag == "" {
			continue
		}
		registry, err := deployer.registries.GetRegistry(ctx, namespaceUID, image.RegistryUID)
		if err != nil {
			return deployer.access.Wrap(err)
		}
		exists, digest, err := deployer.images.TagExists(ctx, *registry, image.Repository, image.Tag)
		if err != nil {
			return deployer.access.Wrap(err)
		}
		if !exists {
			return merr.Conflict("image referenced by this version is no longer available")
		}
		if image.Digest != "" && digest != "" && image.Digest != digest {
			return merr.Conflict("image digest changed since the version was published")
		}
	}
	return nil
}

func validateDeployValues(values bo.ApplicationDeployValues, version *bo.ApplicationVersion) error {
	workloads := map[string]string{}
	for _, resource := range version.Resources {
		if resource.Kind == "Deployment" || resource.Kind == "StatefulSet" {
			workloads[resource.Item] = resource.Kind
		}
	}
	for item, replicas := range values.Replicas {
		if replicas < 0 || replicas > 100 {
			return merr.Invalid("replicas must be between 0 and 100")
		}
		if _, ok := workloads[item]; !ok {
			return merr.Invalid("replicas can only be overridden for Deployment or StatefulSet")
		}
	}
	return nil
}

func selectReleases(releases []bo.ApplicationRelease, clusterUIDs []string) []bo.ApplicationRelease {
	if len(clusterUIDs) == 0 {
		out := make([]bo.ApplicationRelease, 0, len(releases))
		for _, release := range releases {
			if release.Status != "UNINSTALLED" {
				out = append(out, release)
			}
		}
		return out
	}
	wanted := map[string]struct{}{}
	for _, uid := range clusterUIDs {
		wanted[uid] = struct{}{}
	}
	out := make([]bo.ApplicationRelease, 0, len(releases))
	for _, release := range releases {
		if _, ok := wanted[release.ClusterUID]; ok && release.Status != "UNINSTALLED" {
			out = append(out, release)
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
