import type { useClusterOverview } from "./use-cluster-overview"
import {
  ClusterCapacityCard,
  ClusterComponentCard,
  ClusterOverviewKpis,
  ClusterOverviewStatus,
  ClusterPodPhaseCard,
  ClusterTopCard,
  ClusterWorkloadHealthCard,
} from "./overview-cards"

type ClusterOverviewState = ReturnType<typeof useClusterOverview>

interface ClusterOverviewTabProps {
  overview: ClusterOverviewState
  namespace: string
  onOpenWorkloadKind: (kind: string) => void
}

/**
 * 集群概况：概览指标、Pod 分布、工作负载健康、容量、基础组件与 Top 榜单。
 * 数据都来自现有列表接口的浏览器侧汇总，卡片之间互不阻塞。
 */
export function ClusterOverviewTab({
  overview,
  namespace,
  onOpenWorkloadKind,
}: ClusterOverviewTabProps) {
  const data = overview.data

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
          onRetry={overview.reload}
        />
        <ClusterOverviewKpis data={data} loading={overview.loading} />
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
          <ClusterComponentCard
            data={data}
            loading={overview.loading}
            onRetry={overview.reload}
            refreshedAt={data?.refreshedAt ?? 0}
          />
        </div>
        <ClusterTopCard
          data={data}
          loading={overview.loading}
          onRetry={overview.reload}
        />
      </div>
    </div>
  )
}
