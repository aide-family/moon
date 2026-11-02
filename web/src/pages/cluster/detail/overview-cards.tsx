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

import type { KubernetesOverview, KubernetesPodPhaseCount } from "@/api/cluster"
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
import type { ClusterMetricsState } from "./use-cluster-overview"

type KpiTone = "primary" | "success" | "warning" | "danger" | "neutral"

// 颜色只使用语义 token：运行/成功、等待、失败、未知分别映射状态色。
const toneChips: Record<KpiTone, string> = {
  primary: "bg-primary-soft text-primary-ink",
  success: "bg-success-soft text-success",
  warning: "bg-warning-soft text-warning",
  danger: "bg-danger-soft text-danger",
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
  Running: "bg-success",
  Pending: "bg-warning",
  Succeeded: "bg-info",
  Failed: "bg-danger",
  Unknown: "bg-muted-foreground/50",
}

const SECTION_LABEL_KEYS: Record<string, string> = {
  nodes: "kube.sectionNodes",
  workloads: "kube.sectionWorkloads",
  pods: "kube.sectionPods",
  jobs: "kube.sectionJobs",
  services: "kube.sectionServices",
  ingresses: "kube.sectionIngresses",
}

const ALL_SECTIONS = [
  "nodes",
  "workloads",
  "pods",
  "jobs",
  "services",
  "ingresses",
]

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

function readNumber(value: number | string | undefined): number {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

function formatPercent(value: number | undefined): string {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "—"
  }
  return `${Math.round(value * 10) / 10}%`
}

function formatCores(value: number | undefined, t: TFunction): string {
  if (typeof value !== "number" || !Number.isFinite(value)) {
    return "—"
  }
  return `${Math.round(value * 100) / 100} ${t("kube.unitCores")}`
}

function sectionState(
  data: KubernetesOverview | null,
  section: string
): "ok" | "denied" | "failed" {
  if (data?.deniedSections?.includes(section)) {
    return "denied"
  }
  if (data?.failedSections?.includes(section)) {
    return "failed"
  }
  return "ok"
}

function sectionLabel(t: TFunction, section: string): string {
  const key = SECTION_LABEL_KEYS[section]
  return key ? t(key) : section
}

interface SectionNoticeProps {
  state: "denied" | "failed"
  onRetry: () => void
}

/** 区块级降级：无权限只说明；查询失败给重试，不阻塞其他卡片。 */
function SectionNotice({ state, onRetry }: SectionNoticeProps) {
  const { t } = useTranslation()
  if (state === "denied") {
    return (
      <div className="flex flex-wrap items-center gap-2 rounded-lg border border-border/60 bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
        <span>{t("kube.overviewSectionDenied")}</span>
      </div>
    )
  }
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-lg border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive">
      <TriangleAlert aria-hidden="true" className="size-4 shrink-0" />
      <span className="min-w-0 flex-1">{t("kube.overviewSectionFailed")}</span>
      <Button variant="ghost" size="sm" onClick={onRetry}>
        {t("action.retry")}
      </Button>
    </div>
  )
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

interface OverviewSectionProps {
  data: KubernetesOverview | null
  loading: boolean
  onRetry: () => void
}

type ClusterOverviewKpisProps = OverviewSectionProps

/** 四条核心指标：节点就绪、Pod 总数、工作负载就绪/期望、Ingress 数量。 */
export function ClusterOverviewKpis({
  data,
  loading,
  onRetry,
}: ClusterOverviewKpisProps) {
  const { t } = useTranslation()
  const nodesState = sectionState(data, "nodes")
  const podsState = sectionState(data, "pods")
  const workloadsState = sectionState(data, "workloads")
  const ingressesState = sectionState(data, "ingresses")
  const podsRunning =
    data?.podPhases?.find((item) => item.phase === "Running")?.count ?? 0
  const podsPending =
    data?.podPhases?.find((item) => item.phase === "Pending")?.count ?? 0
  const workloadReady = (data?.workloads ?? []).reduce(
    (sum, row) => sum + row.ready,
    0
  )
  const workloadDesired = (data?.workloads ?? []).reduce(
    (sum, row) => sum + row.desired,
    0
  )

  function fallbackHint(state: "ok" | "denied" | "failed"): string {
    if (state === "denied") {
      return t("kube.overviewSectionDenied")
    }
    if (state === "failed") {
      return t("kube.overviewSectionFailed")
    }
    return t("kube.overviewWaiting")
  }

  const tiles: ClusterKpiTileProps[] = [
    {
      label: t("kube.overviewNodesReady"),
      value:
        nodesState === "ok" && data
          ? `${data.nodeReady}/${data.nodeTotal}`
          : "—",
      hint:
        nodesState === "ok" && data
          ? data.nodeNotReady > 0
            ? t("kube.overviewNodesNotReady", { count: data.nodeNotReady })
            : t("kube.overviewNodesAllReady")
          : fallbackHint(nodesState),
      icon: Server,
      tone:
        nodesState === "ok" && data
          ? data.nodeNotReady > 0
            ? "warning"
            : "success"
          : nodesState === "failed"
            ? "danger"
            : "neutral",
      loading,
    },
    {
      label: t("kube.overviewPodsRunning"),
      value: podsState === "ok" && data ? String(podsRunning) : "—",
      hint:
        podsState === "ok" && data
          ? podsPending > 0
            ? t("kube.overviewPodsPending", {
                count: podsPending,
                total: data.podTotal,
              })
            : t("kube.overviewPodsHint", { total: data.podTotal })
          : fallbackHint(podsState),
      icon: Gauge,
      tone:
        podsState === "ok"
          ? podsPending > 0
            ? "warning"
            : "success"
          : podsState === "failed"
            ? "danger"
            : "neutral",
      loading,
    },
    {
      label: t("kube.overviewWorkloadHealthy"),
      value:
        workloadsState === "ok" && data
          ? `${workloadReady}/${workloadDesired}`
          : "—",
      hint:
        workloadsState === "ok" && data
          ? t("kube.overviewWorkloadsHint", {
              ready: workloadReady,
              desired: workloadDesired,
            })
          : fallbackHint(workloadsState),
      icon: Cpu,
      tone:
        workloadsState === "ok"
          ? workloadReady >= workloadDesired
            ? "success"
            : "warning"
          : workloadsState === "failed"
            ? "danger"
            : "neutral",
      loading,
    },
    {
      label: t("kube.overviewIngresses"),
      value: ingressesState === "ok" && data ? String(data.ingressTotal) : "—",
      hint:
        ingressesState === "ok"
          ? t("kube.overviewIngressesHint")
          : fallbackHint(ingressesState),
      icon: Database,
      tone: ingressesState === "failed" ? "danger" : "primary",
      loading,
    },
  ]

  // 失败区块在 KPI 上方汇总成一行提示，点击可重试整页。
  const failedSections = data?.failedSections ?? []
  const deniedSections = data?.deniedSections ?? []

  return (
    <div className="grid gap-2">
      {failedSections.length > 0 || deniedSections.length > 0 ? (
        <div className="flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
          {failedSections.length > 0 ? (
            <>
              <span className="inline-flex items-center gap-1 text-destructive">
                <TriangleAlert aria-hidden="true" className="size-3.5" />
                {t("kube.overviewFailedSections", {
                  count: failedSections.length,
                })}
              </span>
              <Button
                variant="ghost"
                size="sm"
                className="h-6 px-1.5"
                onClick={onRetry}
              >
                {t("action.retry")}
              </Button>
            </>
          ) : null}
          {deniedSections.length > 0 ? (
            <span>
              {t("kube.overviewDeniedSections", {
                count: deniedSections.length,
              })}
            </span>
          ) : null}
        </div>
      ) : null}
      <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
        {tiles.map((tile) => (
          <ClusterKpiTile key={tile.label} {...tile} />
        ))}
      </div>
    </div>
  )
}

type ClusterPodPhaseCardProps = OverviewSectionProps

/** Pod 状态分布：阶段堆叠条 + 计数图例，全量口径（服务端不再采样）。 */
export function ClusterPodPhaseCard({
  data,
  loading,
  onRetry,
}: ClusterPodPhaseCardProps) {
  const { t } = useTranslation()
  const state = sectionState(data, "pods")
  const phases: KubernetesPodPhaseCount[] = data?.podPhases ?? []
  const total = data?.podTotal ?? 0
  const ariaLabel = phases
    .map((item) => `${phaseLabel(t, item.phase)} ${item.count}`)
    .join(", ")

  return (
    <OverviewCard
      title={t("kube.podDistribution")}
      description={t("kube.podDistributionHint")}
      badge={
        state === "ok" && total > 0 ? (
          <span className="rounded border border-border/70 bg-muted/50 px-1.5 py-0.5 text-[10px] text-muted-foreground tabular-nums">
            {t("kube.podDistributionFull", { count: total })}
          </span>
        ) : null
      }
    >
      {loading && !data ? (
        <CardSkeleton rows={4} />
      ) : state !== "ok" ? (
        <SectionNotice state={state} onRetry={onRetry} />
      ) : total === 0 ? (
        <CardEmpty text={t("kube.podDistributionEmpty")} />
      ) : (
        <div className="grid gap-3">
          <div
            role="img"
            aria-label={ariaLabel}
            className="flex h-2.5 w-full overflow-hidden rounded-full bg-muted"
          >
            {phases.map((item) => (
              <div
                key={item.phase}
                className={phaseBar(item.phase)}
                style={{
                  width: `${(item.count / Math.max(total, 1)) * 100}%`,
                }}
                title={`${phaseLabel(t, item.phase)} ${item.count}`}
              />
            ))}
          </div>
          <ul className="grid grid-cols-2 gap-x-4 gap-y-1.5 sm:grid-cols-3">
            {phases.map((item) => (
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
        </div>
      )}
    </OverviewCard>
  )
}

interface ClusterWorkloadHealthCardProps extends OverviewSectionProps {
  onOpenKind: (kind: string) => void
}

/** 工作负载健康：按类型看就绪/期望，点击直接跳到对应类型列表。 */
export function ClusterWorkloadHealthCard({
  data,
  loading,
  onRetry,
  onOpenKind,
}: ClusterWorkloadHealthCardProps) {
  const { t } = useTranslation()
  const state = sectionState(data, "workloads")
  const rows = data?.workloads ?? []
  const total = rows.reduce((sum, row) => sum + row.total, 0)
  const ready = rows.reduce((sum, row) => sum + row.ready, 0)
  const desired = rows.reduce((sum, row) => sum + row.desired, 0)

  return (
    <OverviewCard
      title={t("kube.workloadHealth")}
      description={t("kube.workloadHealthHint")}
      badge={
        state === "ok" ? (
          <span
            className={cn(
              "rounded-full border px-2 py-0.5 text-[11px]",
              ready >= desired
                ? "border-success-border bg-success-soft text-success"
                : "border-warning-border bg-warning-soft text-warning"
            )}
          >
            {ready}/{desired}
          </span>
        ) : null
      }
    >
      {loading && !data ? (
        <CardSkeleton rows={4} />
      ) : state !== "ok" ? (
        <SectionNotice state={state} onRetry={onRetry} />
      ) : total === 0 ? (
        <CardEmpty text={t("kube.workloadHealthEmpty")} />
      ) : (
        <ul className="grid gap-1">
          {rows.map((row) => {
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
                          healthy ? "bg-success" : "bg-warning"
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

type ClusterCapacityCardProps = OverviewSectionProps

/** 可分配 CPU / 内存：只累加 Ready 且可调度节点，口径与列表 Capacity 不同。 */
export function ClusterCapacityCard({
  data,
  loading,
  onRetry,
}: ClusterCapacityCardProps) {
  const { t } = useTranslation()
  const state = sectionState(data, "nodes")
  const cpuMillis = readNumber(data?.allocatableCpuMillis)
  const memoryBytes = readNumber(data?.allocatableMemoryBytes)

  return (
    <OverviewCard
      title={t("kube.capacityTitle")}
      description={t("kube.capacityAllocatableHint")}
      badge={
        state === "ok" && data ? (
          <span className="rounded-full border border-border/70 bg-muted/50 px-2 py-0.5 text-[11px] text-muted-foreground tabular-nums">
            {t("kube.capacityReadyNodes", {
              ready: data.nodeReady,
              total: data.nodeTotal,
            })}
          </span>
        ) : null
      }
    >
      {loading && !data ? (
        <CardSkeleton rows={3} />
      ) : state !== "ok" ? (
        <SectionNotice state={state} onRetry={onRetry} />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          <div className="rounded-lg border border-border/60 bg-background/50 p-3">
            <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              <Cpu aria-hidden="true" className="size-3.5" />
              {t("kube.capacityAllocatableCpu")}
            </p>
            <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
              {formatCpuCapacity(cpuMillis, t)}
            </p>
          </div>
          <div className="rounded-lg border border-border/60 bg-background/50 p-3">
            <p className="flex items-center gap-1.5 text-[11px] text-muted-foreground">
              <MemoryStick aria-hidden="true" className="size-3.5" />
              {t("kube.capacityAllocatableMemory")}
            </p>
            <p className="mt-1 font-heading text-xl font-semibold tabular-nums">
              {formatMemoryCapacity(memoryBytes)}
            </p>
          </div>
        </div>
      )}
    </OverviewCard>
  )
}

interface ClusterMetricsCardProps {
  metrics: ClusterMetricsState
  datasourceName: string
  canBind: boolean
  onBind: () => void
  loading: boolean
}

/** 指标卡：未绑定/部分失败/数据源不可用都用状态说明，不弹服务端错误。 */
export function ClusterMetricsCard({
  metrics,
  datasourceName,
  canBind,
  onBind,
  loading,
}: ClusterMetricsCardProps) {
  const { t } = useTranslation()
  const status = metrics.data?.status ?? ""
  const missing = metrics.data?.unavailableMetrics ?? []

  let body: ReactNode
  if (metrics.loading || (loading && !metrics.data)) {
    body = <CardSkeleton rows={4} />
  } else if (metrics.failed) {
    body = (
      <div className="grid gap-2">
        <p className="text-xs text-destructive">
          {metrics.error || t("kube.metricsLoadFailed")}
        </p>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" size="sm" onClick={metrics.retry}>
            {t("action.retry")}
          </Button>
          {canBind ? (
            <Button variant="outline" size="sm" onClick={onBind}>
              {t("kube.metricsBind")}
            </Button>
          ) : null}
        </div>
      </div>
    )
  } else if (status === "UNBOUND") {
    body = (
      <div className="grid gap-2">
        <p className="text-sm font-medium">{t("kube.metricsUnbound")}</p>
        <p className="text-xs leading-relaxed text-muted-foreground">
          {t("kube.metricsUnboundHint")}
        </p>
        {canBind ? (
          <div>
            <Button variant="outline" size="sm" onClick={onBind}>
              {t("kube.metricsBind")}
            </Button>
          </div>
        ) : null}
      </div>
    )
  } else if (status === "FAILED") {
    body = (
      <div className="grid gap-2">
        <p className="text-xs break-words text-destructive">
          {metrics.data?.message || t("kube.metricsLoadFailed")}
        </p>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" size="sm" onClick={metrics.retry}>
            {t("action.retry")}
          </Button>
          {canBind ? (
            <Button variant="outline" size="sm" onClick={onBind}>
              {t("kube.metricsBind")}
            </Button>
          ) : null}
        </div>
      </div>
    )
  } else {
    const data = metrics.data
    body = (
      <div className="grid gap-3">
        <div className="grid gap-3 sm:grid-cols-2">
          {(
            [
              [
                "kube.metricsCpuUsage",
                formatPercent(data?.nodeCpuUsagePercent),
              ],
              [
                "kube.metricsMemoryUsage",
                formatPercent(data?.nodeMemoryUsagePercent),
              ],
              ["kube.metricsPodCpu", formatCores(data?.podCpuUsageCores, t)],
              [
                "kube.metricsPodMemory",
                typeof data?.podMemoryUsageBytes === "number"
                  ? formatMemoryCapacity(data.podMemoryUsageBytes)
                  : "—",
              ],
            ] as const
          ).map(([label, value]) => (
            <div
              key={label}
              className="rounded-lg border border-border/60 bg-background/50 px-3 py-2"
            >
              <p className="text-[11px] text-muted-foreground">{t(label)}</p>
              <p className="mt-0.5 font-heading text-lg font-semibold tabular-nums">
                {value}
              </p>
            </div>
          ))}
        </div>
        {status === "PARTIAL" ? (
          <div className="grid gap-1 rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-xs text-warning">
            <p className="break-words">
              {metrics.data?.message || t("kube.metricsPartial")}
            </p>
            {missing.length > 0 ? (
              <p className="break-words">
                {t("kube.metricsUnavailable", { items: missing.join(", ") })}
              </p>
            ) : null}
            <div>
              <Button variant="outline" size="sm" onClick={metrics.retry}>
                {t("action.retry")}
              </Button>
            </div>
          </div>
        ) : null}
      </div>
    )
  }

  return (
    <OverviewCard
      title={t("kube.metricsTitle")}
      description={t("kube.metricsHint")}
      badge={
        datasourceName ? (
          <span className="rounded-full border border-border/70 bg-muted/50 px-2 py-0.5 text-[11px] text-muted-foreground">
            {t("kube.metricsBoundTo", { name: datasourceName })}
          </span>
        ) : null
      }
    >
      {body}
    </OverviewCard>
  )
}

type ClusterJobsCardProps = OverviewSectionProps

interface MetricCellProps {
  label: string
  value: number | string
  state: "ok" | "denied" | "failed"
}

/** Job / Service 数量单元格；无权限与失败分别给出明确文案。 */
function MetricCell({ label, value, state }: MetricCellProps) {
  const { t } = useTranslation()
  if (state === "denied") {
    return (
      <div className="rounded-lg border border-border/60 bg-muted/40 px-3 py-2">
        <p className="text-[11px] text-muted-foreground">{label}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">
          {t("kube.overviewSectionDenied")}
        </p>
      </div>
    )
  }
  return (
    <div className="rounded-lg border border-border/60 bg-background/50 px-3 py-2">
      <p className="text-[11px] text-muted-foreground">{label}</p>
      <p className="mt-0.5 font-heading text-lg font-semibold tabular-nums">
        {state === "failed" ? "—" : value}
      </p>
    </div>
  )
}

/** Job 与 Service 的数量汇总；两颗区块各自按权限/失败降级。 */
export function ClusterJobsCard({
  data,
  loading,
  onRetry,
}: ClusterJobsCardProps) {
  const { t } = useTranslation()
  const jobsState = sectionState(data, "jobs")
  const servicesState = sectionState(data, "services")

  return (
    <OverviewCard
      title={t("kube.jobsServicesTitle")}
      description={t("kube.jobsServicesHint")}
    >
      {loading && !data ? (
        <CardSkeleton rows={3} />
      ) : (
        <div className="grid gap-3">
          {jobsState === "failed" || servicesState === "failed" ? (
            <div className="flex flex-wrap items-center gap-2 text-xs text-destructive">
              <TriangleAlert aria-hidden="true" className="size-4 shrink-0" />
              <span className="min-w-0 flex-1">
                {t("kube.overviewSectionFailed")}
              </span>
              <Button variant="ghost" size="sm" onClick={onRetry}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          <div className="grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
            <MetricCell
              label={t("kube.jobTotal")}
              value={data?.jobTotal ?? 0}
              state={jobsState}
            />
            <MetricCell
              label={t("kube.jobActive")}
              value={data?.jobActive ?? 0}
              state={jobsState}
            />
            <MetricCell
              label={t("kube.jobFailed")}
              value={data?.jobFailed ?? 0}
              state={jobsState}
            />
            <MetricCell
              label={t("kube.serviceTotal")}
              value={data?.serviceTotal ?? 0}
              state={servicesState}
            />
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

interface ClusterComponentCardProps extends OverviewSectionProps {
  refreshedAt: number
}

/**
 * 基础组件状态：API Server 版本与资源查询是否可用；kubelet 由节点 Ready 推断；
 * scheduler / controller-manager / etcd 仍明确标注未接入。
 */
export function ClusterComponentCard({
  data,
  loading,
  refreshedAt,
}: ClusterComponentCardProps) {
  const { t } = useTranslation()
  const nodesState = sectionState(data, "nodes")
  const allFailed =
    data !== null &&
    ALL_SECTIONS.every((section) => sectionState(data, section) === "failed")
  const updatedAt =
    refreshedAt > 0 ? new Date(refreshedAt).toLocaleTimeString() : ""

  const rows: ComponentRow[] = [
    {
      name: "kube-apiserver",
      state: allFailed
        ? t("kube.componentUnknown")
        : t("kube.componentReachable"),
      tone: allFailed ? "neutral" : "success",
      hint: data?.apiServerVersion
        ? `${data.apiServerVersion}${updatedAt ? ` · ${updatedAt}` : ""}`
        : allFailed
          ? t("kube.componentUnknownHint")
          : t("kube.componentApiServerHint"),
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
      state:
        nodesState === "ok" && data
          ? `${data.nodeReady}/${data.nodeTotal}`
          : t("kube.componentUnknown"),
      tone:
        nodesState === "ok" && data
          ? data.nodeNotReady > 0
            ? "warning"
            : "success"
          : "neutral",
      hint:
        nodesState === "ok" && data
          ? t("kube.componentKubeletHint")
          : nodesState === "failed"
            ? t("kube.overviewSectionFailed")
            : nodesState === "denied"
              ? t("kube.overviewSectionDenied")
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
      ) : allFailed ? (
        <div className="flex flex-wrap items-center gap-2 py-2">
          <TriangleAlert
            aria-hidden="true"
            className="size-4 text-destructive"
          />
          <p className="text-xs text-destructive">
            {t("kube.overviewLoadFailed")}
          </p>
        </div>
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

interface ClusterOverviewStatusProps {
  data: KubernetesOverview | null
  namespace: string
  refreshing: boolean
  failed: boolean
  error: string
  onRetry: () => void
}

/** 概况顶部：范围与口径说明、生成时间、区块级降级与整体失败提示。 */
export function ClusterOverviewStatus({
  data,
  namespace,
  refreshing,
  failed,
  error,
  onRetry,
}: ClusterOverviewStatusProps) {
  const { t } = useTranslation()
  const generatedAt = readNumber(data?.generatedAt)
  const denied = data?.deniedSections ?? []
  const failedSections = data?.failedSections ?? []

  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2 text-[11px] text-muted-foreground">
        <span className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <span>
            {namespace
              ? t("kube.overviewScopeNamespace", { namespace })
              : t("kube.overviewScopeAll")}
          </span>
          <span aria-hidden="true">·</span>
          <span>{t("kube.overviewScopeClusterScope")}</span>
        </span>
        <span className="flex items-center gap-1.5">
          {refreshing ? (
            <RefreshCw aria-hidden="true" className="size-3 animate-spin" />
          ) : null}
          {generatedAt > 0
            ? t("kube.overviewGeneratedAt", {
                time: new Date(generatedAt * 1000).toLocaleTimeString(),
              })
            : null}
        </span>
      </div>
      {failed ? (
        <div className="flex flex-wrap items-center gap-2 rounded-lg border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs text-destructive">
          <TriangleAlert aria-hidden="true" className="size-4" />
          <span className="min-w-0 flex-1">
            {error || t("kube.overviewUnreachable")}
          </span>
          <Button variant="outline" size="sm" onClick={onRetry}>
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
      {!failed && (denied.length > 0 || failedSections.length > 0) ? (
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-[11px] text-warning">
          {denied.length > 0 ? (
            <span>
              {t("kube.overviewDeniedSectionsDetail", {
                sections: denied
                  .map((section) => sectionLabel(t, section))
                  .join(", "),
              })}
            </span>
          ) : null}
          {failedSections.length > 0 ? (
            <span>
              {t("kube.overviewFailedSectionsDetail", {
                sections: failedSections
                  .map((section) => sectionLabel(t, section))
                  .join(", "),
              })}
            </span>
          ) : null}
          {failedSections.length > 0 ? (
            <Button variant="outline" size="sm" onClick={onRetry}>
              {t("action.retry")}
            </Button>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
