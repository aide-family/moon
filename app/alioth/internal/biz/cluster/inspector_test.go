package cluster

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// fixClusterRepo 返回固定的集群元数据，供 Inspector 和 Registry 的测试使用。
type fixClusterRepo struct {
	repository.ClusterRepo
	cluster *bo.Cluster
	err     error
	// 记录写入参数，验证绑定字段怎么传给仓储。
	createdBinding string
	updatedBinding *string
}

func (r *fixClusterRepo) GetCluster(context.Context, string, string) (*bo.Cluster, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.cluster, nil
}

func (r *fixClusterRepo) CreateCluster(_ context.Context, in bo.Cluster, _ string, metricsDatasourceUID string) (*bo.Cluster, error) {
	r.createdBinding = metricsDatasourceUID
	item := in
	return &item, nil
}

func (r *fixClusterRepo) UpdateCluster(_ context.Context, in bo.Cluster, _ string, metricsDatasourceUID *string) (*bo.Cluster, error) {
	r.updatedBinding = metricsDatasourceUID
	item := in
	return &item, nil
}

// fixKubeResources 只实现 Inspector 用到的只读方法，其余方法由 nil 接口兜住。
type fixKubeResources struct {
	repository.KubernetesRepo
	nodes      []bo.KubeNode
	workloads  []bo.KubeWorkload
	pods       []bo.KubePod
	jobs       []bo.KubeJob
	services   []bo.KubeService
	ingresses  []bo.KubeIngress
	version    string
	versionErr error
	nodeDetail *bo.KubeNodeDetail
	podDetail  *bo.KubePodDetail
	events     []bo.KubeEvent
	eventsErr  error
	// failures 按区块名注入查询失败。
	failures map[string]error
	// namespace 记录最近一次带命名空间的查询入参。概览的区块是并发读取的，
	// 用互斥锁保护这个记录字段，避免测试自身产生数据竞争。
	mu        sync.Mutex
	namespace string
}

// recordNamespace 记录带命名空间的查询入参。
func (r *fixKubeResources) recordNamespace(namespace string) {
	r.mu.Lock()
	r.namespace = namespace
	r.mu.Unlock()
}

// lastNamespace 返回最近一次带命名空间的查询入参。
func (r *fixKubeResources) lastNamespace() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.namespace
}

func (r *fixKubeResources) ListNodes(context.Context, string, string) ([]bo.KubeNode, error) {
	return r.nodes, r.failures["nodes"]
}

func (r *fixKubeResources) ListWorkloads(_ context.Context, _, namespace, _, _ string) ([]bo.KubeWorkload, error) {
	r.recordNamespace(namespace)
	return r.workloads, r.failures["workloads"]
}

func (r *fixKubeResources) ListPods(_ context.Context, _, namespace, _, _, _ string) ([]bo.KubePod, error) {
	r.recordNamespace(namespace)
	return r.pods, r.failures["pods"]
}

func (r *fixKubeResources) ListJobs(_ context.Context, _, namespace, _ string) ([]bo.KubeJob, error) {
	r.recordNamespace(namespace)
	return r.jobs, r.failures["jobs"]
}

func (r *fixKubeResources) ListServices(_ context.Context, _, namespace, _ string) ([]bo.KubeService, error) {
	r.recordNamespace(namespace)
	return r.services, r.failures["services"]
}

func (r *fixKubeResources) ListIngresses(_ context.Context, _, namespace, _ string) ([]bo.KubeIngress, error) {
	r.recordNamespace(namespace)
	return r.ingresses, r.failures["ingresses"]
}

func (r *fixKubeResources) ServerVersion(context.Context, string) (string, error) {
	return r.version, r.versionErr
}

func (r *fixKubeResources) NodeDetail(context.Context, string, string) (*bo.KubeNodeDetail, error) {
	return r.nodeDetail, nil
}

func (r *fixKubeResources) PodDetail(context.Context, string, string, string) (*bo.KubePodDetail, error) {
	return r.podDetail, nil
}

func (r *fixKubeResources) ListEvents(context.Context, string, string, string, string, string, int32) ([]bo.KubeEvent, error) {
	return r.events, r.eventsErr
}

// fixDatasourceRepo 是绑定与指标汇总用的数据源仓储。
type fixDatasourceRepo struct {
	repository.DatasourceRepo
	item   *bo.Datasource
	getErr error
	values map[string]float64
	found  map[string]bool
	errs   map[string]error
}

func (r *fixDatasourceRepo) GetDatasource(context.Context, string, string) (*bo.Datasource, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.item, nil
}

func (r *fixDatasourceRepo) InstantValue(_ context.Context, _ bo.Datasource, expr string) (float64, bool, error) {
	key := metricKeyOfQuery(expr)
	if err := r.errs[key]; err != nil {
		return 0, false, err
	}
	value, found := r.values[key]
	return value, found, nil
}

// metricKeyOfQuery 把固定 PromQL 映射回指标 key，方便测试里按 key 注入结果。
func metricKeyOfQuery(expr string) string {
	for _, item := range metricsSummaryQueries {
		if item.query == expr {
			return item.key
		}
	}
	return ""
}

// fixNamespaceRepo 让测试操作者成为租户负责人：负责人隐式持有全部租户权限。
type fixNamespaceRepo struct{ repository.NamespaceRepo }

func (fixNamespaceRepo) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	return &bo.Namespace{UID: uid, Name: uid, OwnerUID: "fix-owner", Status: "ACTIVE"}, nil
}

// fixFeatureRepo 返回完整的模块与能力集合，避免测试依赖套餐数据。
type fixFeatureRepo struct{ repository.FeatureRepo }

func (fixFeatureRepo) EffectiveModules(context.Context, string) ([]string, error) {
	return perm.FeatureCodes(), nil
}

func (fixFeatureRepo) EffectiveCapabilities(context.Context, string) ([]string, error) {
	items := perm.Capabilities()
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	return codes, nil
}

// fixUserRepo 让非负责人操作者被判定为普通用户，不会走平台只读通道。
type fixUserRepo struct{ repository.UserRepo }

func (fixUserRepo) UserByUID(context.Context, string) (*bo.Account, error) {
	return &bo.Account{UID: "fix-user", Status: "ACTIVE", SystemRole: "USER"}, nil
}

// fixRoleRepo 返回固定的租户权限集合。
type fixRoleRepo struct {
	repository.RoleRepo
	perms map[string]struct{}
}

func (r fixRoleRepo) Permissions(context.Context, string, string) (map[string]struct{}, error) {
	return r.perms, nil
}

// ownerChecker 让操作者成为租户负责人，持有全部租户权限。
func ownerChecker() access.Checker {
	return access.Checker{
		Namespaces: fixNamespaceRepo{},
		Features:   fixFeatureRepo{},
		Log:        slog.Default(),
	}
}

// memberChecker 让操作者只是普通成员，并持有给定的权限集合。
func memberChecker(codes ...string) access.Checker {
	perms := make(map[string]struct{}, len(codes)*2)
	for _, code := range codes {
		perms[code] = struct{}{}
		for _, menu := range perm.MenusOfPermission(code) {
			perms[perm.MenuPermission(menu)] = struct{}{}
		}
	}
	return access.Checker{
		Namespaces: fixNamespaceRepo{},
		Features:   fixFeatureRepo{},
		Users:      fixUserRepo{},
		Roles:      fixRoleRepo{perms: perms},
		Log:        slog.Default(),
	}
}

// ownerContext 提供已认证的租户负责人。
func ownerContext() context.Context {
	return actor.WithUser(actor.WithNamespace(context.Background(), "fix-ns"), "fix-owner")
}

// memberContext 提供已认证的普通成员。
func memberContext() context.Context {
	return actor.WithUser(actor.WithNamespace(context.Background(), "fix-ns"), "fix-user")
}

func fullOverviewResources() *fixKubeResources {
	return &fixKubeResources{
		nodes: []bo.KubeNode{
			{Name: "node-a", Status: "Ready", Schedulable: true, AllocatableCPUMillis: 4000, AllocatableMemoryBytes: 8 << 30},
			{Name: "node-b", Status: "Ready", Schedulable: false, AllocatableCPUMillis: 4000, AllocatableMemoryBytes: 8 << 30},
			{Name: "node-c", Status: "NotReady", Schedulable: true, AllocatableCPUMillis: 2000, AllocatableMemoryBytes: 4 << 30},
		},
		workloads: []bo.KubeWorkload{
			{Kind: "Deployment", DesiredReplicas: 3, ReadyReplicas: 2},
			{Kind: "Deployment", DesiredReplicas: 1, ReadyReplicas: 1},
			{Kind: "StatefulSet", DesiredReplicas: 3, ReadyReplicas: 3},
		},
		pods: []bo.KubePod{
			{Phase: "Running"}, {Phase: "Running"}, {Phase: "Pending"}, {Phase: "Succeeded"}, {Phase: "Failed"},
		},
		jobs:     []bo.KubeJob{{Active: 1, Failed: 0}, {Active: 0, Failed: 2}},
		services: []bo.KubeService{{}, {}},
		version:  "v1.31.2",
	}
}

func TestOverviewAggregatesSections(t *testing.T) {
	resources := fullOverviewResources()
	resources.ingresses = nil
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1", Version: "v1.30.0"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	overview, err := inspector.Overview(ownerContext(), "c1", "default")
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if overview.APIServerVersion != "v1.31.2" {
		t.Fatalf("version = %q", overview.APIServerVersion)
	}
	if overview.NodeTotal != 3 || overview.NodeReady != 2 || overview.NodeNotReady != 1 {
		t.Fatalf("node counts = %d/%d/%d", overview.NodeTotal, overview.NodeReady, overview.NodeNotReady)
	}
	// 只累加 Ready 且可调度节点：node-b 被 cordon，node-c 未 Ready。
	if overview.AllocatableCPUMillis != 4000 || overview.AllocatableMemoryBytes != 8<<30 {
		t.Fatalf("allocatable = %d/%d", overview.AllocatableCPUMillis, overview.AllocatableMemoryBytes)
	}
	if overview.PodTotal != 5 || len(overview.PodPhases) != 5 {
		t.Fatalf("pods = %d phases=%+v", overview.PodTotal, overview.PodPhases)
	}
	if overview.PodPhases[0].Phase != "Running" || overview.PodPhases[0].Count != 2 {
		t.Fatalf("running phase = %+v", overview.PodPhases[0])
	}
	if overview.PodPhases[4].Phase != "Unknown" || overview.PodPhases[4].Count != 0 {
		t.Fatalf("unknown phase = %+v", overview.PodPhases[4])
	}
	if len(overview.Workloads) != 3 || overview.Workloads[0].Kind != "Deployment" {
		t.Fatalf("workloads = %+v", overview.Workloads)
	}
	if overview.Workloads[0].Total != 2 || overview.Workloads[0].Desired != 4 || overview.Workloads[0].Ready != 3 {
		t.Fatalf("deployment summary = %+v", overview.Workloads[0])
	}
	if overview.Workloads[2].Kind != "DaemonSet" || overview.Workloads[2].Total != 0 {
		t.Fatalf("daemonset summary = %+v", overview.Workloads[2])
	}
	if overview.JobTotal != 2 || overview.JobActive != 1 || overview.JobFailed != 2 {
		t.Fatalf("jobs = %+v", overview)
	}
	if overview.ServiceTotal != 2 || overview.IngressTotal != 0 {
		t.Fatalf("services/ingress = %d/%d", overview.ServiceTotal, overview.IngressTotal)
	}
	if len(overview.DeniedSections) != 0 || len(overview.FailedSections) != 0 {
		t.Fatalf("sections = %+v %+v", overview.DeniedSections, overview.FailedSections)
	}
	if overview.GeneratedAt == 0 {
		t.Fatal("generated at is zero")
	}
	if resources.lastNamespace() != "default" {
		t.Fatalf("namespace = %q, want default", resources.lastNamespace())
	}
}

func TestOverviewVersionFallsBackToCluster(t *testing.T) {
	resources := fullOverviewResources()
	resources.versionErr = merr.Server("discovery failed")
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1", Version: "v1.29.4"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	overview, err := inspector.Overview(ownerContext(), "c1", "")
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if overview.APIServerVersion != "v1.29.4" {
		t.Fatalf("version = %q", overview.APIServerVersion)
	}
}

func TestOverviewDeniedSections(t *testing.T) {
	resources := fullOverviewResources()
	// 只有 cluster:read 和节点读权限：节点区块有数据，其余五个区块进 denied_sections。
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, resources, &fixDatasourceRepo{},
		memberChecker(perm.ClusterRead, perm.K8sNodeRead))
	overview, err := inspector.Overview(memberContext(), "c1", "")
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if overview.NodeTotal != 3 {
		t.Fatalf("node total = %d, want 3", overview.NodeTotal)
	}
	wantDenied := []string{overviewSectionWorkloads, overviewSectionPods, overviewSectionJobs, overviewSectionServices, overviewSectionIngress}
	if len(overview.DeniedSections) != len(wantDenied) {
		t.Fatalf("denied = %+v", overview.DeniedSections)
	}
	for index, name := range wantDenied {
		if overview.DeniedSections[index] != name {
			t.Fatalf("denied = %+v", overview.DeniedSections)
		}
	}
	if overview.PodTotal != 0 || overview.ServiceTotal != 0 {
		t.Fatalf("denied sections must stay zero: pods=%d services=%d", overview.PodTotal, overview.ServiceTotal)
	}
	if len(overview.FailedSections) != 0 {
		t.Fatalf("failed = %+v", overview.FailedSections)
	}
}

func TestOverviewFailedSections(t *testing.T) {
	resources := fullOverviewResources()
	resources.failures = map[string]error{"services": merr.Server("list services failed")}
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	overview, err := inspector.Overview(ownerContext(), "c1", "")
	if err != nil {
		t.Fatalf("Overview: %v", err)
	}
	if len(overview.FailedSections) != 1 || overview.FailedSections[0] != overviewSectionServices {
		t.Fatalf("failed = %+v", overview.FailedSections)
	}
	if overview.ServiceTotal != 0 || overview.PodTotal != 5 {
		t.Fatalf("overview = %+v", overview)
	}
}

func TestOverviewAllSectionsFailed(t *testing.T) {
	resources := fullOverviewResources()
	connectionErr := merr.Client("cluster is unreachable")
	resources.failures = map[string]error{
		"nodes": connectionErr, "workloads": connectionErr, "pods": connectionErr,
		"jobs": connectionErr, "services": connectionErr, "ingresses": connectionErr,
	}
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	overview, err := inspector.Overview(ownerContext(), "c1", "")
	if err == nil {
		t.Fatalf("want error, got overview %+v", overview)
	}
	if !strings.Contains(err.Error(), "unreachable") {
		t.Fatalf("err = %v", err)
	}
}

func TestNodeDetailWarnsWhenEventsFail(t *testing.T) {
	resources := &fixKubeResources{
		nodeDetail: &bo.KubeNodeDetail{
			Node:       bo.KubeNode{Name: "node-a", UID: "node-uid"},
			Conditions: []bo.KubeCondition{{Type: "Ready", Status: "True"}},
		},
		eventsErr: merr.Client("cluster credential is not allowed to perform this operation"),
	}
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	detail, err := inspector.NodeDetail(ownerContext(), "c1", "node-a")
	if err != nil {
		t.Fatalf("NodeDetail: %v", err)
	}
	if len(detail.Warnings) != 1 || detail.Warnings[0] != eventsUnavailableWarning {
		t.Fatalf("warnings = %+v", detail.Warnings)
	}
	if detail.Events != nil {
		t.Fatalf("events = %+v, want nil", detail.Events)
	}
	if len(detail.Conditions) != 1 {
		t.Fatalf("conditions = %+v", detail.Conditions)
	}
}

func TestNodeDetailReturnsEvents(t *testing.T) {
	resources := &fixKubeResources{
		nodeDetail: &bo.KubeNodeDetail{Node: bo.KubeNode{Name: "node-a", UID: "node-uid"}},
		events:     []bo.KubeEvent{{Reason: "NodeNotReady", Count: 2}},
	}
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	detail, err := inspector.NodeDetail(ownerContext(), "c1", "node-a")
	if err != nil {
		t.Fatalf("NodeDetail: %v", err)
	}
	if len(detail.Events) != 1 || detail.Events[0].Reason != "NodeNotReady" {
		t.Fatalf("events = %+v", detail.Events)
	}
	if len(detail.Warnings) != 0 {
		t.Fatalf("warnings = %+v", detail.Warnings)
	}
}

func TestPodDetailWarnsWhenEventsFail(t *testing.T) {
	resources := &fixKubeResources{
		podDetail: &bo.KubePodDetail{Pod: bo.KubePod{Name: "web-1", UID: "pod-uid"}},
		eventsErr: merr.Client("forbidden"),
	}
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, resources, &fixDatasourceRepo{}, ownerChecker())
	detail, err := inspector.PodDetail(ownerContext(), "c1", "default", "web-1")
	if err != nil {
		t.Fatalf("PodDetail: %v", err)
	}
	if len(detail.Warnings) != 1 {
		t.Fatalf("warnings = %+v", detail.Warnings)
	}
}

func TestPodDetailRequiresNamespace(t *testing.T) {
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, &fixKubeResources{}, &fixDatasourceRepo{}, ownerChecker())
	if _, err := inspector.PodDetail(ownerContext(), "c1", " ", "web-1"); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

// metricsInspector 组装绑定与指标汇总的测试对象。
func metricsInspector(binding string, sources *fixDatasourceRepo) *Inspector {
	cluster := &bo.Cluster{UID: "c1", MetricsDatasourceUID: binding}
	return NewInspector(&fixClusterRepo{cluster: cluster}, &fixKubeResources{}, sources, ownerChecker())
}

func TestMetricsSummaryUnbound(t *testing.T) {
	inspector := metricsInspector("", &fixDatasourceRepo{})
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusUnbound || summary.Message == "" {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestMetricsSummaryDatasourceGone(t *testing.T) {
	// 数据源被删除后绑定仍然保留，汇总返回 FAILED 而不是 UNBOUND。
	inspector := metricsInspector("ds-1", &fixDatasourceRepo{getErr: merr.NotFound("not found")})
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusFailed || summary.Message == "" {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestMetricsSummaryDisabled(t *testing.T) {
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Status: "DISABLED"}}
	inspector := metricsInspector("ds-1", sources)
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusFailed {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestMetricsSummaryOK(t *testing.T) {
	sources := &fixDatasourceRepo{
		item: &bo.Datasource{UID: "ds-1", Status: "ACTIVE"},
		values: map[string]float64{
			MetricNodeCPUUsagePercent: 12.5, MetricNodeMemoryUsagePercent: 60,
			MetricPodCPUUsageCores: 1.5, MetricPodMemoryUsageBytes: 1024,
		},
		found: map[string]bool{
			MetricNodeCPUUsagePercent: true, MetricNodeMemoryUsagePercent: true,
			MetricPodCPUUsageCores: true, MetricPodMemoryUsageBytes: true,
		},
	}
	inspector := metricsInspector("ds-1", sources)
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusOK || summary.Message != "" {
		t.Fatalf("summary = %+v", summary)
	}
	if summary.NodeCPUUsagePercent == nil || *summary.NodeCPUUsagePercent != 12.5 {
		t.Fatalf("cpu = %+v", summary.NodeCPUUsagePercent)
	}
	if summary.PodMemoryUsageBytes == nil || *summary.PodMemoryUsageBytes != 1024 {
		t.Fatalf("memory = %+v", summary.PodMemoryUsageBytes)
	}
}

func TestMetricsSummaryPartial(t *testing.T) {
	sources := &fixDatasourceRepo{
		item: &bo.Datasource{UID: "ds-1", Status: "ACTIVE"},
		values: map[string]float64{
			MetricNodeCPUUsagePercent: 12.5, MetricNodeMemoryUsagePercent: 60, MetricPodCPUUsageCores: 1.5,
		},
		found: map[string]bool{
			MetricNodeCPUUsagePercent: true, MetricNodeMemoryUsagePercent: true, MetricPodCPUUsageCores: true,
		},
	}
	inspector := metricsInspector("ds-1", sources)
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusPartial {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.UnavailableMetrics) != 1 || summary.UnavailableMetrics[0] != MetricPodMemoryUsageBytes {
		t.Fatalf("unavailable = %+v", summary.UnavailableMetrics)
	}
	if summary.PodMemoryUsageBytes != nil {
		t.Fatal("missing metric must stay nil")
	}
	if summary.Message == "" {
		t.Fatal("message is empty")
	}
}

func TestMetricsSummaryFailed(t *testing.T) {
	// 四条查询都报错：FAILED，message 说明失败原因。
	errs := map[string]error{}
	for _, item := range metricsSummaryQueries {
		errs[item.key] = merr.Server("upstream query failed")
	}
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Status: "ACTIVE"}, errs: errs}
	inspector := metricsInspector("ds-1", sources)
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusFailed || !strings.Contains(summary.Message, "metrics query failed") {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestMetricsSummaryAllEmpty(t *testing.T) {
	// 查询成功但没有任何采样：说明数据源没有 node-exporter / cAdvisor 指标。
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Status: "ACTIVE"}}
	inspector := metricsInspector("ds-1", sources)
	summary, err := inspector.MetricsSummary(ownerContext(), "c1")
	if err != nil {
		t.Fatalf("MetricsSummary: %v", err)
	}
	if summary.Status != MetricsStatusFailed || !strings.Contains(summary.Message, "node-exporter") {
		t.Fatalf("summary = %+v", summary)
	}
}

// ---------- 绑定校验 ----------

func registryFor(clusters *fixClusterRepo, sources *fixDatasourceRepo, checker access.Checker) *Registry {
	return NewRegistry(clusters, sources, checker)
}

func TestCreateClusterBindingValid(t *testing.T) {
	clusters := &fixClusterRepo{}
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Type: "METRICS", Driver: metricsDriverPrometheus, Status: "ACTIVE"}}
	registry := registryFor(clusters, sources, ownerChecker())
	if _, err := registry.Create(ownerContext(), "prod", "apiVersion: v1", "", strPtr("ds-1")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if clusters.createdBinding != "ds-1" {
		t.Fatalf("binding = %q", clusters.createdBinding)
	}
}

func TestCreateClusterWithoutBinding(t *testing.T) {
	clusters := &fixClusterRepo{}
	registry := registryFor(clusters, &fixDatasourceRepo{}, ownerChecker())
	if _, err := registry.Create(ownerContext(), "prod", "apiVersion: v1", "", nil); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if clusters.createdBinding != "" {
		t.Fatalf("binding = %q, want empty", clusters.createdBinding)
	}
}

func TestCreateClusterRejectsNonMetricsDatasource(t *testing.T) {
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Type: "LOGS", Driver: "LOGS_LOKI", Status: "ACTIVE"}}
	registry := registryFor(&fixClusterRepo{}, sources, ownerChecker())
	if _, err := registry.Create(ownerContext(), "prod", "apiVersion: v1", "", strPtr("ds-1")); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestCreateClusterRejectsUnsupportedMetricsDriver(t *testing.T) {
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Type: "METRICS", Driver: "METRICS_OTHER", Status: "ACTIVE"}}
	registry := registryFor(&fixClusterRepo{}, sources, ownerChecker())
	if _, err := registry.Create(ownerContext(), "prod", "apiVersion: v1", "", strPtr("ds-1")); !merr.IsInvalid(err) {
		t.Fatalf("err = %v, want invalid", err)
	}
}

func TestCreateClusterRejectsForeignDatasource(t *testing.T) {
	sources := &fixDatasourceRepo{getErr: merr.NotFound("not found")}
	registry := registryFor(&fixClusterRepo{}, sources, ownerChecker())
	if _, err := registry.Create(ownerContext(), "prod", "apiVersion: v1", "", strPtr("ds-1")); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestCreateClusterBindingNeedsDatasourceRead(t *testing.T) {
	clusters := &fixClusterRepo{}
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-1", Type: "METRICS", Driver: metricsDriverVictoriaMetrics, Status: "ACTIVE"}}
	// 只有 cluster:write，没有 datasource:read。
	registry := registryFor(clusters, sources, memberChecker(perm.ClusterWrite))
	_, err := registry.Create(memberContext(), "prod", "apiVersion: v1", "", strPtr("ds-1"))
	if !isForbidden(err) {
		t.Fatalf("err = %v, want forbidden", err)
	}
	if clusters.createdBinding != "" {
		t.Fatalf("binding = %q, want no write", clusters.createdBinding)
	}
}

func TestUpdateClusterBindingSemantics(t *testing.T) {
	sources := &fixDatasourceRepo{item: &bo.Datasource{UID: "ds-2", Type: "METRICS", Driver: metricsDriverVictoriaMetrics, Status: "ACTIVE"}}
	clusters := &fixClusterRepo{}
	registry := registryFor(clusters, sources, ownerChecker())

	// nil 表示保持原值。
	if _, err := registry.Update(ownerContext(), "c1", "prod", "", "", nil); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if clusters.updatedBinding != nil {
		t.Fatalf("binding = %v, want nil", *clusters.updatedBinding)
	}

	// 显式空串表示解绑，不需要读数据源。
	if _, err := registry.Update(ownerContext(), "c1", "prod", "", "", strPtr("")); err != nil {
		t.Fatalf("Update unbind: %v", err)
	}
	if clusters.updatedBinding == nil || *clusters.updatedBinding != "" {
		t.Fatalf("binding = %v, want empty", clusters.updatedBinding)
	}

	// 绑定非空数据源。
	if _, err := registry.Update(ownerContext(), "c1", "prod", "", "", strPtr("ds-2")); err != nil {
		t.Fatalf("Update bind: %v", err)
	}
	if clusters.updatedBinding == nil || *clusters.updatedBinding != "ds-2" {
		t.Fatalf("binding = %v", clusters.updatedBinding)
	}
}

func strPtr(value string) *string { return &value }

// isForbidden 判断错误是否是平台无权限。
func isForbidden(err error) bool {
	return err != nil && strings.Contains(err.Error(), "permission denied")
}

func TestNodeDetailRequiresNodeRead(t *testing.T) {
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, &fixKubeResources{}, &fixDatasourceRepo{},
		memberChecker(perm.ClusterRead))
	if _, err := inspector.NodeDetail(memberContext(), "c1", "node-a"); !isForbidden(err) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func TestPodDetailRequiresPodRead(t *testing.T) {
	inspector := NewInspector(&fixClusterRepo{cluster: &bo.Cluster{UID: "c1"}}, &fixKubeResources{}, &fixDatasourceRepo{},
		memberChecker(perm.ClusterRead))
	if _, err := inspector.PodDetail(memberContext(), "c1", "default", "web-1"); !isForbidden(err) {
		t.Fatalf("err = %v, want forbidden", err)
	}
}

func TestNodeDetailForeignCluster(t *testing.T) {
	inspector := NewInspector(&fixClusterRepo{err: merr.NotFound("not found")}, &fixKubeResources{}, &fixDatasourceRepo{}, ownerChecker())
	if _, err := inspector.NodeDetail(ownerContext(), "missing", "node-a"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestMetricsSummaryForeignCluster(t *testing.T) {
	inspector := NewInspector(&fixClusterRepo{err: merr.NotFound("not found")}, &fixKubeResources{}, &fixDatasourceRepo{}, ownerChecker())
	if _, err := inspector.MetricsSummary(ownerContext(), "missing"); !merr.IsNotFound(err) {
		t.Fatalf("err = %v, want not found", err)
	}
}
