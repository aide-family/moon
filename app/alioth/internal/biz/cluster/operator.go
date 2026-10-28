package cluster

import (
	"context"
	"strings"

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
