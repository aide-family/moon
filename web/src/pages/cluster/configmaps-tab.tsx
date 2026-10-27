import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getKubernetesConfigMapDetail,
  listKubernetesConfigMaps,
} from "@/api/cluster"
import type {
  GetKubernetesConfigMapDetailResult,
  KubernetesConfigMapItem,
  KubernetesWorkloadRefItem,
} from "@/api/cluster"
import { DetailFlag, DetailTagList } from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions } from "@/components/row-actions"
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
  ClusterRelatedWorkloadList,
  ClusterRelationBody,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

interface KubernetesConfigMapsTabProps {
  clusterUid: string
  namespace: string
  name: string
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

/** ConfigMap 列表：只展示键名与不可变标记，值永远不出现在页面上。 */
export function KubernetesConfigMapsTab({
  clusterUid,
  namespace,
  name,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesConfigMapsTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesConfigMapItem | null>(null)
  const configMapRelation =
    useClusterResourceDetail<GetKubernetesConfigMapDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesConfigMapDetail(uid, { namespace: ns, name: nm }),
    })
  const jumpTo = useClusterRelationJump()

  function openConfigMapWorkload(ref: KubernetesWorkloadRefItem) {
    setDetail(null)
    jumpTo({
      tab: "workloads",
      namespace: ref.namespace,
      name: ref.name,
      kind: ref.kind,
    })
  }

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesConfigMapItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesConfigMaps(clusterUid, {
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
            title={t("kube.configmapsEmpty")}
            hint={t("kube.configmapsEmptyHint")}
          />
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
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
                  {t("kube.dataKeys")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.binaryDataKeys")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.immutable")}
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
              {items.map((configMap) => (
                <TableRow key={`${configMap.namespace}/${configMap.name}`}>
                  <TableCell className="font-medium break-words">
                    {configMap.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {configMap.namespace}
                  </TableCell>
                  <TableCell
                    className="font-mono text-xs tabular-nums"
                    title={configMap.dataKeys?.join(", ")}
                  >
                    {configMap.dataKeys?.length ?? 0}
                  </TableCell>
                  <TableCell
                    className="font-mono text-xs tabular-nums"
                    title={configMap.binaryDataKeys?.join(", ")}
                  >
                    {configMap.binaryDataKeys?.length ?? 0}
                  </TableCell>
                  <TableCell className="text-xs">
                    {configMap.immutable ? t("kube.yes") : t("kube.no")}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(configMap.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(configMap)}
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
            <span className="font-mono text-xs">{detail.namespace}</span>
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
                      label: t("kube.immutable"),
                      value: (
                        <DetailFlag
                          value={detail.immutable}
                          trueLabel={t("kube.yes")}
                          falseLabel={t("kube.no")}
                        />
                      ),
                    },
                  ],
                },
                {
                  title: t("kube.sectionConfigData"),
                  description: t("kube.keysCount", {
                    count:
                      (detail.dataKeys?.length ?? 0) +
                      (detail.binaryDataKeys?.length ?? 0),
                  }),
                  fields: [
                    {
                      label: t("kube.dataKeys"),
                      value: (
                        <DetailTagList
                          items={detail.dataKeys ?? []}
                          copyable
                          copyLabel={t("kube.dataKeys")}
                        />
                      ),
                      span: 2,
                    },
                    {
                      label: t("kube.binaryDataKeys"),
                      value: (
                        <DetailTagList
                          items={detail.binaryDataKeys ?? []}
                          copyable
                          copyLabel={t("kube.binaryDataKeys")}
                        />
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionRelatedWorkloads"),
                  description: configMapRelation.data?.relatedWorkloads?.length
                    ? t("kube.relationsCount", {
                        count: configMapRelation.data.relatedWorkloads.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={configMapRelation.loading}
                      failed={configMapRelation.failed}
                      error={configMapRelation.error}
                      onRetry={configMapRelation.retry}
                      empty={
                        !configMapRelation.loading &&
                        !configMapRelation.failed &&
                        (configMapRelation.data?.relatedWorkloads?.length ??
                          0) === 0
                      }
                    >
                      <ClusterRelatedWorkloadList
                        items={configMapRelation.data?.relatedWorkloads ?? []}
                        onOpen={openConfigMapWorkload}
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
                kind: "ConfigMap",
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
