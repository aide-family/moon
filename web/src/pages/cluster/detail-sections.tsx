import { TriangleAlert } from "lucide-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import type {
  KubernetesConditionItem,
  KubernetesContainerStatusItem,
  KubernetesEventItem,
  KubernetesOwnerReferenceItem,
  KubernetesTaintItem,
  KubernetesVolumeItem,
} from "@/api/cluster"
import { StatusBadge, type BadgeTone } from "@/components/status-badge"
import { cn } from "@/lib/class-names"

import { formatClusterTime } from "./format"

/** 空状态占位，与关联资源区块保持一致。 */
function EmptyHint() {
  const { t } = useTranslation()
  return (
    <p className="text-sm text-muted-foreground">{t("kube.detailEmpty")}</p>
  )
}

const rowClass =
  "flex min-w-0 items-start justify-between gap-3 rounded-lg border border-border/60 bg-muted/20 px-2.5 py-2"

/** 条件列表：类型 + 状态标签 + 原因/说明 + 最近变更时间。 */
export function ConditionList({
  conditions,
}: {
  conditions: KubernetesConditionItem[] | undefined
}) {
  const { t } = useTranslation()
  if (!conditions || conditions.length === 0) {
    return <EmptyHint />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {conditions.map((condition) => (
        <li
          key={`${condition.type}-${condition.lastTransitionAt}`}
          className={rowClass}
        >
          <div className="grid min-w-0 flex-1 gap-1">
            <div className="flex min-w-0 flex-wrap items-center gap-2">
              <span className="font-mono text-xs font-medium break-all">
                {condition.type}
              </span>
              <StatusBadge
                value={condition.status}
                label={condition.status}
                tone={conditionTone(condition.status)}
              />
            </div>
            {condition.message || condition.reason ? (
              <p className="text-xs break-words text-muted-foreground">
                {condition.reason ? `${condition.reason}: ` : ""}
                {condition.message || "—"}
              </p>
            ) : null}
          </div>
          <span className="shrink-0 text-[11px] text-muted-foreground tabular-nums">
            {formatClusterTime(condition.lastTransitionAt) ||
              t("kube.neverUpdated")}
          </span>
        </li>
      ))}
    </ul>
  )
}

function conditionTone(status: string): BadgeTone {
  if (status === "True") {
    return "success"
  }
  if (status === "False") {
    return "danger"
  }
  return "warning"
}

/** 节点污点列表。 */
export function TaintList({
  taints,
}: {
  taints: KubernetesTaintItem[] | undefined
}) {
  if (!taints || taints.length === 0) {
    return <EmptyHint />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {taints.map((taint) => (
        <li key={`${taint.key}-${taint.effect}`} className={rowClass}>
          <div className="grid min-w-0 flex-1 gap-0.5">
            <span className="font-mono text-xs break-all">{taint.key}</span>
            <span className="font-mono text-[11px] break-all text-muted-foreground">
              {taint.value || "—"}
            </span>
          </div>
          <StatusBadge
            value={taint.effect}
            label={taint.effect}
            tone="warning"
            className="shrink-0"
          />
        </li>
      ))}
    </ul>
  )
}

/** 容量 / 可分配量：按资源键并排展示 capacity 与 allocatable。 */
export function ResourceQuantityList({
  capacity,
  allocatable,
}: {
  capacity: Record<string, string> | undefined
  allocatable: Record<string, string> | undefined
}) {
  const { t } = useTranslation()
  const keys = [
    ...new Set([
      ...Object.keys(capacity ?? {}),
      ...Object.keys(allocatable ?? {}),
    ]),
  ].sort()
  if (keys.length === 0) {
    return <EmptyHint />
  }
  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[24rem] text-sm">
        <thead>
          <tr className="border-b border-border/70 text-left text-xs text-muted-foreground">
            <th scope="col" className="py-1.5 pr-3 font-medium">
              {t("kube.resourceName")}
            </th>
            <th scope="col" className="py-1.5 pr-3 text-right font-medium">
              {t("kube.capacity")}
            </th>
            <th scope="col" className="py-1.5 text-right font-medium">
              {t("kube.allocatable")}
            </th>
          </tr>
        </thead>
        <tbody>
          {keys.map((key) => (
            <tr key={key} className="border-b border-border/60 last:border-0">
              <th
                scope="row"
                className="py-1.5 pr-3 text-left font-mono text-xs font-normal break-all"
              >
                {key}
              </th>
              <td className="py-1.5 pr-3 text-right font-mono text-xs tabular-nums">
                {capacity?.[key] || "—"}
              </td>
              <td className="py-1.5 text-right font-mono text-xs tabular-nums">
                {allocatable?.[key] || "—"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function eventTone(type: string): BadgeTone {
  return type === "Warning" ? "danger" : "info"
}

/** 最近事件：类型、原因、说明、次数、来源与最近发生时间。 */
export function EventList({
  events,
}: {
  events: KubernetesEventItem[] | undefined
}) {
  const { t } = useTranslation()
  if (!events || events.length === 0) {
    return <EmptyHint />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {events.map((event, index) => (
        <li
          key={`${event.reason}-${event.lastSeenAt}-${index}`}
          className={cn(rowClass, "grid gap-1")}
        >
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <StatusBadge value={event.type} tone={eventTone(event.type)} />
            <span className="min-w-0 text-xs font-medium break-words">
              {event.reason || "—"}
            </span>
            <span className="ml-auto shrink-0 font-mono text-[11px] text-muted-foreground tabular-nums">
              {formatClusterTime(event.lastSeenAt) || "—"}
            </span>
          </div>
          <p className="text-xs leading-relaxed break-words text-muted-foreground">
            {event.message || "—"}
          </p>
          <p className="flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-muted-foreground">
            <span>
              {t("kube.eventCount")}: {event.count}
            </span>
            {event.source ? (
              <span>
                {t("kube.eventSource")}: {event.source}
              </span>
            ) : null}
            {event.firstSeenAt ? (
              <span>
                {t("kube.eventFirstSeen")}:{" "}
                {formatClusterTime(event.firstSeenAt)}
              </span>
            ) : null}
          </p>
        </li>
      ))}
    </ul>
  )
}

/** 容器状态列表：普通容器与初始化容器共用。 */
export function ContainerStatusList({
  containers,
}: {
  containers: KubernetesContainerStatusItem[] | undefined
}) {
  const { t } = useTranslation()
  if (!containers || containers.length === 0) {
    return <EmptyHint />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {containers.map((container) => (
        <li key={container.name} className={cn(rowClass, "grid gap-1")}>
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <span className="min-w-0 font-mono text-xs font-medium break-all">
              {container.name}
            </span>
            <StatusBadge
              value={container.state}
              label={container.state || "—"}
              tone={containerStateTone(container.state)}
            />
            <span className="ml-auto shrink-0 text-[11px] text-muted-foreground tabular-nums">
              {t("kube.restarts")}: {container.restarts}
            </span>
          </div>
          {container.image ? (
            <p className="font-mono text-[11px] break-all text-muted-foreground">
              {container.image}
            </p>
          ) : null}
          {container.stateReason || container.stateMessage ? (
            <p className="text-xs break-words text-muted-foreground">
              {container.stateReason ? `${container.stateReason}: ` : ""}
              {container.stateMessage || "—"}
            </p>
          ) : null}
          <p className="flex flex-wrap gap-x-3 gap-y-0.5 text-[11px] text-muted-foreground">
            <span>
              {t("kube.ready")}:{" "}
              {container.ready ? t("kube.yes") : t("kube.no")}
            </span>
            {container.startedAt ? (
              <span>
                {t("kube.containerStartedAt")}:{" "}
                {formatClusterTime(container.startedAt)}
              </span>
            ) : null}
            {container.ports?.length ? (
              <span>
                {t("kube.containerPorts")}: {container.ports.join(", ")}
              </span>
            ) : null}
          </p>
        </li>
      ))}
    </ul>
  )
}

function containerStateTone(state: string): BadgeTone {
  if (state === "RUNNING") {
    return "success"
  }
  if (state === "WAITING") {
    return "warning"
  }
  if (state === "TERMINATED") {
    return "danger"
  }
  return "neutral"
}

/** 存储卷列表：只展示类型与来源摘要，不读取任何敏感值。 */
export function VolumeList({
  volumes,
}: {
  volumes: KubernetesVolumeItem[] | undefined
}) {
  const { t } = useTranslation()
  if (!volumes || volumes.length === 0) {
    return <EmptyHint />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {volumes.map((volume) => (
        <li key={volume.name} className={rowClass}>
          <div className="grid min-w-0 flex-1 gap-0.5">
            <span className="font-mono text-xs break-all">{volume.name}</span>
            <span className="text-[11px] break-words text-muted-foreground">
              {volume.summary || "—"}
            </span>
          </div>
          <div className="flex shrink-0 flex-wrap items-center justify-end gap-1.5">
            <span className="rounded border border-border/70 bg-muted/50 px-1.5 py-0.5 text-[10px] text-muted-foreground">
              {volume.type || "—"}
            </span>
            {volume.readOnly ? (
              <StatusBadge
                value="READONLY"
                label={t("kube.volumeReadOnly")}
                tone="warning"
              />
            ) : null}
          </div>
        </li>
      ))}
    </ul>
  )
}

/** 归属（ownerReferences）列表。 */
export function OwnerReferenceList({
  owners,
}: {
  owners: KubernetesOwnerReferenceItem[] | undefined
}) {
  const { t } = useTranslation()
  if (!owners || owners.length === 0) {
    return <EmptyHint />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {owners.map((owner) => (
        <li key={`${owner.kind}/${owner.uid}`} className={rowClass}>
          <div className="grid min-w-0 flex-1 gap-0.5">
            <span className="text-sm break-words">
              <span className="font-mono text-xs text-muted-foreground">
                {owner.kind}/
              </span>
              {owner.name}
            </span>
            <span className="font-mono text-[11px] break-all text-muted-foreground">
              {owner.uid}
            </span>
          </div>
          {owner.controller ? (
            <StatusBadge
              value="CONTROLLER"
              label={t("kube.ownerController")}
              tone="primary"
              className="shrink-0"
            />
          ) : null}
        </li>
      ))}
    </ul>
  )
}

/** 详情接口的降级提示：events 等非致命问题。 */
export function DetailWarnings({
  warnings,
}: {
  warnings: string[] | undefined
}) {
  const { t } = useTranslation()
  if (!warnings || warnings.length === 0) {
    return null
  }
  return (
    <div className="flex flex-wrap items-start gap-2 rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-xs text-warning">
      <TriangleAlert aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
      <div className="grid min-w-0 gap-1">
        <p className="font-medium">{t("kube.detailWarnings")}</p>
        <ul className="grid gap-0.5">
          {warnings.map((warning) => (
            <li key={warning} className="break-words">
              {warning}
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

/** 详情区块统一加载/失败外壳：失败可重试，成功交给调用方渲染。 */
export function DetailSectionBody({
  loading,
  failed,
  error,
  onRetry,
  children,
}: {
  loading: boolean
  failed: boolean
  error: string
  onRetry: () => void
  children: ReactNode
}) {
  const { t } = useTranslation()
  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
  }
  if (failed) {
    return (
      <div className="flex flex-wrap items-center gap-2">
        <p className="min-w-0 text-sm break-words text-destructive">
          {error || t("kube.detailLoadFailed")}
        </p>
        <button
          type="button"
          className="text-sm text-primary-ink underline-offset-4 hover:underline"
          onClick={onRetry}
        >
          {t("action.retry")}
        </button>
      </div>
    )
  }
  return <>{children}</>
}
