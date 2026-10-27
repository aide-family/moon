import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getKubernetesIngressDetail,
  listKubernetesIngresses,
} from "@/api/cluster"
import type {
  GetKubernetesIngressDetailResult,
  KubernetesIngressItem,
  KubernetesServiceItem,
} from "@/api/cluster"
import {
  DetailCopyText,
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
import {
  ClusterRelationBody,
  ClusterRelationList,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

interface KubernetesIngressesTabProps {
  clusterUid: string
  namespace: string
  name: string
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

function serviceTypeTone(type: string): BadgeTone {
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

/** Ingress 列表：IngressClass、主机、转发规则与 TLS Secret。 */
export function KubernetesIngressesTab({
  clusterUid,
  namespace,
  name,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesIngressesTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesIngressItem | null>(null)
  const ingressRelation =
    useClusterResourceDetail<GetKubernetesIngressDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesIngressDetail(uid, { namespace: ns, name: nm }),
    })
  const jumpTo = useClusterRelationJump()

  function openService(service: KubernetesServiceItem) {
    setDetail(null)
    jumpTo({
      tab: "services",
      namespace: service.namespace || detail?.namespace,
      name: service.name,
    })
  }

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesIngressItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesIngresses(clusterUid, {
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
            title={t("kube.ingressesEmpty")}
            hint={t("kube.ingressesEmptyHint")}
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
                  {t("kube.className")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.hosts")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.uid}>
                  {t("kube.paths")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.tls")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("kube.age")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((ingress) => (
                <TableRow key={`${ingress.namespace}/${ingress.name}`}>
                  <TableCell className="font-medium break-words">
                    {ingress.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {ingress.namespace}
                  </TableCell>
                  <TableCell>
                    {ingress.className ? (
                      <StatusBadge value={ingress.className} tone="neutral" />
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell className="max-w-[14rem] font-mono text-xs break-words whitespace-normal">
                    {ingress.hosts?.length ? ingress.hosts.join(", ") : "—"}
                  </TableCell>
                  <TableCell className="whitespace-normal">
                    {ingress.paths?.length ? (
                      <ul className="grid gap-1">
                        {ingress.paths.map((path, index) => (
                          <li
                            key={`${path.host}|${path.path}|${path.serviceName}|${path.servicePort}|${index}`}
                            className="text-xs break-all"
                          >
                            <span className="font-mono">
                              {path.host || "*"}
                              {path.path || "/"}
                            </span>
                            <span
                              aria-hidden="true"
                              className="mx-1 text-muted-foreground"
                            >
                              →
                            </span>
                            <span className="font-mono">
                              {path.serviceName || "—"}:
                              {path.servicePort || "—"}
                            </span>
                            {path.pathType ? (
                              <span className="ml-1 text-muted-foreground">
                                ({path.pathType})
                              </span>
                            ) : null}
                          </li>
                        ))}
                      </ul>
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell className="max-w-[12rem] font-mono text-xs break-all whitespace-normal text-muted-foreground">
                    {ingress.tls?.length ? ingress.tls.join(", ") : "—"}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(ingress.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(ingress)}
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
              <StatusBadge value={detail.className || "—"} tone="neutral" />
            </>
          ) : null
        }
        info={
          detail
            ? [
                {
                  title: t("kube.sectionBasic"),
                  fields: [
                    { label: t("kube.namespace"), value: detail.namespace },
                    {
                      label: t("kube.className"),
                      value: (
                        <DetailCopyText
                          value={detail.className}
                          label={t("kube.className")}
                        />
                      ),
                    },
                    {
                      label: t("kube.tls"),
                      value: (
                        <DetailTagList
                          items={detail.tls ?? []}
                          copyable
                          copyLabel={t("kube.tls")}
                        />
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionRouting"),
                  description: detail.paths?.length
                    ? t("kube.rulesCount", { count: detail.paths.length })
                    : undefined,
                  fields: [
                    {
                      label: t("kube.hosts"),
                      value: (
                        <DetailTagList
                          items={detail.hosts ?? []}
                          copyable
                          copyLabel={t("kube.hosts")}
                        />
                      ),
                      span: 2,
                    },
                    {
                      label: t("kube.paths"),
                      value: (
                        <DetailItemList
                          items={(detail.paths ?? []).map((path, index) => ({
                            key: `${path.host}|${path.path}|${path.serviceName}|${path.servicePort}|${index}`,
                            title: (
                              <span className="font-mono text-xs break-all">
                                {path.host || "*"}
                                {path.path || "/"}
                              </span>
                            ),
                            subtitle: (
                              <span className="font-mono">
                                <span
                                  aria-hidden="true"
                                  className="mr-1 text-muted-foreground"
                                >
                                  →
                                </span>
                                {path.serviceName || "—"}:
                                {path.servicePort || "—"}
                              </span>
                            ),
                            meta: path.pathType ? (
                              <span>
                                {t("kube.pathType")}: {path.pathType}
                              </span>
                            ) : undefined,
                          }))}
                        />
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionRelatedService"),
                  description: ingressRelation.data?.services?.length
                    ? t("kube.relationsCount", {
                        count: ingressRelation.data.services.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={ingressRelation.loading}
                      failed={ingressRelation.failed}
                      error={ingressRelation.error}
                      onRetry={ingressRelation.retry}
                      empty={
                        !ingressRelation.loading &&
                        !ingressRelation.failed &&
                        (ingressRelation.data?.services?.length ?? 0) === 0
                      }
                    >
                      <ClusterRelationList
                        items={(ingressRelation.data?.services ?? []).map(
                          (service) => ({
                            key: `${service.namespace}/${service.name}`,
                            title: (
                              <span className="font-mono text-xs break-all">
                                {service.name}
                              </span>
                            ),
                            status: (
                              <StatusBadge
                                value={service.type || "—"}
                                tone={serviceTypeTone(service.type)}
                              />
                            ),
                            subtitle: service.clusterIp
                              ? service.clusterIp
                              : undefined,
                            meta: service.ports?.length ? (
                              <span className="font-mono">
                                {service.ports
                                  .map(
                                    (port) =>
                                      `${port.port}/${port.protocol || "TCP"}`
                                  )
                                  .join(" · ")}
                              </span>
                            ) : undefined,
                            onSelect: () => openService(service),
                            ariaLabel: t("kube.openResource", {
                              name: service.name,
                            }),
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
                kind: "Ingress",
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
