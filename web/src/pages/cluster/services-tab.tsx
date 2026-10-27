import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getKubernetesServiceDetail,
  listKubernetesServices,
} from "@/api/cluster"
import type {
  GetKubernetesServiceDetailResult,
  KubernetesServiceItem,
  KubernetesServicePortItem,
} from "@/api/cluster"
import { CodeBlock } from "@/components/code-block"
import {
  DetailCopyText,
  DetailEmpty,
  DetailItemList,
  DetailTagList,
} from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions } from "@/components/row-actions"
import { StatusBadge, type BadgeTone } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { usePagedList } from "@/hooks/use-paged-list"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { ClusterDetailSheet } from "./cluster-detail-sheet"
import { formatClusterAge, formatClusterTime } from "./format"
import { ClusterRelationBody, ClusterRelationList } from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

interface KubernetesServicesTabProps {
  clusterUid: string
  namespace: string
  name: string
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

function serviceTone(type: string): BadgeTone {
  switch (type) {
    case "ClusterIP":
      return "info"
    case "NodePort":
      return "warning"
    case "LoadBalancer":
      return "primary"
    case "ExternalName":
      return "violet"
    default:
      return "neutral"
  }
}

function portLabel(port: KubernetesServicePortItem): string {
  const target = port.targetPort || "—"
  const protocol = port.protocol || "TCP"
  const nodePort = port.nodePort ? ` :${port.nodePort}` : ""
  return `${port.port} → ${target}/${protocol}${nodePort}`
}

/** 选择器是 JSON 对象字符串，展示前格式化；解析失败时保留原文。 */
function formatSelector(selector: string): string {
  try {
    return JSON.stringify(JSON.parse(selector), null, 2)
  } catch {
    return selector
  }
}

/** 把选择器 JSON 解析成 `key=value` 标签；不是键值对象时返回 null。 */
function selectorEntries(selector: string): string[] | null {
  try {
    const parsed = JSON.parse(selector) as unknown
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return null
    }
    return Object.entries(parsed as Record<string, unknown>).map(
      ([key, value]) => `${key}=${String(value)}`
    )
  } catch {
    return null
  }
}

function SelectorValue({ selector }: { selector: string }) {
  const { t } = useTranslation()
  if (!selector) {
    return <DetailEmpty />
  }
  const entries = selectorEntries(selector)
  if (entries === null) {
    return (
      <CodeBlock
        code={formatSelector(selector)}
        language="json"
        label="JSON"
        copyable
        wrap
        maxHeightClass="max-h-48"
      />
    )
  }
  if (entries.length === 0) {
    return <DetailEmpty />
  }
  return (
    <DetailTagList items={entries} copyable copyLabel={t("kube.selector")} />
  )
}

/** Service 列表：类型、ClusterIP、外部入口与端口映射，详情展示选择器。 */
export function KubernetesServicesTab({
  clusterUid,
  namespace,
  name,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesServicesTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesServiceItem | null>(null)
  const serviceRelation =
    useClusterResourceDetail<GetKubernetesServiceDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesServiceDetail(uid, { namespace: ns, name: nm }),
    })

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesServiceItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesServices(clusterUid, {
          namespace: namespace || undefined,
          name: name || undefined,
          page: nextPage,
          pageSize: nextPageSize,
        }),
    })

  return (
    <>
      <ListFrame
        total={total}
        page={page}
        pageSize={pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => onPageChange(Math.max(1, page - 1))}
        onNext={() => onPageChange(page + 1)}
      >
        {loading ? (
          <ClusterResourceState state="loading" title={t("state.loading")} />
        ) : null}
        {!loading && failed ? (
          <ClusterResourceState
            state="error"
            title={error || t("state.failed")}
            hint={t("kube.resourceErrorHint")}
            onRetry={() => reload()}
          />
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <ClusterResourceState
            state="empty"
            title={t("kube.servicesEmpty")}
            hint={t("kube.servicesEmptyHint")}
          />
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("kube.name")}</TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.identity} ${TABLE_NOWRAP}`}
                >
                  {t("kube.namespace")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.type")}
                </TableHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("kube.clusterIp")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.externalIps")}
                </TableHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("kube.loadBalancerIp")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.uid}>
                  {t("kube.ports")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.time}>
                  {t("kube.age")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((service) => (
                <TableRow key={`${service.namespace}/${service.name}`}>
                  <TableCell className="font-medium break-words">
                    {service.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {service.namespace}
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={service.type || "—"}
                      tone={serviceTone(service.type)}
                    />
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {service.clusterIp || "—"}
                  </TableCell>
                  <TableCell className="max-w-[12rem] font-mono text-xs break-all whitespace-normal">
                    {service.externalIps?.length
                      ? service.externalIps.join(", ")
                      : "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {service.loadBalancerIp || "—"}
                  </TableCell>
                  <TableCell className="whitespace-normal">
                    {service.ports?.length ? (
                      <ul className="grid gap-0.5">
                        {service.ports.map((port) => (
                          <li
                            key={`${port.name}-${port.port}-${port.protocol}`}
                            className="font-mono text-xs break-all"
                          >
                            {portLabel(port)}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(service.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(service)}
                      actions={[]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <ClusterDetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        size="wide"
        description={
          detail ? (
            <>
              <span className="font-mono text-xs">{detail.namespace}</span>
              <StatusBadge
                value={detail.type || "—"}
                tone={serviceTone(detail.type)}
              />
            </>
          ) : null
        }
        info={
          detail
            ? [
                {
                  title: t("kube.sectionBasic"),
                  fields: [
                    {
                      label: t("kube.type"),
                      value: (
                        <StatusBadge
                          value={detail.type || "—"}
                          tone={serviceTone(detail.type)}
                        />
                      ),
                    },
                    {
                      label: t("kube.clusterIp"),
                      value: (
                        <DetailCopyText
                          value={detail.clusterIp}
                          label={t("kube.clusterIp")}
                        />
                      ),
                    },
                    {
                      label: t("kube.externalIps"),
                      value: (
                        <DetailTagList
                          items={detail.externalIps ?? []}
                          copyable
                          copyLabel={t("kube.externalIps")}
                        />
                      ),
                      span: 2,
                    },
                    {
                      label: t("kube.loadBalancerIp"),
                      value: (
                        <DetailCopyText
                          value={detail.loadBalancerIp}
                          label={t("kube.loadBalancerIp")}
                        />
                      ),
                    },
                  ],
                },
                {
                  title: t("kube.ports"),
                  description: detail.ports?.length
                    ? t("kube.portsCount", { count: detail.ports.length })
                    : undefined,
                  content: (
                    <DetailItemList
                      items={(detail.ports ?? []).map((port) => ({
                        key: `${port.name}-${port.port}-${port.protocol}`,
                        title: (
                          <span className="font-mono text-xs break-all">
                            {portLabel(port)}
                          </span>
                        ),
                        meta: (
                          <>
                            <span>
                              {t("kube.protocol")}: {port.protocol || "TCP"}
                            </span>
                            {port.name ? (
                              <span>
                                {t("kube.name")}:{" "}
                                <span className="font-mono">{port.name}</span>
                              </span>
                            ) : null}
                            {port.nodePort ? (
                              <span>
                                {t("kube.nodePort")}:{" "}
                                <span className="font-mono tabular-nums">
                                  {port.nodePort}
                                </span>
                              </span>
                            ) : null}
                          </>
                        ),
                      }))}
                    />
                  ),
                },
                {
                  title: t("kube.selector"),
                  content: <SelectorValue selector={detail.selector} />,
                },
                {
                  title: t("kube.sectionEndpoints"),
                  description: serviceRelation.data?.endpoints?.length
                    ? t("kube.relationsCount", {
                        count: serviceRelation.data.endpoints.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={serviceRelation.loading}
                      failed={serviceRelation.failed}
                      error={serviceRelation.error}
                      onRetry={serviceRelation.retry}
                      empty={
                        !serviceRelation.loading &&
                        !serviceRelation.failed &&
                        (serviceRelation.data?.endpoints?.length ?? 0) === 0
                      }
                    >
                      <ClusterRelationList
                        items={(serviceRelation.data?.endpoints ?? []).map(
                          (endpoint, index) => ({
                            key: `${endpoint.address}|${index}`,
                            title: (
                              <span className="font-mono text-xs break-all">
                                {endpoint.address}
                                {endpoint.hostname
                                  ? ` (${endpoint.hostname})`
                                  : ""}
                              </span>
                            ),
                            status: (
                              <StatusBadge
                                value={endpoint.ready ? "OK" : "FAILED"}
                                label={
                                  endpoint.ready
                                    ? t("kube.endpointReady")
                                    : t("kube.endpointNotReady")
                                }
                                tone={endpoint.ready ? "success" : "warning"}
                              />
                            ),
                            subtitle: [
                              endpoint.nodeName,
                              endpoint.targetName
                                ? `${endpoint.targetKind || "—"}/${endpoint.targetName}`
                                : "",
                            ]
                              .filter(Boolean)
                              .join(" · "),
                            meta: endpoint.ports?.length ? (
                              <span className="font-mono">
                                {endpoint.ports
                                  .map(
                                    (port) =>
                                      `${port.name ? `${port.name}:` : ""}${port.port}/${port.protocol || "TCP"}`
                                  )
                                  .join(" · ")}
                              </span>
                            ) : undefined,
                          })
                        )}
                      />
                    </ClusterRelationBody>
                  ),
                },
                {
                  title: t("kube.sectionTime"),
                  fields: [
                    {
                      label: t("kube.age"),
                      value: formatClusterAge(detail.createdAt, t),
                    },
                    {
                      label: t("kube.createdAt"),
                      value: formatClusterTime(detail.createdAt) || "—",
                    },
                  ],
                },
              ]
            : []
        }
        yaml={
          detail
            ? {
                clusterUid,
                kind: "Service",
                namespace: detail.namespace,
                name: detail.name,
              }
            : null
        }
        onOpenChange={(open) => {
          if (!open) {
            setDetail(null)
          }
        }}
      />
    </>
  )
}
