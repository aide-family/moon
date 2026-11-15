package application

import (
	"context"
	"log/slog"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
)

const operationLease = 10 * time.Minute

// Executor 认领一条部署操作并在目标集群执行。
type Executor struct {
	apps      repository.ApplicationRepo
	releases  repository.ApplicationReleaseRepo
	applier   repository.ApplicationApplier
	inspector *Inspector
	log       *slog.Logger
	owner     string
}

func NewExecutor(apps repository.ApplicationRepo, releases repository.ApplicationReleaseRepo, applier repository.ApplicationApplier, inspector *Inspector, logger *slog.Logger) *Executor {
	return &Executor{apps: apps, releases: releases, applier: applier, inspector: inspector, log: logger, owner: id.New()}
}

func (executor *Executor) ExecuteNext(ctx context.Context) error {
	now := time.Now().UTC()
	operation, err := executor.releases.ClaimOperation(ctx, executor.owner, now, now.Add(operationLease))
	if err != nil || operation == nil {
		return err
	}
	executor.execute(ctx, *operation)
	return nil
}

func (executor *Executor) execute(ctx context.Context, operation bo.ApplicationOperation) {
	deployment, err := executor.releases.GetDeployment(ctx, operation.NamespaceUID, operation.DeploymentUID)
	if err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	release, err := executor.releases.GetRelease(ctx, operation.NamespaceUID, operation.ReleaseUID)
	if err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	if operation.Action == "UNINSTALL" {
		executor.uninstall(ctx, operation, deployment, release)
		return
	}
	version, err := executor.apps.GetVersion(ctx, operation.NamespaceUID, deployment.ApplicationUID, operation.ToVersionUID)
	if err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	secrets, _ := executor.apps.GetVersionSecrets(ctx, operation.NamespaceUID, version.UID)
	resources := applyReplicaOverrides(version.Resources, operation.Values)
	refs, err := executor.applier.Apply(ctx, release.ClusterUID, release.TargetNamespace, deployment.UID, release.UID, release.Name, resources, secrets)
	if err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	previous, _ := executor.releases.LastSuccessfulOperation(ctx, release.UID)
	if previous != nil {
		executor.prune(ctx, release, previous, refs)
	}
	now := time.Now().UTC()
	release.Status = "RUNNING"
	release.VersionUID = version.UID
	release.Version = version.Version
	release.Images = version.Images
	release.Message = ""
	release.LastOperationUID = operation.UID
	release.DeployedAt = &now
	if err := executor.releases.UpdateRelease(ctx, *release); err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	operation.Status = "SUCCEEDED"
	operation.Resources = refs
	operation.Diff = diffRefs(previous, refs)
	operation.FinishedAt = &now
	operation.Message = ""
	if err := executor.releases.UpdateOperation(ctx, operation); err != nil {
		executor.log.Error("update application operation: " + err.Error())
	}
	executor.inspector.Recompute(ctx, operation.NamespaceUID, operation.DeploymentUID)
}

func (executor *Executor) uninstall(ctx context.Context, operation bo.ApplicationOperation, deployment *bo.ApplicationDeployment, release *bo.ApplicationRelease) {
	previous, err := executor.releases.LastSuccessfulOperation(ctx, release.UID)
	if err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	if previous != nil {
		if err := executor.applier.Delete(ctx, release.ClusterUID, release.UID, previous.Resources, operation.KeepPVC, operation.DeleteClusterScoped); err != nil {
			executor.fail(ctx, operation, err)
			return
		}
	}
	now := time.Now().UTC()
	release.Status = "UNINSTALLED"
	release.Message = ""
	release.LastOperationUID = operation.UID
	if err := executor.releases.UpdateRelease(ctx, *release); err != nil {
		executor.fail(ctx, operation, err)
		return
	}
	operation.Status = "SUCCEEDED"
	operation.FinishedAt = &now
	operation.Message = ""
	if previous != nil {
		operation.Resources = previous.Resources
	}
	if err := executor.releases.UpdateOperation(ctx, operation); err != nil {
		executor.log.Error("update application operation: " + err.Error())
	}
	executor.inspector.Recompute(ctx, operation.NamespaceUID, deployment.UID)
}

func (executor *Executor) prune(ctx context.Context, release *bo.ApplicationRelease, previous *bo.ApplicationOperation, current []bo.ResourceRef) {
	keep := map[string]struct{}{}
	for _, ref := range current {
		keep[refKey(ref)] = struct{}{}
	}
	stale := make([]bo.ResourceRef, 0)
	for _, ref := range previous.Resources {
		if _, ok := keep[refKey(ref)]; !ok {
			stale = append(stale, ref)
		}
	}
	if len(stale) == 0 {
		return
	}
	if err := executor.applier.Delete(ctx, release.ClusterUID, release.UID, stale, true, false); err != nil {
		executor.log.Error("prune application resources: " + err.Error())
	}
}

func (executor *Executor) fail(ctx context.Context, operation bo.ApplicationOperation, err error) {
	now := time.Now().UTC()
	operation.Status = "FAILED"
	operation.Message = err.Error()
	operation.FinishedAt = &now
	_ = executor.releases.UpdateOperation(ctx, operation)
	if release, getErr := executor.releases.GetRelease(ctx, operation.NamespaceUID, operation.ReleaseUID); getErr == nil {
		release.Status = "FAILED"
		release.Message = err.Error()
		_ = executor.releases.UpdateRelease(ctx, *release)
	}
	executor.inspector.Recompute(ctx, operation.NamespaceUID, operation.DeploymentUID)
}

// applyReplicaOverrides 只覆盖标准工作负载的 spec.replicas；其它字段不开放覆盖。
func applyReplicaOverrides(resources []bo.ApplicationResource, values bo.ApplicationDeployValues) []bo.ApplicationResource {
	if len(values.Replicas) == 0 {
		return resources
	}
	out := make([]bo.ApplicationResource, 0, len(resources))
	for _, resource := range resources {
		replicas, ok := values.Replicas[resource.Item]
		if !ok || (resource.Kind != "Deployment" && resource.Kind != "StatefulSet") {
			out = append(out, resource)
			continue
		}
		var manifest map[string]any
		if err := yaml.Unmarshal([]byte(resource.Manifest), &manifest); err != nil {
			out = append(out, resource)
			continue
		}
		spec, _ := manifest["spec"].(map[string]any)
		if spec == nil {
			spec = map[string]any{}
			manifest["spec"] = spec
		}
		spec["replicas"] = replicas
		raw, err := yaml.Marshal(manifest)
		if err != nil {
			out = append(out, resource)
			continue
		}
		resource.Manifest = string(raw)
		out = append(out, resource)
	}
	return out
}

func refKey(ref bo.ResourceRef) string {
	return ref.APIVersion + "|" + ref.Kind + "|" + ref.Namespace + "|" + ref.Name
}

func diffRefs(previous *bo.ApplicationOperation, current []bo.ResourceRef) map[string]int32 {
	old := map[string]struct{}{}
	if previous != nil {
		for _, ref := range previous.Resources {
			old[refKey(ref)] = struct{}{}
		}
	}
	added, updated := 0, 0
	for _, ref := range current {
		if _, ok := old[refKey(ref)]; ok {
			updated++
		} else {
			added++
		}
	}
	deleted := 0
	if previous != nil {
		for _, ref := range previous.Resources {
			found := false
			for _, item := range current {
				if refKey(item) == refKey(ref) {
					found = true
					break
				}
			}
			if !found {
				deleted++
			}
		}
	}
	return map[string]int32{"added": int32(added), "updated": int32(updated), "deleted": int32(deleted)}
}
