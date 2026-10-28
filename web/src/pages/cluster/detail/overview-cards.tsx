import {
  Cpu,
  Database,
  Gauge,
  MemoryStick,
  RefreshCw,
  Server,
  TriangleAlert,
} from "lucide-react"
import type { ReactNode } from "react"
import type { TFunction } from "i18next"
import type { LucideIcon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { StatusBadge, type BadgeTone } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/class-names"

import {
  formatCpuCapacity,
  formatMemoryCapacity,
  formatRatio,
  readReadyRatio,
} from "../format"
import { POD_SAMPLE_LIMIT } from "./use-cluster-overview"
import type {
  ClusterOverviewSnapshot,
  ClusterPodSummary,
} from "./use-cluster-overview"

type KpiTone = "primary" | "success" | "warning" | "danger" | "neutral"

const toneChips: Record<KpiTone, string> = {
  primary: "bg-primary/10 text-primary-ink",
  success: "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400",
  warning: "bg-amber-500/10 text-amber-600 dark:text-amber-400",
  danger: "bg-red-500/10 text-red-600 dark:text-red-400",
  neutral: "bg-muted text-muted-foreground",
}

const PHASE_KEYS: Record<string, string> = {
  Running: "kube.phaseRunning",
  Pending: "kube.phasePending",
  Succeeded: "kube.phaseSucceeded",
  Failed: "kube.phaseFailed",
  Unknown: "kube.phaseUnknown",
}

const PHASE_BARS: Record<string, string> = {
  Running: "bg-emerald-500",
  Pending: "bg-amber-500",
  Succeeded: "bg-sky-500",
  Failed: "bg-red-500",
  Unknown: "bg-muted-foreground/50",
}

function phaseLabel(t: TFunction, phase: string): string {
  const key = PHASE_KEYS[phase]
  return key ? t(key) : phase
}

function phaseBar(phase: string): string {
  return PHASE_BARS[phase] ?? "bg-muted-foreground/40"
}

function workloadKindTone(kind: string): BadgeTone {
  if (kind === "Deployment") {
    return "info"
  }
  if (kind === "StatefulSet") {
    return "primary"
  }
  return "warning"
}

interface OverviewCardProps {
  title: string
  description: string
  badge?: ReactNode
  className?: string
  children: ReactNode
}

/** 概况卡片统一外壳：标题、口径说明与可选的右上角结论。 */
function OverviewCard({
  title,
  description,
  badge,
  className,
  children,
}: OverviewCardProps) {
  return (
    <Card size="sm" className={cn("min-w-0 gap-3", className)}>
      <CardHeader className="pb-0">
        <div className="flex min-w-0 items-start justify-between gap-3">
          <div className="min-w-0">
            <CardTitle className="text-sm">{title}</CardTitle>
            <CardDescription className="mt-0.5 text-xs leading-5">
              {description}
            </CardDescription>
          </div>
          {badge ? <div className="shrink-0">{badge}</div> : null}
        </div>
      </CardHeader>
      <CardContent className="min-w-0">{children}</CardContent>
    </Card>
  )
}

function CardSkeleton({ rows = 3 }: { rows?: number }) {
  return (
    <div className="grid gap-2 py-1">
      {Array.from({ length: rows }, (_, index) => (
        <Skeleton
          key={index}
          className={cn(
            "h-4",
            index === 0 ? "w-2/3" : index === 1 ? "w-full" : "w-1/2"
          )}
        />
      ))}
    </div>
  )
}

function CardFailure({ onRetry }: { onRetry: () => void }) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-2 py-2">
      <TriangleAlert aria-hidden="true" className="size-4 text-destructive" />
      <p className="text-xs text-destructive">{t("kube.overviewLoadFailed")}</p>
      <Button variant="ghost" size="sm" onClick={onRetry}>
        {t("action.retry")}
      </Button>
    </div>
  )
}

function CardEmpty({ text }: { text: string }) {
  return (
    <p className="py-4 text-center text-xs text-muted-foreground">{text}</p>
  )
}

interface ClusterKpiTileProps {
  label: string
  value: string
  hint: string
  icon: LucideIcon
  tone: KpiTone
  loading: boolean
}

function ClusterKpiTile({
  label,
  value,
  hint,
  icon: Icon,
  tone,
  loading,
}: ClusterKpiTileProps) {
  return (
    <Card size="sm" className="min-w-0">
      <div className="flex min-w-0 items-start justify-between gap-3 px-3">
        <div className="min-w-0 space-y-1">
          <p className="text-[11px] font-medium tracking-[0.08em] text-muted-foreground uppercase">
            {label}
          </p>
          {loading ? (
            <Skeleton className="h-8 w-16" />
          ) : (
            <p className="font-heading text-2xl leading-tight font-semibold tracking-tight tabular-nums">
              {value}
            </p>
          )}
          <p className="text-[11px] leading-4 text-muted-foreground">{hint}</p>
        </div>
        <span
          aria-hidden="true"
          className={cn(
            "grid size-8 shrink-0 place-items-center rounded-lg",
            toneChips[tone]
          )}
        >
          <Icon className="size-4" />
        </span>
      </div>
    </Card>
  )
}

interface ClusterOverviewKpisProps {
  data: ClusterOverviewSnapshot | null
  loading: boolean
}

/** 四条核心指标：节点就绪、Pod 运行中、工作负载健康、Ingress 数量。 */
export function ClusterOverviewKpis({
  data,
  loading,
}: ClusterOverviewKpisProps) {
  const { t } = useTranslation()
  const nodes = data?.nodes ?? null
  const pods = data?.pods ?? null
  const workloads = data?.workloads ?? null
  const ingresses = data?.ingresses ?? null
  const running =
    pods?.byPhase.find((item) => item.phase === "Running")?.count ?? 0
  const pending =
    pods?.byPhase.find((item) => item.phase === "Pending")?.count ?? 0

  const tiles: ClusterKpiTileProps[] = [
    {
      label: t("kube.overviewNodesReady"),
      value: nodes ? `${nodes.ready}/${nodes.total}` : "—",
      hint: nodes
        ? nodes.notReady > 0
          ? t("kube.overviewNodesNotReady", { count: nodes.notReady })
          : t("kube.overviewNodesAllReady")
        : data?.failed.nodes
          ? t("kube.overviewFailed")
          : t("kube.overviewWaiting"),
      icon: Server,
      tone: nodes ? (nodes.notReady > 0 ? "warning" : "success") : "neutral",
      loading: loading,
    },
    {
      label: t("kube.overviewPodsRunning"),
      value: pods ? String(running) : "—",
      hint: pods
        ? pending > 0
          ? t("kube.overviewPodsPending", {
              count: pending,
              total: pods.total,
            })
          : t("kube.overviewPodsHint", { total: pods.total })
        : data?.failed.pods
          ? t("kube.overviewFailed")
          : t("kube.overviewWaiting"),
      icon: Gauge,
      tone: pods ? (pending > 0 ? "warning" : "success") : "neutral",
      loading: loading,
    },
    {
      label: t("kube.overviewWorkloadHealthy"),
      value: workloads ? `${workloads.healthy}/${workloads.total}` : "—",
      hint: workloads
        ? t("kube.overviewWorkloadsHint", {
            healthy: workloads.healthy,
            total: workloads.total,
          })
        : data?.failed.workloads
          ? t("kube.overviewFailed")
          : t("kube.overviewWaiting"),
      icon: Cpu,
      tone: workloads
        ? workloads.healthy === workloads.total
          ? "success"
          : "warning"
        : "neutral",
      loading: loading,
    },
    {
      label: t("kube.overviewIngresses"),
      value: ingresses ? String(ingresses.total) : "—",
      hint: ingresses
        ? t("kube.overviewIngressesHint")
        : data?.failed.ingresses
          ? t("kube.overviewFailed")
          : t("kube.overviewWaiting"),
      icon: Database,
      tone: "primary",
      loading: loading,
    },
  ]

  return (
    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
      {tiles.map((tile) => (
        <ClusterKpiTile key={tile.label} {...tile} />
      ))}
    </div>
  )
}

interface ClusterPodPhaseCardProps {
  data: ClusterOverviewSnapshot | null
  loading: boolean
  onRetry: () => void
}

/** Pod 状态分布：阶段堆叠条 + 计数图例，样本数不足时明确标注。 */
export function ClusterPodPhaseCard({
  data,
  loading,
  onRetry,
}: ClusterPodPhaseCardProps) {
  const { t } = useTranslation()
  const summary: ClusterPodSummary | null = data?.pods ?? null
  const failed = data?.failed.pods ?? false
  const sampled = summary?.sampled ?? 0
  const ariaLabel = summary
    ? summary.byPhase
        .map((item) => `${phaseLabel(t, item.phase)} ${item.count}`)
        .join(", ")
    : ""

  return (
    <OverviewCard
      title={t("kube.podDistribution")}
      description={t("kube.podDistributionHint")}
    >
      {loading && !data ? (
        <CardSkeleton rows={4} />
      ) : failed ? (
        <CardFailure onRetry={onRetry} />
      ) : !summary || summary.total === 0 ? (
        <CardEmpty text={t("kube.podDistributionEmpty")} />
      ) : (
        <div className="grid gap-3">
          <div
            role="img"
            aria-label={ariaLabel}
            className="flex h-2.5 w-full overflow-hidden rounded-full bg-muted"
          >
            {summary.byPhase.map((item) => (
              <div
                key={item.phase}
                className={phaseBar(item.phase)}
                style={{
                  width: `${(item.count / Math.max(sampled, 1)) * 100}%`,
                }}
                title={`${phaseLabel(t, item.phase)} ${item.count}`}
              />
            ))}
          </div>
          <ul className="grid grid-cols-2 gap-x-4 gap-y-1.5 sm:grid-cols-3">
            {summary.byPhase.map((item) => (
              <li
                key={item.phase}
                className="flex min-w-0 items-center gap-1.5 text-xs"
              >
                <span
                  aria-hidden="true"
                  className={cn(
                    "size-2 shrink-0 rounded-full",
                    phaseBar(item.phase)
                  )}
                />
                <span className="min-w-0 truncate text-muted-foreground">
                  {phaseLabel(t, item.phase)}
                </span>
                <span className="ml-auto font-mono tabular-nums">
                  {item.count}
                </span>
              </li>
            ))}
          </ul>
          <p className="flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-muted-foreground">
            <span>{t("kube.podNotReady", { count: summary.notReady })}</span>
            <span aria-hidden="true">·</span>
            <span>{t("kube.podRestarts", { count: summary.restarts })}</span>
            {summary.truncated ? (
              <>
                <span aria-hidden="true">·</span>
                <span>
                  {t("kube.podDistributionSample", {
                    sampled: summary.sampled,
                    total: summary.total,
                    limit: POD_SAMPLE_LIMIT,
                  })}
                </span>
              </>
            ) : null}
          </p>
        </div>
      )}
    </OverviewCard>
  )
}

interface ClusterWorkloadHealthCardProps {
  data: ClusterOverviewSnapshot | null
  loading: boolean
  onRetry: () => void
  onOpenKind: (kind: string) => void
}

/** 工作负载健康：按类型看就绪副本，点击直接跳到对应类型列表。 */
export function ClusterWorkloadHealthCard({
  data,
  loading,
  onRetry,
  onOpenKind,
}: ClusterWorkloadHealthCardProps) {
  const { t } = useTranslation()
  const summary = data?.workloads ?? null
  const failed = data?.failed.workloads ?? false

  return (
    <OverviewCard
      title={t("kube.workloadHealth")}
      description={t("kube.workloadHealthHint")}
      badge={
        summary ? (
          <span
            className={cn(
              "rounded-full border px-2 py-0.5 text-[11px]",
              summary.healthy === summary.total
                ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300"
                : "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300"
            )}
          >
            {summary.healthy}/{summary.total}
          </span>
        ) : null
      }
    >
      {loading && !data ? (
        <CardSkeleton rows={4} />
      ) : failed ? (
        <CardFailure onRetry={onRetry} />
      ) : !summary || summary.total === 0 ? (
        <CardEmpty text={t("kube.workloadHealthEmpty")} />
      ) : (
        <ul className="grid gap-1">
          {summary.rows.map((row) => {
            const ratio = readReadyRatio(row.ready, row.desired)
            const healthy = row.desired <= 0 || row.ready >= row.desired
            return (
              <li key={row.kind}>
                <button
                  type="button"
                  onClick={() => onOpenKind(row.kind)}
                  title={t("kube.workloadHealthOpen")}
                  className="flex w-full min-w-0 items-center gap-3 rounded-lg px-1.5 py-2 text-left transition-colors hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
                >
                  <StatusBadge
                    value={row.kind}
                    tone={workloadKindTone(row.kind)}
                    className="w-[6.5rem] shrink-0 justify-center"
                  />
                  <span className="min-w-0 flex-1">
                    <span className="flex items-center justify-between gap-2 text-[11px] text-muted-foreground">
                      <span>
                        {t("kube.ready")} {row.ready}/{row.desired}
                      </span>
                      <span className="font-mono tabular-nums">
                        {formatRatio(ratio)}
                      </span>
                    </span>
                    <span
                      role="progressbar"
                      aria-label={row.kind}
                      aria-valuemin={0}
                      aria-valuemax={100}
                      aria-valuenow={Math.round(ratio * 100)}
                      className="mt-1 block h-1.5 overflow-hidden rounded-full bg-muted"
                    >
                      <span
                        className={cn(
                          "block h-full rounded-full",
                          healthy ? "bg-emerald-500" : "bg-amber-500"
                        )}
                        style={{ width: `${ratio * 100}%` }}
                      />
                    </span>
                  </span>
                  <span className="shrink-0 font-mono text-xs text-muted-foreground tabular-nums">
                    {row.total}
                  </span>
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </OverviewCard>
  )
}

interface ClusterCapacityCardProps {
  data: ClusterOverviewSnapshot | null
  loading: boolean
  onRetry: () => void
}

/** CPU / 内存容量。用量依赖指标数据源，未绑定时给出明确降级说明。 */
export function ClusterCapacityCard({
  data,
  loading,
  onRetry,
}: ClusterCapacityCardProps) {
  const { t } = useTranslation()
  const nodes = data?.nodes ?? null
  const failed = data?.failed.nodes ?? false

  return (
    <OverviewCard
      title={t("kube.capacityTitle")}
      description={t("kube.capacityHint")}
      badge={
        nodes && nodes.cordoned > 0 ? (
          <span className="rounded-full border border-amber-500/30 bg-amber-500/10 px-2 py-0.5 text-[11px] text-amber-700 dark:text-amber-300">
            {t("kube.capacityCordoned", { count: nodes.cordoned })}
          </span>
        ) : null
      }
    >
      {loading && !data ? (
        <CardSkeleton rows={4} />
      ) : failed ? (
        <CardFailure onRetry={onRetry} />
      ) : !nodes || nodes.total === 0 ? (
        <CardEmpty text={t("kube.capacityEmpty")} />
      ) : (
        <div className="grid gap-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="rounded-lg border border-border/60 bg-background/50 p-3">
              <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                <Cpu aria-hidden="true" className="size-3.5" />
                {t("kube.capacityCpu")}
              </p>
              <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
                {formatCpuCapacity(nodes.cpuMillis, t)}
              </p>
              <p className="mt-0.5 text-[11px] text-muted-foreground">
                {t("kube.capacityNodes", { count: nodes.total })}
              </p>
            </div>
            <div className="rounded-lg border border-border/60 bg-background/50 p-3">
              <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
                <MemoryStick aria-hidden="true" className="size-3.5" />
                {t("kube.capacityMemory")}
              </p>
              <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
                {formatMemoryCapacity(nodes.memoryBytes)}
              </p>
              <p className="mt-0.5 text-[11px] text-muted-foreground">
                {t("kube.capacityPods", { count: nodes.pods })}
              </p>
            </div>
          </div>
          <div className="flex flex-wrap items-center justify-between gap-2 rounded-lg border border-dashed border-border/70 px-2.5 py-2">
            <p className="text-[11px] leading-4 text-muted-foreground">
              {t("kube.capacityUsageHint")}
            </p>
            <span className="rounded border border-border/70 bg-muted/50 px-1.5 py-0.5 text-[10px] text-muted-foreground">
              {t("kube.noMetrics")}
            </span>
          </div>
        </div>
      )}
    </OverviewCard>
  )
}

interface ComponentRow {
  name: string
  state: string
  tone: BadgeTone
  hint: string
}

interface ClusterComponentCardProps {
  data: ClusterOverviewSnapshot | null
  loading: boolean
  onRetry: () => void
  refreshedAt: number
}

/**
 * 基础组件状态。当前接口没有控制面指标，只能给出可推断的结论：
 * API Server 由资源查询是否成功推断，kubelet 由节点 Ready 数量推断；
 * scheduler / controller-manager / etcd 需要指标数据源，明确标注未接入。
 */
export function ClusterComponentCard({
  data,
  loading,
  onRetry,
  refreshedAt,
}: ClusterComponentCardProps) {
  const { t } = useTranslation()
  const nodes = data?.nodes ?? null
  const reachable = data
    ? Object.values(data.failed).some((value) => !value)
    : false
  const updatedAt =
    refreshedAt > 0 ? new Date(refreshedAt).toLocaleTimeString() : ""

  const rows: ComponentRow[] = [
    {
      name: "kube-apiserver",
      state: reachable
        ? t("kube.componentReachable")
        : t("kube.componentUnknown"),
      tone: reachable ? "success" : "neutral",
      hint: reachable
        ? updatedAt
          ? `${t("kube.componentApiServerHint")} · ${updatedAt}`
          : t("kube.componentApiServerHint")
        : t("kube.componentUnknownHint"),
    },
    {
      name: "kube-scheduler",
      state: t("kube.componentUnknown"),
      tone: "neutral",
      hint: t("kube.componentMetricsHint"),
    },
    {
      name: "kube-controller-manager",
      state: t("kube.componentUnknown"),
      tone: "neutral",
      hint: t("kube.componentMetricsHint"),
    },
    {
      name: "etcd",
      state: t("kube.componentUnknown"),
      tone: "neutral",
      hint: t("kube.componentMetricsHint"),
    },
    {
      name: "kubelet",
      state: nodes
        ? `${nodes.ready}/${nodes.total}`
        : t("kube.componentUnknown"),
      tone: nodes
        ? nodes.ready === nodes.total
          ? "success"
          : "warning"
        : "neutral",
      hint: nodes
        ? t("kube.componentKubeletHint")
        : data?.failed.nodes
          ? t("kube.overviewLoadFailed")
          : t("kube.componentUnknownHint"),
    },
  ]

  return (
    <OverviewCard
      title={t("kube.components")}
      description={t("kube.componentsHint")}
    >
      {loading && !data ? (
        <CardSkeleton rows={5} />
      ) : data && Object.values(data.failed).every(Boolean) ? (
        <CardFailure onRetry={onRetry} />
      ) : (
        <div className="grid gap-2">
          <ul className="grid gap-1.5">
            {rows.map((row) => (
              <li
                key={row.name}
                className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-xs"
              >
                <span className="shrink-0 font-mono text-[11px] text-foreground/90">
                  {row.name}
                </span>
                <StatusBadge
                  value={row.state}
                  label={row.state}
                  tone={row.tone}
                  className="px-1.5 py-0"
                />
                <span className="min-w-0 flex-1 text-right text-[11px] break-words text-muted-foreground">
                  {row.hint}
                </span>
              </li>
            ))}
          </ul>
          <p className="text-[11px] leading-4 text-muted-foreground">
            {t("kube.componentsFooter")}
          </p>
        </div>
      )}
    </OverviewCard>
  )
}

interface ClusterTopCardProps {
  data: ClusterOverviewSnapshot | null
  loading: boolean
  onRetry: () => void
}

/** Top 榜单：节点 Pod 数与重启次数来自现有接口；CPU/内存用量留给指标数据源。 */
export function ClusterTopCard({
  data,
  loading,
  onRetry,
}: ClusterTopCardProps) {
  const { t } = useTranslation()
  const nodes = data?.nodes ?? null
  const pods = data?.pods ?? null
  const failed = data ? data.failed.nodes && data.failed.pods : false
  const maxPods = nodes?.topPods[0]?.podCount ?? 0

  return (
    <OverviewCard
      title={t("kube.topTitle")}
      description={t("kube.topHint")}
      badge={
        <span className="rounded border border-border/70 bg-muted/50 px-1.5 py-0.5 text-[10px] text-muted-foreground">
          {t("kube.noMetrics")}
        </span>
      }
    >
      {loading && !data ? (
        <CardSkeleton rows={5} />
      ) : failed ? (
        <CardFailure onRetry={onRetry} />
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          <section className="min-w-0">
            <h3 className="text-[11px] font-medium text-muted-foreground">
              {t("kube.topNodePods")}
            </h3>
            {nodes && nodes.topPods.length > 0 ? (
              <ul className="mt-2 grid gap-1.5">
                {nodes.topPods.map((node) => (
                  <li key={node.name} className="grid gap-1">
                    <div className="flex min-w-0 items-center gap-2 text-xs">
                      <span className="min-w-0 flex-1 truncate font-mono text-[11px]">
                        {node.name}
                      </span>
                      <span className="shrink-0 font-mono text-[11px] text-muted-foreground tabular-nums">
                        {node.podCount}
                      </span>
                    </div>
                    <div className="h-1 overflow-hidden rounded-full bg-muted">
                      <div
                        className="h-full rounded-full bg-primary/70"
                        style={{
                          width: `${maxPods > 0 ? Math.max(4, (node.podCount / maxPods) * 100) : 0}%`,
                        }}
                      />
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-2 text-xs text-muted-foreground">
                {t("kube.topEmpty")}
              </p>
            )}
          </section>
          <section className="min-w-0">
            <h3 className="text-[11px] font-medium text-muted-foreground">
              {t("kube.topRestarts")}
            </h3>
            {pods && pods.topRestarts.length > 0 ? (
              <ul className="mt-2 grid gap-1.5">
                {pods.topRestarts.map((pod) => (
                  <li
                    key={`${pod.namespace}/${pod.name}`}
                    className="flex min-w-0 items-center gap-2 rounded-md border border-border/60 bg-background/50 px-2 py-1"
                  >
                    <span className="min-w-0 flex-1 truncate font-mono text-[11px]">
                      {pod.namespace}/{pod.name}
                    </span>
                    <span className="sr-only">{t("kube.restarts")}</span>
                    <StatusBadge
                      value="RESTARTS"
                      label={String(pod.restarts)}
                      tone="warning"
                      className="px-1.5 py-0"
                    />
                  </li>
                ))}
              </ul>
            ) : (
              <p className="mt-2 text-xs text-muted-foreground">
                {t("kube.topEmpty")}
              </p>
            )}
          </section>
        </div>
      )}
    </OverviewCard>
  )
}

interface ClusterOverviewStatusProps {
  data: ClusterOverviewSnapshot | null
  namespace: string
  refreshing: boolean
  onRetry: () => void
}

/** 概况顶部的一行范围说明与全部失败时的整体提示。 */
export function ClusterOverviewStatus({
  data,
  namespace,
  refreshing,
  onRetry,
}: ClusterOverviewStatusProps) {
  const { t } = useTranslation()
  const allFailed = data ? Object.values(data.failed).every(Boolean) : false

  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2 text-[11px] text-muted-foreground">
        <span>
          {namespace
            ? t("kube.overviewScopeNamespace", { namespace })
            : t("kube.overviewScopeAll")}
        </span>
        <span className="flex items-center gap-1.5">
          {refreshing ? (
            <RefreshCw aria-hidden="true" className="size-3 animate-spin" />
          ) : null}
          {data
            ? t("kube.overviewRefreshedAt", {
                time: new Date(data.refreshedAt).toLocaleTimeString(),
              })
            : null}
        </span>
      </div>
      {allFailed ? (
        <div className="flex flex-wrap items-center gap-2 rounded-lg border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          <TriangleAlert aria-hidden="true" className="size-4" />
          <span className="min-w-0 flex-1">
            {t("kube.overviewUnreachable")}
          </span>
          <Button variant="outline" size="sm" onClick={onRetry}>
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
    </div>
  )
}
