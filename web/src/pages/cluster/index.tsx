import { Search } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, Navigate, useNavigate } from "react-router-dom"

import { deleteCluster, listClusters, testCluster } from "@/api/cluster"
import type { ClusterItem, ClusterProbe } from "@/api/cluster"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { FilterChips, type FilterChip } from "@/components/filter-chips"
import { FilterToggle } from "@/components/filter-toggle"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { TableActionsHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formErrorMessage } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { ServerErrorPage } from "@/pages/server-error"

import { formatClusterTime } from "./format"
import { ClusterKubeconfigDialog } from "./kubeconfig-dialog"
import { ClusterProbeDialog } from "./probe-dialog"

/** 集群列表：名称检索、接入/编辑、删除，以及保存前的连通性校验。 */
export function ClusterPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const [draftName, setDraftName] = useState(query.name)
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editing, setEditing] = useState<ClusterItem | null>(null)
  const [deleting, setDeleting] = useState<ClusterItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [testingUid, setTestingUid] = useState("")
  const [probeDialog, setProbeDialog] = useState<{
    name: string
    probe: ClusterProbe | null
    error: string
  } | null>(null)

  const canWrite = session.can("cluster:write")
  const canDelete = session.can("cluster:delete")
  // 测试连通性只读集群元数据，服务端按 cluster:read 校验，按钮也按读权限展示。
  const canTest = session.can("cluster:read")
  const { items, total, loading, failed, failedStatus, error, reload } =
    usePagedList<ClusterItem>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${query.name}`,
      fetch: (page, pageSize) =>
        listClusters({ page, pageSize, name: query.name }),
    })

  if (!session.loading && !session.can("cluster:read")) {
    return <Navigate to="/forbidden" replace />
  }

  // 网关/后端不可用时展示标准 502/503 页，而不是把浏览器的 "Failed to fetch" 留在列表内。
  if (!loading && failed && failedStatus >= 500) {
    return (
      <ServerErrorPage
        fullPage={false}
        status={failedStatus === 503 ? 503 : 502}
        detail={error === t("error.badGatewayTitle") ? "" : error}
        onRetry={() => reload()}
      />
    )
  }

  const filterChips: FilterChip[] = query.name
    ? [
        {
          key: "name",
          label: `${t("cluster.name")}: ${query.name}`,
          onRemove: () => {
            setDraftName("")
            query.update({ name: "" })
          },
        },
      ]
    : []

  function openCreate() {
    setEditing(null)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  function openEdit(item: ClusterItem) {
    setEditing(item)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  async function runTest(item: ClusterItem) {
    if (testingUid) {
      return
    }
    setTestingUid(item.uid)
    try {
      const result = await testCluster(item.uid)
      setProbeDialog({ name: item.name, probe: result.probe, error: "" })
    } catch (err) {
      setProbeDialog({
        name: item.name,
        probe: null,
        error: formErrorMessage(err, t("cluster.testFailed")),
      })
    } finally {
      setTestingUid("")
      // 后端会刷新最近检测时间和最近错误，重载列表让两列保持同步。
      reload()
    }
  }

  async function runDelete() {
    if (!deleting) {
      return
    }
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteCluster(deleting.uid)
      setDeleting(null)
      reload()
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

  function rowActions(item: ClusterItem): RowAction[] {
    const actions: RowAction[] = []
    if (canWrite) {
      actions.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => openEdit(item),
      })
    }
    if (canTest) {
      actions.push({
        key: "test",
        label: t("cluster.testConnection"),
        onSelect: () => void runTest(item),
      })
    }
    if (canDelete) {
      actions.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setDeleteError("")
          setDeleting(item)
        },
      })
    }
    return actions
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("cluster.title")}
        description={t("cluster.description")}
      >
        <div className="flex flex-wrap items-center justify-end gap-2">
          <FilterToggle
            open={filtersOpen}
            count={filterChips.length}
            controls="cluster-filters"
            onClick={() => setFiltersOpen((open) => !open)}
          />
          {canWrite ? (
            <Button onClick={openCreate}>{t("action.create")}</Button>
          ) : null}
        </div>
      </PageHeader>
      <div
        className={`min-w-0 flex-col gap-2 ${filtersOpen || filterChips.length > 0 ? "flex" : "hidden"}`}
      >
        {filterChips.length > 0 ? (
          <FilterChips
            items={filterChips}
            onClearAll={() => {
              setDraftName("")
              query.update({ name: "" })
            }}
          />
        ) : null}
        <form
          id="cluster-filters"
          className={`flex min-w-0 flex-nowrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4 ${filtersOpen ? "" : "hidden"}`}
          onSubmit={(event) => {
            event.preventDefault()
            query.update({ name: draftName.trim() })
          }}
        >
          <div className="min-w-0 flex-1 basis-48">
            <Input
              id="cluster-name-filter"
              value={draftName}
              placeholder={t("cluster.searchPlaceholder")}
              aria-label={t("cluster.name")}
              onChange={(event) => setDraftName(event.target.value)}
            />
          </div>
          <div className="flex shrink-0 flex-nowrap items-center gap-2">
            {draftName.trim() !== query.name ? (
              <span
                className="text-xs whitespace-nowrap text-amber-600 dark:text-amber-400"
                role="status"
              >
                {t("filter.unapplied")}
              </span>
            ) : null}
            {query.name ? (
              <Button
                type="button"
                variant="ghost"
                onClick={() => {
                  setDraftName("")
                  query.update({ name: "" })
                }}
              >
                {t("filter.clear")}
              </Button>
            ) : null}
            <Button type="submit" variant="outline">
              <Search aria-hidden="true" data-icon="inline-start" />
              {t("filter.apply")}
            </Button>
          </div>
        </form>
      </div>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">
              {error || t("state.failed")}
            </p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("cluster.empty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("cluster.name")}</TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.uid} ${TABLE_NOWRAP}`}
                >
                  {t("cluster.server")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.status} ${TABLE_NOWRAP}`}
                >
                  {t("cluster.version")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("cluster.lastCheck")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("cluster.lastError")}
                </TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("cluster.remark")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-medium break-words">
                    <Link
                      to={`/clusters/${item.uid}`}
                      className="underline-offset-4 hover:underline"
                    >
                      {item.name}
                    </Link>
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs break-all ${TABLE_COLUMN_WIDTH.uid}`}
                  >
                    {item.server || "—"}
                  </TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>
                    {item.version || "—"}
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                  >
                    {formatClusterTime(item.lastCheckedAt) ||
                      t("cluster.neverChecked")}
                  </TableCell>
                  <TableCell
                    className={`break-all ${TABLE_COLUMN_WIDTH.identity}`}
                  >
                    {item.lastError ? (
                      <span
                        className="text-xs break-words text-destructive"
                        title={item.lastError}
                      >
                        {item.lastError}
                      </span>
                    ) : (
                      "—"
                    )}
                  </TableCell>
                  <TableCell
                    className={`break-words text-muted-foreground ${TABLE_COLUMN_WIDTH.identity}`}
                  >
                    {item.remark || "—"}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => navigate(`/clusters/${item.uid}`)}
                      actions={rowActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <ClusterKubeconfigDialog
        key={`${editing?.uid ?? "create"}-${editorSession}`}
        open={editorOpen}
        cluster={editing}
        onOpenChange={setEditorOpen}
        onSaved={() => {
          setEditorOpen(false)
          reload()
        }}
      />
      <ClusterProbeDialog
        open={probeDialog !== null}
        name={probeDialog?.name ?? ""}
        probe={probeDialog?.probe ?? null}
        error={probeDialog?.error ?? ""}
        onOpenChange={(open) => {
          if (!open) {
            setProbeDialog(null)
          }
        }}
      />
      <ConfirmAction
        open={deleting !== null}
        title={t("cluster.deleteTitle")}
        description={t("cluster.deleteDescription")}
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
    </section>
  )
}
