import { Search } from "lucide-react"
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
import {
  DateTimeRangePicker,
  type DateTimeRange,
} from "@/components/date-time-range-picker"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { FormField, KeyValueRowLabels } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
import { LevelBadge, StatusBadge } from "@/components/status-badge"
import { ListFilter } from "@/components/list-filter"
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
import { SearchableSelect } from "@/components/searchable-select"
import { useLabelOptions } from "@/hooks/use-label-options"
import { Input } from "@/components/ui/input"
import { Field, FieldLabel } from "@/components/ui/field"
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
import { useFilterDraft } from "@/hooks/use-filter-draft"
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
import { formatDateTimeValue, parseDateTimeValue } from "@/lib/date-range"
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

/** RFC3339 → 选择器使用的本地 YYYY-MM-DDTHH:mm（精确到分钟）。 */
function toDateTimeValue(value: string): string {
  const date = parseDateTimeValue(value)
  return date ? formatDateTimeValue(date) : ""
}

/** 选择器的本地 YYYY-MM-DDTHH:mm → RFC3339（提交给后端）。 */
function toRFC3339(value: string): string {
  const date = parseDateTimeValue(value)
  return date ? date.toISOString() : ""
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
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty, reset } = useFilterDraft({
    name: query.name,
    status: query.status,
  })
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
  const labelOptions = useLabelOptions()
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

  function applyFilters() {
    query.update({ name: draft.name.trim(), status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", status: "" })
    reset()
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
    setStartsAt(toDateTimeValue(item.startsAt))
    setEndsAt(toDateTimeValue(item.endsAt))
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

  /**
   * 开始/结束共用一个控件：值直接写回原有状态，组件给的是本地
   * YYYY-MM-DDTHH:mm，提交时再由 toRFC3339 转成 RFC3339。
   * 同时撤掉上一次提交留下的时间类错误，避免改了还挂着红字。
   */
  function updateTimeRange(range: DateTimeRange) {
    setStartsAt(range.from)
    setEndsAt(range.to)
    setFieldErrors((current) => {
      if (!current.startsAt && !current.endsAt && !current.timeRange) {
        return current
      }
      const next = { ...current }
      delete next.startsAt
      delete next.endsAt
      delete next.timeRange
      return next
    })
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

  // 开始/结束共用一个控件，校验结果合成一条就地提示：
  // 区间顺序（两端都填了才可能出错）优先，其次是两端的必填。
  const timeFieldError =
    fieldErrors.timeRange ?? fieldErrors.startsAt ?? fieldErrors.endsAt

  const filterChips: FilterChip[] = []
  if (query.name) {
    filterChips.push({
      key: "name",
      label: `${t("maintenanceWindow.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (query.status) {
    filterChips.push({
      key: "status",
      label: `${t("maintenanceWindow.statusFilter")}: ${
        query.status === "ACTIVE"
          ? t("maintenanceWindow.filterActive")
          : t("maintenanceWindow.filterDisabled")
      }`,
      onRemove: () => query.update({ status: "" }),
    })
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
        <ListFilter
          id="maintenance-filters"
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
              htmlFor="maintenance-name"
              className="text-xs text-muted-foreground"
            >
              {t("maintenanceWindow.name")}
            </FieldLabel>
            <div className="relative">
              <Search
                aria-hidden="true"
                className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                id="maintenance-name"
                className="pl-9"
                value={draft.name}
                placeholder={t("maintenanceWindow.searchPlaceholder")}
                onChange={(event) => update("name", event.target.value)}
              />
            </div>
          </Field>
          <Field className="w-40 shrink-0">
            <FieldLabel
              htmlFor="maintenance-status"
              className="text-xs text-muted-foreground"
            >
              {t("maintenanceWindow.statusFilter")}
            </FieldLabel>
            <Select
              value={draft.status || "ALL"}
              onValueChange={(value) =>
                update("status", value === "ALL" ? "" : value)
              }
            >
              <SelectTrigger id="maintenance-status" className="w-full">
                <SelectValue
                  placeholder={t("maintenanceWindow.allStatuses")}
                />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  <SelectItem value="ALL">
                    {t("maintenanceWindow.allStatuses")}
                  </SelectItem>
                  <SelectItem value="ACTIVE">
                    <span className="flex items-center gap-2">
                      <EnumIcon value="ACTIVE" />
                      <span>{t("maintenanceWindow.filterActive")}</span>
                    </span>
                  </SelectItem>
                  <SelectItem value="DISABLED">
                    <span className="flex items-center gap-2">
                      <EnumIcon value="DISABLED" />
                      <span>{t("maintenanceWindow.filterDisabled")}</span>
                    </span>
                  </SelectItem>
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
            {/* 自带 px-4 py-4 的焦点环净空，用 m-0 抵消 DialogBody 的 -m-1.5，字段与标题保持同一左边界。 */}
            <DialogBody className="m-0 grid gap-4 px-4 py-4">
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
                error={timeFieldError}
              >
                {/* 值由 DateTimeRangePicker 统一给出（本地 YYYY-MM-DDTHH:mm），
                    写回原有的 startsAt / endsAt 状态，提交时再转 RFC3339。
                    校验失败的原因显示在 label 行右侧。 */}
                <DateTimeRangePicker
                  id="maintenance-time-range"
                  ariaLabel={t("maintenanceWindow.timeRange")}
                  from={startsAt}
                  to={endsAt}
                  invalid={Boolean(timeFieldError)}
                  onChange={updateTimeRange}
                />
              </FormField>
              <FormField
                label={t("maintenanceWindow.labels")}
                hint={t("maintenanceWindow.labelsHint")}
                error={
                  labelPairErrors.find((item) => item.key)?.key ??
                  labelPairErrors.find((item) => item.value)?.value
                }
              >
                <div className="grid gap-2">
                  <KeyValueRowLabels
                    inline
                    keyLabel={t("escalation.labelKey")}
                    valueLabel={t("escalation.labelValue")}
                    trailing={<span aria-hidden="true" className="w-[4.5rem] shrink-0" />}
                  />
                  {labels.map((pair, index) => {
                    const keyOptions = labelOptions.map((option) => ({
                      value: option.key,
                      label: option.key,
                    }))
                    if (
                      pair.key !== "" &&
                      !keyOptions.some((option) => option.value === pair.key)
                    ) {
                      keyOptions.push({ value: pair.key, label: pair.key })
                    }
                    const values =
                      labelOptions.find((option) => option.key === pair.key)
                        ?.values ?? []
                    const valueOptions = values.map((value) => ({
                      value,
                      label: value,
                    }))
                    if (pair.value !== "" && !values.includes(pair.value)) {
                      valueOptions.push({ value: pair.value, label: pair.value })
                    }
                    return (
                    <div key={index} className="grid gap-1">
                      <div className="flex items-center gap-2">
                        <SearchableSelect
                          value={pair.key}
                          options={keyOptions}
                          placeholder={t("maintenanceWindow.labelKey")}
                          searchPlaceholder={t(
                            "alertRule.datasourceLabelKeySearch"
                          )}
                          emptyText={t("alertRule.customKeyHint")}
                          ariaLabel={t("maintenanceWindow.labelKey")}
                          allowCustom
                          customLabel={(value) =>
                            t("alertRule.customKey", { value })
                          }
                          customHint={t("alertRule.customKeyHint")}
                          clearLabel={t("alertRule.datasourceLabelClear")}
                          className="min-w-0 flex-1"
                          onChange={(key) => {
                            if (key === pair.key) {
                              return
                            }
                            const nextValues =
                              labelOptions.find((option) => option.key === key)
                                ?.values ?? []
                            const next = [...labels]
                            next[index] = {
                              key,
                              value: nextValues.includes(pair.value)
                                ? pair.value
                                : "",
                            }
                            setLabels(next)
                            setLabelPairErrors([])
                          }}
                        />
                        <SearchableSelect
                          value={pair.value}
                          options={valueOptions}
                          placeholder={t("maintenanceWindow.labelValue")}
                          searchPlaceholder={t(
                            "alertRule.datasourceLabelValueSearch"
                          )}
                          emptyText={t("alertRule.datasourceLabelValueEmpty")}
                          ariaLabel={t("maintenanceWindow.labelValue")}
                          allowCustom
                          customLabel={(value) =>
                            t("alertRule.datasourceLabelCustomValue", { value })
                          }
                          customHint={t("alertRule.datasourceLabelCustomHint")}
                          clearLabel={t("alertRule.datasourceLabelClear")}
                          className="min-w-0 flex-1"
                          onChange={(value) => {
                            const next = [...labels]
                            next[index] = { ...pair, value }
                            setLabels(next)
                            setLabelPairErrors([])
                          }}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          className="w-[4.5rem] shrink-0"
                          onClick={() =>
                            setLabels(
                              labels.filter((_, current) => current !== index)
                            )
                          }
                        >
                          {t("action.delete")}
                        </Button>
                      </div>
                    </div>
                    )
                  })}
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
