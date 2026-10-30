import { EllipsisVertical } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import {
  createDatasource,
  deleteDatasource,
  getDatasource,
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
import { FormField } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
import { MultiSelect } from "@/components/multi-select"
import { PageHeader } from "@/components/page-header"
import { ListFrame } from "@/components/list-frame"
import { StatusBadge } from "@/components/status-badge"
import { StepForm } from "@/components/step-form"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
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
  const [params, setParams] = useSearchParams()
  const type = params.get("type") || ""
  const selected = params.get("uid") || ""
  const [draftName, setDraftName] = useState(query.name)
  const [current, setCurrent] = useState<DatasourceItem | null>(null)
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [notificationGroups, setNotificationGroups] = useState<
    NotificationGroupSelectItem[]
  >([])
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState(false)
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

  useEffect(() => {
    let active = true
    async function load() {
      if (!selected) {
        setCurrent(null)
        return
      }
      try {
        const result = await getDatasource(selected)
        if (active) {
          setCurrent(result.datasource)
        }
      } catch {
        if (active) {
          setCurrent(null)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [selected, version, session.namespaceUid])

  // 进入数据源菜单时如果 URL 没有 uid，默认选中列表里的第一个数据源，
  // 让右侧详情面板直接可用，而不是停在“从左侧选择一个数据源”的空态。
  // 已有 uid（深链、上一页选择）优先，创建流程保存后写入的 uid 也不会被覆盖；
  // 列表为空（含筛选后无结果）、加载中或加载失败时保持原来的占位提示。
  // 条件是幂等的：选中一旦写入 URL，selected 非空，本 effect 不再动作。
  useEffect(() => {
    if (selected || loading || failed || items.length === 0) {
      return
    }
    // 自动补选不写历史记录：否则用户在无 uid 的地址上点后退，本 effect 会
    // 立刻把 uid 写回去，后退看起来永远停在同一页。
    setParam("uid", items[0].uid, { replace: true })
    // setParam 每次渲染都是新函数；只依赖“是否已有选中项 + 列表状态”，
    // 写入 uid 后 selected 非空，本 effect 自然停下，不会重复触发。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selected, loading, failed, items])

  useEffect(() => {
    let active = true
    async function load() {
      if (!canWrite) {
        return
      }
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
      // 可用性告警只能绑定启用中的通知组；编辑时停用组的名称由详情响应补上。
      try {
        const groups = await loadAllSelectItems(
          (lastUid) =>
            selectNotificationGroups({ limit: 100, lastUid, activeOnly: true }),
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
    void load()
    return () => {
      active = false
    }
  }, [canWrite, session.namespaceUid, version])

  function setParam(
    key: string,
    value: string,
    options: { replace?: boolean } = {}
  ) {
    const next = new URLSearchParams(params)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    if (key !== "uid") {
      next.set("page", "1")
    }
    setParams(next, { replace: options.replace })
  }

  function openCreate() {
    setEditing(false)
    setForm(draftForAllowedTypes())
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setOpen(true)
  }

  // 编辑时把当前数据源已绑定的通知组补进选项：select 接口带 activeOnly，
  // 停用的通知组不会返回，不补的话下拉框只能显示 uid。
  // 名称取自详情响应的 notificationGroupNames（只用于展示）；只有名称与
  // uid 数量一致时才按下标补齐，避免把名称错配到别的通知组上。
  const notificationGroupOptions = useMemo(() => {
    const options = notificationGroups.map((item) => ({
      value: item.uid,
      label: item.name,
    }))
    const known = new Set(options.map((option) => option.value))
    // 新建时不要把上一次编辑的数据源绑定补进选项。
    const uids = editing ? current?.notificationGroupUids ?? [] : []
    const names = editing ? current?.notificationGroupNames ?? [] : []
    if (uids.length === names.length) {
      uids.forEach((uid, index) => {
        const name = names[index]
        if (name && !known.has(uid)) {
          options.push({ value: uid, label: name })
        }
      })
    }
    return options
  }, [notificationGroups, current, editing])

  function openEdit(item: DatasourceItem) {
    setParam("uid", item.uid)
    setEditing(true)
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
      if (editing && current) {
        await updateDatasource({ uid: current.uid, ...payload })
      } else {
        const created = await createDatasource(payload)
        setParam("uid", created.datasource.uid)
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
        if (selected === confirm.uid) {
          setParam("uid", "")
        }
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

  // 详情里的通知组名称来自响应，仅用于展示；空数组表示未绑定通知组。
  const availabilityGroupNames = current?.notificationGroupNames ?? []

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
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draftName })
        }}
      >
        <Input
          className="max-w-sm"
          value={draftName}
          placeholder={t("datasource.namePlaceholder")}
          onChange={(event) => setDraftName(event.target.value)}
        />
        <Select
          value={type || "ALL"}
          onValueChange={(value) =>
            setParam("type", value === "ALL" ? "" : value)
          }
        >
          <SelectTrigger className="w-40" aria-label={t("datasource.type")}>
            <SelectValue placeholder={t("datasource.typePlaceholder")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="ALL">{t("datasource.typeAll")}</SelectItem>
            {allowedTypes.map((item) => (
              <SelectItem key={item} value={item}>
                {enumLabel(t, item)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
      <div className="grid min-h-0 min-w-0 flex-1 gap-6 lg:grid-cols-[minmax(0,18rem)_minmax(0,1fr)]">
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && items.length > 0}
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
              {t("datasource.empty")}
            </p>
          ) : null}
          <div className="grid min-h-0 min-w-0 flex-1 content-start gap-2 overflow-x-hidden overflow-y-auto">
            {items.map((item) => (
              <div key={item.uid} className="flex min-w-0 items-center gap-1">
                <Button
                  variant={item.uid === selected ? "secondary" : "ghost"}
                  className="h-auto min-w-0 flex-1 justify-start overflow-hidden px-2 py-2"
                  onClick={() => setParam("uid", item.uid)}
                >
                  <span className="flex w-full min-w-0 items-center gap-2">
                    <BrandIcon
                      driver={item.driver}
                      className="size-5 shrink-0 rounded-md border bg-background p-0.5"
                    />
                    <span className="flex min-w-0 flex-1 flex-col items-start gap-1">
                      <span
                        className="w-full truncate text-left"
                        title={item.name}
                      >
                        {item.name}
                      </span>
                      <span className="flex flex-wrap gap-1">
                        <StatusBadge value={item.type} />
                        <StatusBadge value={item.status} />
                      </span>
                    </span>
                  </span>
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={t("action.more")}
                    >
                      <EllipsisVertical />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {canWrite ? (
                      <DropdownMenuItem onSelect={() => openEdit(item)}>
                        {t("action.edit")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite && item.status !== "ACTIVE" ? (
                      <DropdownMenuItem onSelect={() => void enableItem(item)}>
                        {t("action.enable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite ? <DropdownMenuSeparator /> : null}
                    {canWrite && item.status === "ACTIVE" ? (
                      <DropdownMenuItem
                        variant="destructive"
                        onSelect={() => askConfirm("disable", item.uid)}
                      >
                        {t("action.disable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite ? (
                      <DropdownMenuItem
                        variant="destructive"
                        onSelect={() => askConfirm("delete", item.uid)}
                      >
                        {t("action.delete")}
                      </DropdownMenuItem>
                    ) : null}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
          </div>
        </ListFrame>
        <Card className="flex min-h-0 flex-1 flex-col overflow-hidden">
          <CardContent className="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
            {!current ? (
              <p className="text-sm text-muted-foreground">
                {t("datasource.pick")}
              </p>
            ) : null}
            {current ? (
              <Tabs
                defaultValue="detail"
                className="flex min-h-0 flex-1 flex-col"
              >
                <TabsList className="shrink-0">
                  <TabsTrigger value="detail">
                    {t("datasource.tabDetail")}
                  </TabsTrigger>
                  <TabsTrigger value="metadata">
                    {t("datasource.tabMetadata")}
                  </TabsTrigger>
                  <TabsTrigger value="query">
                    {t("datasource.tabQuery")}
                  </TabsTrigger>
                </TabsList>
                <TabsContent
                  value="detail"
                  className="min-h-0 flex-1 overflow-y-auto"
                >
                  <div className="grid gap-4">
                    <div className="grid gap-2">
                      <h2 className="flex min-w-0 items-center gap-2 text-lg font-semibold">
                        <BrandIcon
                          driver={current.driver}
                          className="size-7 rounded-md border bg-background p-0.5"
                        />
                        <span className="min-w-0 truncate" title={current.name}>
                          {current.name}
                        </span>
                      </h2>
                      <p className="text-sm text-muted-foreground">
                        {current.endpoint}
                      </p>
                      <div className="flex flex-wrap items-center gap-1.5">
                        <StatusBadge value={current.type} />
                        <StatusBadge value={current.driver} />
                        <StatusBadge value={current.status} />
                      </div>
                    </div>
                    <dl className="grid min-w-0 gap-4 sm:grid-cols-2">
                      <div className="grid min-w-0 gap-1">
                        <dt className="text-sm text-muted-foreground">
                          {t("datasource.level")}
                        </dt>
                        <dd className="min-w-0 break-words">
                          {current.levelName || current.levelUid || "—"}
                        </dd>
                      </div>
                      <div className="grid min-w-0 gap-1">
                        <dt className="text-sm text-muted-foreground">
                          {t("datasource.secretSet")}
                        </dt>
                        <dd>
                          {current.secretSet
                            ? t("datasource.yes")
                            : t("datasource.no")}
                        </dd>
                      </div>
                      <div className="grid min-w-0 gap-1 sm:col-span-full">
                        <dt className="text-sm text-muted-foreground">
                          {t("datasource.labels")}
                        </dt>
                        <dd className="min-w-0">
                          <LabelTags labels={current.labels} />
                        </dd>
                      </div>
                      <div className="grid min-w-0 gap-1 sm:col-span-full">
                        <dt className="text-sm text-muted-foreground">
                          {t("datasource.remark")}
                        </dt>
                        <dd className="min-w-0 break-words whitespace-pre-wrap">
                          {current.remark || "—"}
                        </dd>
                      </div>
                    </dl>
                    <section className="grid min-w-0 gap-3 border-t pt-4">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <h3 className="text-sm font-medium">
                          {t("datasource.availabilityAlertTitle")}
                        </h3>
                        <span className="text-sm text-muted-foreground">
                          {current.availabilityAlertEnabled
                            ? t("datasource.availabilityAlertOn")
                            : t("datasource.availabilityAlertOff")}
                        </span>
                      </div>
                      <dl className="grid min-w-0 gap-4 sm:grid-cols-2">
                        <div className="grid min-w-0 gap-1">
                          <dt className="text-sm text-muted-foreground">
                            {t("datasource.failureWindow")}
                          </dt>
                          <dd className="tabular-nums">
                            {t("datasource.failureWindowValue", {
                              count: current.failureWindowMinutes,
                            })}
                          </dd>
                        </div>
                        <div className="grid min-w-0 gap-1">
                          <dt className="text-sm text-muted-foreground">
                            {t("datasource.failureThreshold")}
                          </dt>
                          <dd className="tabular-nums">
                            {t("datasource.failureThresholdValue", {
                              count: current.failureThreshold,
                            })}
                          </dd>
                        </div>
                        <div className="grid min-w-0 gap-1 sm:col-span-full">
                          <dt className="text-sm text-muted-foreground">
                            {t("datasource.notificationGroups")}
                          </dt>
                          <dd className="min-w-0 break-words">
                            {availabilityGroupNames.length > 0
                              ? availabilityGroupNames.join(", ")
                              : t("datasource.notificationGroupsUnbound")}
                          </dd>
                        </div>
                      </dl>
                      {current.availabilityAlertEnabled &&
                      availabilityGroupNames.length === 0 ? (
                        <p className="rounded-md border border-warning-border bg-warning-soft px-2.5 py-1.5 text-xs text-warning">
                          {t("datasource.noNotificationGroupsHint")}
                        </p>
                      ) : null}
                    </section>
                  </div>
                </TabsContent>
                <TabsContent
                  value="metadata"
                  className="flex min-h-0 flex-1 flex-col overflow-hidden"
                >
                  <MetadataPanel datasource={current} />
                </TabsContent>
                <TabsContent
                  value="query"
                  className="flex min-h-0 flex-1 flex-col overflow-hidden"
                >
                  <QueryPanel key={current.uid} datasource={current} />
                </TabsContent>
              </Tabs>
            ) : null}
          </CardContent>
        </Card>
      </div>
      <Dialog
        open={open}
        onOpenChange={(value) => {
          setOpen(value)
          if (!value) {
            setError("")
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editing ? t("action.edit") : t("action.create")}
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
