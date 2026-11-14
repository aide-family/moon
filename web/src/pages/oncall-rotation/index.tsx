import { Search } from "lucide-react"
import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import { listMembers } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import {
  deleteOnCallRotation,
  getOnCallRotationSchedule,
  listOnCallRotations,
  updateOnCallRotationStatus,
} from "@/api/oncall-rotation"
import type {
  OnCallRotationItem,
  OnCallRotationShiftItem,
} from "@/api/oncall-rotation/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import type { MultiSelectOption } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectGroup,
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
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { loadAllPagedItems } from "@/lib/select-pages"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { RotationDialog } from "./rotation-dialog"

// 排班默认看未来 14 天，与服务端缺省一致；服务端上限 30 天。
const SCHEDULE_DAYS = 14

interface ScheduleState {
  /** 请求签名：detail.uid + attempt，用它与当前签名比较派生出 loading。 */
  key: string
  shifts: OnCallRotationShiftItem[]
  failed: boolean
}

function normalizeStatus(value: string): string {
  return value === "ACTIVE" || value === "DISABLED" ? value : ""
}

function periodLabelKey(period: string): string {
  return period === "DAILY" ? "rotation.periodDaily" : "rotation.periodWeekly"
}

/** 值班轮换列表：配置 CRUD、启停与未来 14 天排班查看。 */
export function OnCallRotationPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({
    name: query.name,
    status: normalizeStatus(query.status),
  })
  const [members, setMembers] = useState<MultiSelectOption[]>([])
  const [open, setOpen] = useState(false)
  const [editorSession, setEditorSession] = useState(0)
  const [editing, setEditing] = useState<OnCallRotationItem | null>(null)
  const [detail, setDetail] = useState<OnCallRotationItem | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [confirm, setConfirm] = useState<OnCallRotationItem | null>(null)
  const [pending, setPending] = useState(false)
  const [confirmError, setConfirmError] = useState("")
  const [scheduleAttempt, setScheduleAttempt] = useState(0)
  const [schedule, setSchedule] = useState<ScheduleState | null>(null)
  const canWrite = session.can("notification_rotation:write")
  const canDelete = session.can("notification_rotation:delete")
  // 服务端列表接口没有 status 参数：状态筛选时取全量后在浏览器过滤再本地分页，
  // 保证分页总数与筛选结果一致；数据量按值班轮换的规模可以接受。
  const status = normalizeStatus(query.status)
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<OnCallRotationItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${status}`,
    fetch: async (page, pageSize) => {
      if (!status) {
        return listOnCallRotations({
          page,
          pageSize,
          name: query.name,
        })
      }
      const all = await loadAllPagedItems((current, currentSize) =>
        listOnCallRotations({
          page: current,
          pageSize: currentSize,
          name: query.name,
        })
      )
      const filtered = all.filter((item) => item.status === status)
      const start = (page - 1) * pageSize
      return {
        items: filtered.slice(start, start + pageSize),
        total: filtered.length,
      }
    },
  })

  // 成员候选只保留已加入的租户成员，与服务端校验一致。
  useEffect(() => {
    let active = true
    loadAllPagedItems<MemberItem>(async (page, pageSize) => {
      const result = await listMembers({ page, pageSize })
      return { items: result.items ?? [], total: result.total ?? 0 }
    })
      .then((items) => {
        if (!active) {
          return
        }
        setMembers(
          items
            .filter((item) => item.status === "JOINED")
            .map((item) => ({
              value: item.uid,
              label: item.name || item.email || item.uid,
              description: item.email,
            }))
        )
      })
      .catch(() => {
        if (active) {
          setMembers([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  // 删除或筛选收缩后页码可能超过最后一页：回到最后一页，避免空白列表与失效分页。
  useEffect(() => {
    if (loading || failed || query.page <= 1) {
      return
    }
    const lastPage = Math.max(1, Math.ceil(total / query.pageSize))
    if (
      lastPage !== query.page &&
      (items.length === 0 || query.page > lastPage)
    ) {
      query.update({ page: lastPage })
    }
    // query.update 每次渲染都是新函数，这里用页码与总数做真正的依赖。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, failed, items.length, total, query.page, query.pageSize])

  // 排班按详情 uid + 重试序号加载；loading 由签名比较派生，不在 effect 里同步改状态。
  const scheduleKey =
    detailOpen && detail ? `${detail.uid}:${scheduleAttempt}` : ""
  useEffect(() => {
    if (!scheduleKey || !detail) {
      return
    }
    let active = true
    const uid = detail.uid
    getOnCallRotationSchedule({ uid, days: SCHEDULE_DAYS })
      .then((result) => {
        if (active) {
          setSchedule({
            key: scheduleKey,
            shifts: result.shifts ?? [],
            failed: false,
          })
        }
      })
      .catch(() => {
        if (active) {
          setSchedule({ key: scheduleKey, shifts: [], failed: true })
        }
      })
    return () => {
      active = false
    }
  }, [scheduleKey, detail])
  const scheduleLoading = scheduleKey !== "" && schedule?.key !== scheduleKey

  if (!session.loading && !session.can("notification_rotation:read")) {
    return <Navigate to="/forbidden" replace />
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", status: "" })
  }

  /** 状态筛选只认 ACTIVE / DISABLED，其它值按「全部状态」展示。 */
  function statusLabel(value: string): string {
    if (value === "ACTIVE") {
      return t("rotation.filterActive")
    }
    if (value === "DISABLED") {
      return t("rotation.filterDisabled")
    }
    return t("rotation.allStatuses")
  }

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("rotation.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (status) {
    filterChips.push({
      key: "status",
      label: `${t("rotation.statusFilter")}: ${statusLabel(status)}`,
      onRemove: () => query.update({ status: "" }),
    })
  }

  function openCreate() {
    setEditing(null)
    setEditorSession((value) => value + 1)
    setOpen(true)
  }

  function openEdit(item: OnCallRotationItem) {
    setEditing(item)
    setEditorSession((value) => value + 1)
    setOpen(true)
  }

  function openDetail(item: OnCallRotationItem) {
    setDetail(item)
    setDetailOpen(true)
  }

  async function toggleStatus(item: OnCallRotationItem) {
    const next = item.status === "ACTIVE" ? "DISABLED" : "ACTIVE"
    setPending(true)
    try {
      await updateOnCallRotationStatus(item.uid, next)
      reload()
    } catch (err) {
      // 403 由 http 层跳 /forbidden，404/5xx 已由 http 层全局 toast；这里补 400/409 与网络错误。
      if (
        err instanceof ApiError &&
        (err.status === 403 || err.status === 404 || err.status >= 500)
      ) {
        return
      }
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runDelete() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await deleteOnCallRotation(confirm.uid)
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: OnCallRotationItem): RowAction[] {
    const list: RowAction[] = []
    if (canWrite) {
      list.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => openEdit(item),
      })
      list.push(
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
      list.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirm(item)
        },
      })
    }
    return list
  }

  function memberName(uid: string): string {
    return members.find((member) => member.value === uid)?.label ?? uid
  }

  function detailSections(item: OnCallRotationItem): DetailSection[] {
    const shifts = schedule?.key === scheduleKey ? schedule.shifts : []
    const scheduleFailedHere = schedule?.key === scheduleKey && schedule.failed

    let scheduleContent: ReactNode
    if (scheduleLoading) {
      scheduleContent = (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      )
    } else if (scheduleFailedHere) {
      scheduleContent = (
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-sm text-destructive">
            {t("rotation.scheduleFailed")}
          </p>
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setSchedule(null)
              setScheduleAttempt((value) => value + 1)
            }}
          >
            {t("action.retry")}
          </Button>
        </div>
      )
    } else if (shifts.length === 0) {
      scheduleContent = (
        <p className="text-sm text-muted-foreground">
          {t("rotation.scheduleEmpty")}
        </p>
      )
    } else {
      scheduleContent = (
        <ol className="grid gap-2">
          {shifts.map((shift) => (
            <li
              key={`${shift.startAt}-${shift.memberUid}`}
              className={cn(
                "grid gap-1 rounded-lg border px-3 py-2",
                shift.current
                  ? "border-primary/40 bg-primary-soft/40"
                  : "border-border/70 bg-muted/30"
              )}
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="min-w-0 text-sm font-medium break-words">
                  {shift.memberName || memberName(shift.memberUid)}
                </span>
                {shift.current ? (
                  <Badge
                    variant="outline"
                    className="h-5 shrink-0 rounded-md border-primary/40 bg-primary-soft px-2 font-normal text-primary-ink"
                  >
                    {t("rotation.currentShift")}
                  </Badge>
                ) : null}
              </div>
              <p className="font-mono text-xs text-muted-foreground tabular-nums">
                {formatDateTime(shift.startAt)} — {formatDateTime(shift.endAt)}
              </p>
              {shift.email ? (
                <p className="text-xs break-all text-muted-foreground">
                  {shift.email}
                </p>
              ) : null}
            </li>
          ))}
        </ol>
      )
    }

    return [
      {
        title: t("rotation.basicInfo"),
        layout: "inline",
        fields: [
          {
            label: t("rotation.status"),
            value: <StatusBadge value={item.status} />,
          },
          {
            label: t("rotation.period"),
            value: t(periodLabelKey(item.period)),
          },
          { label: t("rotation.timezone"), value: item.timezone || "—" },
          {
            label: t("rotation.handoffHour"),
            value: t("rotation.handoffHourValue", { hour: item.handoffHour }),
          },
          {
            label: t("rotation.startsAt"),
            value: formatDateTime(item.startsAt) || "—",
          },
          {
            label: t("rotation.updatedAt"),
            value: formatDateTime(item.updatedAt) || "—",
          },
        ],
      },
      {
        title: t("rotation.members"),
        description: t("rotation.memberCount", { count: item.members.length }),
        content:
          item.members.length === 0 ? (
            <p className="text-sm text-muted-foreground">—</p>
          ) : (
            <ol className="grid gap-2">
              {[...item.members]
                .sort((left, right) => left.position - right.position)
                .map((member, index) => (
                  <li
                    key={member.memberUid}
                    className="flex flex-wrap items-center gap-2 rounded-lg border border-border/70 bg-muted/30 px-3 py-2"
                  >
                    <span className="font-mono text-xs text-muted-foreground tabular-nums">
                      {index + 1}
                    </span>
                    <span className="min-w-0 text-sm font-medium break-words">
                      {member.memberName || memberName(member.memberUid)}
                    </span>
                    {member.email ? (
                      <span className="min-w-0 text-xs break-all text-muted-foreground">
                        {member.email}
                      </span>
                    ) : null}
                  </li>
                ))}
            </ol>
          ),
      },
      {
        title: t("rotation.scheduleTitle", { days: SCHEDULE_DAYS }),
        content: scheduleContent,
      },
      ...(item.remark
        ? [
            {
              title: t("rotation.remark"),
              content: (
                <p className="text-sm leading-relaxed break-words whitespace-pre-wrap">
                  {item.remark}
                </p>
              ),
            },
          ]
        : []),
    ]
  }

  return (
    <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("rotation.title")}
        description={t("rotation.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("rotation.createAction")}</Button>
        ) : null}
      </PageHeader>
      <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-4">
        <ListFilter
          id="rotation-filters"
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
              htmlFor="rotation-filter-name"
              className="text-xs text-muted-foreground"
            >
              {t("rotation.name")}
            </FieldLabel>
            <div className="relative">
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                id="rotation-filter-name"
                className="pl-9"
                value={draft.name}
                placeholder={t("rotation.searchPlaceholder")}
                onChange={(event) => update("name", event.target.value)}
              />
            </div>
          </Field>
          <Field className="w-40 shrink-0">
            <FieldLabel
              htmlFor="rotation-filter-status"
              className="text-xs text-muted-foreground"
            >
              {t("rotation.statusFilter")}
            </FieldLabel>
            <Select
              value={draft.status || "ALL"}
              onValueChange={(value) =>
                update("status", value === "ALL" ? "" : value)
              }
            >
              <SelectTrigger id="rotation-filter-status" className="w-full">
                <SelectValue placeholder={t("rotation.allStatuses")}>
                  <span className="flex items-center gap-2">
                    {draft.status ? <EnumIcon value={draft.status} /> : null}
                    <span>{statusLabel(draft.status)}</span>
                  </span>
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="ALL">
                    {t("rotation.allStatuses")}
                  </SelectItem>
                  {["ACTIVE", "DISABLED"].map((item) => (
                    <SelectItem key={item} value={item}>
                      <span className="flex items-center gap-2">
                        <EnumIcon value={item} />
                        <span>{statusLabel(item)}</span>
                      </span>
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>
        </ListFilter>
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && total > 0}
          onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
          onNext={() => query.update({ page: query.page + 1 })}
        >
          {loading ? (
            <p className="text-sm text-muted-foreground">
              {t("state.loading")}
            </p>
          ) : null}
          {!loading && failed ? (
            <div className="flex items-center gap-2">
              <p className="text-sm text-destructive">
                {loadError || t("state.failed")}
              </p>
              <Button variant="outline" onClick={() => reload()}>
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? (
            <div className="grid gap-1">
              <p className="text-sm text-muted-foreground">
                {t("rotation.empty")}
              </p>
              <p className="text-sm text-muted-foreground">
                {t("rotation.emptyHint")}
              </p>
            </div>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("rotation.name")}</TableHead>
                  <TableHead>{t("rotation.timezone")}</TableHead>
                  <TableHead
                    className={`${TABLE_COLUMN_WIDTH.type} ${TABLE_NOWRAP}`}
                  >
                    {t("rotation.period")}
                  </TableHead>
                  <TableHead
                    className={`${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                  >
                    {t("rotation.memberCountLabel")}
                  </TableHead>
                  <TableStatusHead>{t("rotation.status")}</TableStatusHead>
                  <TableHead
                    className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP}`}
                  >
                    {t("rotation.updatedAt")}
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
                    <TableCell className="font-mono text-xs break-all">
                      {item.timezone || "—"}
                    </TableCell>
                    <TableCell className={`text-sm ${TABLE_NOWRAP}`}>
                      {t(periodLabelKey(item.period))}
                    </TableCell>
                    <TableCell
                      className={`font-mono text-xs tabular-nums ${TABLE_COLUMN_WIDTH.level} ${TABLE_NOWRAP}`}
                    >
                      {item.members.length}
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
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => openDetail(item)}
                        actions={actions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
        </ListFrame>
      </div>
      <RotationDialog
        key={`${editing?.uid ?? "create"}-${editorSession}`}
        open={open}
        rotation={editing}
        members={members}
        onOpenChange={setOpen}
        onSaved={() => {
          setOpen(false)
          reload()
        }}
      />
      <DetailSheet
        open={detailOpen}
        size="40vw"
        title={detail?.name ?? ""}
        description={detail ? <StatusBadge value={detail.status} /> : undefined}
        sections={detail ? detailSections(detail) : []}
        onOpenChange={(value) => {
          setDetailOpen(value)
          if (!value) {
            setDetail(null)
          }
        }}
      />
      <ConfirmAction
        open={confirm !== null}
        title={t("rotation.deleteTitle")}
        description={t("rotation.deleteDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runDelete()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
