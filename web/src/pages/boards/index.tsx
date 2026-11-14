import { Search } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, Navigate, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import {
  deleteBoard,
  getBoard,
  listBoards,
  updateBoardStatus,
} from "@/api/board"
import type { BoardItem, BoardListItem, BoardStatus } from "@/api/board/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
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
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { ServerErrorPage } from "@/pages/server-error"

import { BoardDialog } from "./board-dialog"

// Radix Select 不允许空字符串选项，用哨兵值表示「全部状态」。
const ALL_STATUS = "ALL"

function normalizeStatus(value: string): BoardStatus | "" {
  return value === "ACTIVE" || value === "DISABLED" ? value : ""
}

/** 大盘列表：名称/状态筛选、分页，以及查看、编辑、启停、删除。 */
export function BoardPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    name: query.name,
    status: normalizeStatus(query.status),
  })
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editingBoard, setEditingBoard] = useState<BoardItem | null>(null)
  const [deleting, setDeleting] = useState<BoardListItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [togglingUid, setTogglingUid] = useState("")

  const canWrite = session.can("board:write")
  const canDelete = session.can("board:delete")
  // 状态筛选只认后端枚举，手改 URL 的非法值按不筛选处理。
  const status = normalizeStatus(query.status)
  const { items, total, loading, failed, failedStatus, error, reload } =
    usePagedList<BoardListItem>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${query.name}|${status}`,
      fetch: (page, pageSize) =>
        listBoards({ page, pageSize, name: query.name, status }),
    })

  if (!session.loading && !session.can("board:read")) {
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

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("board.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (status) {
    filterChips.push({
      key: "status",
      label: `${t("board.status")}: ${t(`enum.${status}`)}`,
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
    setEditingBoard(null)
    setEditorSession((value) => value + 1)
    setEditorOpen(true)
  }

  // 列表项不带面板，编辑前先取详情，避免整体保存时把已有面板当成删除。
  async function openEdit(item: BoardListItem) {
    try {
      const result = await getBoard(item.uid)
      setEditingBoard(result.board)
      setEditorSession((value) => value + 1)
      setEditorOpen(true)
    } catch (err) {
      // 403 由 http 层跳 /forbidden，404/5xx 已由 http 层全局 toast。
      if (
        err instanceof ApiError &&
        (err.status === 403 || err.status === 404 || err.status >= 500)
      ) {
        return
      }
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function toggleStatus(item: BoardListItem) {
    if (togglingUid) {
      return
    }
    setTogglingUid(item.uid)
    const next: BoardStatus = item.status === "ACTIVE" ? "DISABLED" : "ACTIVE"
    try {
      await updateBoardStatus(item.uid, next)
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

  async function runDelete() {
    if (!deleting) {
      return
    }
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteBoard(deleting.uid)
      setDeleting(null)
      reload()
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

  function rowActions(item: BoardListItem): RowAction[] {
    const actions: RowAction[] = []
    if (canWrite) {
      actions.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => void openEdit(item),
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
      <PageHeader title={t("board.title")} description={t("board.description")}>
        {canWrite ? (
          <Button onClick={openCreate}>{t("action.create")}</Button>
        ) : null}
      </PageHeader>
      <ListFilter
        id="board-filters"
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
            htmlFor="board-name"
            className="text-xs text-muted-foreground"
          >
            {t("board.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="board-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("board.searchPlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="board-status"
            className="text-xs text-muted-foreground"
          >
            {t("board.status")}
          </FieldLabel>
          <Select
            value={draft.status || ALL_STATUS}
            onValueChange={(value) =>
              update(
                "status",
                value === ALL_STATUS ? "" : (value as BoardStatus)
              )
            }
          >
            <SelectTrigger id="board-status" className="w-full">
              <SelectValue placeholder={t("board.allStatus")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.status} />
                  <span>
                    {draft.status
                      ? t(`enum.${draft.status}`)
                      : t("board.allStatus")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_STATUS}>{t("board.allStatus")}</SelectItem>
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
            <p className="text-sm text-muted-foreground">{t("board.empty")}</p>
            <p className="text-sm text-muted-foreground">
              {t("board.emptyHint")}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("board.name")}</TableHead>
                <TableHead>{t("board.descriptionLabel")}</TableHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                >
                  {t("board.panelCount")}
                </TableHead>
                <TableStatusHead>{t("board.status")}</TableStatusHead>
                <TableHead
                  className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                >
                  {t("board.updatedAt")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-medium break-words">
                    <Link
                      to={`/boards/${item.uid}`}
                      className="underline-offset-4 hover:underline"
                    >
                      {item.name}
                    </Link>
                  </TableCell>
                  <TableCell className="max-w-md text-sm text-muted-foreground">
                    <span className="break-words">
                      {item.description || "—"}
                    </span>
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                  >
                    {item.panelCount}
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell
                    className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                  >
                    {formatDateTime(item.updatedAt) || "—"}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("board.viewAction")}
                      moreLabel={t("action.more")}
                      onDetail={() => navigate(`/boards/${item.uid}`)}
                      actions={rowActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <BoardDialog
        key={`${editingBoard?.uid ?? "create"}-${editorSession}`}
        open={editorOpen}
        board={editingBoard}
        onOpenChange={setEditorOpen}
        onSaved={() => {
          setEditorOpen(false)
          reload()
        }}
      />
      <ConfirmAction
        open={deleting !== null}
        title={t("board.deleteTitle")}
        description={t("board.deleteDescription")}
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
