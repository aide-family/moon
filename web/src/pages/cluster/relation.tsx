import { ChevronRight } from "lucide-react"
import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { useLocation, useNavigate } from "react-router-dom"

import type { KubernetesWorkloadRefItem } from "@/api/cluster"
import { DetailEmpty } from "@/components/detail-values"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/class-names"

import { WORKLOAD_KINDS } from "./detail/sections"

export interface ClusterRelationItem {
  key: string
  /** 主标识，例如 Pod / Service / 工作负载名称。 */
  title: ReactNode
  subtitle?: ReactNode
  status?: ReactNode
  meta?: ReactNode
  /** 给出后整行可点击，用于跳到对应资源视图。 */
  onSelect?: () => void
  /** 可点击行的可访问名称；未给出时退化为资源名文本。 */
  ariaLabel?: string
}

const rowClass =
  "flex min-w-0 items-start justify-between gap-3 rounded-lg border border-border/60 px-2.5 py-2"

/** 关联资源列表：结构同详情子项，可点击项带右箭头并跳转到对应资源视图。 */
export function ClusterRelationList({
  items,
}: {
  items: ClusterRelationItem[]
}) {
  if (items.length === 0) {
    return <DetailEmpty />
  }
  return (
    <ul className="grid min-w-0 gap-1.5">
      {items.map((item) => (
        <li key={item.key} className="min-w-0">
          {item.onSelect ? (
            <button
              type="button"
              aria-label={item.ariaLabel}
              onClick={item.onSelect}
              className={cn(
                rowClass,
                "w-full bg-muted/20 text-left transition-colors hover:border-primary/40 hover:bg-muted/60 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
              )}
            >
              <RelationContent item={item} />
              <ChevronRight
                aria-hidden="true"
                className="mt-1 size-4 shrink-0 text-muted-foreground/70"
              />
            </button>
          ) : (
            <div className={cn(rowClass, "bg-muted/20")}>
              <RelationContent item={item} />
            </div>
          )}
        </li>
      ))}
    </ul>
  )
}

function RelationContent({ item }: { item: ClusterRelationItem }) {
  return (
    <div className="grid min-w-0 flex-1 gap-1">
      <div className="flex min-w-0 flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <div className="min-w-0 text-sm leading-snug">{item.title}</div>
        {item.status ? (
          <div className="shrink-0 text-xs text-muted-foreground">
            {item.status}
          </div>
        ) : null}
      </div>
      {item.subtitle ? (
        <div className="min-w-0 font-mono text-[11px] leading-4 break-all text-muted-foreground">
          {item.subtitle}
        </div>
      ) : null}
      {item.meta ? (
        <div className="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-0.5 text-[11px] text-muted-foreground">
          {item.meta}
        </div>
      ) : null}
    </div>
  )
}

interface ClusterRelationBodyProps {
  loading: boolean
  failed: boolean
  error: string
  onRetry: () => void
  /** 加载成功后没有任何关联资源。 */
  empty: boolean
  children: ReactNode
}

/** 关联区块的加载 / 失败 / 空状态；成功时渲染调用方提供的列表。 */
export function ClusterRelationBody({
  loading,
  failed,
  error,
  onRetry,
  empty,
  children,
}: ClusterRelationBodyProps) {
  const { t } = useTranslation()
  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
  }
  if (failed) {
    return (
      <div className="flex flex-wrap items-center gap-2">
        <p className="min-w-0 text-sm break-words text-destructive">
          {error || t("kube.relationLoadFailed")}
        </p>
        <Button type="button" variant="outline" onClick={onRetry}>
          {t("action.retry")}
        </Button>
      </div>
    )
  }
  if (empty) {
    return <DetailEmpty />
  }
  return <>{children}</>
}

function isWorkloadKind(kind: string): boolean {
  return (WORKLOAD_KINDS as readonly string[]).includes(kind)
}

/** ConfigMap / Secret / PVC 详情里的「关联工作负载」列表。 */
export function ClusterRelatedWorkloadList({
  items,
  onOpen,
}: {
  items: KubernetesWorkloadRefItem[]
  onOpen: (item: KubernetesWorkloadRefItem) => void
}) {
  const { t } = useTranslation()
  return (
    <ClusterRelationList
      items={items.map((item) => ({
        key: `${item.kind}/${item.namespace}/${item.name}`,
        title: (
          <span className="font-mono text-xs break-all">
            {item.kind}/{item.name}
          </span>
        ),
        subtitle: <span className="font-mono">{item.namespace}</span>,
        onSelect: isWorkloadKind(item.kind) ? () => onOpen(item) : undefined,
        ariaLabel: isWorkloadKind(item.kind)
          ? t("kube.openResource", { name: item.name })
          : undefined,
      }))}
    />
  )
}

export interface ClusterRelationTarget {
  tab: string
  namespace?: string
  name?: string
  kind?: string
}

/** 从详情抽屉跳到另一个资源视图，把身份写进 URL 查询串。 */
export function useClusterRelationJump() {
  const navigate = useNavigate()
  const location = useLocation()
  return function jumpTo(target: ClusterRelationTarget) {
    const search = new URLSearchParams()
    search.set("tab", target.tab)
    if (target.namespace) {
      search.set("namespace", target.namespace)
    }
    if (target.name) {
      search.set("name", target.name)
    }
    if (target.kind) {
      search.set("kind", target.kind)
    }
    navigate({ pathname: location.pathname, search: search.toString() })
  }
}
