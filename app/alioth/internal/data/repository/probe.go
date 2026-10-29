package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// ProbeTargetRepo 读写探测目标组、组内地址与探测记录。
type ProbeTargetRepo interface {
	CreateProbeTargetGroup(ctx context.Context, in bo.ProbeTargetGroup) (*bo.ProbeTargetGroup, error)
	UpdateProbeTargetGroup(ctx context.Context, in bo.ProbeTargetGroup) (*bo.ProbeTargetGroup, error)
	UpdateProbeTargetGroupStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.ProbeTargetGroup, error)
	DeleteProbeTargetGroup(ctx context.Context, namespaceUID, uid string) error
	GetProbeTargetGroup(ctx context.Context, namespaceUID, uid string) (*bo.ProbeTargetGroup, error)
	ListProbeTargetGroups(ctx context.Context, namespaceUID string, page, pageSize int, name, driver string) ([]bo.ProbeTargetGroup, int, error)
	// ListProbeTargetGroupsByMatchers 按驱动与标签匹配器返回启用中的目标组，
	// 语义与数据源标签匹配一致：匹配器为空表示该驱动下全部启用目标组。
	ListProbeTargetGroupsByMatchers(ctx context.Context, namespaceUID, driver, matchersJSON string) ([]bo.ProbeTargetGroup, error)
	// ListProbeTargetGroupsByExecutor 返回某个执行器负责的目标组：executor_labels
	// 为空的目标组由集群内执行器负责。
	ListProbeTargetGroupsByExecutor(ctx context.Context, executorKind, executorLabels string) ([]bo.ProbeTargetGroup, error)
	// ListProbeTargets 返回目标组内启用中的地址。
	ListProbeTargets(ctx context.Context, namespaceUID, groupUID string) ([]bo.ProbeTarget, error)

	// SaveProbeResults 按 (target_uid, slot) 幂等写入探测结果。
	SaveProbeResults(ctx context.Context, results []bo.ProbeResult) error
	// ListProbeResults 按目标与时间窗口返回探测结果，按 Slot 升序。
	ListProbeResults(ctx context.Context, namespaceUID string, targetUIDs []string, from, to time.Time) ([]bo.ProbeResult, error)
	// ListProbeTargetResults 分页返回某个目标组的探测记录。
	ListProbeTargetResults(ctx context.Context, namespaceUID, groupUID, targetUID string, page, pageSize int) ([]bo.ProbeResult, int, error)
	// LatestProbeResults 返回一批目标各自最近一条探测结果。
	LatestProbeResults(ctx context.Context, targetUIDs []string) (map[string]bo.ProbeResult, error)
	// PruneProbeResults 清理超过保留期的探测记录，返回删除条数。
	PruneProbeResults(ctx context.Context, before time.Time, keepPerTarget int) (int, error)
}
