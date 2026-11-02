import type { useClusterOverview } from "./use-cluster-overview"
import {
  ClusterCapacityCard,
  ClusterComponentCard,
  ClusterJobsCard,
  ClusterMetricsCard,
  ClusterOverviewKpis,
  ClusterOverviewStatus,
  ClusterPodPhaseCard,
  ClusterWorkloadHealthCard,
} from "./overview-cards"

type ClusterOverviewState = ReturnType<typeof useClusterOverview>

interface ClusterOverviewTabProps {
  overview: ClusterOverviewState
  namespace: string
  onOpenWorkloadKind: (kind: string) => void
  /** cluster:write + datasource:read 才显示指标绑定入口。 */
  canBindMetrics: boolean
  /** 集群当前绑定的指标数据源名称，用于指标卡右上角说明。 */
  metricsDatasourceName: string
  onBindMetrics: () => void
}

/**
 * 集群概况：一次 /overview 聚合六区块，指标汇总独立加载。
 * 卡片之间互不阻塞，区块缺权限/失败按 deniedSections / failedSections 降级。
 */
export function ClusterOverviewTab({
  overview,
  namespace,
  onOpenWorkloadKind,
  canBindMetrics,
  metricsDatasourceName,
  onBindMetrics,
}: ClusterOverviewTabProps) {
  const data = overview.data
  const refreshedAt = data ? Number(data.generatedAt) * 1000 : 0

  return (
    <div
      data-slot="cluster-overview"
      className="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto pr-1"
    >
      <div className="grid gap-3">
        <ClusterOverviewStatus
          data={data}
          namespace={namespace}
          refreshing={overview.refreshing}
          failed={overview.failed}
          error={overview.error}
          onRetry={overview.reload}
        />
        <ClusterOverviewKpis
          data={data}
          loading={overview.loading}
          onRetry={overview.reload}
        />
        <div className="grid gap-3 xl:grid-cols-2">
          <ClusterPodPhaseCard
            data={data}
            loading={overview.loading}
            onRetry={overview.reload}
          />
          <ClusterWorkloadHealthCard
            data={data}
            loading={overview.loading}
            onRetry={overview.reload}
            onOpenKind={onOpenWorkloadKind}
          />
        </div>
        <div className="grid gap-3 xl:grid-cols-2">
          <ClusterCapacityCard
            data={data}
            loading={overview.loading}
            onRetry={overview.reload}
          />
          <ClusterMetricsCard
            metrics={overview.metrics}
            datasourceName={metricsDatasourceName}
            canBind={canBindMetrics}
            onBind={onBindMetrics}
            loading={overview.loading}
          />
        </div>
        <div className="grid gap-3 xl:grid-cols-2">
          <ClusterJobsCard
            data={data}
            loading={overview.loading}
            onRetry={overview.reload}
          />
          <ClusterComponentCard
            data={data}
            loading={overview.loading}
            onRetry={overview.reload}
            refreshedAt={refreshedAt}
          />
        </div>
      </div>
    </div>
  )
}
