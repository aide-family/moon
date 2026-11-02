package cluster

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewRegistry, NewExplorer, NewInspector, NewOperator, NewTerminal)

// maxClusterName 与 proto 校验保持一致，避免绕过 HTTP 的调用写入超长名称。
const maxClusterName = 64

// Registry 管理租户接入的集群。kubeconfig 只进不出：
// 接入和替换时提交，读取接口只返回元数据。绑定指标数据源时还要读一次数据源，
// 确认它属于同一租户、未删除且是指标类驱动。
type Registry struct {
	clusters    repository.ClusterRepo
	datasources repository.DatasourceRepo
	access      access.Checker
}

func NewRegistry(clusters repository.ClusterRepo, datasources repository.DatasourceRepo, checker access.Checker) *Registry {
	return &Registry{clusters: clusters, datasources: datasources, access: checker}
}

// Create 用 kubeconfig 接入集群。仓储会先探测再落库，探测失败不会留下记录。
// metricsDatasourceUID 为 nil 表示不绑定指标数据源。
func (registry *Registry) Create(ctx context.Context, name, kubeconfig, remark string, metricsDatasourceUID *string) (*bo.Cluster, error) {
	ns, err := registry.access.Allow(ctx, perm.ClusterWrite)
	if err != nil {
		return nil, err
	}
	in, err := clusterInput(ns, "", name, remark)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(kubeconfig) == "" {
		return nil, merr.Invalid("kubeconfig is required")
	}
	binding, err := registry.bindMetricsDatasource(ctx, ns, metricsDatasourceUID)
	if err != nil {
		return nil, err
	}
	// 创建时字段缺失等同于不绑定，与 proto3 optional 的缺省语义一致。
	createdBinding := ""
	if binding != nil {
		createdBinding = *binding
	}
	item, err := registry.clusters.CreateCluster(ctx, in, kubeconfig, createdBinding)
	return item, registry.access.Wrap(err)
}

// Update 改名称和备注；kubeconfig 非空时先探测再替换，空表示保持原凭据。
// metricsDatasourceUID 为 nil 表示保持原绑定，指向空串表示解绑。
func (registry *Registry) Update(ctx context.Context, uid, name, kubeconfig, remark string, metricsDatasourceUID *string) (*bo.Cluster, error) {
	ns, err := registry.access.Allow(ctx, perm.ClusterWrite)
	if err != nil {
		return nil, err
	}
	in, err := clusterInput(ns, uid, name, remark)
	if err != nil {
		return nil, err
	}
	binding, err := registry.bindMetricsDatasource(ctx, ns, metricsDatasourceUID)
	if err != nil {
		return nil, err
	}
	item, err := registry.clusters.UpdateCluster(ctx, in, kubeconfig, binding)
	return item, registry.access.Wrap(err)
}

// bindMetricsDatasource 校验要绑定的指标数据源，返回写进集群的 UID。
// nil 表示保持原值，返回的指针原样透传给仓储；空串表示解绑，不需要读数据源。
// 绑定时要求 cluster:write（调用方已经校验）和 datasource:read。
func (registry *Registry) bindMetricsDatasource(ctx context.Context, namespaceUID string, uid *string) (*string, error) {
	if uid == nil {
		return nil, nil
	}
	binding := strings.TrimSpace(*uid)
	if binding == "" {
		return &binding, nil
	}
	if _, err := registry.access.Allow(ctx, perm.DatasourceRead); err != nil {
		return nil, err
	}
	item, err := registry.datasources.GetDatasource(ctx, namespaceUID, binding)
	if err != nil {
		return nil, registry.access.Wrap(err)
	}
	if item.Type != datasourceTypeMetrics || !isMetricsDriver(item.Driver) {
		return nil, merr.Invalid("metrics datasource must be Prometheus or VictoriaMetrics")
	}
	// 数据源状态不在绑定处拦截：临时停用的数据源仍允许保存绑定，
	// 汇总按失败状态返回，避免停用期间无法更新集群配置。
	return &binding, nil
}

// Delete 软删除集群，同时丢弃缓存的客户端，避免继续用已删除的凭据。
func (registry *Registry) Delete(ctx context.Context, uid string) error {
	ns, err := registry.access.Allow(ctx, perm.ClusterDelete)
	if err != nil {
		return err
	}
	return registry.access.Wrap(registry.clusters.DeleteCluster(ctx, ns, uid))
}

func (registry *Registry) Get(ctx context.Context, uid string) (*bo.Cluster, error) {
	ns, err := registry.access.Allow(ctx, perm.ClusterRead)
	if err != nil {
		return nil, err
	}
	item, err := registry.clusters.GetCluster(ctx, ns, uid)
	return item, registry.access.Wrap(err)
}

func (registry *Registry) List(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Cluster, int, int, int, error) {
	ns, err := registry.access.Allow(ctx, perm.ClusterRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := registry.clusters.ListClusters(ctx, ns, strings.TrimSpace(name), p, s)
	return items, total, p, s, registry.access.Wrap(err)
}

// Test 用已保存的凭据探测连通性，并刷新最近检测时间、最近错误和版本元数据。
func (registry *Registry) Test(ctx context.Context, uid string) (*bo.ClusterProbe, *bo.Cluster, error) {
	ns, err := registry.access.Allow(ctx, perm.ClusterRead)
	if err != nil {
		return nil, nil, err
	}
	probe, item, err := registry.clusters.TestCluster(ctx, ns, uid)
	return probe, item, registry.access.Wrap(err)
}

// Probe 供接入表单在保存前校验 kubeconfig，不创建集群。
func (registry *Registry) Probe(ctx context.Context, kubeconfig string) (*bo.ClusterProbe, error) {
	if _, err := registry.access.Allow(ctx, perm.ClusterWrite); err != nil {
		return nil, err
	}
	probe, err := registry.clusters.ProbeKubeconfig(ctx, kubeconfig)
	return probe, registry.access.Wrap(err)
}

// clusterInput 清理名称和备注。名称在同一租户内唯一，仓储用数据库唯一索引兜底冲突。
func clusterInput(namespaceUID, uid, name, remark string) (bo.Cluster, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return bo.Cluster{}, merr.Invalid("name is required")
	}
	if utf8.RuneCountInString(name) > maxClusterName {
		return bo.Cluster{}, merr.Invalid("name is too long")
	}
	return bo.Cluster{
		UID:          strings.TrimSpace(uid),
		NamespaceUID: namespaceUID,
		Name:         name,
		Remark:       strings.TrimSpace(remark),
	}, nil
}
