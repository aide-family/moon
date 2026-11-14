import { Copy, Info, Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import { listProbeTargetGroups } from "@/api/probe-target"
import {
  deleteStatusPage,
  listStatusPages,
  regenerateStatusPageToken,
  updateStatusPageStatus,
} from "@/api/status-page"
import type { StatusPageItem, StatusPageStatus } from "@/api/status-page/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import type { MultiSelectOption } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
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
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { loadAllPagedItems } from "@/lib/select-pages"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { ServerErrorPage } from "@/pages/server-error"

import { StatusPageDialog } from "./status-page-dialog"

// Radix Select 不允许空字符串选项，用哨兵值表示「全部状态」。
const ALL_STATUS = "ALL"

function normalizeStatus(value: string): StatusPageStatus | "" {
  return value === "ACTIVE" || value === "DISABLED" ? value : ""
}

/** 公开链接只在浏览器里拼，后端返回的 token 即访问凭证。 */
function publicURL(item: StatusPageItem): string {
  return `${window.location.origin}/status/${encodeURIComponent(item.token)}`
}

/** 状态页列表：配置 CRUD、启停、公开链接复制与重新生成。 */
export function StatusPageListPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    name: query.name,
    status: normalizeStatus(query.status),
  })
  const [groups, setGroups] = useState<MultiSelectOption[]>([])
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editing, setEditing] = useState<StatusPageItem | null>(null)
  const [confirm, setConfirm] = useState<{
    action: "delete" | "regenerate"
    item: StatusPageItem
  } | null>(null)
  const [confirmPending, setConfirmPending] = useState(false)
  const [confirmError, setConfirmError] = useState("")
  const [togglingUid, setTogglingUid] = useState("")

  const canWrite = session.can("status_page:write")
  const canDelete = session.can("status_page:delete")
  const status = normalizeStatus(query.status)
  const { items, total, loading, failed, failedStatus, error, reload } =
    usePagedList<StatusPageItem>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${query.name}|${status}`,
      fetch: (page, pageSize) =>
        listStatusPages({ page, pageSize, name: query.name, status }),
    })

  // 目标组下拉数据只在租户变化时取一次；状态页只能展示当前租户已授权的组。
  useEffect(() => {
    let active = true
    loadAllPagedItems((page, pageSize) =>
      listProbeTargetGroups({ page, pageSize })
    )
      .then((items) => {
        if (active) {
          setGroups(
            items.map((item) => ({ value: item.uid, label: item.name }))
          )
        }
      })
      .catch(() => {
        if (active) {
          setGroups([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  if (!session.loading && !session.can("status_page:read")) {
    return <Navigate to="/forbidden" replace />
  }

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

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("statusPage.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (status) {
    filterChips.push({
      key: "status",
      label: `${t("statusPage.status")}: ${t(`enum.${status}`)}`,
      onRemove: () => query.update({ status: "" }),
    })
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", status: "" })
  }

  function openCreate() {
    setEditing(null)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  function openEdit(item: StatusPageItem) {
    setEditing(item)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  async function copyLink(item: StatusPageItem) {
    try {
      await navigator.clipboard.writeText(publicURL(item))
      toast.success(t("action.copied"))
    } catch {
      toast.error(t("action.copyFailed"))
    }
  }

  async function toggleStatus(item: StatusPageItem) {
    if (togglingUid) {
      return
    }
    setTogglingUid(item.uid)
    const next: StatusPageStatus =
      item.status === "ACTIVE" ? "DISABLED" : "ACTIVE"
    try {
      await updateStatusPageStatus(item.uid, next)
      reload()
    } catch (err) {
      if (
        err instanceof ApiError &&
        (err.status === 403 || err.status === 404 || err.status >= 500)
      ) {
        return
      }
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setTogglingUid("")
    }
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setConfirmPending(true)
    setConfirmError("")
    try {
      if (confirm.action === "delete") {
        await deleteStatusPage(confirm.item.uid)
      } else {
        await regenerateStatusPageToken(confirm.item.uid)
        toast.success(t("statusPage.regenerated"))
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setConfirmPending(false)
    }
  }

  function rowActions(item: StatusPageItem): RowAction[] {
    const actions: RowAction[] = []
    if (canWrite) {
      actions.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => openEdit(item),
      })
      actions.push(
        item.status === "ACTIVE"
          ? {
              key: "disable",
              label: t("action.disable"),
              onSelect: () => void toggleStatus(item),
            }
          : {
              key: "enable",
              label: t("action.enable"),
              onSelect: () => void toggleStatus(item),
            }
      )
      actions.push({
        key: "regenerate",
        label: t("statusPage.regenerateAction"),
        onSelect: () => {
          setConfirmError("")
          setConfirm({ action: "regenerate", item })
        },
      })
    }
    if (canDelete) {
      actions.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirm({ action: "delete", item })
        },
      })
    }
    return actions
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("statusPage.title")}
        description={t("statusPage.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("action.create")}</Button>
        ) : null}
      </PageHeader>
      <div
        role="note"
        className="flex items-start gap-2 rounded-lg border border-info-border bg-info-soft px-3 py-2 text-sm text-info"
      >
        <Info aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
        <p className="min-w-0 break-words">{t("statusPage.credentialHint")}</p>
      </div>
      <ListFilter
        id="status-page-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={filterChips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={filterChips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel
            htmlFor="status-page-name"
            className="text-xs text-muted-foreground"
          >
            {t("statusPage.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="status-page-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("statusPage.searchPlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="status-page-status"
            className="text-xs text-muted-foreground"
          >
            {t("statusPage.status")}
          </FieldLabel>
          <Select
            value={draft.status || ALL_STATUS}
            onValueChange={(value) =>
              update(
                "status",
                value === ALL_STATUS ? "" : (value as StatusPageStatus)
              )
            }
          >
            <SelectTrigger id="status-page-status" className="w-full">
              <SelectValue placeholder={t("statusPage.allStatus")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.status} />
                  <span>
                    {draft.status
                      ? t(`enum.${draft.status}`)
                      : t("statusPage.allStatus")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_STATUS}>
                {t("statusPage.allStatus")}
              </SelectItem>
              <SelectItem value="ACTIVE">
                <span className="flex items-center gap-2">
                  <EnumIcon value="ACTIVE" />
                  <span>{t("enum.ACTIVE")}</span>
                </span>
              </SelectItem>
              <SelectItem value="DISABLED">
                <span className="flex items-center gap-2">
                  <EnumIcon value="DISABLED" />
                  <span>{t("enum.DISABLED")}</span>
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </Field>
      </ListFilter>
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
          <div className="grid gap-2">
            <p className="text-sm text-muted-foreground">
              {t("statusPage.empty")}
            </p>
            <p className="text-sm text-muted-foreground">
              {t("statusPage.emptyHint")}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("statusPage.name")}</TableHead>
                <TableHead>{t("statusPage.titleLabel")}</TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                >
                  {t("statusPage.componentCount")}
                </TableHead>
                <TableStatusHead>{t("statusPage.status")}</TableStatusHead>
                <TableHead className={TABLE_NOWRAP}>
                  {t("statusPage.publicLink")}
                </TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("statusPage.updatedAt")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-medium break-words">
                    {item.name}
                  </TableCell>
                  <TableCell className="max-w-md text-sm break-words">
                    {item.title}
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                  >
                    {item.targetGroupUids.length}
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => void copyLink(item)}
                    >
                      <Copy aria-hidden="true" data-icon="inline-start" />
                      {t("statusPage.copyLink")}
                    </Button>
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                  >
                    {formatDateTime(item.updatedAt) || "—"}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("statusPage.viewPublic")}
                      moreLabel={t("action.more")}
                      onDetail={() =>
                        window.open(
                          publicURL(item),
                          "_blank",
                          "noopener,noreferrer"
                        )
                      }
                      actions={rowActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <StatusPageDialog
        key={`${editing?.uid ?? "create"}-${editorSession}`}
        open={editorOpen}
        page={editing}
        groups={groups}
        onOpenChange={setEditorOpen}
        onSaved={() => {
          setEditorOpen(false)
          reload()
        }}
      />
      <ConfirmAction
        open={confirm !== null}
        title={
          confirm?.action === "delete"
            ? t("statusPage.deleteTitle")
            : t("statusPage.regenerateTitle")
        }
        description={
          confirm?.action === "delete"
            ? t("statusPage.deleteDescription")
            : t("statusPage.regenerateDescription")
        }
        confirmLabel={
          confirm?.action === "delete"
            ? t("action.delete")
            : t("statusPage.regenerateAction")
        }
        cancelLabel={t("action.cancel")}
        pending={confirmPending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(open) => {
          if (!open) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
