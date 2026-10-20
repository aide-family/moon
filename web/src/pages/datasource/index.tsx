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
import { ConfirmAction } from "@/components/confirm-action"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { ListFrame } from "@/components/list-frame"
import { StatusDot } from "@/components/status-dot"
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
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
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
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { MetadataPanel } from "@/pages/datasource/metadata-panel"
import { QueryPanel } from "@/pages/datasource/query-panel"

const types = ["METRICS", "LOGS", "TRACE"]
const drivers: Record<string, string[]> = {
  METRICS: ["METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS"],
  LOGS: ["LOGS_ELASTICSEARCH", "LOGS_LOKI"],
  TRACE: ["TRACE_JAEGER"],
}

interface Draft {
  name: string
  type: string
  driver: string
  endpoint: string
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
  const [detailItem, setDetailItem] = useState<DatasourceItem | null>(null)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [formStep, setFormStep] = useState(0)
  const [form, setForm] = useState<Draft>(emptyDraft)
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [confirm, setConfirm] = useState<{ action: "disable" | "delete"; uid: string } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [shownAction, setShownAction] = useState<"disable" | "delete">("disable")
  const canWrite = session.can("datasource:write")
  const { items, total, loading, failed, error: loadError, version, reload, fail } = usePagedList<DatasourceItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}|${type}`,
    fetch: (page, pageSize) => listDatasources({ page, pageSize, name: query.name, type }),
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
        const result = await listAlertDictionaries({ page: 1, pageSize: 100, type: "DATASOURCE" })
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
      return collectFieldErrors([
        ["name", requireTrimmed(form.name, t("form.required"))],
        ["type", requireTrimmed(form.type, t("form.required"))],
        ["driver", requireTrimmed(form.driver, t("form.required"))],
        ["endpoint", requireTrimmed(form.endpoint, t("form.required"))],
      ])
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
      <PageHeader title={t("datasource.title")} description={t("datasource.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
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
        <Select value={type || "ALL"} onValueChange={(value) => setParam("type", value === "ALL" ? "" : value)}>
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
      <div className="grid min-h-0 flex-1 gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
        <ListFrame
          total={total}
          page={query.page}
          pageSize={query.pageSize}
          showPager={!loading && !failed && items.length > 0}
          onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
          onNext={() => query.update({ page: query.page + 1 })}
        >
            {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
            {!loading && failed ? (
              <div className="flex items-center gap-2">
                <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
                <Button variant="outline" onClick={() => reload()}>
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!loading && !failed && items.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("datasource.empty")}</p>
            ) : null}
            <div className="grid min-h-0 flex-1 content-start gap-2 overflow-y-auto">
            {items.map((item) => (
              <div key={item.uid} className="flex items-center gap-1">
                <Button
                  variant={item.uid === selected ? "secondary" : "ghost"}
                  className="h-auto min-w-0 flex-1 justify-start px-2 py-2"
                  onClick={() => setParam("uid", item.uid)}
                >
                  <span className="flex min-w-0 items-center gap-2">
                    <StatusDot active={item.status === "ACTIVE"} />
                    <span className="flex min-w-0 flex-col items-start">
                      <span className="truncate">{item.name}</span>
                      <span className="text-xs text-muted-foreground">{enumLabel(t, item.type)}</span>
                    </span>
                  </span>
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" aria-label={t("action.more")}>
                      <EllipsisVertical />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onSelect={() => setDetailItem(item)}>{t("action.detail")}</DropdownMenuItem>
                    {canWrite ? (
                      <DropdownMenuItem onSelect={() => openEdit(item)}>{t("action.edit")}</DropdownMenuItem>
                    ) : null}
                    {canWrite && item.status === "ACTIVE" ? (
                      <DropdownMenuItem variant="destructive" onSelect={() => askConfirm("disable", item.uid)}>
                        {t("action.disable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite && item.status !== "ACTIVE" ? (
                      <DropdownMenuItem onSelect={() => void enableItem(item)}>{t("action.enable")}</DropdownMenuItem>
                    ) : null}
                    {canWrite ? (
                      <DropdownMenuItem variant="destructive" onSelect={() => askConfirm("delete", item.uid)}>
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
            {!current ? <p className="text-sm text-muted-foreground">{t("datasource.pick")}</p> : null}
            {current ? (
              <Tabs defaultValue="detail" className="flex min-h-0 flex-1 flex-col">
                <TabsList className="shrink-0">
                  <TabsTrigger value="detail">{t("datasource.tabDetail")}</TabsTrigger>
                  <TabsTrigger value="metadata">{t("datasource.tabMetadata")}</TabsTrigger>
                  <TabsTrigger value="query">{t("datasource.tabQuery")}</TabsTrigger>
                </TabsList>
                <TabsContent value="detail" className="min-h-0 flex-1 overflow-y-auto">
                  <div className="grid gap-2">
                    <h2 className="flex items-center gap-2 text-lg font-semibold">
                      <StatusDot active={current.status === "ACTIVE"} />
                      {current.name}
                    </h2>
                    <p className="text-sm text-muted-foreground">
                      {enumLabel(t, current.driver)} · {current.endpoint}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {enumLabel(t, current.type)} · {enumLabel(t, current.status)}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {t("datasource.secretSet")}: {current.secretSet ? t("datasource.yes") : t("datasource.no")}
                    </p>
                  </div>
                </TabsContent>
                <TabsContent value="metadata" className="flex min-h-0 flex-1 flex-col overflow-hidden">
                  {current.type === "METRICS" ? (
                    <MetadataPanel uid={current.uid} />
                  ) : (
                    <p className="text-sm text-muted-foreground">{t("datasource.metadataMetricsOnly")}</p>
                  )}
                </TabsContent>
                <TabsContent value="query" className="flex min-h-0 flex-1 flex-col overflow-hidden">
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
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
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
              <Input id="ds-name" value={form.name} placeholder={t("datasource.namePlaceholder")} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
            </FormField>
            <FormField
              label={t("datasource.type")}
              hint={t("datasource.typeHint")}
              required
              error={fieldErrors.type}
            >
              <Select
                value={form.type}
                onValueChange={(value) => setForm({ ...form, type: value, driver: drivers[value][0] })}
              >
                <SelectTrigger aria-label={t("datasource.type")}>
                  <SelectValue placeholder={t("datasource.typePlaceholder")} />
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
              <Select value={form.driver} onValueChange={(value) => setForm({ ...form, driver: value })}>
                <SelectTrigger aria-label={t("datasource.driver")}>
                  <SelectValue placeholder={t("datasource.driverPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {(drivers[form.type] ?? []).map((item) => (
                    <SelectItem key={item} value={item}>
                      {enumLabel(t, item)}
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
              <Input id="ds-endpoint" value={form.endpoint} placeholder="http://127.0.0.1:9090" onChange={(event) => setForm({ ...form, endpoint: event.target.value })} required />
            </FormField>
              </>
            ) : null}
            {formStep === 1 ? (
              <>
            <FormField label={t("datasource.username")} htmlFor="ds-user">
              <Input id="ds-user" value={form.username} placeholder={t("datasource.usernamePlaceholder")} onChange={(event) => setForm({ ...form, username: event.target.value })} />
            </FormField>
            <FormField label={t("datasource.password")} htmlFor="ds-password" hint={t("datasource.secretHint")}>
              <Input id="ds-password" type="password" value={form.password} placeholder={t("datasource.passwordPlaceholder")} onChange={(event) => setForm({ ...form, password: event.target.value })} />
            </FormField>
            <FormField label={t("datasource.secret")} htmlFor="ds-secret" hint={t("datasource.secretHint")} className="sm:col-span-full">
              <Input id="ds-secret" type="password" value={form.secret} placeholder={t("datasource.secretPlaceholder")} onChange={(event) => setForm({ ...form, secret: event.target.value })} />
            </FormField>
              </>
            ) : null}
            {formStep === 2 ? (
              <>
            <FormField label={t("datasource.level")}>
              <Select value={form.levelUid || "NONE"} onValueChange={(value) => setForm({ ...form, levelUid: value === "NONE" ? "" : value })}>
                <SelectTrigger aria-label={t("datasource.level")}>
                  <SelectValue placeholder={t("datasource.levelPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="NONE">{t("datasource.levelNone")}</SelectItem>
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
                <span className="text-sm font-medium">{t("datasource.labels")}</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setForm({ ...form, labels: [...form.labels, { key: "", value: "" }] })}
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
            <FormField label={t("datasource.remark")} htmlFor="ds-remark" className="sm:col-span-full">
              <Input id="ds-remark" value={form.remark} placeholder={t("datasource.remarkPlaceholder")} onChange={(event) => setForm({ ...form, remark: event.target.value })} />
            </FormField>
              </>
            ) : null}
          </StepForm>
        </DialogContent>
      </Dialog>
      <ConfirmAction
        open={confirm !== null}
        title={shownAction === "delete" ? t("confirm.deleteTitle") : t("confirm.disableTitle")}
        description={shownAction === "delete" ? t("confirm.deleteDescription") : t("confirm.disableDescription")}
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
      <Sheet open={detailItem !== null} onOpenChange={(value) => { if (!value) setDetailItem(null) }}>
        <SheetContent>
          <SheetHeader>
            <SheetTitle>{detailItem?.name}</SheetTitle>
            <SheetDescription>{detailItem ? enumLabel(t, detailItem.status) : ""}</SheetDescription>
          </SheetHeader>
          {detailItem ? (
            <dl className="grid gap-3 px-4 text-sm">
              <div>
                <dt className="text-muted-foreground">{t("datasource.type")}</dt>
                <dd>{enumLabel(t, detailItem.type)}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">{t("datasource.driver")}</dt>
                <dd>{enumLabel(t, detailItem.driver)}</dd>
              </div>
              <div>
                <dt className="text-muted-foreground">{t("datasource.endpoint")}</dt>
                <dd className="break-all">{detailItem.endpoint}</dd>
              </div>
            </dl>
          ) : null}
        </SheetContent>
      </Sheet>
    </section>
  )
}
