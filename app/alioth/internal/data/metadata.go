package data

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// esClusterHealthPath 和 esAllocationPath 是元数据页读取的 Elasticsearch 接口。
// allocation 返回每个节点的分片和磁盘分配，按 JSON 解析后直接展示。
const (
	lokiBuildInfoPath   = "/loki/api/v1/status/buildinfo"
	vlogsHealthPath     = "/health"
	esClusterHealthPath = "/_cluster/health"
	esAllocationPath    = "/_cat/allocation?format=json"
)

// DatasourceMetadata 按日志驱动读取关键信息。指标数据源不使用这条路径，
// 元数据页仍然通过 ListMetrics 读取指标说明。
func (r *datasourceRepo) DatasourceMetadata(ctx context.Context, item bo.Datasource) (*bo.DatasourceMetadata, error) {
	switch item.Driver {
	case driverLogsLoki:
		return r.lokiMetadata(ctx, item)
	case driverLogsElasticsearch:
		return r.elasticsearchMetadata(ctx, item)
	case driverLogsVlogs:
		return r.vlogsMetadata(ctx, item)
	default:
		return nil, merr.Invalid("unsupported datasource driver")
	}
}

// lokiMetadata 读取 Loki 的构建信息。Loki 没有统一的集群概览接口，
// 版本、提交和构建时间是配置数据源时最有用的信息。
func (r *datasourceRepo) lokiMetadata(ctx context.Context, item bo.Datasource) (*bo.DatasourceMetadata, error) {
	body, err := r.callUpstream(ctx, item, http.MethodGet, lokiBuildInfoPath, "")
	if err != nil {
		return nil, err
	}
	var payload struct {
		Version   string `json:"version"`
		Revision  string `json:"revision"`
		Branch    string `json:"branch"`
		BuildDate string `json:"buildDate"`
		BuildUser string `json:"buildUser"`
		GoVersion string `json:"goVersion"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		r.log.Error(fmt.Sprintf("datasource %s loki metadata", item.UID))
		return nil, merr.Client("upstream query failed")
	}
	return &bo.DatasourceMetadata{
		Driver: item.Driver,
		Items: []bo.DatasourceMetadataItem{
			{Key: "status", Value: "up"},
			{Key: "version", Value: payload.Version},
			{Key: "revision", Value: payload.Revision},
			{Key: "branch", Value: payload.Branch},
			{Key: "buildDate", Value: payload.BuildDate},
			{Key: "buildUser", Value: payload.BuildUser},
			{Key: "goVersion", Value: payload.GoVersion},
		},
	}, nil
}

// vlogsMetadata 读取 VictoriaLogs 的健康检查。VictoriaLogs 未提供构建信息接口，
// 元数据页展示连通状态和查询入口说明。
func (r *datasourceRepo) vlogsMetadata(ctx context.Context, item bo.Datasource) (*bo.DatasourceMetadata, error) {
	body, err := r.callUpstream(ctx, item, http.MethodGet, vlogsHealthPath, "")
	if err != nil {
		return nil, err
	}
	health := strings.TrimSpace(body)
	if health == "" {
		health = "up"
	}
	return &bo.DatasourceMetadata{
		Driver: item.Driver,
		Items: []bo.DatasourceMetadataItem{
			{Key: "status", Value: "up"},
			{Key: "health", Value: health},
		},
	}, nil
}

// elasticsearchMetadata 读取集群健康状态和节点分配。
// 集群状态取 _cluster/health，节点分片和磁盘取 _cat/allocation。
func (r *datasourceRepo) elasticsearchMetadata(ctx context.Context, item bo.Datasource) (*bo.DatasourceMetadata, error) {
	healthBody, err := r.callUpstream(ctx, item, http.MethodGet, esClusterHealthPath, "")
	if err != nil {
		return nil, err
	}
	var health struct {
		ClusterName                 string  `json:"cluster_name"`
		Status                      string  `json:"status"`
		NumberOfNodes               int     `json:"number_of_nodes"`
		NumberOfDataNodes           int     `json:"number_of_data_nodes"`
		ActivePrimaryShards         int     `json:"active_primary_shards"`
		ActiveShards                int     `json:"active_shards"`
		RelocatingShards            int     `json:"relocating_shards"`
		InitializingShards          int     `json:"initializing_shards"`
		UnassignedShards            int     `json:"unassigned_shards"`
		DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
		NumberOfPendingTasks        int     `json:"number_of_pending_tasks"`
		ActiveShardsPercentAsNumber float64 `json:"active_shards_percent_as_number"`
	}
	if err := json.Unmarshal([]byte(healthBody), &health); err != nil {
		r.log.Error(fmt.Sprintf("datasource %s elasticsearch health", item.UID))
		return nil, merr.Client("upstream query failed")
	}

	allocationBody, err := r.callUpstream(ctx, item, http.MethodGet, esAllocationPath, "")
	if err != nil {
		return nil, err
	}
	var allocation []struct {
		Shards      string `json:"shards"`
		DiskUsed    string `json:"disk.used"`
		DiskAvail   string `json:"disk.avail"`
		DiskPercent string `json:"disk.percent"`
		Host        string `json:"host"`
		IP          string `json:"ip"`
		Node        string `json:"node"`
		NodeRole    string `json:"node.role"`
	}
	if err := json.Unmarshal([]byte(allocationBody), &allocation); err != nil {
		r.log.Error(fmt.Sprintf("datasource %s elasticsearch allocation", item.UID))
		return nil, merr.Client("upstream query failed")
	}

	nodes := make([]bo.DatasourceMetadataNode, 0, len(allocation))
	for _, entry := range allocation {
		address := strings.TrimSpace(entry.IP)
		if address == "" {
			address = strings.TrimSpace(entry.Host)
		}
		name := strings.TrimSpace(entry.Node)
		if name == "" {
			name = strings.TrimSpace(entry.Host)
		}
		nodes = append(nodes, bo.DatasourceMetadataNode{
			Name:        name,
			Address:     address,
			Role:        strings.TrimSpace(entry.NodeRole),
			Shards:      strings.TrimSpace(entry.Shards),
			DiskUsed:    strings.TrimSpace(entry.DiskUsed),
			DiskAvail:   strings.TrimSpace(entry.DiskAvail),
			DiskPercent: strings.TrimSpace(entry.DiskPercent),
		})
	}

	return &bo.DatasourceMetadata{
		Driver: item.Driver,
		Items: []bo.DatasourceMetadataItem{
			{Key: "clusterName", Value: health.ClusterName},
			{Key: "status", Value: health.Status},
			{Key: "numberOfNodes", Value: strconv.Itoa(health.NumberOfNodes)},
			{Key: "numberOfDataNodes", Value: strconv.Itoa(health.NumberOfDataNodes)},
			{Key: "activePrimaryShards", Value: strconv.Itoa(health.ActivePrimaryShards)},
			{Key: "activeShards", Value: strconv.Itoa(health.ActiveShards)},
			{Key: "relocatingShards", Value: strconv.Itoa(health.RelocatingShards)},
			{Key: "initializingShards", Value: strconv.Itoa(health.InitializingShards)},
			{Key: "unassignedShards", Value: strconv.Itoa(health.UnassignedShards)},
			{Key: "delayedUnassignedShards", Value: strconv.Itoa(health.DelayedUnassignedShards)},
			{Key: "numberOfPendingTasks", Value: strconv.Itoa(health.NumberOfPendingTasks)},
			{Key: "activeShardsPercent", Value: strconv.FormatFloat(health.ActiveShardsPercentAsNumber, 'f', -1, 64) + "%"},
		},
		Nodes: nodes,
	}, nil
}
