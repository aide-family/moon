import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAlertLevels } from "@/api/alert-level"
import type { AlertLevelItem } from "@/api/alert-level/types"
import {
  createDatasource,
  deleteDatasource,
  getDatasource,
  getDatasourceStatus,
  listDatasources,
  updateDatasource,
  updateDatasourceStatus,
} from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { proxyQuery, queryMetric, queryMetricRange } from "@/api/metric-query"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { ListFrame } from "@/components/list-frame"
import { StepForm } from "@/components/step-form"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useListQuery } from "@/lib/list-query"

const types = ["METRICS", "LOGS", "TRACE"]
const drivers: Record<string, string[]> = {
  METRICS: ["METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS"],
  LOGS: ["LOGS_ELASTICSEARCH"],
  TRACE: ["TRACE_JAEGER"],
}

const areaClass =
  "min-h-24 w-full rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"

interface Draft {
  name: string
  type: string
  driver: string
  endpoint: string
  username: string
  password: string
  secret: string
  levelUid: string
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
  const [items, setItems] = useState<DatasourceItem[]>([])
  const [total, setTotal] = useState(0)
  const [current, setCurrent] = useState<DatasourceItem | null>(null)
  const [levels, setLevels] = useState<AlertLevelItem[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState(false)
  const [formStep, setFormStep] = useState(0)
  const [form, setForm] = useState<Draft>(emptyDraft)
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [expr, setExpr] = useState("")
  const [start, setStart] = useState("")
  const [end, setEnd] = useState("")
  const [step, setStep] = useState("60")
  const [result, setResult] = useState("")
  const [queryError, setQueryError] = useState("")
  const [proxyMethod, setProxyMethod] = useState("GET")
  const [proxyPath, setProxyPath] = useState("/")
  const [proxyBody, setProxyBody] = useState("")
  const [confirm, setConfirm] = useState<"disable" | "delete" | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canWrite = session.can("datasource:write")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listDatasources({
          page: query.page,
          pageSize: query.pageSize,
          name: query.name,
          type,
        })
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
      } catch {
        if (active) {
          setFailed(true)
          setItems([])
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [query.page, query.pageSize, query.name, type, attempt, session.namespaceUid])

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
  }, [selected, attempt, session.namespaceUid])

  useEffect(() => {
    let active = true
    async function load() {
      if (!canWrite) {
        return
      }
      try {
        const result = await listAlertLevels({ page: 1, pageSize: 100, type: "DATASOURCE" })
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
  }, [canWrite, session.namespaceUid, attempt])

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
    setFormStep(0)
    setOpen(true)
  }

  function openEdit() {
    if (!current) {
      return
    }
    setEditing(true)
    setForm({
      name: current.name,
      type: current.type,
      driver: current.driver,
      endpoint: current.endpoint,
      username: current.username,
      password: "",
      secret: "",
      levelUid: current.levelUid,
      remark: current.remark,
    })
    setError("")
    setFormStep(0)
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      if (editing && current) {
        await updateDatasource({ uid: current.uid, ...form })
      } else {
        const created = await createDatasource(form)
        setParam("uid", created.datasource.uid)
      }
      setOpen(false)
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function runQuery(kind: "instant" | "range" | "status" | "proxy") {
    if (!current) {
      return
    }
    setPending(true)
    setQueryError("")
    try {
      if (kind === "status") {
        const response = await getDatasourceStatus(current.uid)
        setResult(response.status)
      } else if (kind === "instant") {
        const response = await queryMetric({ datasourceUid: current.uid, expr, time: 0 })
        setResult(response.body)
      } else if (kind === "range") {
        const response = await queryMetricRange({
          datasourceUid: current.uid,
          expr,
          start: Number(start),
          end: Number(end),
          step: Number(step),
        })
        setResult(response.body)
      } else {
        const response = await proxyQuery({
          datasourceUid: current.uid,
          method: proxyMethod,
          path: proxyPath,
          body: proxyBody,
        })
        setResult(response.body)
      }
    } catch (err) {
      setQueryError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function runConfirm() {
    if (!current || !confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm === "delete") {
        await deleteDatasource(current.uid)
        setParam("uid", "")
      } else {
        await updateDatasourceStatus(current.uid, "DISABLED")
      }
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  const actions: RowAction[] = current && canWrite
    ? [
        { key: "edit", label: t("action.edit"), onSelect: openEdit },
        current.status === "ACTIVE"
          ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => setConfirm("disable") }
          : {
              key: "enable",
              label: t("action.enable"),
              onSelect: () => {
                void updateDatasourceStatus(current.uid, "ACTIVE").then(() => setAttempt((value) => value + 1))
              },
            },
        { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm("delete") },
      ]
    : []

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
                {item}
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
                <p className="text-sm text-destructive">{t("state.failed")}</p>
                <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!loading && !failed && items.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("datasource.empty")}</p>
            ) : null}
            <div className="grid min-h-0 flex-1 content-start gap-2 overflow-y-auto">
            {items.map((item) => (
              <Button
                key={item.uid}
                variant={item.uid === selected ? "secondary" : "ghost"}
                className="h-auto justify-start px-2 py-2"
                onClick={() => setParam("uid", item.uid)}
              >
                <span className="flex flex-col items-start">
                  <span>{item.name}</span>
                  <span className="text-xs text-muted-foreground">
                    {item.type} · {item.status}
                  </span>
                </span>
              </Button>
            ))}
            </div>
        </ListFrame>
        <Card className="min-h-0 flex-1">
          <CardContent className="grid min-h-0 flex-1 content-start gap-6 overflow-y-auto">
            {!current ? <p className="text-sm text-muted-foreground">{t("datasource.pick")}</p> : null}
            {current ? (
              <>
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <h2 className="text-lg font-semibold">{current.name}</h2>
                    <p className="text-sm text-muted-foreground">
                      {current.driver} · {current.endpoint}
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {t("datasource.secretSet")}: {current.secretSet ? t("datasource.yes") : t("datasource.no")}
                    </p>
                  </div>
                  {actions.length > 0 ? (
                    <RowActions
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setResult(`${current.name}\n${current.endpoint}`)}
                      actions={actions}
                    />
                  ) : null}
                </div>
                {current.type === "METRICS" ? (
                  <div className="grid gap-4">
                    <FormField label={t("datasource.expr")} htmlFor="metric-expr" hint={t("datasource.exprHint")}>
                      <textarea
                        id="metric-expr"
                        className={areaClass}
                        value={expr}
                        placeholder={t("datasource.exprPlaceholder")}
                        onChange={(event) => setExpr(event.target.value)}
                      />
                    </FormField>
                    <div className="flex flex-wrap gap-2">
                      <Button type="button" disabled={pending || expr === ""} onClick={() => void runQuery("instant")}>
                        {t("datasource.instant")}
                      </Button>
                      <Button type="button" variant="outline" disabled={pending} onClick={() => void runQuery("status")}>
                        {t("datasource.check")}
                      </Button>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      <Input className="max-w-40" value={start} placeholder={t("datasource.start")} onChange={(event) => setStart(event.target.value)} />
                      <Input className="max-w-40" value={end} placeholder={t("datasource.end")} onChange={(event) => setEnd(event.target.value)} />
                      <Input className="max-w-28" value={step} placeholder={t("datasource.step")} onChange={(event) => setStep(event.target.value)} />
                      <Button type="button" variant="outline" disabled={pending || expr === ""} onClick={() => void runQuery("range")}>
                        {t("datasource.range")}
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="grid gap-4">
                    <div className="flex flex-wrap gap-2">
                      <Select value={proxyMethod} onValueChange={setProxyMethod}>
                        <SelectTrigger className="w-28" aria-label={t("datasource.method")}>
                          <SelectValue placeholder={t("datasource.method")} />
                        </SelectTrigger>
                        <SelectContent>
                          {["GET", "POST", "PUT"].map((item) => (
                            <SelectItem key={item} value={item}>
                              {item}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <Input className="max-w-sm" value={proxyPath} placeholder={t("datasource.pathPlaceholder")} onChange={(event) => setProxyPath(event.target.value)} />
                      <Button type="button" disabled={pending} onClick={() => void runQuery("proxy")}>
                        {t("datasource.proxy")}
                      </Button>
                    </div>
                    <textarea
                      className={areaClass}
                      value={proxyBody}
                      placeholder={t("datasource.bodyPlaceholder")}
                      onChange={(event) => setProxyBody(event.target.value)}
                    />
                  </div>
                )}
                {queryError ? <p className="text-sm text-destructive">{queryError}</p> : null}
                {result ? <pre className="overflow-auto rounded-lg border p-3 text-sm">{result}</pre> : null}
              </>
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
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <StepForm
            step={formStep}
            stepCount={3}
            pending={pending}
            error={error}
            onBack={() => setFormStep((value) => Math.max(0, value - 1))}
            onNext={() => setFormStep((value) => Math.min(2, value + 1))}
            onSubmit={onSave}
          >
            {formStep === 0 ? (
              <>
            <FormField label={t("datasource.name")} htmlFor="ds-name">
              <Input id="ds-name" value={form.name} placeholder={t("datasource.namePlaceholder")} onChange={(event) => setForm({ ...form, name: event.target.value })} />
            </FormField>
            <FormField label={t("datasource.type")} hint={t("datasource.typeHint")}>
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
                      {item}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <FormField label={t("datasource.driver")}>
              <Select value={form.driver} onValueChange={(value) => setForm({ ...form, driver: value })}>
                <SelectTrigger aria-label={t("datasource.driver")}>
                  <SelectValue placeholder={t("datasource.driverPlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {(drivers[form.type] ?? []).map((item) => (
                    <SelectItem key={item} value={item}>
                      {item}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            <FormField label={t("datasource.endpoint")} htmlFor="ds-endpoint" hint={t("datasource.endpointHint")}>
              <Input id="ds-endpoint" value={form.endpoint} placeholder="http://127.0.0.1:9090" onChange={(event) => setForm({ ...form, endpoint: event.target.value })} />
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
            <FormField label={t("datasource.secret")} htmlFor="ds-secret" hint={t("datasource.secretHint")}>
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
            <FormField label={t("datasource.remark")} htmlFor="ds-remark">
              <Input id="ds-remark" value={form.remark} placeholder={t("datasource.remarkPlaceholder")} onChange={(event) => setForm({ ...form, remark: event.target.value })} />
            </FormField>
              </>
            ) : null}
          </StepForm>
        </DialogContent>
      </Dialog>
      <ConfirmAction
        open={confirm !== null}
        title={shown.current === "delete" ? t("confirm.deleteTitle") : t("confirm.disableTitle")}
        description={shown.current === "delete" ? t("confirm.deleteDescription") : t("confirm.disableDescription")}
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
