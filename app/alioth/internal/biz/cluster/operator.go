package cluster

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	workloadDeployment  = "Deployment"
	workloadStatefulSet = "StatefulSet"
	workloadDaemonSet   = "DaemonSet"
	// maxWorkloadImageLength 与 proto 校验一致，避免绕过 HTTP 的调用写入超长镜像。
	maxWorkloadImageLength = 1024
)

// Operator 修改集群资源。每个方法都先确认调用方有 cluster:write，
// 再确认集群属于当前租户，最后才访问集群 API Server。
type Operator struct {
	clusters  repository.ClusterRepo
	resources repository.KubernetesRepo
	access    access.Checker
}

func NewOperator(clusters repository.ClusterRepo, resources repository.KubernetesRepo, checker access.Checker) *Operator {
	return &Operator{clusters: clusters, resources: resources, access: checker}
}

// SetNodeUnschedulable 设置节点是否可调度：true 为 cordon，false 为 uncordon。
func (operator *Operator) SetNodeUnschedulable(ctx context.Context, clusterUID, name string, unschedulable bool) error {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return merr.Invalid("node name is required")
	}
	return operator.access.Wrap(operator.resources.SetNodeUnschedulable(ctx, clusterUID, name, unschedulable))
}

// ScaleWorkload 调整副本数。DaemonSet 的副本数由节点数决定，不能手动改。
func (operator *Operator) ScaleWorkload(ctx context.Context, clusterUID, namespace, kind, name string, replicas int32) (*bo.KubeWorkload, error) {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, kind, name = strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name)
	if namespace == "" || name == "" || (kind != workloadDeployment && kind != workloadStatefulSet) {
		return nil, merr.Invalid("invalid request")
	}
	item, err := operator.resources.ScaleWorkload(ctx, clusterUID, namespace, kind, name, replicas)
	return item, operator.access.Wrap(err)
}

// RestartWorkload 给 Pod 模板打重启时间注解，三种工作负载都支持。
func (operator *Operator) RestartWorkload(ctx context.Context, clusterUID, namespace, kind, name string) (*bo.KubeWorkload, error) {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, kind, name = strings.TrimSpace(namespace), strings.TrimSpace(kind), strings.TrimSpace(name)
	if namespace == "" || name == "" || !validWorkloadKind(kind) {
		return nil, merr.Invalid("invalid request")
	}
	item, err := operator.resources.RestartWorkload(ctx, clusterUID, namespace, kind, name)
	return item, operator.access.Wrap(err)
}

// DeletePod 删除一个 Pod。由工作负载管理的 Pod 会被控制器重新创建。
func (operator *Operator) DeletePod(ctx context.Context, clusterUID, namespace, name string) error {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return err
	}
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return merr.Invalid("invalid request")
	}
	return operator.access.Wrap(operator.resources.DeletePod(ctx, clusterUID, namespace, name))
}

// SetWorkloadImage 只改目标容器的镜像。容器必须存在，init 决定查找普通容器还是初始化容器。
func (operator *Operator) SetWorkloadImage(ctx context.Context, clusterUID, namespace, kind, name, container string, init bool, image string) (*bo.KubeWorkload, error) {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, kind, name, err := normalizeWorkloadTarget(namespace, kind, name)
	if err != nil {
		return nil, err
	}
	container, image = strings.TrimSpace(container), strings.TrimSpace(image)
	if container == "" {
		return nil, merr.Invalid("container is required")
	}
	if image == "" {
		return nil, merr.Invalid("image is required")
	}
	if utf8.RuneCountInString(image) > maxWorkloadImageLength {
		return nil, merr.Invalid("image is too long")
	}
	item, err := operator.resources.SetWorkloadImage(ctx, clusterUID, namespace, kind, name, container, init, image)
	return item, operator.access.Wrap(err)
}

// RollbackWorkload 把工作负载回滚到指定历史版本，revision 从 1 开始。
func (operator *Operator) RollbackWorkload(ctx context.Context, clusterUID, namespace, kind, name string, revision int32) (*bo.KubeWorkload, error) {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, kind, name, err := normalizeWorkloadTarget(namespace, kind, name)
	if err != nil {
		return nil, err
	}
	if revision < 1 {
		return nil, merr.Invalid("revision is required")
	}
	item, err := operator.resources.RollbackWorkload(ctx, clusterUID, namespace, kind, name, revision)
	return item, operator.access.Wrap(err)
}

// DeleteJob 删除 Job，它的 Pod 由默认的后台级联删除清理。
func (operator *Operator) DeleteJob(ctx context.Context, clusterUID, namespace, name string) error {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return err
	}
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return merr.Invalid("invalid request")
	}
	return operator.access.Wrap(operator.resources.DeleteJob(ctx, clusterUID, namespace, name))
}

// SetCronJobSuspend 暂停或恢复 CronJob 的调度。
func (operator *Operator) SetCronJobSuspend(ctx context.Context, clusterUID, namespace, name string, suspend bool) (*bo.KubeCronJob, error) {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("invalid request")
	}
	item, err := operator.resources.SetCronJobSuspend(ctx, clusterUID, namespace, name, suspend)
	return item, operator.access.Wrap(err)
}

// TriggerCronJob 用 jobTemplate 立即创建一个独立 Job，返回新建的 Job。
func (operator *Operator) TriggerCronJob(ctx context.Context, clusterUID, namespace, name string) (*bo.KubeJob, error) {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return nil, err
	}
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("invalid request")
	}
	item, err := operator.resources.TriggerCronJob(ctx, clusterUID, namespace, name)
	return item, operator.access.Wrap(err)
}

// DeleteCronJob 删除 CronJob 和它创建的 Job。
func (operator *Operator) DeleteCronJob(ctx context.Context, clusterUID, namespace, name string) error {
	if _, err := operator.scope(ctx, clusterUID); err != nil {
		return err
	}
	namespace, name = strings.TrimSpace(namespace), strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return merr.Invalid("invalid request")
	}
	return operator.access.Wrap(operator.resources.DeleteCronJob(ctx, clusterUID, namespace, name))
}

func (operator *Operator) scope(ctx context.Context, clusterUID string) (string, error) {
	ns, err := operator.access.Allow(ctx, perm.ClusterWrite)
	if err != nil {
		return "", err
	}
	if _, err := operator.clusters.GetCluster(ctx, ns, clusterUID); err != nil {
		return "", operator.access.Wrap(err)
	}
	return ns, nil
}

func validWorkloadKind(kind string) bool {
	switch kind {
	case workloadDeployment, workloadStatefulSet, workloadDaemonSet:
		return true
	default:
		return false
	}
}
