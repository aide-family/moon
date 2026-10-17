import { EllipsisVertical } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import type { AlertDictionaryItem } from "@/api/alert-dictionary/types"
import {
  createAlertRule,
  createAlertRuleGroup,
  deleteAlertRule,
  deleteAlertRuleGroup,
  listAlertRuleGroups,
  listAlertRules,
  updateAlertRule,
  updateAlertRuleGroup,
  updateAlertRuleGroupStatus,
  updateAlertRuleStatus,
} from "@/api/alert-rule"
import type { AlertRuleGroupItem, AlertRuleItem, AlertRuleLevelInput } from "@/api/alert-rule/types"
import { listDatasources } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { selectNotificationGroups } from "@/api/notification-group"
import type { SelectItem as NotifyGroupOption } from "@/api/notification-group/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { PromqlEditor } from "@/components/promql-editor"
import { RowActions, type RowAction } from "@/components/row-actions"
import { ListFrame } from "@/components/list-frame"
import { StatusDot } from "@/components/status-dot"
import { StepForm } from "@/components/step-form"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
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
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
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
import { enumLabel } from "@/lib/enum-label"
import { firstError, requireTrimmed } from "@/lib/form"
import { failureText, readFormError } from "@/lib/form-error"
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import { useListQuery } from "@/lib/list-query"

const types = ["METRICS", "LOGS", "TRACE"]
const drivers: Record<string, string[]> = {
  METRICS: ["METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS"],
  LOGS: ["LOGS_ELASTICSEARCH"],
  TRACE: ["TRACE_JAEGER"],
}
const conditions = ["EQ", "NE", "GT", "GTE", "LT", "LTE", "IN", "NOT_IN"]
const samples = ["FOR", "MAX", "MIN"]

interface RuleDraft {
  name: string
  type: string
  driver: string
  expr: string
  labels: string
  datasourceLabels: LabelPair[]
  summary: string
  description: string
  levels: AlertRuleLevelInput[]
  notificationGroupUids: string[]
}

const emptyLevel: AlertRuleLevelInput = {
  condition: "GT",
  sample: "FOR",
  threshold: "",
  forSeconds: 0,
  levelUid: "",
}

const emptyRule: RuleDraft = {
  name: "",
  type: "METRICS",
  driver: "METRICS_PROMETHEUS",
  expr: "",
  labels: "",
  datasourceLabels: [{ key: "", value: "" }],
  summary: "",
  description: "",
  levels: [{ ...emptyLevel }],
  notificationGroupUids: [],
}

function matchDatasourceCount(sources: DatasourceItem[], type: string, matchers: LabelPair[]): number {
  const matcherEntries = matchers
    .map((pair) => [pair.key.trim(), pair.value.trim()] as const)
    .filter(([key]) => key.length > 0)
  return sources.filter((item) => {
    if (item.type !== type || item.status !== "ACTIVE") {
      return false
    }
    let labels: Record<string, string> = {}
    try {
      labels = JSON.parse(item.labels || "{}") as Record<string, string>
    } catch {
      return false
    }
    return matcherEntries.every(([key, value]) => labels[key] === value)
  }).length
}

function matchedMetricsUid(sources: DatasourceItem[], type: string, matchers: LabelPair[]): string {
  if (type !== "METRICS") {
    return ""
  }
  const matcherEntries = matchers
    .map((pair) => [pair.key.trim(), pair.value.trim()] as const)
    .filter(([key]) => key.length > 0)
  return (
    sources.find((item) => {
      if (item.type !== "METRICS" || item.status !== "ACTIVE") {
        return false
      }
      let labels: Record<string, string> = {}
      try {
        labels = JSON.parse(item.labels || "{}") as Record<string, string>
      } catch {
        return false
      }
      return matcherEntries.every(([key, value]) => labels[key] === value)
    })?.uid ?? ""
  )
}

export function AlertRulePage() {
  const { t } = useTranslation()
  const session = useSession()
  const groupsQuery = useListQuery()
  const [params, setParams] = useSearchParams()
  const groupUid = params.get("groupUid") || ""
  const rulePage = Number(params.get("rulePage") || "1")
  const ruleName = params.get("ruleName") || ""
  const [groupDraft, setGroupDraft] = useState(groupsQuery.name)
  const [ruleDraft, setRuleDraft] = useState(ruleName)
  const [groups, setGroups] = useState<AlertRuleGroupItem[]>([])
  const [groupTotal, setGroupTotal] = useState(0)
  const [rules, setRules] = useState<AlertRuleItem[]>([])
  const [ruleTotal, setRuleTotal] = useState(0)
  const [levels, setLevels] = useState<AlertDictionaryItem[]>([])
  const [sources, setSources] = useState<DatasourceItem[]>([])
  const [notifyGroups, setNotifyGroups] = useState<NotifyGroupOption[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [groupDetail, setGroupDetail] = useState<AlertRuleGroupItem | null>(null)
  const [attempt, setAttempt] = useState(0)
  const [groupOpen, setGroupOpen] = useState(false)
  const [groupName, setGroupName] = useState("")
  const [editingGroup, setEditingGroup] = useState<AlertRuleGroupItem | null>(null)
  const [ruleOpen, setRuleOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRuleItem | null>(null)
  const [form, setForm] = useState<RuleDraft>(emptyRule)
  const [formStep, setFormStep] = useState(0)
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertRuleItem | null>(null)
  const [confirm, setConfirm] = useState<{ kind: "group" | "rule"; action: "disable" | "delete"; uid: string } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [shownAction, setShownAction] = useState<"disable" | "delete">("disable")
  const canWrite = session.can("alert_rule:write")

  function askConfirm(kind: "group" | "rule", action: "disable" | "delete", uid: string) {
    setShownAction(action)
    setConfirm({ kind, action, uid })
  }

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listAlertRuleGroups({
          page: groupsQuery.page,
          pageSize: groupsQuery.pageSize,
          name: groupsQuery.name,
        })
        if (!active) {
          return
        }
        setGroups(result.items ?? [])
        setGroupTotal(result.total ?? 0)
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureText(err, ""))
          setGroups([])
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
  }, [groupsQuery.page, groupsQuery.pageSize, groupsQuery.name, attempt, session.namespaceUid])

  useEffect(() => {
    let active = true
    async function load() {
      if (!groupUid) {
        setRules([])
        setRuleTotal(0)
        return
      }
      try {
        const result = await listAlertRules({
          page: rulePage,
          pageSize: groupsQuery.pageSize,
          groupUid,
          name: ruleName,
        })
        if (active) {
          setRules(result.items ?? [])
          setRuleTotal(result.total ?? 0)
        }
      } catch {
        if (active) {
          setRules([])
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [groupUid, rulePage, ruleName, groupsQuery.pageSize, attempt, session.namespaceUid])

  useEffect(() => {
    let active = true
    async function load() {
      try {
        const [levelResult, sourceResult, notifyResult] = await Promise.all([
          listAlertDictionaries({ page: 1, pageSize: 100, type: "ALERT" }),
          listDatasources({ page: 1, pageSize: 100 }),
          selectNotificationGroups({ limit: 100, activeOnly: true }),
        ])
        if (!active) {
          return
        }
        setLevels(levelResult.items ?? [])
        setSources(sourceResult.items ?? [])
        setNotifyGroups(notifyResult.items ?? [])
      } catch {
        if (active) {
          setLevels([])
          setSources([])
          setNotifyGroups([])
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [session.namespaceUid, attempt])

  function setParam(key: string, value: string) {
    const next = new URLSearchParams(params)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    setParams(next)
  }

  function openRule(item?: AlertRuleItem) {
    setEditingRule(item ?? null)
    setForm(
      item
        ? {
            name: item.name,
            type: item.type,
            driver: item.driver,
            expr: item.expr,
            labels: item.labels,
            datasourceLabels: parseLabelPairs(item.datasourceLabels),
            summary: item.summary,
            description: item.description,
            levels: item.levels.length
              ? item.levels.map((level) => ({
                  condition: level.condition,
                  sample: level.sample,
                  threshold: level.threshold,
                  forSeconds: level.forSeconds,
                  levelUid: level.levelUid,
                }))
              : [{ ...emptyLevel }],
            notificationGroupUids: item.notificationGroupUids ?? [],
          }
        : emptyRule,
    )
    setError("")
    setFormStep(0)
    setRuleOpen(true)
  }

  async function onSaveGroup(event: React.FormEvent) {
    event.preventDefault()
    const invalid = requireTrimmed(groupName, t("form.required"))
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      const trimmed = groupName.trim()
      if (editingGroup) {
        await updateAlertRuleGroup(editingGroup.uid, trimmed)
      } else {
        const created = await createAlertRuleGroup(trimmed)
        setParam("groupUid", created.group.uid)
      }
      setGroupOpen(false)
      setEditingGroup(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  function validateRuleStep(step: number): string | null {
    if (step === 0) {
      return firstError(
        requireTrimmed(form.name, t("form.required")),
        requireTrimmed(form.type, t("form.required")),
        requireTrimmed(form.driver, t("form.required")),
      )
    }
    return null
  }

  async function onSaveRule(event: React.FormEvent) {
    event.preventDefault()
    if (!groupUid) {
      return
    }
    const invalid = firstError(validateRuleStep(0), validateRuleStep(1), validateRuleStep(2))
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      const body = {
        ...form,
        name: form.name.trim(),
        groupUid,
        datasourceLabels: serializeLabelPairs(form.datasourceLabels),
      }
      if (editingRule) {
        await updateAlertRule({ ...body, uid: editingRule.uid })
      } else {
        await createAlertRule(body)
      }
      setRuleOpen(false)
      setAttempt((value) => value + 1)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
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
      if (confirm.kind === "group" && confirm.action === "delete") {
        await deleteAlertRuleGroup(confirm.uid)
        if (groupUid === confirm.uid) {
          setParam("groupUid", "")
        }
      } else if (confirm.kind === "group") {
        await updateAlertRuleGroupStatus(confirm.uid, "DISABLED")
      } else if (confirm.action === "delete") {
        await deleteAlertRule(confirm.uid)
      } else {
        await updateAlertRuleStatus(confirm.uid, "DISABLED")
      }
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setConfirmError(failure.form)
    } finally {
      setPending(false)
    }
  }

  function ruleActions(item: AlertRuleItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openRule(item) },
      item.status === "ACTIVE"
        ? { key: "disable", label: t("action.disable"), destructive: true, onSelect: () => askConfirm("rule", "disable", item.uid) }
        : { key: "enable", label: t("action.enable"), onSelect: () => void updateAlertRuleStatus(item.uid, "ACTIVE").then(() => setAttempt((value) => value + 1)) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => askConfirm("rule", "delete", item.uid) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("alertRule.title")} description={t("alertRule.description")} />
      <div className="grid min-h-0 flex-1 gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
        <div className="flex min-h-0 flex-col gap-4">
            <form
              className="flex gap-2"
              onSubmit={(event) => {
                event.preventDefault()
                groupsQuery.update({ name: groupDraft })
              }}
            >
              <Input value={groupDraft} placeholder={t("alertRule.groupPlaceholder")} onChange={(event) => setGroupDraft(event.target.value)} />
              <Button type="submit" variant="outline">
                {t("action.search")}
              </Button>
            </form>
            {canWrite ? (
              <Button
                onClick={() => {
                  setEditingGroup(null)
                  setGroupName("")
                  setError("")
                  setGroupOpen(true)
                }}
              >
                {t("alertRule.createGroup")}
              </Button>
            ) : null}
            <ListFrame
              total={groupTotal}
              page={groupsQuery.page}
              pageSize={groupsQuery.pageSize}
              showPager={!loading && !failed && groups.length > 0}
              onPrev={() => groupsQuery.update({ page: Math.max(1, groupsQuery.page - 1) })}
              onNext={() => groupsQuery.update({ page: groupsQuery.page + 1 })}
            >
            {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
            {!loading && failed ? (
              <div className="flex items-center gap-2">
                <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
                <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!loading && !failed && groups.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("alertRule.groupEmpty")}</p>
            ) : null}
            <div className="grid min-h-0 flex-1 content-start gap-2 overflow-y-auto">
            {groups.map((item) => (
              <div key={item.uid} className="flex items-center gap-1">
                <Button
                  variant={item.uid === groupUid ? "secondary" : "ghost"}
                  className="h-auto min-w-0 flex-1 justify-start px-2 py-2"
                  onClick={() => setParam("groupUid", item.uid)}
                >
                  <span className="flex min-w-0 items-center gap-2">
                    <StatusDot active={item.status === "ACTIVE"} />
                    <span className="truncate">{item.name}</span>
                  </span>
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon" aria-label={t("action.more")}>
                      <EllipsisVertical />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem onSelect={() => setGroupDetail(item)}>{t("action.detail")}</DropdownMenuItem>
                    {canWrite ? (
                      <DropdownMenuItem
                        onSelect={() => {
                          setEditingGroup(item)
                          setGroupName(item.name)
                          setError("")
                          setGroupOpen(true)
                        }}
                      >
                        {t("action.edit")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite && item.status === "ACTIVE" ? (
                      <DropdownMenuItem variant="destructive" onSelect={() => askConfirm("group", "disable", item.uid)}>
                        {t("action.disable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite && item.status !== "ACTIVE" ? (
                      <DropdownMenuItem
                        onSelect={() => {
                          void updateAlertRuleGroupStatus(item.uid, "ACTIVE").then(() => setAttempt((value) => value + 1))
                        }}
                      >
                        {t("action.enable")}
                      </DropdownMenuItem>
                    ) : null}
                    {canWrite ? (
                      <DropdownMenuItem variant="destructive" onSelect={() => askConfirm("group", "delete", item.uid)}>
                        {t("action.delete")}
                      </DropdownMenuItem>
                    ) : null}
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            ))}
            </div>
            </ListFrame>
        </div>
        <div className="flex min-h-0 flex-col gap-4">
            {!groupUid ? <p className="text-sm text-muted-foreground">{t("alertRule.pickGroup")}</p> : null}
            {groupUid ? (
              <div className="flex flex-wrap items-center justify-between gap-2">
                  <form
                    className="flex gap-2"
                    onSubmit={(event) => {
                      event.preventDefault()
                      const next = new URLSearchParams(params)
                      if (ruleDraft) {
                        next.set("ruleName", ruleDraft)
                      } else {
                        next.delete("ruleName")
                      }
                      next.set("rulePage", "1")
                      setParams(next)
                    }}
                  >
                    <Input value={ruleDraft} placeholder={t("alertRule.namePlaceholder")} onChange={(event) => setRuleDraft(event.target.value)} />
                    <Button type="submit" variant="outline">
                      {t("action.search")}
                    </Button>
                  </form>
                  {canWrite ? <Button onClick={() => openRule()}>{t("action.create")}</Button> : null}
              </div>
            ) : null}
            <ListFrame
              total={ruleTotal}
              page={rulePage}
              pageSize={groupsQuery.pageSize}
              showPager={Boolean(groupUid) && rules.length > 0}
              onPrev={() => setParam("rulePage", String(Math.max(1, rulePage - 1)))}
              onNext={() => setParam("rulePage", String(rulePage + 1))}
            >
                {groupUid && rules.length === 0 ? <p className="text-sm text-muted-foreground">{t("alertRule.empty")}</p> : null}
                {rules.length > 0 ? (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t("alertRule.name")}</TableHead>
                        <TableHead>{t("datasource.type")}</TableHead>
                        <TableHead>{t("namespace.status")}</TableHead>
                        <TableHead />
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {rules.map((item) => (
                        <TableRow key={item.uid}>
                          <TableCell>{item.name}</TableCell>
                          <TableCell>{enumLabel(t, item.type)}</TableCell>
                          <TableCell>{enumLabel(t, item.status)}</TableCell>
                          <TableCell>
                            <RowActions
                              detailLabel={t("action.detail")}
                              moreLabel={t("action.more")}
                              onDetail={() => setDetail(item)}
                              actions={ruleActions(item)}
                            />
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                ) : null}
            </ListFrame>
        </div>
      </div>
      <Dialog open={groupOpen} onOpenChange={setGroupOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingGroup ? t("action.edit") : t("alertRule.createGroup")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onSaveGroup}>
            <FormField label={t("alertRule.group")} htmlFor="group-name" error={error}>
              <Input id="group-name" value={groupName} placeholder={t("alertRule.groupPlaceholder")} onChange={(event) => setGroupName(event.target.value)} required />
            </FormField>
            <DialogFooter>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={ruleOpen} onOpenChange={setRuleOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editingRule ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <StepForm
            step={formStep}
            stepCount={3}
            pending={pending}
            error={error}
            fieldsClassName="sm:grid-cols-2"
            onBack={() => setFormStep((value) => Math.max(0, value - 1))}
            onNext={() => {
              const invalid = validateRuleStep(formStep)
              if (invalid) {
                setError(invalid)
                return
              }
              setError("")
              setFormStep((value) => Math.min(2, value + 1))
            }}
            onSubmit={onSaveRule}
          >
            {formStep === 0 ? (
              <>
            <FormField label={t("alertRule.name")} htmlFor="rule-name" className="sm:col-span-full">
              <Input id="rule-name" value={form.name} placeholder={t("alertRule.namePlaceholder")} onChange={(event) => setForm({ ...form, name: event.target.value })} required />
            </FormField>
            <FormField label={t("datasource.type")}>
              <Select value={form.type} onValueChange={(value) => setForm({ ...form, type: value, driver: drivers[value][0] })}>
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
            <FormField label={t("datasource.driver")}>
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
            <div className="grid gap-2 sm:col-span-full">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm font-medium">{t("alertRule.datasourceLabels")}</span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setForm({ ...form, datasourceLabels: [...form.datasourceLabels, { key: "", value: "" }] })}
                >
                  {t("alertRule.addDatasourceLabel")}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">{t("alertRule.datasourceLabelsHint")}</p>
              {form.datasourceLabels.map((pair, index) => (
                <div key={index} className="grid gap-2 sm:grid-cols-2">
                  <Input
                    value={pair.key}
                    placeholder={t("alertRule.datasourceLabelKeyPlaceholder")}
                    onChange={(event) => {
                      const datasourceLabels = [...form.datasourceLabels]
                      datasourceLabels[index] = { ...pair, key: event.target.value }
                      setForm({ ...form, datasourceLabels })
                    }}
                  />
                  <Input
                    value={pair.value}
                    placeholder={t("alertRule.datasourceLabelValuePlaceholder")}
                    onChange={(event) => {
                      const datasourceLabels = [...form.datasourceLabels]
                      datasourceLabels[index] = { ...pair, value: event.target.value }
                      setForm({ ...form, datasourceLabels })
                    }}
                  />
                </div>
              ))}
              <p className="text-xs text-muted-foreground">
                {t("alertRule.datasourceMatchPreview", {
                  count: matchDatasourceCount(sources, form.type, form.datasourceLabels),
                })}
              </p>
            </div>
            <FormField label={t("alertRule.summary")} htmlFor="rule-summary" className="sm:col-span-full">
              <Textarea
                id="rule-summary"
                value={form.summary}
                placeholder={t("alertRule.summaryPlaceholder")}
                onChange={(event) => setForm({ ...form, summary: event.target.value })}
              />
            </FormField>
            <FormField label={t("alertRule.descriptionField")} htmlFor="rule-description" className="sm:col-span-full">
              <Textarea
                id="rule-description"
                value={form.description}
                placeholder={t("alertRule.descriptionPlaceholder")}
                onChange={(event) => setForm({ ...form, description: event.target.value })}
              />
            </FormField>
              </>
            ) : null}
            {formStep === 1 ? (
              <>
            <FormField label={t("datasource.expr")} htmlFor="rule-expr" hint={t("datasource.exprHint")} className="sm:col-span-full">
              <PromqlEditor
                value={form.expr}
                datasourceUid={matchedMetricsUid(sources, form.type, form.datasourceLabels)}
                onChange={(value) => setForm({ ...form, expr: value })}
              />
            </FormField>
            <FormField label={t("alertRule.labels")} htmlFor="rule-labels" hint={t("alertRule.labelsHint")} className="sm:col-span-full">
              <Input id="rule-labels" value={form.labels} placeholder='{"job":"api"}' onChange={(event) => setForm({ ...form, labels: event.target.value })} />
            </FormField>
              </>
            ) : null}
            {formStep === 2 ? (
              <>
            <div className="grid gap-2 sm:col-span-full">
              <Label>{t("alertRule.levels")}</Label>
              {form.levels.map((level, index) => (
                <div key={index} className="grid gap-2 rounded-lg border p-3 sm:grid-cols-4">
                  <Select
                    value={level.condition}
                    onValueChange={(value) => {
                      const levels = [...form.levels]
                      levels[index] = { ...level, condition: value }
                      setForm({ ...form, levels })
                    }}
                  >
                    <SelectTrigger className="w-full" aria-label={t("alertRule.condition")}>
                      <SelectValue placeholder={t("alertRule.condition")} />
                    </SelectTrigger>
                    <SelectContent>
                      {conditions.map((item) => (
                        <SelectItem key={item} value={item}>
                          {enumLabel(t, item)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Select
                    value={level.sample}
                    onValueChange={(value) => {
                      const levels = [...form.levels]
                      levels[index] = { ...level, sample: value }
                      setForm({ ...form, levels })
                    }}
                  >
                    <SelectTrigger className="w-full" aria-label={t("alertRule.sample")}>
                      <SelectValue placeholder={t("alertRule.sample")} />
                    </SelectTrigger>
                    <SelectContent>
                      {samples.map((item) => (
                        <SelectItem key={item} value={item}>
                          {enumLabel(t, item)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Input
                    value={level.threshold}
                    placeholder={t("alertRule.threshold")}
                    onChange={(event) => {
                      const levels = [...form.levels]
                      levels[index] = { ...level, threshold: event.target.value }
                      setForm({ ...form, levels })
                    }}
                  />
                  <Input
                    value={String(level.forSeconds)}
                    placeholder={t("alertRule.forSeconds")}
                    onChange={(event) => {
                      const levels = [...form.levels]
                      levels[index] = { ...level, forSeconds: Number(event.target.value || "0") }
                      setForm({ ...form, levels })
                    }}
                  />
                  <div className="sm:col-span-full">
                    <Select
                      value={level.levelUid || "NONE"}
                      onValueChange={(value) => {
                        const levels = [...form.levels]
                        levels[index] = { ...level, levelUid: value === "NONE" ? "" : value }
                        setForm({ ...form, levels })
                      }}
                    >
                      <SelectTrigger className="w-full" aria-label={t("alertRule.levelPlaceholder")}>
                        <SelectValue placeholder={t("alertRule.levelPlaceholder")} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="NONE">{t("alertRule.levelPlaceholder")}</SelectItem>
                        {levels.map((item) => (
                          <SelectItem key={item.uid} value={item.uid}>
                            {item.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              ))}
              <Button type="button" variant="outline" onClick={() => setForm({ ...form, levels: [...form.levels, { ...emptyLevel }] })}>
                {t("alertRule.addLevel")}
              </Button>
            </div>
            <div className="grid gap-2 sm:col-span-full">
              <Label>{t("alertRule.notify")}</Label>
              {notifyGroups.length === 0 ? (
                <p className="text-sm text-muted-foreground">{t("alertRule.notifyEmpty")}</p>
              ) : (
                <div className="grid gap-2">
                  {notifyGroups.map((item) => {
                    const checked = form.notificationGroupUids.includes(item.uid)
                    return (
                      <label key={item.uid} className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(event) => {
                            const next = event.target.checked
                              ? [...form.notificationGroupUids, item.uid]
                              : form.notificationGroupUids.filter((uid) => uid !== item.uid)
                            setForm({ ...form, notificationGroupUids: next })
                          }}
                        />
                        {item.name}
                      </label>
                    )
                  })}
                </div>
              )}
            </div>
              </>
            ) : null}
          </StepForm>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detail !== null}
        title={detail?.name ?? ""}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                { label: t("datasource.type"), value: enumLabel(t, detail.type) },
                { label: t("namespace.status"), value: enumLabel(t, detail.status) },
                { label: t("datasource.expr"), value: detail.expr },
                { label: t("alertRule.datasourceLabels"), value: detail.datasourceLabels || "{}" },
                { label: t("alertRule.summary"), value: detail.summary },
                { label: t("alertRule.forSeconds"), value: String(detail.levels[0]?.forSeconds ?? 0) },
              ]
            : []
        }
      />
      <DetailSheet
        open={groupDetail !== null}
        title={groupDetail?.name ?? ""}
        description={groupDetail ? enumLabel(t, groupDetail.status) : ""}
        onOpenChange={(value) => {
          if (!value) {
            setGroupDetail(null)
          }
        }}
        fields={groupDetail ? [{ label: t("detail.uid"), value: groupDetail.uid }] : []}
      />
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
    </section>
  )
}
