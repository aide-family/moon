import { ShieldAlert } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import { getKubernetesSecretDetail, listKubernetesSecrets } from "@/api/cluster"
import type {
  GetKubernetesSecretDetailResult,
  KubernetesSecretItem,
  KubernetesWorkloadRefItem,
} from "@/api/cluster"
import { DetailCopyText, DetailFlag } from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
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
import { KubernetesKeyList } from "./key-list"
import {
  ClusterRelatedWorkloadList,
  ClusterRelationBody,
  useClusterRelationJump,
} from "./relation"
import { ClusterResourceState } from "./resource-state"
import { useClusterResourceDetail } from "./use-resource-detail"

interface KubernetesSecretsTabProps {
  clusterUid: string
  namespace: string
  name: string
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

/** Secret 列表：只读展示类型、键名与不可变标记；密钥值永远不会返回。 */
export function KubernetesSecretsTab({
  clusterUid,
  namespace,
  name,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesSecretsTabProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<KubernetesSecretItem | null>(null)
  const secretRelation =
    useClusterResourceDetail<GetKubernetesSecretDetailResult>({
      enabled: detail !== null,
      clusterUid,
      namespace: detail?.namespace ?? "",
      name: detail?.name ?? "",
      fetch: (uid, ns, nm) =>
        getKubernetesSecretDetail(uid, { namespace: ns, name: nm }),
    })
  const jumpTo = useClusterRelationJump()

  function openSecretWorkload(ref: KubernetesWorkloadRefItem) {
    setDetail(null)
    jumpTo({
      tab: "workloads",
      namespace: ref.namespace,
      name: ref.name,
      kind: ref.kind,
    })
  }

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesSecretItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesSecrets(clusterUid, {
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
            title={t("kube.secretsEmpty")}
            hint={t("kube.secretsEmptyHint")}
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
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.type")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.dataKeys")}
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
              {items.map((secret) => (
                <TableRow key={`${secret.namespace}/${secret.name}`}>
                  <TableCell className="font-medium break-words">
                    {secret.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {secret.namespace}
                  </TableCell>
                  <TableCell className="max-w-[16rem] whitespace-normal">
                    {secret.type ? (
                      <StatusBadge value={secret.type} tone="neutral" />
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell
                    className="font-mono text-xs tabular-nums"
                    title={secret.dataKeys?.join(", ")}
                  >
                    {secret.dataKeys?.length ?? 0}
                  </TableCell>
                  <TableCell className="text-xs">
                    {secret.immutable ? t("kube.yes") : t("kube.no")}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(secret.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(secret)}
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
              {detail.type ? (
                <StatusBadge value={detail.type} tone="neutral" />
              ) : null}
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
                      label: t("kube.type"),
                      value: (
                        <DetailCopyText
                          value={detail.type}
                          label={t("kube.type")}
                        />
                      ),
                      span: 2,
                    },
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
                    count: detail.dataKeys?.length ?? 0,
                  }),
                  fields: [
                    {
                      label: t("kube.dataKeys"),
                      value: <KubernetesKeyList keys={detail.dataKeys ?? []} />,
                      span: 2,
                    },
                    {
                      label: t("kube.values"),
                      value: (
                        <span className="flex items-start gap-1.5 rounded-md border border-border/60 bg-muted/30 px-2.5 py-2 text-xs leading-5 text-muted-foreground">
                          <ShieldAlert
                            aria-hidden="true"
                            className="mt-0.5 size-3.5 shrink-0"
                          />
                          <span className="min-w-0 break-words">
                            {t("kube.secretValuesHidden")}
                          </span>
                        </span>
                      ),
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionRelatedWorkloads"),
                  description: secretRelation.data?.relatedWorkloads?.length
                    ? t("kube.relationsCount", {
                        count: secretRelation.data.relatedWorkloads.length,
                      })
                    : undefined,
                  content: (
                    <ClusterRelationBody
                      loading={secretRelation.loading}
                      failed={secretRelation.failed}
                      error={secretRelation.error}
                      onRetry={secretRelation.retry}
                      empty={
                        !secretRelation.loading &&
                        !secretRelation.failed &&
                        (secretRelation.data?.relatedWorkloads?.length ?? 0) ===
                          0
                      }
                    >
                      <ClusterRelatedWorkloadList
                        items={secretRelation.data?.relatedWorkloads ?? []}
                        onOpen={openSecretWorkload}
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
                kind: "Secret",
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
