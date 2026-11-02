package cluster

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// datasourceTypeMetrics 是集群概况只接受的指标类数据源。
	datasourceTypeMetrics = "METRICS"
	// metricsDriverPrometheus 和 metricsDriverVictoriaMetrics 是允许绑定的指标驱动。
	metricsDriverPrometheus      = "METRICS_PROMETHEUS"
	metricsDriverVictoriaMetrics = "METRICS_VICTORIA_METRICS"

	// 指标汇总状态。响应统一 200，由状态表达结果，避免前端把「没绑数据源」当成接口错误。
	MetricsStatusUnbound = "UNBOUND"
	MetricsStatusOK      = "OK"
	MetricsStatusPartial = "PARTIAL"
	MetricsStatusFailed  = "FAILED"

	// 概况里的区块名，与 Web 的降级提示一一对应。
	overviewSectionNodes     = "nodes"
	overviewSectionWorkloads = "workloads"
	overviewSectionPods      = "pods"
	overviewSectionJobs      = "jobs"
	overviewSectionServices  = "services"
	overviewSectionIngress   = "ingresses"

	// maxDetailEvents 是节点 / Pod 详情附带的最近事件条数。
	maxDetailEvents int32 = 20
	// metricsSummaryTimeout 是四条汇总查询的整体超时，避免慢数据源拖住概况页。
	metricsSummaryTimeout = 5 * time.Second
	// metricSampleInterval 是四条固定 PromQL 的统计区间。
	metricSampleInterval = "[5m]"
)

// Pod 阶段分布固定返回这五行，缺的阶段补 0。
var overviewPodPhases = []string{"Running", "Pending", "Succeeded", "Failed", "Unknown"}

// 概况里的工作负载固定返回这三类。
var overviewWorkloadKinds = []string{"Deployment", "StatefulSet", "DaemonSet"}

// 概况汇总的四条固定 PromQL。不允许用户输入表达式，避免把概况变成查询入口。
const (
	metricNodeCPUUsage = `100 - avg(rate(node_cpu_seconds_total{mode="idle"}` + metricSampleInterval + `)) * 100`
	metricNodeMemory   = `100 * (1 - sum(node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes))`
	metricPodCPU       = `sum(rate(container_cpu_usage_seconds_total{image!=""}` + metricSampleInterval + `))`
	metricPodMemory    = `sum(container_memory_working_set_bytes{image!=""})`

	// 四条汇总指标在 unavailable_metrics 里的 key，与 Web 的字段名一致。
	MetricNodeCPUUsagePercent    = "node_cpu_usage_percent"
	MetricNodeMemoryUsagePercent = "node_memory_usage_percent"
	MetricPodCPUUsageCores       = "pod_cpu_usage_cores"
	MetricPodMemoryUsageBytes    = "pod_memory_usage_bytes"

	// eventsUnavailableWarning 是事件读取失败时的降级提示。
	eventsUnavailableWarning = "events are not available for this cluster credential"
)

// metricsSummaryQueries 按固定顺序描述四条汇总查询，结果字段由 assign 写入。
type metricsSummaryQuery struct {
	key   string
	query string
}

var metricsSummaryQueries = []metricsSummaryQuery{
	{key: MetricNodeCPUUsagePercent, query: metricNodeCPUUsage},
	{key: MetricNodeMemoryUsagePercent, query: metricNodeMemory},
	{key: MetricPodCPUUsageCores, query: metricPodCPU},
	{key: MetricPodMemoryUsageBytes, query: metricPodMemory},
}

// Inspector 是集群的只读巡检：概况聚合、指标汇总与节点 / Pod 详情。
// 所有方法都先校验资源级读权限与集群归属，单个区块失败只降级该区块。
type Inspector struct {
	clusters    repository.ClusterRepo
	resources   repository.KubernetesRepo
	datasources repository.DatasourceRepo
	access      access.Checker
}

func NewInspector(clusters repository.ClusterRepo, resources repository.KubernetesRepo, datasources repository.DatasourceRepo, checker access.Checker) *Inspector {
	return &Inspector{clusters: clusters, resources: resources, datasources: datasources, access: checker}
}

// isMetricsDriver 判断驱动是否走 Prometheus 兼容的指标查询接口。
func isMetricsDriver(driver string) bool {
	return driver == metricsDriverPrometheus || driver == metricsDriverVictoriaMetrics
}

// clusterScope 校验集群读权限和集群归属，返回租户和集群元数据。
func (inspector *Inspector) clusterScope(ctx context.Context, clusterUID, code string) (string, *bo.Cluster, error) {
	ns, err := inspector.access.Allow(ctx, code)
	if err != nil {
		return "", nil, err
	}
	item, err := inspector.clusters.GetCluster(ctx, ns, clusterUID)
	if err != nil {
		return "", nil, inspector.access.Wrap(err)
	}
	return ns, item, nil
}

// Overview 一次聚合六个区块。每个区块单独校验读权限、单独执行：缺权限进
// denied_sections，查询失败进 failed_sections，接口仍然成功。
// 六个区块都因同一个连接错误失败时返回该错误，避免把不可达集群伪装成全空。
func (inspector *Inspector) Overview(ctx context.Context, clusterUID, namespace string) (*bo.KubeOverview, error) {
	_, cluster, err := inspector.clusterScope(ctx, clusterUID, perm.ClusterRead)
	if err != nil {
		return nil, err
	}
	namespace = strings.TrimSpace(namespace)
	overview := &bo.KubeOverview{
		PodPhases:   make([]bo.KubePodPhaseCount, 0, len(overviewPodPhases)),
		Workloads:   make([]bo.KubeWorkloadSummary, 0, len(overviewWorkloadKinds)),
		GeneratedAt: time.Now().Unix(),
	}
	// 固定行的骨架先建好：缺权限或失败时这几行仍然返回 0 值。
	knownPhases := make(map[string]struct{}, len(overviewPodPhases))
	for _, phase := range overviewPodPhases {
		knownPhases[phase] = struct{}{}
	}
	phaseCounts := make(map[string]int32, len(overviewPodPhases))
	workloadTotals := make(map[string]*bo.KubeWorkloadSummary, len(overviewWorkloadKinds))
	for _, kind := range overviewWorkloadKinds {
		workloadTotals[kind] = &bo.KubeWorkloadSummary{Kind: kind}
	}

	type section struct {
		name string
		code string
		run  func(context.Context) error
	}
	sections := []section{
		{name: overviewSectionNodes, code: perm.K8sNodeRead, run: func(ctx context.Context) error {
			items, err := inspector.resources.ListNodes(ctx, clusterUID, "")
			if err != nil {
				return err
			}
			for _, item := range items {
				overview.NodeTotal++
				if item.Status == "Ready" {
					overview.NodeReady++
					// 可分配量只累加 Ready 且可调度的节点，口径与「容量」卡片不同。
					if item.Schedulable {
						overview.AllocatableCPUMillis += item.AllocatableCPUMillis
						overview.AllocatableMemoryBytes += item.AllocatableMemoryBytes
					}
				} else {
					overview.NodeNotReady++
				}
			}
			return nil
		}},
		{name: overviewSectionWorkloads, code: perm.K8sWorkloadRead, run: func(ctx context.Context) error {
			items, err := inspector.resources.ListWorkloads(ctx, clusterUID, namespace, "", "")
			if err != nil {
				return err
			}
			for _, item := range items {
				total, ok := workloadTotals[item.Kind]
				if !ok {
					continue
				}
				total.Total++
				total.Desired += item.DesiredReplicas
				total.Ready += item.ReadyReplicas
			}
			return nil
		}},
		{name: overviewSectionPods, code: perm.K8sPodRead, run: func(ctx context.Context) error {
			items, err := inspector.resources.ListPods(ctx, clusterUID, namespace, "", "", "")
			if err != nil {
				return err
			}
			overview.PodTotal = int32(len(items))
			for _, item := range items {
				phase := item.Phase
				if _, ok := knownPhases[phase]; !ok {
					phase = "Unknown"
				}
				phaseCounts[phase]++
			}
			return nil
		}},
		{name: overviewSectionJobs, code: perm.K8sJobRead, run: func(ctx context.Context) error {
			items, err := inspector.resources.ListJobs(ctx, clusterUID, namespace, "")
			if err != nil {
				return err
			}
			overview.JobTotal = int32(len(items))
			for _, item := range items {
				overview.JobActive += item.Active
				overview.JobFailed += item.Failed
			}
			return nil
		}},
		{name: overviewSectionServices, code: perm.K8sServiceRead, run: func(ctx context.Context) error {
			items, err := inspector.resources.ListServices(ctx, clusterUID, namespace, "")
			if err != nil {
				return err
			}
			overview.ServiceTotal = int32(len(items))
			return nil
		}},
		{name: overviewSectionIngress, code: perm.K8sIngressRead, run: func(ctx context.Context) error {
			items, err := inspector.resources.ListIngresses(ctx, clusterUID, namespace, "")
			if err != nil {
				return err
			}
			overview.IngressTotal = int32(len(items))
			return nil
		}},
	}

	type sectionResult struct {
		name string
		err  error
	}
	results := make(chan sectionResult, len(sections)+1)
	var wg sync.WaitGroup
	for _, item := range sections {
		if err := inspector.access.Require(ctx, item.code); err != nil {
			overview.DeniedSections = append(overview.DeniedSections, item.name)
			continue
		}
		wg.Add(1)
		go func(section section) {
			defer wg.Done()
			results <- sectionResult{name: section.name, err: section.run(ctx)}
		}(item)
	}
	// API Server 版本优先实时 discovery，失败回落最近一次连通性探测的缓存值。
	wg.Add(1)
	go func() {
		defer wg.Done()
		version, err := inspector.resources.ServerVersion(ctx, clusterUID)
		if err != nil || strings.TrimSpace(version) == "" {
			version = strings.TrimSpace(cluster.Version)
		}
		overview.APIServerVersion = version
		results <- sectionResult{name: "api_server_version", err: nil}
	}()
	wg.Wait()
	close(results)

	var sectionErrs []error
	for result := range results {
		if result.err == nil {
			continue
		}
		overview.FailedSections = append(overview.FailedSections, result.name)
		sectionErrs = append(sectionErrs, result.err)
	}
	// 并发收集的顺序不稳定，按固定区块顺序输出，便于前端和验收对齐。
	sort.Slice(overview.FailedSections, func(i, j int) bool {
		return overviewSectionOrder(overview.FailedSections[i]) < overviewSectionOrder(overview.FailedSections[j])
	})
	// 六个区块全部失败且原因相同：连接错误，不能返回一份全 0 的概况。
	if len(overview.DeniedSections) == 0 && len(overview.FailedSections) == len(sections) && sameErrors(sectionErrs) {
		return nil, sectionErrs[0]
	}
	for _, phase := range overviewPodPhases {
		overview.PodPhases = append(overview.PodPhases, bo.KubePodPhaseCount{Phase: phase, Count: phaseCounts[phase]})
	}
	for _, kind := range overviewWorkloadKinds {
		overview.Workloads = append(overview.Workloads, *workloadTotals[kind])
	}
	return overview, nil
}

// overviewSectionOrder 返回区块在固定顺序里的位置，用于稳定排序。
func overviewSectionOrder(name string) int {
	order := []string{
		overviewSectionNodes, overviewSectionWorkloads, overviewSectionPods,
		overviewSectionJobs, overviewSectionServices, overviewSectionIngress,
	}
	for index, section := range order {
		if section == name {
			return index
		}
	}
	return len(order)
}

// sameErrors 判断所有错误是否来自同一个失败原因。错误由 kubeError 归一化成固定文案，
// 比较文案即可；少于两个错误时不认为「同一个连接错误」。
func sameErrors(errs []error) bool {
	if len(errs) < 2 {
		return false
	}
	first := errs[0].Error()
	for _, err := range errs[1:] {
		if err.Error() != first {
			return false
		}
	}
	return true
}

// NodeDetail 返回节点详情，并附带最近事件。事件读取失败只写 warnings，
// 不影响节点主体：events 是独立资源，常见只读角色没有 list 权限。
func (inspector *Inspector) NodeDetail(ctx context.Context, clusterUID, name string) (*bo.KubeNodeDetail, error) {
	_, _, err := inspector.clusterScope(ctx, clusterUID, perm.K8sNodeRead)
	if err != nil {
		return nil, err
	}
	detail, err := inspector.resources.NodeDetail(ctx, clusterUID, name)
	if err != nil {
		return nil, inspector.access.Wrap(err)
	}
	// 节点事件按约定写在 default 命名空间。
	events, err := inspector.resources.ListEvents(ctx, clusterUID, "default", "Node", detail.Node.Name, detail.Node.UID, maxDetailEvents)
	if err != nil {
		detail.Warnings = append(detail.Warnings, eventsUnavailableWarning)
	} else {
		detail.Events = events
	}
	return detail, nil
}

// PodDetail 返回 Pod 详情，并附带最近事件。事件读取失败同样只降级为 warnings。
func (inspector *Inspector) PodDetail(ctx context.Context, clusterUID, namespace, name string) (*bo.KubePodDetail, error) {
	_, _, err := inspector.clusterScope(ctx, clusterUID, perm.K8sPodRead)
	if err != nil {
		return nil, err
	}
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	if namespace == "" || name == "" {
		return nil, merr.Invalid("namespace and name are required")
	}
	detail, err := inspector.resources.PodDetail(ctx, clusterUID, namespace, name)
	if err != nil {
		return nil, inspector.access.Wrap(err)
	}
	events, err := inspector.resources.ListEvents(ctx, clusterUID, namespace, "Pod", detail.Pod.Name, detail.Pod.UID, maxDetailEvents)
	if err != nil {
		detail.Warnings = append(detail.Warnings, eventsUnavailableWarning)
	} else {
		detail.Events = events
	}
	return detail, nil
}

// MetricsSummary 读取集群绑定数据源的四条汇总指标。未绑定、数据源停用或删除、
// 上游失败都通过 status 表达；只有权限和集群归属错误返回给接口层。
func (inspector *Inspector) MetricsSummary(ctx context.Context, clusterUID string) (*bo.ClusterMetricsSummary, error) {
	ns, cluster, err := inspector.clusterScope(ctx, clusterUID, perm.ClusterRead)
	if err != nil {
		return nil, err
	}
	uid := strings.TrimSpace(cluster.MetricsDatasourceUID)
	if uid == "" {
		return &bo.ClusterMetricsSummary{Status: MetricsStatusUnbound, Message: "no metrics datasource is bound to this cluster"}, nil
	}
	datasource, err := inspector.datasources.GetDatasource(ctx, ns, uid)
	if err != nil {
		// 数据源被删除时 GetDatasource 返回 NotFound；绑定保留，汇总改成失败状态。
		return &bo.ClusterMetricsSummary{Status: MetricsStatusFailed, Message: "metrics datasource is not available"}, nil
	}
	if datasource.Status != "ACTIVE" {
		return &bo.ClusterMetricsSummary{Status: MetricsStatusFailed, Message: "metrics datasource is disabled"}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, metricsSummaryTimeout)
	defer cancel()
	type metricResult struct {
		key   string
		value float64
		found bool
		err   error
	}
	results := make(chan metricResult, len(metricsSummaryQueries))
	var wg sync.WaitGroup
	for _, item := range metricsSummaryQueries {
		wg.Add(1)
		go func(item metricsSummaryQuery) {
			defer wg.Done()
			value, found, err := inspector.datasources.InstantValue(ctx, *datasource, item.query)
			results <- metricResult{key: item.key, value: value, found: found, err: err}
		}(item)
	}
	wg.Wait()
	close(results)

	values := make(map[string]float64, len(metricsSummaryQueries))
	unavailable := make([]string, 0, len(metricsSummaryQueries))
	var firstErr error
	for result := range results {
		if result.err != nil || !result.found {
			unavailable = append(unavailable, result.key)
			if result.err != nil && firstErr == nil {
				firstErr = result.err
			}
			continue
		}
		values[result.key] = result.value
	}
	summary := &bo.ClusterMetricsSummary{}
	assign := func(key string) *float64 {
		value, ok := values[key]
		if !ok {
			return nil
		}
		return &value
	}
	summary.NodeCPUUsagePercent = assign(MetricNodeCPUUsagePercent)
	summary.NodeMemoryUsagePercent = assign(MetricNodeMemoryUsagePercent)
	summary.PodCPUUsageCores = assign(MetricPodCPUUsageCores)
	summary.PodMemoryUsageBytes = assign(MetricPodMemoryUsageBytes)
	switch {
	case len(unavailable) == 0:
		summary.Status = MetricsStatusOK
	case len(values) > 0:
		summary.Status = MetricsStatusPartial
		summary.UnavailableMetrics = unavailable
		summary.Message = "some metrics are not available: " + strings.Join(unavailable, ", ")
	case firstErr != nil:
		summary.Status = MetricsStatusFailed
		summary.Message = "metrics query failed: " + errorMessage(firstErr)
	default:
		// 四条查询都成功但都没有采样：数据源可达但没有 node-exporter / cAdvisor 指标。
		summary.Status = MetricsStatusFailed
		summary.Message = "no metric samples were returned; node-exporter and cAdvisor metrics are required"
	}
	return summary, nil
}

// errorMessage 取 Kratos 错误的文案，避免把状态码拼进给用户看的理由里。
func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return kerrors.FromError(err).Message
}
