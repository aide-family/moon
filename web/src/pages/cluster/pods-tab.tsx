import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  deleteKubernetesPod,
  getKubernetesPodLogs,
  listKubernetesPods,
} from "@/api/cluster"
import type { KubernetesPodItem } from "@/api/cluster"
import { CodeBlock } from "@/components/code-block"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { DetailCopyText, DetailItemList } from "@/components/detail-values"
import { ListFrame } from "@/components/list-frame"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge, type BadgeTone } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Field, FieldLabel } from "@/components/ui/field"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { usePagedList } from "@/hooks/use-paged-list"
import { formErrorMessage } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { formatClusterAge, formatClusterTime, isPodReady } from "./format"
import { ClusterDetailSheet } from "./cluster-detail-sheet"
import { ClusterResourceState } from "./resource-state"

const TAIL_LINE_OPTIONS = [200, 500, 1000, 2000]

function podTone(phase: string): BadgeTone {
  switch (phase) {
    case "Running":
      return "success"
    case "Pending":
      return "warning"
    case "Failed":
      return "danger"
    case "Succeeded":
      return "info"
    default:
      return "neutral"
  }
}

interface KubernetesPodsTabProps {
  clusterUid: string
  namespace: string
  name: string
  /** k8s:pod:write：只控制删除，读取日志是独立权限。 */
  canWrite: boolean
  /** k8s:pod:logs：控制日志行操作与日志面板入口。 */
  canLogs: boolean
  page: number
  pageSize: number
  onPageChange: (page: number) => void
  /** 工具栏刷新按钮递增；变化时静默重载。 */
  refreshKey?: number
}

/** Pod 列表：状态、重启次数、所属工作负载，以及删除与日志读取。 */
export function KubernetesPodsTab({
  clusterUid,
  namespace,
  name,
  canWrite,
  canLogs,
  page,
  pageSize,
  onPageChange,
  refreshKey = 0,
}: KubernetesPodsTabProps) {
  const { t } = useTranslation()
  const [deleting, setDeleting] = useState<KubernetesPodItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [detail, setDetail] = useState<KubernetesPodItem | null>(null)
  const [logsPod, setLogsPod] = useState<KubernetesPodItem | null>(null)
  const [logContainer, setLogContainer] = useState("")
  const [logContainerActual, setLogContainerActual] = useState("")
  const [tailLines, setTailLines] = useState("200")
  const [previous, setPrevious] = useState(false)
  const [logContent, setLogContent] = useState("")
  const [logLoading, setLogLoading] = useState(false)
  const [logFailed, setLogFailed] = useState(false)
  const [logError, setLogError] = useState("")
  const [logRefreshKey, setLogRefreshKey] = useState(0)

  const { items, total, loading, failed, error, reload } =
    usePagedList<KubernetesPodItem>({
      page,
      pageSize,
      filterKey: `${clusterUid}|${namespace}|${name}`,
      refreshKey,
      fetch: (nextPage, nextPageSize) =>
        listKubernetesPods(clusterUid, {
          namespace: namespace || undefined,
          name: name || undefined,
          page: nextPage,
          pageSize: nextPageSize,
        }),
    })

  useEffect(() => {
    if (!logsPod) {
      return
    }
    let active = true
    async function loadLogs() {
      setLogLoading(true)
      setLogFailed(false)
      setLogError("")
      try {
        const result = await getKubernetesPodLogs(clusterUid, {
          namespace: logsPod!.namespace,
          name: logsPod!.name,
          container: logContainer || undefined,
          tailLines: Number(tailLines),
          previous,
        })
        if (!active) {
          return
        }
        setLogContent(result.content)
        setLogContainerActual(result.container)
      } catch (err) {
        if (!active) {
          return
        }
        setLogFailed(true)
        setLogError(formErrorMessage(err, t("state.failed")))
      } finally {
        if (active) {
          setLogLoading(false)
        }
      }
    }
    void loadLogs()
    return () => {
      active = false
    }
  }, [clusterUid, logsPod, logContainer, tailLines, previous, logRefreshKey, t])

  function podActions(pod: KubernetesPodItem): RowAction[] {
    const actions: RowAction[] = []
    if (canLogs) {
      actions.push({
        key: "logs",
        label: t("kube.logs"),
        onSelect: () => openLogs(pod),
      })
    }
    if (canWrite) {
      actions.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setDeleteError("")
          setDeleting(pod)
        },
      })
    }
    return actions
  }

  function openLogs(pod: KubernetesPodItem) {
    setLogsPod(pod)
    setLogContainer(pod.containers?.[0]?.name ?? "")
    setLogContainerActual("")
    setTailLines("200")
    setPrevious(false)
    setLogContent("")
    setLogFailed(false)
    setLogError("")
  }

  async function runDelete() {
    if (!deleting) {
      return
    }
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteKubernetesPod(clusterUid, deleting.namespace, deleting.name)
      setDeleting(null)
      reload()
      toast.success(t("kube.actionDone"))
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

  const logContainers = logsPod?.containers ?? []

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
            title={t("kube.podsEmpty")}
            hint={t("kube.podsEmptyHint")}
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
                  {t("kube.phase")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.ready")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("kube.restarts")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.node")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("kube.owner")}
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
              {items.map((pod) => (
                <TableRow key={`${pod.namespace}/${pod.name}`}>
                  <TableCell className="font-mono text-xs font-medium break-all">
                    {pod.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {pod.namespace}
                  </TableCell>
                  <TableCell>
                    <StatusBadge value={pod.phase} tone={podTone(pod.phase)} />
                  </TableCell>
                  <TableCell className="font-mono text-xs tabular-nums">
                    {pod.ready || "—"}
                  </TableCell>
                  <TableCell className="font-mono text-xs tabular-nums">
                    {pod.restarts}
                  </TableCell>
                  <TableCell className="max-w-[12rem] font-mono text-xs break-words whitespace-normal text-muted-foreground">
                    {pod.nodeName || "—"}
                  </TableCell>
                  <TableCell className="max-w-[14rem] text-xs break-words whitespace-normal text-muted-foreground">
                    {pod.ownerName ? `${pod.ownerKind}/${pod.ownerName}` : "—"}
                  </TableCell>
                  <TableCell className={`text-xs tabular-nums ${TABLE_NOWRAP}`}>
                    {formatClusterAge(pod.createdAt, t)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      forceMenu
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(pod)}
                      actions={podActions(pod)}
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
              <StatusBadge value={detail.phase} tone={podTone(detail.phase)} />
              <span className="font-mono text-xs">{detail.namespace}</span>
            </>
          ) : null
        }
        info={
          detail
            ? [
                {
                  title: t("kube.sectionStatus"),
                  fields: [
                    {
                      label: t("kube.phase"),
                      value: (
                        <StatusBadge
                          value={detail.phase}
                          tone={podTone(detail.phase)}
                        />
                      ),
                    },
                    {
                      label: t("kube.ready"),
                      value: (
                        <span
                          className={
                            isPodReady(detail.ready)
                              ? "font-mono text-xs tabular-nums"
                              : "font-mono text-xs text-warning tabular-nums"
                          }
                        >
                          {detail.ready || "—"}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.restarts"),
                      value: (
                        <span
                          className={
                            detail.restarts > 0
                              ? "font-mono text-xs text-warning tabular-nums"
                              : "font-mono text-xs tabular-nums"
                          }
                        >
                          {detail.restarts}
                        </span>
                      ),
                    },
                    {
                      label: t("kube.node"),
                      value: (
                        <DetailCopyText
                          value={detail.nodeName}
                          label={t("kube.node")}
                        />
                      ),
                    },
                    {
                      label: t("kube.podIp"),
                      value: (
                        <DetailCopyText
                          value={detail.podIp}
                          label={t("kube.podIp")}
                        />
                      ),
                    },
                    {
                      label: t("kube.owner"),
                      value: detail.ownerName ? (
                        <DetailCopyText
                          value={`${detail.ownerKind}/${detail.ownerName}`}
                          label={t("kube.owner")}
                        />
                      ) : null,
                      span: 2,
                    },
                  ],
                },
                {
                  title: t("kube.sectionContainers"),
                  description: detail.containers?.length
                    ? t("kube.containersCount", {
                        count: detail.containers.length,
                      })
                    : undefined,
                  content: (
                    <DetailItemList
                      items={(detail.containers ?? []).map((container) => ({
                        key: container.name,
                        title: (
                          <DetailCopyText
                            value={container.name}
                            label={t("kube.container")}
                          />
                        ),
                        subtitle: container.image ? (
                          <DetailCopyText
                            value={container.image}
                            label={t("kube.images")}
                            muted
                          />
                        ) : undefined,
                        status: (
                          <span
                            className={
                              container.ready
                                ? "text-emerald-600 dark:text-emerald-400"
                                : "text-muted-foreground"
                            }
                          >
                            {t("kube.ready")}:{" "}
                            {container.ready ? t("kube.yes") : t("kube.no")}
                          </span>
                        ),
                        meta: (
                          <span>
                            {t("kube.restarts")}:{" "}
                            <span
                              className={
                                container.restarts > 0
                                  ? "font-mono text-warning tabular-nums"
                                  : "font-mono tabular-nums"
                              }
                            >
                              {container.restarts}
                            </span>
                          </span>
                        ),
                      }))}
                    />
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
                kind: "Pod",
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
      <DetailSheet
        open={logsPod !== null}
        title={
          logsPod ? `${logsPod.namespace}/${logsPod.name}` : t("kube.logsTitle")
        }
        size="wide"
        description={
          logsPod ? (
            <>
              <StatusBadge
                value={logsPod.phase}
                tone={podTone(logsPod.phase)}
              />
              <span className="font-mono text-xs">
                {logsPod.nodeName || "—"}
              </span>
              {logContainerActual ? (
                <span className="font-mono text-xs">{logContainerActual}</span>
              ) : null}
            </>
          ) : null
        }
        sections={[
          {
            title: t("kube.logs"),
            description: logsPod
              ? t("kube.logsHint", { lines: tailLines })
              : undefined,
            content: (
              <div className="grid gap-3">
                <div className="flex flex-wrap items-end gap-3">
                  <Field className="w-52 shrink-0">
                    <FieldLabel htmlFor="pod-log-container">
                      {t("kube.container")}
                    </FieldLabel>
                    <Select
                      value={logContainer}
                      disabled={logContainers.length <= 1}
                      onValueChange={setLogContainer}
                    >
                      <SelectTrigger id="pod-log-container" className="w-full">
                        <SelectValue placeholder={t("kube.container")} />
                      </SelectTrigger>
                      <SelectContent>
                        {logContainers.map((container) => (
                          <SelectItem
                            key={container.name}
                            value={container.name}
                          >
                            {container.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </Field>
                  <Field className="w-36 shrink-0">
                    <FieldLabel htmlFor="pod-log-tail">
                      {t("kube.tailLines")}
                    </FieldLabel>
                    <Select value={tailLines} onValueChange={setTailLines}>
                      <SelectTrigger id="pod-log-tail" className="w-full">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {TAIL_LINE_OPTIONS.map((option) => (
                          <SelectItem key={option} value={String(option)}>
                            {option}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </Field>
                  <label className="flex h-8 items-center gap-2 text-sm">
                    <Checkbox
                      checked={previous}
                      onCheckedChange={(value) => setPrevious(value === true)}
                    />
                    {t("kube.previous")}
                  </label>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={logLoading}
                    onClick={() => setLogRefreshKey((value) => value + 1)}
                  >
                    {t("action.refresh")}
                  </Button>
                </div>
                {logLoading ? (
                  <p className="text-sm text-muted-foreground">
                    {t("state.loading")}
                  </p>
                ) : null}
                {!logLoading && logFailed ? (
                  <p className="text-sm text-destructive">
                    {logError || t("state.failed")}
                  </p>
                ) : null}
                {!logLoading && !logFailed ? (
                  <CodeBlock
                    code={logContent}
                    language="log"
                    label={t("kube.logs")}
                    copyable
                    wrap
                    maxHeightClass="max-h-[60vh]"
                    placeholder={t("kube.logsEmpty")}
                    ariaLabel={t("kube.logs")}
                  />
                ) : null}
              </div>
            ),
          },
        ]}
        onOpenChange={(open) => {
          if (!open) {
            setLogsPod(null)
          }
        }}
      />
      <ConfirmAction
        open={deleting !== null}
        title={t("kube.deletePodTitle")}
        description={
          deleting
            ? t("kube.deletePodDescription", {
                namespace: deleting.namespace,
                name: deleting.name,
              })
            : ""
        }
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={deletePending}
        error={deleteError}
        onConfirm={() => void runDelete()}
        onOpenChange={(open) => {
          if (!open) {
            setDeleting(null)
            setDeleteError("")
          }
        }}
      />
    </>
  )
}
