import { Search } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import {
  createDatasource,
  deleteDatasource,
  listDatasources,
  updateDatasource,
  updateDatasourceStatus,
} from "@/api/datasource"
import type {
  CreateDatasourceParams,
  DatasourceItem,
} from "@/api/datasource/types"
import { selectNotificationGroups } from "@/api/notification-group"
import type { SelectItem as NotificationGroupSelectItem } from "@/api/notification-group/types"
import { useSession } from "@/auth/session"
import { authorizedDatasourceDrivers } from "@/lib/feature-catalog"
import { BrandIcon } from "@/components/brand-icon"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { FormField } from "@/components/form-field"
import { ListFilter } from "@/components/list-filter"
import { MultiSelect } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { ListFrame } from "@/components/list-frame"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { StepForm } from "@/components/step-form"
import {
  TableActionsHead,
  TableStatusHead,
  TableTypeHead,
} from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
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
import { cn } from "@/lib/class-names"
import { enumLabel } from "@/lib/enum-label"
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"
import {
  parseLabelPairs,
  serializeLabelPairs,
  type LabelPair,
} from "@/lib/labels"
import { loadAllSelectItems } from "@/lib/select-pages"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { DatasourceDetailSheet } from "@/pages/datasource/detail"
import { MetadataPanel } from "@/pages/datasource/metadata-panel"
import { QueryPanel } from "@/pages/datasource/query-panel"

const types = ["METRICS", "LOGS", "TRACE"]
const drivers: Record<string, string[]> = {
  METRICS: ["METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS"],
  LOGS: ["LOGS_ELASTICSEARCH", "LOGS_LOKI", "LOGS_VLOGS"],
  TRACE: ["TRACE_JAEGER"],
}

// 与 proto 的 buf.validate 对齐：窗口 0..1440 分钟，阈值 0..1000 次，
// 绑定的通知组最多 20 个。0 在服务端表示“不触发”，所以只在启用时要求大于 0。
const MAX_AVAILABILITY_WINDOW_MINUTES = 1440
const MAX_AVAILABILITY_THRESHOLD = 1000
const MAX_AVAILABILITY_GROUPS = 20

/** 整数范围校验：空值、非整数和越界都返回同一条字段错误。 */
function availabilityCountError(
  value: string,
  minimum: number,
  maximum: number,
  message: string
): string | null {
  const trimmed = value.trim()
  if (!/^\d+$/.test(trimmed)) {
    return message
  }
  const count = Number(trimmed)
  return count < minimum || count > maximum ? message : null
}

interface Draft {
  name: string
  type: string
  driver: string
  endpoint: string
  indexPattern: string
  timeField: string
  username: string
  password: string
  secret: string
  levelUid: string
  labels: LabelPair[]
  remark: string
  availabilityAlertEnabled: boolean
  failureWindowMinutes: string
  failureThreshold: string
  notificationGroupUids: string[]
}

const emptyDraft: Draft = {
  name: "",
  type: "METRICS",
  driver: "METRICS_PROMETHEUS",
  endpoint: "",
  indexPattern: "",
  timeField: "",
  username: "",
  password: "",
  secret: "",
  levelUid: "",
  labels: [{ key: "", value: "" }],
  remark: "",
  // 新数据源默认关闭告警：5 分钟窗口、3 次失败，不绑定通知组。
  availabilityAlertEnabled: false,
  failureWindowMinutes: "5",
  failureThreshold: "3",
  notificationGroupUids: [],
}

export function DatasourcePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const type = query.type
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({ name: query.name, type })
  // 三个抽屉共用列表项：列表响应已经包含详情字段，不需要再按 uid 拉一次详情。
  const [detailItem, setDetailItem] = useState<DatasourceItem | null>(null)
  const [metadataItem, setMetadataItem] = useState<DatasourceItem | null>(null)
  const [queryItem, setQueryItem] = useState<DatasourceItem | null>(null)
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [notificationGroups, setNotificationGroups] = useState<
    NotificationGroupSelectItem[]
  >([])
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [open, setOpen] = useState(false)
  const [editingItem, setEditingItem] = useState<DatasourceItem | null>(null)
  const [formStep, setFormStep] = useState(0)
  const [form, setForm] = useState<Draft>(emptyDraft)
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [confirm, setConfirm] = useState<{
    action: "disable" | "delete"
    uid: string
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [shownAction, setShownAction] = useState<"disable" | "delete">(
    "disable"
  )
  const canWrite = session.can("datasource:write")
  // 套餐按驱动细分：只展示租户已开通的数据源类型，未授权类型既不出现也不能新建。
  const allowedDrivers = useMemo(
    () => authorizedDatasourceDrivers(session.features),
    [session.features]
  )
  const allowedTypes = useMemo(
    () =>
      types.filter((item) =>
        (drivers[item] ?? []).some((driver) => allowedDrivers.includes(driver))
      ),
    [allowedDrivers]
  )

  function draftForAllowedTypes(): Draft {
    const type = allowedTypes[0] ?? "METRICS"
    const driver =
      (drivers[type] ?? []).find((item) => allowedDrivers.includes(item)) ?? ""
    return { ...emptyDraft, type, driver }
  }
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    version,
    reload,
    fail,
  } = usePagedList<DatasourceItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${type}`,
    fetch: (page, pageSize) =>
      listDatasources({ page, pageSize, name: query.name, type }),
  })

  function askConfirm(action: "disable" | "delete", uid: string) {
    setShownAction(action)
    setConfirm({ action, uid })
  }

  const canReadLevels = session.can("alert_dictionary:read")
  const canReadNotificationGroups = session.can("notification_group:read")

  useEffect(() => {
    let active = true
    async function load() {
      if (!canWrite) {
        return
      }
      // 等级和通知组是跨模块的可选配置：当前租户没开通对应模块时跳过请求，
      // 否则 403 会被全局拦截成「无权限」页。
      if (canReadLevels) {
        try {
          const result = await listAlertDictionaries({
            page: 1,
            pageSize: 100,
            type: "DATASOURCE",
          })
          if (active) {
            setLevels(result.items ?? [])
          }
        } catch {
          if (active) {
            setLevels([])
          }
        }
      }
      // 可用性告警只能绑定启用中的通知组；编辑时停用组的名称由详情响应补上。
      if (canReadNotificationGroups) {
        try {
          const groups = await loadAllSelectItems(
            (lastUid) =>
              selectNotificationGroups({
                limit: 100,
                lastUid,
                activeOnly: true,
              }),
            100
          )
          if (active) {
            setNotificationGroups(groups)
          }
        } catch {
          if (active) {
            setNotificationGroups([])
          }
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [
    canWrite,
    canReadLevels,
    canReadNotificationGroups,
    session.namespaceUid,
    version,
  ])

  function openCreate() {
    setEditingItem(null)
    setForm(draftForAllowedTypes())
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setOpen(true)
  }

  // 编辑时把当前数据源已绑定的通知组补进选项：select 接口带 activeOnly，
  // 停用的通知组不会返回，不补的话下拉框只能显示 uid。
  // 名称取自列表响应的 notificationGroupNames（只用于展示）；只有名称与
  // uid 数量一致时才按下标补齐，避免把名称错配到别的通知组上。
  const notificationGroupOptions = useMemo(() => {
    const options = notificationGroups.map((item) => ({
      value: item.uid,
      label: item.name,
    }))
    const known = new Set(options.map((option) => option.value))
    // 新建时不要把上一次编辑的数据源绑定补进选项。
    const uids = editingItem?.notificationGroupUids ?? []
    const names = editingItem?.notificationGroupNames ?? []
    if (uids.length === names.length) {
      uids.forEach((uid, index) => {
        const name = names[index]
        if (name && !known.has(uid)) {
          options.push({ value: uid, label: name })
        }
      })
    }
    return options
  }, [notificationGroups, editingItem])

  function openEdit(item: DatasourceItem) {
    setEditingItem(item)
    setForm({
      name: item.name,
      type: item.type,
      driver: item.driver,
      endpoint: item.endpoint,
      indexPattern: item.indexPattern ?? "",
      timeField: item.timeField ?? "",
      username: item.username,
      password: "",
      secret: "",
      levelUid: item.levelUid,
      labels: parseLabelPairs(item.labels),
      remark: item.remark,
      availabilityAlertEnabled: Boolean(item.availabilityAlertEnabled),
      failureWindowMinutes: String(item.failureWindowMinutes ?? 5),
      failureThreshold: String(item.failureThreshold ?? 3),
      notificationGroupUids: item.notificationGroupUids ?? [],
    })
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setOpen(true)
  }

  function validateAvailabilityStep(): FieldErrors {
    // 服务端只在启用、窗口大于 0 且阈值大于 0 时触发。关闭时允许保存 0
    // （语义为不触发）；启用后必须大于 0，否则开关打开却永远不会产生事件。
    const minimum = form.availabilityAlertEnabled ? 1 : 0
    return collectFieldErrors([
      [
        "failureWindowMinutes",
        availabilityCountError(
          form.failureWindowMinutes,
          minimum,
          MAX_AVAILABILITY_WINDOW_MINUTES,
          t("datasource.failureWindowMinutesInvalid")
        ),
      ],
      [
        "failureThreshold",
        availabilityCountError(
          form.failureThreshold,
          minimum,
          MAX_AVAILABILITY_THRESHOLD,
          t("datasource.failureThresholdInvalid")
        ),
      ],
      [
        "notificationGroupUids",
        form.notificationGroupUids.length > MAX_AVAILABILITY_GROUPS
          ? t("datasource.notificationGroupsTooMany")
          : null,
      ],
    ])
  }

  function validateDatasourceStep(step: number): FieldErrors {
    if (step === 0) {
      const errors = collectFieldErrors([
        ["name", requireTrimmed(form.name, t("form.required"))],
        ["name", maxLength(form.name, 64, t("form.nameTooLong"))],
        ["type", requireTrimmed(form.type, t("form.required"))],
        ["driver", requireTrimmed(form.driver, t("form.required"))],
        ["endpoint", requireTrimmed(form.endpoint, t("form.required"))],
      ])
      // 后端只接受 http(s) 根地址；`htp://` 这类拼写错误要在保存时拦住，
      // 否则要等到查询时才由数据层报一个没有上下文的 invalid request。
      const endpoint = form.endpoint.trim()
      if (!errors.endpoint && endpoint && !/^https?:\/\/\S+/i.test(endpoint)) {
        errors.endpoint = t("form.invalidUrl")
      }
      return errors
    }
    if (step === 2) {
      return validateAvailabilityStep()
    }
    return {}
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    // 保存入口在最后一步，但基础信息可能被回退修改过，两步一起校验；
    // 出错时跳回对应步骤，字段错误仍显示在输入框下。
    const errors = {
      ...validateDatasourceStep(0),
      ...validateDatasourceStep(2),
    }
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setFormStep(
        errors.failureWindowMinutes ||
          errors.failureThreshold ||
          errors.notificationGroupUids
          ? 2
          : 0
      )
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload: CreateDatasourceParams = {
        name: form.name.trim(),
        type: form.type,
        driver: form.driver,
        endpoint: form.endpoint.trim(),
        indexPattern: form.indexPattern,
        timeField: form.timeField,
        username: form.username,
        password: form.password,
        secret: form.secret,
        levelUid: form.levelUid,
        remark: form.remark,
        labels: serializeLabelPairs(form.labels),
        availabilityAlertEnabled: form.availabilityAlertEnabled,
        failureWindowMinutes: Number(form.failureWindowMinutes),
        failureThreshold: Number(form.failureThreshold),
        notificationGroupUids: form.notificationGroupUids,
      }
      if (editingItem) {
        await updateDatasource({ uid: editingItem.uid, ...payload })
      } else {
        await createDatasource(payload)
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

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm.action === "delete") {
        await deleteDatasource(confirm.uid)
        // 删除后列表会少一行，顺手关掉指向该数据源的抽屉，避免继续操作已删除资源。
        setDetailItem((item) => (item?.uid === confirm.uid ? null : item))
        setMetadataItem((item) => (item?.uid === confirm.uid ? null : item))
        setQueryItem((item) => (item?.uid === confirm.uid ? null : item))
      } else {
        await updateDatasourceStatus(confirm.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setConfirmError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function enableItem(item: DatasourceItem) {
    try {
      await updateDatasourceStatus(item.uid, "ACTIVE")
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      if (failure.form) {
        fail(failure.form)
      }
    }
  }

  // 更多菜单：详情之外的入口全部收进菜单；有危险操作时 RowActions 会自动加分隔线。
  function rowActions(item: DatasourceItem): RowAction[] {
    const actions: RowAction[] = [
      {
        key: "metadata",
        label: t("datasource.tabMetadata"),
        onSelect: () => setMetadataItem(item),
      },
      {
        key: "query",
        label: t("datasource.tabQuery"),
        onSelect: () => setQueryItem(item),
      },
    ]
    if (!canWrite) {
      return actions
    }
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
            destructive: true,
            onSelect: () => askConfirm("disable", item.uid),
          }
        : {
            key: "enable",
            label: t("action.enable"),
            onSelect: () => void enableItem(item),
          }
    )
    actions.push({
      key: "delete",
      label: t("action.delete"),
      destructive: true,
      onSelect: () => askConfirm("delete", item.uid),
    })
    return actions
  }

  const chips: FilterChip[] = []
  if (query.name) {
    chips.push({
      key: "name",
      label: `${t("datasource.name")}: ${query.name}`,
      onRemove: () => query.update({ name: "" }),
    })
  }
  if (type) {
    chips.push({
      key: "type",
      label: `${t("datasource.type")}: ${enumLabel(t, type)}`,
      onRemove: () => query.update({ type: "" }),
    })
  }

  function applyFilters() {
    query.update({ name: draft.name.trim(), type: draft.type })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ name: "", type: "" })
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("datasource.title")}
        description={t("datasource.description")}
      >
        {canWrite && allowedDrivers.length > 0 ? (
          <Button onClick={openCreate}>{t("action.create")}</Button>
        ) : null}
      </PageHeader>
      <ListFilter
        id="datasource-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={chips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={chips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="min-w-0 w-full max-w-[500px] flex-1 basis-56">
          <FieldLabel
            htmlFor="datasource-filter-name"
            className="text-xs text-muted-foreground"
          >
            {t("datasource.name")}
          </FieldLabel>
          <div className="relative">
            <Search
              aria-hidden="true"
              className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
            />
            <Input
              id="datasource-filter-name"
              className="pl-9"
              value={draft.name}
              placeholder={t("datasource.namePlaceholder")}
              onChange={(event) => update("name", event.target.value)}
            />
          </div>
        </Field>
        <Field className="w-40 shrink-0">
          <FieldLabel
            htmlFor="datasource-filter-type"
            className="text-xs text-muted-foreground"
          >
            {t("datasource.type")}
          </FieldLabel>
          <Select
            value={draft.type || "ALL"}
            onValueChange={(value) =>
              update("type", value === "ALL" ? "" : value)
            }
          >
            <SelectTrigger id="datasource-filter-type" className="w-full">
              <SelectValue placeholder={t("datasource.typePlaceholder")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.type} />
                  <span>
                    {draft.type
                      ? enumLabel(t, draft.type)
                      : t("datasource.typeAll")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="ALL">{t("datasource.typeAll")}</SelectItem>
              {allowedTypes.map((item) => (
                <SelectItem key={item} value={item}>
                  <span className="flex items-center gap-2">
                    <EnumIcon value={item} />
                    <span>{enumLabel(t, item)}</span>
                  </span>
                </SelectItem>
              ))}
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
              {loadError || t("state.failed")}
            </p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {t("datasource.empty")}
          </p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("datasource.name")}</TableHead>
                <TableTypeHead>{t("datasource.type")}</TableTypeHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("datasource.driver")}
                </TableHead>
                <TableHead>{t("datasource.endpoint")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="min-w-0 whitespace-normal">
                    <span className="flex min-w-0 items-center gap-2">
                      <BrandIcon
                        driver={item.driver}
                        className="size-5 shrink-0 rounded-md border bg-background p-0.5"
                      />
                      <span className="min-w-0 break-words font-medium">
                        {item.name}
                      </span>
                    </span>
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge value={item.type} />
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge value={item.driver} />
                  </TableCell>
                  <TableCell className="max-w-[20rem] text-sm break-all whitespace-normal text-muted-foreground">
                    {item.endpoint}
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetailItem(item)}
                      actions={rowActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <Dialog
        open={open}
        onOpenChange={(value) => {
          setOpen(value)
          if (!value) {
            setError("")
            setEditingItem(null)
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingItem ? t("action.edit") : t("action.create")}
            </DialogTitle>
          </DialogHeader>
          <StepForm
            step={formStep}
            stepCount={3}
            pending={pending}
            error={error}
            fieldsClassName="sm:grid-cols-2"
            onBack={() => setFormStep((value) => Math.max(0, value - 1))}
            onNext={() => {
              const errors = validateDatasourceStep(formStep)
              if (hasFieldErrors(errors)) {
                setFieldErrors(errors)
                setError("")
                return
              }
              setError("")
              setFieldErrors({})
              setFormStep((value) => Math.min(2, value + 1))
            }}
            onSubmit={onSave}
          >
            {formStep === 0 ? (
              <>
                <FormField
                  label={t("datasource.name")}
                  htmlFor="ds-name"
                  className="sm:col-span-full"
                  required
                  error={fieldErrors.name}
                >
                  <Input
                    id="ds-name"
                    value={form.name}
                    maxLength={64}
                    placeholder={t("datasource.namePlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, name: event.target.value })
                    }
                    required
                  />
                </FormField>
                <FormField
                  label={t("datasource.type")}
                  hint={t("datasource.typeHint")}
                  required
                  error={fieldErrors.type}
                >
                  <Select
                    value={form.type}
                    onValueChange={(value) =>
                      setForm({
                        ...form,
                        type: value,
                        driver:
                          (drivers[value] ?? []).find((item) =>
                            allowedDrivers.includes(item)
                          ) ?? "",
                      })
                    }
                  >
                    <SelectTrigger aria-label={t("datasource.type")}>
                      <SelectValue
                        placeholder={t("datasource.typePlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {allowedTypes.map((item) => (
                        <SelectItem key={item} value={item}>
                          {enumLabel(t, item)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField
                  label={t("datasource.driver")}
                  required
                  error={fieldErrors.driver}
                >
                  <Select
                    value={form.driver}
                    onValueChange={(value) =>
                      setForm({ ...form, driver: value })
                    }
                  >
                    <SelectTrigger aria-label={t("datasource.driver")}>
                      <SelectValue
                        placeholder={t("datasource.driverPlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {(drivers[form.type] ?? [])
                        .filter((item) => allowedDrivers.includes(item))
                        .map((item) => (
                        <SelectItem key={item} value={item}>
                          <span className="flex items-center gap-2">
                            <BrandIcon driver={item} className="size-4" />
                            <span>{enumLabel(t, item)}</span>
                          </span>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField
                  label={t("datasource.endpoint")}
                  htmlFor="ds-endpoint"
                  hint={t("datasource.endpointHint")}
                  className="sm:col-span-full"
                  required
                  error={fieldErrors.endpoint}
                >
                  <Input
                    id="ds-endpoint"
                    value={form.endpoint}
                    placeholder="http://127.0.0.1:9090"
                    onChange={(event) =>
                      setForm({ ...form, endpoint: event.target.value })
                    }
                    required
                  />
                </FormField>
              </>
            ) : null}
            {formStep === 1 ? (
              <>
                {form.driver === "LOGS_ELASTICSEARCH" ? (
                  <>
                    <FormField
                      label={t("datasource.indexPattern")}
                      htmlFor="ds-index-pattern"
                      hint={t("datasource.indexPatternHint")}
                    >
                      <Input
                        id="ds-index-pattern"
                        value={form.indexPattern}
                        placeholder={t("datasource.indexPatternPlaceholder")}
                        onChange={(event) =>
                          setForm({ ...form, indexPattern: event.target.value })
                        }
                      />
                    </FormField>
                    <FormField
                      label={t("datasource.timeField")}
                      htmlFor="ds-time-field"
                      hint={t("datasource.timeFieldHint")}
                    >
                      <Input
                        id="ds-time-field"
                        value={form.timeField}
                        placeholder={t("datasource.timeFieldPlaceholder")}
                        onChange={(event) =>
                          setForm({ ...form, timeField: event.target.value })
                        }
                      />
                    </FormField>
                  </>
                ) : null}
                <FormField label={t("datasource.username")} htmlFor="ds-user">
                  <Input
                    id="ds-user"
                    value={form.username}
                    placeholder={t("datasource.usernamePlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, username: event.target.value })
                    }
                  />
                </FormField>
                <FormField
                  label={t("datasource.password")}
                  htmlFor="ds-password"
                  hint={t("datasource.secretHint")}
                >
                  <Input
                    id="ds-password"
                    type="password"
                    value={form.password}
                    placeholder={t("datasource.passwordPlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, password: event.target.value })
                    }
                  />
                </FormField>
                <FormField
                  label={t("datasource.secret")}
                  htmlFor="ds-secret"
                  hint={t("datasource.secretHint")}
                  className="sm:col-span-full"
                >
                  <Input
                    id="ds-secret"
                    type="password"
                    value={form.secret}
                    placeholder={t("datasource.secretPlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, secret: event.target.value })
                    }
                  />
                </FormField>
              </>
            ) : null}
            {formStep === 2 ? (
              <>
                <FormField label={t("datasource.level")}>
                  <Select
                    value={form.levelUid || "NONE"}
                    onValueChange={(value) =>
                      setForm({
                        ...form,
                        levelUid: value === "NONE" ? "" : value,
                      })
                    }
                  >
                    <SelectTrigger aria-label={t("datasource.level")}>
                      <SelectValue
                        placeholder={t("datasource.levelPlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="NONE">
                        {t("datasource.levelNone")}
                      </SelectItem>
                      {levels.map((item) => (
                        <SelectItem key={item.uid} value={item.uid}>
                          {item.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <div className="grid gap-2 sm:col-span-full">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-medium">
                      {t("datasource.labels")}
                    </span>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setForm({
                          ...form,
                          labels: [...form.labels, { key: "", value: "" }],
                        })
                      }
                    >
                      {t("datasource.addLabel")}
                    </Button>
                  </div>
                  {form.labels.map((pair, index) => (
                    <div key={index} className="grid gap-2 sm:grid-cols-2">
                      <Input
                        value={pair.key}
                        placeholder={t("datasource.labelKeyPlaceholder")}
                        onChange={(event) => {
                          const labels = [...form.labels]
                          labels[index] = { ...pair, key: event.target.value }
                          setForm({ ...form, labels })
                        }}
                      />
                      <Input
                        value={pair.value}
                        placeholder={t("datasource.labelValuePlaceholder")}
                        onChange={(event) => {
                          const labels = [...form.labels]
                          labels[index] = { ...pair, value: event.target.value }
                          setForm({ ...form, labels })
                        }}
                      />
                    </div>
                  ))}
                </div>
                <FormField
                  label={t("datasource.remark")}
                  htmlFor="ds-remark"
                  className="sm:col-span-full"
                >
                  <Input
                    id="ds-remark"
                    value={form.remark}
                    placeholder={t("datasource.remarkPlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, remark: event.target.value })
                    }
                  />
                </FormField>
                <section className="grid gap-4 rounded-lg border bg-muted/10 p-4 sm:col-span-full">
                  <div className="grid gap-1">
                    <h3 className="text-sm font-medium">
                      {t("datasource.availabilityAlertTitle")}
                    </h3>
                    <p className="text-xs text-muted-foreground">
                      {t("datasource.availabilityAlertDescription")}
                    </p>
                  </div>
                  <label className="flex items-start gap-2">
                    <Checkbox
                      id="ds-availability-enabled"
                      className="mt-0.5"
                      checked={form.availabilityAlertEnabled}
                      onCheckedChange={(value) =>
                        setForm({
                          ...form,
                          availabilityAlertEnabled: value === true,
                        })
                      }
                    />
                    <span className="grid gap-1">
                      <span className="text-sm font-medium">
                        {t("datasource.availabilityAlertEnabled")}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {t("datasource.availabilityAlertEnabledHint")}
                      </span>
                    </span>
                  </label>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <FormField
                      label={t("datasource.failureWindowMinutes")}
                      htmlFor="ds-availability-window"
                      hint={t("datasource.failureWindowMinutesHint")}
                      required={form.availabilityAlertEnabled}
                      error={fieldErrors.failureWindowMinutes}
                    >
                      <Input
                        id="ds-availability-window"
                        type="number"
                        min={form.availabilityAlertEnabled ? 1 : 0}
                        max={MAX_AVAILABILITY_WINDOW_MINUTES}
                        step={1}
                        inputMode="numeric"
                        value={form.failureWindowMinutes}
                        onChange={(event) =>
                          setForm({
                            ...form,
                            failureWindowMinutes: event.target.value,
                          })
                        }
                        required={form.availabilityAlertEnabled}
                      />
                    </FormField>
                    <FormField
                      label={t("datasource.failureThreshold")}
                      htmlFor="ds-availability-threshold"
                      hint={t("datasource.failureThresholdHint")}
                      required={form.availabilityAlertEnabled}
                      error={fieldErrors.failureThreshold}
                    >
                      <Input
                        id="ds-availability-threshold"
                        type="number"
                        min={form.availabilityAlertEnabled ? 1 : 0}
                        max={MAX_AVAILABILITY_THRESHOLD}
                        step={1}
                        inputMode="numeric"
                        value={form.failureThreshold}
                        onChange={(event) =>
                          setForm({
                            ...form,
                            failureThreshold: event.target.value,
                          })
                        }
                        required={form.availabilityAlertEnabled}
                      />
                    </FormField>
                  </div>
                  <FormField
                    label={t("datasource.notificationGroups")}
                    htmlFor="ds-availability-groups"
                    hint={t("datasource.notificationGroupsHint")}
                    error={fieldErrors.notificationGroupUids}
                  >
                    <MultiSelect
                      id="ds-availability-groups"
                      invalid={Boolean(fieldErrors.notificationGroupUids)}
                      values={form.notificationGroupUids}
                      options={notificationGroupOptions}
                      onChange={(values) =>
                        setForm({ ...form, notificationGroupUids: values })
                      }
                      placeholder={t("multiSelect.placeholder")}
                      searchPlaceholder={t("multiSelect.searchPlaceholder")}
                      emptyText={t("datasource.notificationGroupsEmpty")}
                      allLabel={t("multiSelect.all")}
                      clearLabel={t("multiSelect.clear")}
                    />
                  </FormField>
                  {form.availabilityAlertEnabled &&
                  form.notificationGroupUids.length === 0 ? (
                    <p className="rounded-md border border-warning-border bg-warning-soft px-2.5 py-1.5 text-xs text-warning">
                      {t("datasource.noNotificationGroupsHint")}
                    </p>
                  ) : null}
                </section>
              </>
            ) : null}
          </StepForm>
        </DialogContent>
      </Dialog>
      <DatasourceDetailSheet
        item={detailItem}
        onOpenChange={(value) => {
          if (!value) {
            setDetailItem(null)
          }
        }}
      />
      <DetailSheet
        open={metadataItem !== null}
        title={metadataItem?.name ?? ""}
        description={
          metadataItem ? t("datasource.tabMetadata") : undefined
        }
        size="80vw"
        onOpenChange={(value) => {
          if (!value) {
            setMetadataItem(null)
          }
        }}
      >
        {metadataItem ? <MetadataPanel datasource={metadataItem} /> : null}
      </DetailSheet>
      <DetailSheet
        open={queryItem !== null}
        title={queryItem?.name ?? ""}
        description={queryItem ? t("datasource.tabQuery") : undefined}
        size="80vw"
        contentMode="fill"
        onOpenChange={(value) => {
          if (!value) {
            setQueryItem(null)
          }
        }}
      >
        {queryItem ? (
          <div className="flex min-h-0 min-w-0 flex-1 flex-col px-4 py-4 sm:px-5">
            <QueryPanel key={queryItem.uid} datasource={queryItem} />
          </div>
        ) : null}
      </DetailSheet>
      <ConfirmAction
        open={confirm !== null}
        title={
          shownAction === "delete"
            ? t("confirm.deleteTitle")
            : t("confirm.disableTitle")
        }
        description={
          shownAction === "delete"
            ? t("confirm.deleteDescription")
            : t("confirm.disableDescription")
        }
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
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
