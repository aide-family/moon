import { EllipsisVertical } from "lucide-react"
import { useEffect, useState } from "react"
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
import type { DatasourceItem } from "@/api/datasource/types"
import { useSession } from "@/auth/session"
import { BrandIcon } from "@/components/brand-icon"
import { ConfirmAction } from "@/components/confirm-action"
import { FormField } from "@/components/form-field"
import { LabelTags } from "@/components/label-tags"
import { PageHeader } from "@/components/page-header"
import { ListFrame } from "@/components/list-frame"
import { StatusBadge } from "@/components/status-badge"
import { StepForm } from "@/components/step-form"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
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
    }
    void load()
    return () => {
      active = false
    }
  }, [canWrite, session.namespaceUid, version])

  function setParam(key: string, value: string) {
    const next = new URLSearchParams(params)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    if (key !== "uid") {
      next.set("page", "1")
    }
    setParams(next)
  }

  function openCreate() {
    setEditing(false)
    setForm(emptyDraft)
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setOpen(true)
  }

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
    })
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setOpen(true)
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
    return {}
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = validateDatasourceStep(0)
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setFormStep(0)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const payload = {
        ...form,
        name: form.name.trim(),
        endpoint: form.endpoint.trim(),
        labels: serializeLabelPairs(form.labels),
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

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("datasource.title")}
        description={t("datasource.description")}
      >
        {canWrite ? (
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
            {types.map((item) => (
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
                        driver: drivers[value][0],
                      })
                    }
                  >
                    <SelectTrigger aria-label={t("datasource.type")}>
                      <SelectValue
                        placeholder={t("datasource.typePlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {types.map((item) => (
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
                      {(drivers[form.type] ?? []).map((item) => (
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
