import { RefreshCw, Search, X } from "lucide-react"
import { useEffect, useState, type FormEvent, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import { listAlertRules } from "@/api/alert-rule"
import type { AlertRuleItem } from "@/api/alert-rule/types"
import {
  createAlertMaintenanceWindow,
  deleteAlertMaintenanceWindow,
  listAlertMaintenanceWindows,
  updateAlertMaintenanceWindow,
  updateAlertMaintenanceWindowStatus,
} from "@/api/alert-maintenance"
import type {
  AlertMaintenanceWindowItem,
  SaveAlertMaintenanceWindowParams,
} from "@/api/alert-maintenance/types"
import { selectDatasources } from "@/api/datasource"
import type { SelectItem as DatasourceOption } from "@/api/datasource/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
import { LevelBadge, StatusBadge } from "@/components/status-badge"
import { ListFrame } from "@/components/list-frame"
import { MultiSelect } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
import { Textarea } from "@/components/ui/textarea"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"
import {
  parseLabelPairs,
  serializeLabelPairs,
  type LabelPair,
} from "@/lib/labels"
import { loadAllPagedItems, loadAllSelectItems } from "@/lib/select-pages"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

// 名称、描述长度与后端 buf.validate 保持一致。
const NAME_MAX = 64
const DESCRIPTION_MAX = 1024

type WindowPhase = "NOT_STARTED" | "ONGOING" | "ENDED" | "DISABLED"

interface LabelPairErrors {
  key?: string
  value?: string
}

/** RFC3339 → datetime-local 输入值（本地时区，精确到分钟）。 */
function toDateTimeInput(value: string): string {
  if (!value) {
    return ""
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ""
  }
  const pad = (part: number) => String(part).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/** datetime-local 输入值按本地时区解释，提交前转成 RFC3339。 */
function toRFC3339(value: string): string {
  if (!value) {
    return ""
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? "" : date.toISOString()
}

/**
 * 展示状态由「启停开关 + 当前时间」共同决定，停用优先：
 * 未开始 / 进行中 / 已结束只在启用状态下按生效区间判断。
 */
function windowPhase(item: AlertMaintenanceWindowItem, now: number): WindowPhase {
  if (item.status === "DISABLED") {
    return "DISABLED"
  }
  const startsAt = new Date(item.startsAt).getTime()
  const endsAt = new Date(item.endsAt).getTime()
  if (Number.isFinite(startsAt) && now < startsAt) {
    return "NOT_STARTED"
  }
  if (Number.isFinite(endsAt) && now >= endsAt) {
    return "ENDED"
  }
  return "ONGOING"
}

function phaseLabelKey(phase: WindowPhase): string {
  switch (phase) {
    case "NOT_STARTED":
      return "maintenanceWindow.statusNotStarted"
    case "ONGOING":
      return "maintenanceWindow.statusOngoing"
    case "ENDED":
      return "maintenanceWindow.statusEnded"
    default:
      return "maintenanceWindow.statusDisabled"
  }
}

function phaseTone(phase: WindowPhase) {
  switch (phase) {
    case "ONGOING":
      return "success" as const
    case "NOT_STARTED":
      return "warning" as const
    case "ENDED":
      return "neutral" as const
    default:
      return "neutral" as const
  }
}

function toOptions<T extends { uid: string; name: string }>(items: T[]) {
  return items.map((item) => ({ value: item.uid, label: item.name }))
}

export function AlertMaintenancePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [nameDraft, setNameDraft] = useState(query.name)
  const [statusDraft, setStatusDraft] = useState(query.status)
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<AlertMaintenanceWindowItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${query.status}`,
    fetch: (page, pageSize) =>
      listAlertMaintenanceWindows({
        page,
        pageSize,
        name: query.name,
        status: query.status,
      }),
  })
  // 列表里的「未开始 / 进行中 / 已结束」跟随当前时间，定时重算。
  const [now, setNow] = useState(() => Date.now())
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [datasources, setDatasources] = useState<DatasourceOption[]>([])
  const [rules, setRules] = useState<AlertRuleItem[]>([])
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AlertMaintenanceWindowItem | null>(
    null
  )
  const [detail, setDetail] = useState<AlertMaintenanceWindowItem | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [startsAt, setStartsAt] = useState("")
  const [endsAt, setEndsAt] = useState("")
  const [labels, setLabels] = useState<LabelPair[]>([{ key: "", value: "" }])
  const [labelPairErrors, setLabelPairErrors] = useState<LabelPairErrors[]>([])
  const [levelUids, setLevelUids] = useState<string[]>([])
  const [datasourceUids, setDatasourceUids] = useState<string[]>([])
  const [ruleUids, setRuleUids] = useState<string[]>([])
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [confirm, setConfirm] = useState<AlertMaintenanceWindowItem | null>(
    null
  )
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("alert_maintenance:write")
  const canDelete = session.can("alert_maintenance:delete")

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 60000)
    return () => window.clearInterval(timer)
  }, [])

  // 匹配条件选项：级别、数据源、规则三份数据互相独立，任一失败不影响其他两项。
  useEffect(() => {
    let active = true
    void Promise.allSettled([
      loadAllPagedItems((page, pageSize) =>
        listAlertDictionaries({ page, pageSize, type: "ALERT" })
      ),
      loadAllSelectItems(
        (lastUid) => selectDatasources({ limit: 100, lastUid }),
        100
      ),
      loadAllPagedItems((page, pageSize) => listAlertRules({ page, pageSize })),
    ]).then(([levelResult, datasourceResult, ruleResult]) => {
      if (!active) {
        return
      }
      setLevels(levelResult.status === "fulfilled" ? levelResult.value : [])
      setDatasources(
        datasourceResult.status === "fulfilled" ? datasourceResult.value : []
      )
      setRules(ruleResult.status === "fulfilled" ? ruleResult.value : [])
    })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  // 删除或筛选收缩后页码可能超过最后一页：回到最后一页，避免空白列表和失效的分页。
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

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  function applyFilters(event: FormEvent) {
    event.preventDefault()
    query.update({ name: nameDraft.trim(), status: statusDraft })
  }

  function clearFilters() {
    setNameDraft("")
    setStatusDraft("")
    query.update({ name: "", status: "" })
  }

  function resetForm() {
    setEditing(null)
    setName("")
    setDescription("")
    setStartsAt("")
    setEndsAt("")
    setLabels([{ key: "", value: "" }])
    setLabelPairErrors([])
    setLevelUids([])
    setDatasourceUids([])
    setRuleUids([])
    setError("")
    setFieldErrors({})
  }

  function openCreate() {
    resetForm()
    setOpen(true)
  }

  function openEdit(item: AlertMaintenanceWindowItem) {
    setEditing(item)
    setName(item.name)
    setDescription(item.description ?? "")
    setStartsAt(toDateTimeInput(item.startsAt))
    setEndsAt(toDateTimeInput(item.endsAt))
    setLabels(parseLabelPairs(item.labels))
    setLabelPairErrors([])
    setLevelUids(item.levelUids ?? [])
    setDatasourceUids(item.datasourceUids ?? [])
    setRuleUids(item.ruleUids ?? [])
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openDetail(item: AlertMaintenanceWindowItem) {
    setDetail(item)
    setDetailOpen(true)
  }

  /** 标签行逐项校验：键不能有值无键、不能重复，值必须是合法正则。 */
  function validateLabelPairs(pairs: LabelPair[]): LabelPairErrors[] {
    const seen = new Set<string>()
    const errors = pairs.map((pair) => {
      const key = pair.key.trim()
      const value = pair.value.trim()
      const next: LabelPairErrors = {}
      if (key && value) {
        try {
          new RegExp(`^(?:${value})$`)
        } catch {
          next.value = t("maintenanceWindow.labelValueInvalid")
        }
      }
      if (key) {
        if (seen.has(key)) {
          next.key = t("maintenanceWindow.labelDuplicateKey")
        }
        seen.add(key)
      } else if (value) {
        next.key = t("maintenanceWindow.labelKeyRequired")
      }
      return next
    })
    return errors
  }

  async function onSave(event: FormEvent) {
    event.preventDefault()
    const startsAtValue = toRFC3339(startsAt)
    const endsAtValue = toRFC3339(endsAt)
    const pairErrors = validateLabelPairs(labels)
    const pairInvalid = pairErrors.some((item) => item.key || item.value)
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("maintenanceWindow.nameRequired"))],
      ["name", maxLength(name, NAME_MAX, t("maintenanceWindow.nameTooLong"))],
      [
        "description",
        maxLength(
          description,
          DESCRIPTION_MAX,
          t("maintenanceWindow.descriptionTooLong")
        ),
      ],
      [
        "startsAt",
        requireTrimmed(startsAt, t("maintenanceWindow.startsAtRequired")),
      ],
      ["endsAt", requireTrimmed(endsAt, t("maintenanceWindow.endsAtRequired"))],
      [
        "timeRange",
        startsAtValue && endsAtValue && endsAtValue <= startsAtValue
          ? t("maintenanceWindow.timeRangeInvalid")
          : null,
      ],
    ])
    setLabelPairErrors(pairErrors)
    if (hasFieldErrors(errors) || pairInvalid) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload: SaveAlertMaintenanceWindowParams = {
        name: name.trim(),
        description: description.trim(),
        labels: serializeLabelPairs(labels),
        levelUids,
        datasourceUids,
        ruleUids,
        startsAt: startsAtValue,
        endsAt: endsAtValue,
      }
      if (editing) {
        await updateAlertMaintenanceWindow({ ...payload, uid: editing.uid })
      } else {
        await createAlertMaintenanceWindow(payload)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function toggleStatus(item: AlertMaintenanceWindowItem) {
    const next = item.status === "ACTIVE" ? "DISABLED" : "ACTIVE"
    setPending(true)
    try {
      await updateAlertMaintenanceWindowStatus(item.uid, next)
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
      await deleteAlertMaintenanceWindow(confirm.uid)
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AlertMaintenanceWindowItem): RowAction[] {
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
              label: t("maintenanceWindow.filterDisabled"),
              onSelect: () => void toggleStatus(item),
            }
          : {
              key: "enable",
              label: t("maintenanceWindow.filterActive"),
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

  function matcherSummary(item: AlertMaintenanceWindowItem): string {
    const labelCount = parseLabelPairs(item.labels).filter(
      (pair) => pair.key.trim() !== ""
    ).length
    const levelCount = item.levelUids?.length ?? 0
    const datasourceCount = item.datasourceUids?.length ?? 0
    const ruleCount = item.ruleUids?.length ?? 0
    if (
      labelCount === 0 &&
      levelCount === 0 &&
      datasourceCount === 0 &&
      ruleCount === 0
    ) {
      return t("maintenanceWindow.allEvents")
    }
    return t("maintenanceWindow.matchSummary", {
      labels: labelCount,
      levels: levelCount,
      datasources: datasourceCount,
      rules: ruleCount,
    })
  }

  function levelTags(uids: string[]): ReactNode {
    if (uids.length === 0) {
      return null
    }
    return (
      <span className="flex flex-wrap gap-1.5">
        {uids.map((uid) => {
          const level = levels.find((entry) => entry.uid === uid)
          return (
            <LevelBadge key={uid} name={level?.name ?? uid} color={level?.color} />
          )
        })}
      </span>
    )
  }

  function nameTags(
    uids: string[],
    dictionary: Array<{ uid: string; name: string }>
  ): ReactNode {
    if (uids.length === 0) {
      return null
    }
    return (
      <span className="flex flex-wrap gap-1.5">
        {uids.map((uid) => (
          <Badge
            key={uid}
            variant="outline"
            className="h-auto max-w-full rounded-md px-2 py-0.5 font-normal break-all whitespace-normal"
          >
            {dictionary.find((entry) => entry.uid === uid)?.name ?? uid}
          </Badge>
        ))}
      </span>
    )
  }

  function detailSections(item: AlertMaintenanceWindowItem): DetailSection[] {
    const phase = windowPhase(item, now)
    return [
      {
        title: t("maintenanceWindow.matchConditions"),
        description: matcherSummary(item),
        layout: "inline",
        fields: [
          {
            label: t("maintenanceWindow.labels"),
            value: <LabelTags labels={item.labels} />,
          },
          {
            label: t("maintenanceWindow.levels"),
            value: levelTags(item.levelUids ?? []),
          },
          {
            label: t("maintenanceWindow.datasources"),
            value: nameTags(item.datasourceUids ?? [], datasources),
          },
          {
            label: t("maintenanceWindow.rules"),
            value: nameTags(item.ruleUids ?? [], rules),
          },
        ],
      },
      {
        title: t("maintenanceWindow.timeRange"),
        layout: "inline",
        fields: [
          {
            label: t("maintenanceWindow.startsAt"),
            value: formatDateTime(item.startsAt) || "—",
          },
          {
            label: t("maintenanceWindow.endsAt"),
            value: formatDateTime(item.endsAt) || "—",
          },
          {
            label: t("maintenanceWindow.status"),
            value: (
              <StatusBadge
                value={phase}
                label={t(phaseLabelKey(phase))}
                tone={phaseTone(phase)}
              />
            ),
          },
          {
            label: t("maintenanceWindow.createdAt"),
            value: formatDateTime(item.createdAt) || "—",
          },
          {
            label: t("maintenanceWindow.createdBy"),
            value: item.createdBy || "—",
          },
        ],
      },
      ...(item.description
        ? [
            {
              title: t("maintenanceWindow.descriptionLabel"),
              content: (
                <p className="text-sm leading-relaxed break-words whitespace-pre-wrap">
                  {item.description}
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
        title={t("maintenanceWindow.title")}
        description={t("maintenanceWindow.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("maintenanceWindow.createAction")}</Button>
        ) : null}
      </PageHeader>
      <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-4">
        <form
          className="flex min-w-0 flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3 sm:p-4"
          onSubmit={applyFilters}
        >
          <div className="flex min-w-0 flex-1 basis-56 items-center gap-2">
            <Input
              className="min-w-0 flex-1"
              value={nameDraft}
              placeholder={t("maintenanceWindow.searchPlaceholder")}
              aria-label={t("maintenanceWindow.name")}
              onChange={(event) => setNameDraft(event.target.value)}
            />
          </div>
          <Select
            value={statusDraft || "ALL"}
            onValueChange={(value) =>
              setStatusDraft(value === "ALL" ? "" : value)
            }
          >
            <SelectTrigger
              className="w-40 shrink-0"
              aria-label={t("maintenanceWindow.statusFilter")}
            >
              <SelectValue placeholder={t("maintenanceWindow.allStatuses")} />
            </SelectTrigger>
            <SelectContent>
              <SelectGroup>
                <SelectItem value="ALL">
                  {t("maintenanceWindow.allStatuses")}
                </SelectItem>
                <SelectItem value="ACTIVE">
                  {t("maintenanceWindow.filterActive")}
                </SelectItem>
                <SelectItem value="DISABLED">
                  {t("maintenanceWindow.filterDisabled")}
                </SelectItem>
              </SelectGroup>
            </SelectContent>
          </Select>
          <div className="flex shrink-0 items-center gap-2">
            {query.name || query.status ? (
              <Button type="button" variant="ghost" onClick={clearFilters}>
                <X aria-hidden data-icon="inline-start" />
                {t("filter.clear")}
              </Button>
            ) : null}
            <Button type="submit" variant="outline">
              <Search aria-hidden data-icon="inline-start" />
              {t("action.search")}
            </Button>
          </div>
        </form>
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && total > 0}
          actions={
            <Button
              type="button"
              variant="outline"
              size="icon"
              className="size-8"
              aria-label={t("action.refresh")}
              title={t("action.refresh")}
              disabled={loading}
              onClick={() => reload()}
            >
              <RefreshCw
                aria-hidden
                className={cn(loading && "animate-spin")}
              />
            </Button>
          }
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
            <p className="text-sm text-muted-foreground">
              {t("maintenanceWindow.empty")}
            </p>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("maintenanceWindow.name")}</TableHead>
                  <TableHead>{t("maintenanceWindow.matchConditions")}</TableHead>
                  <TableHead
                    className={cn(TABLE_COLUMN_WIDTH.time, TABLE_NOWRAP)}
                  >
                    {t("maintenanceWindow.startsAt")}
                  </TableHead>
                  <TableHead
                    className={cn(TABLE_COLUMN_WIDTH.time, TABLE_NOWRAP)}
                  >
                    {t("maintenanceWindow.endsAt")}
                  </TableHead>
                  <TableStatusHead>
                    {t("maintenanceWindow.status")}
                  </TableStatusHead>
                  <TableActionsHead>
                    {t("action.actions")}
                  </TableActionsHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => {
                  const phase = windowPhase(item, now)
                  return (
                    <TableRow key={item.uid}>
                      <TableCell className="font-medium break-words">
                        {item.name}
                      </TableCell>
                      <TableCell className="max-w-md text-sm text-muted-foreground">
                        <span className="break-words">
                          {matcherSummary(item)}
                        </span>
                      </TableCell>
                      <TableCell
                        className={cn(
                          TABLE_COLUMN_WIDTH.time,
                          TABLE_NOWRAP,
                          "font-mono text-xs text-muted-foreground tabular-nums"
                        )}
                      >
                        {formatDateTime(item.startsAt) || "—"}
                      </TableCell>
                      <TableCell
                        className={cn(
                          TABLE_COLUMN_WIDTH.time,
                          TABLE_NOWRAP,
                          "font-mono text-xs text-muted-foreground tabular-nums"
                        )}
                      >
                        {formatDateTime(item.endsAt) || "—"}
                      </TableCell>
                      <TableCell className={TABLE_NOWRAP}>
                        <StatusBadge
                          value={phase}
                          label={t(phaseLabelKey(phase))}
                          tone={phaseTone(phase)}
                        />
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
                  )
                })}
              </TableBody>
            </Table>
          ) : null}
        </ListFrame>
      </div>
      <Dialog
        open={open}
        onOpenChange={(value) => {
          if (!value) {
            setOpen(false)
          }
        }}
      >
        <DialogContent className="gap-0 p-0 sm:max-w-3xl">
          <DialogHeader className="shrink-0 border-b px-4 py-3 pr-12">
            <DialogTitle>
              {editing
                ? t("maintenanceWindow.editTitle")
                : t("maintenanceWindow.createTitle")}
            </DialogTitle>
          </DialogHeader>
          <form
            noValidate
            className="flex min-h-0 flex-1 flex-col"
            onSubmit={onSave}
          >
            <DialogBody className="grid gap-4 px-4 py-4">
              <FormField
                label={t("maintenanceWindow.name")}
                htmlFor="maintenance-name"
                required
                error={fieldErrors.name}
              >
                <Input
                  id="maintenance-name"
                  value={name}
                  placeholder={t("maintenanceWindow.namePlaceholder")}
                  onChange={(event) => setName(event.target.value)}
                  required
                />
              </FormField>
              <FormField
                label={t("maintenanceWindow.descriptionLabel")}
                htmlFor="maintenance-description"
                error={fieldErrors.description}
              >
                <Textarea
                  id="maintenance-description"
                  value={description}
                  placeholder={t("maintenanceWindow.descriptionPlaceholder")}
                  onChange={(event) => setDescription(event.target.value)}
                />
              </FormField>
              <FormField
                label={t("maintenanceWindow.timeRange")}
                required
                error={fieldErrors.timeRange}
              >
                <div className="grid gap-2 sm:grid-cols-2">
                  <FormField
                    label={t("maintenanceWindow.startsAt")}
                    htmlFor="maintenance-starts-at"
                    required
                    error={fieldErrors.startsAt}
                  >
                    <Input
                      id="maintenance-starts-at"
                      type="datetime-local"
                      value={startsAt}
                      onChange={(event) => setStartsAt(event.target.value)}
                    />
                  </FormField>
                  <FormField
                    label={t("maintenanceWindow.endsAt")}
                    htmlFor="maintenance-ends-at"
                    required
                    error={fieldErrors.endsAt}
                  >
                    <Input
                      id="maintenance-ends-at"
                      type="datetime-local"
                      value={endsAt}
                      onChange={(event) => setEndsAt(event.target.value)}
                    />
                  </FormField>
                </div>
              </FormField>
              <FormField
                label={t("maintenanceWindow.labels")}
                hint={t("maintenanceWindow.labelsHint")}
              >
                <div className="grid gap-2">
                  {labels.map((pair, index) => (
                    <div key={index} className="grid gap-1">
                      <div className="flex items-center gap-2">
                        <Input
                          value={pair.key}
                          placeholder={t("maintenanceWindow.labelKey")}
                          aria-label={t("maintenanceWindow.labelKey")}
                          onChange={(event) => {
                            const next = [...labels]
                            next[index] = { ...pair, key: event.target.value }
                            setLabels(next)
                            setLabelPairErrors([])
                          }}
                        />
                        <Input
                          value={pair.value}
                          placeholder={t("maintenanceWindow.labelValue")}
                          aria-label={t("maintenanceWindow.labelValue")}
                          onChange={(event) => {
                            const next = [...labels]
                            next[index] = {
                              ...pair,
                              value: event.target.value,
                            }
                            setLabels(next)
                            setLabelPairErrors([])
                          }}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          onClick={() =>
                            setLabels(
                              labels.filter((_, current) => current !== index)
                            )
                          }
                        >
                          {t("action.delete")}
                        </Button>
                      </div>
                      {labelPairErrors[index]?.key ? (
                        <p role="alert" className="text-xs text-destructive">
                          {labelPairErrors[index].key}
                        </p>
                      ) : null}
                      {labelPairErrors[index]?.value ? (
                        <p role="alert" className="text-xs text-destructive">
                          {labelPairErrors[index].value}
                        </p>
                      ) : null}
                    </div>
                  ))}
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      setLabels([...labels, { key: "", value: "" }])
                    }
                  >
                    {t("maintenanceWindow.addLabel")}
                  </Button>
                </div>
              </FormField>
              <FormField
                label={t("maintenanceWindow.levels")}
                htmlFor="maintenance-levels"
                hint={t("maintenanceWindow.levelsHint")}
              >
                <MultiSelect
                  id="maintenance-levels"
                  values={levelUids}
                  options={toOptions(levels)}
                  onChange={setLevelUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("maintenanceWindow.noLevels")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("maintenanceWindow.datasources")}
                htmlFor="maintenance-datasources"
                hint={t("maintenanceWindow.datasourcesHint")}
              >
                <MultiSelect
                  id="maintenance-datasources"
                  values={datasourceUids}
                  options={toOptions(datasources)}
                  onChange={setDatasourceUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("maintenanceWindow.noDatasources")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
              <FormField
                label={t("maintenanceWindow.rules")}
                htmlFor="maintenance-rules"
                hint={t("maintenanceWindow.rulesHint")}
              >
                <MultiSelect
                  id="maintenance-rules"
                  values={ruleUids}
                  options={toOptions(rules)}
                  onChange={setRuleUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("maintenanceWindow.noRules")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                />
              </FormField>
            </DialogBody>
            {error ? (
              <p className="text-sm text-destructive">{error}</p>
            ) : null}
            <DialogFooter className="mx-0 mb-0 shrink-0 border-t px-4 py-3">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
              >
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detailOpen}
        size="40vw"
        title={detail?.name ?? ""}
        description={
          detail ? (
            <StatusBadge
              value={windowPhase(detail, now)}
              label={t(phaseLabelKey(windowPhase(detail, now)))}
              tone={phaseTone(windowPhase(detail, now))}
            />
          ) : undefined
        }
        sections={detail ? detailSections(detail) : []}
        onOpenChange={setDetailOpen}
      />
      <ConfirmAction
        open={confirm !== null}
        title={t("maintenanceWindow.deleteTitle")}
        description={t("maintenanceWindow.deleteDescription")}
        confirmLabel={t("action.confirm")}
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
