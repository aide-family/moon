import { EllipsisVertical, Trash2 } from "lucide-react"
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
  getAlertRule,
  listAlertRuleGroups,
  listAlertRules,
  updateAlertRule,
  updateAlertRuleGroup,
  updateAlertRuleGroupStatus,
  updateAlertRuleStatus,
} from "@/api/alert-rule"
import type {
  AlertRuleGroupItem,
  AlertRuleItem,
  AlertRuleLevelInput,
  AlertRuleLevelItem,
} from "@/api/alert-rule/types"
import { listDatasources } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import {
  LogQLEditor,
  LogsQLEditor,
  PromQLEditor,
} from "@/components/query-expression-editor"
import { RowActions, type RowAction } from "@/components/row-actions"
import { ListFrame } from "@/components/list-frame"
import { StatusBadge } from "@/components/status-badge"
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
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
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
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireSelected,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { failureMessage, readFormError } from "@/lib/form-error"
import {
  labelMatchersMatch,
  parseLabelPairs,
  serializeLabelPairs,
  type LabelPair,
} from "@/lib/labels"
import { useListQuery } from "@/hooks/use-list-query"

// 规则不再保存类型，只保存驱动；数据源匹配也只按驱动。
const drivers = [
  "METRICS_PROMETHEUS",
  "METRICS_VICTORIA_METRICS",
  "LOGS_ELASTICSEARCH",
  "LOGS_LOKI",
  "LOGS_VLOGS",
  "TRACE_JAEGER",
]
const conditions = ["EQ", "NE", "GT", "GTE", "LT", "LTE", "IN", "NOT_IN"]
const samples = ["FOR", "MAX", "MIN"]

interface RuleDraft {
  name: string
  driver: string
  expr: string
  labels: string
  datasourceLabels: LabelPair[]
  summary: string
  description: string
  levels: AlertRuleLevelInput[]
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
  driver: "METRICS_PROMETHEUS",
  expr: "",
  labels: "",
  datasourceLabels: [{ key: "", value: "" }],
  summary: "",
  description: "",
  levels: [{ ...emptyLevel }],
}

function parseDatasourceLabels(raw: string): Record<string, string> | null {
  try {
    return JSON.parse(raw || "{}") as Record<string, string>
  } catch {
    return null
  }
}

// datasourceMatchLabels 与后端保持一致：数据源自身标签之外，
// datasource_level 和 datasource 也可以被规则匹配。
function datasourceMatchLabels(
  item: DatasourceItem
): Record<string, string> | null {
  const labels = parseDatasourceLabels(item.labels)
  if (labels === null) {
    return null
  }
  return {
    ...labels,
    datasource: item.name,
    ...(item.levelName ? { datasource_level: item.levelName } : {}),
  }
}

function matchDatasourceCount(
  sources: DatasourceItem[],
  driver: string,
  matchers: LabelPair[]
): number {
  return sources.filter((item) => {
    if (item.driver !== driver || item.status !== "ACTIVE") {
      return false
    }
    const labels = datasourceMatchLabels(item)
    return labels !== null && labelMatchersMatch(matchers, labels)
  }).length
}

function matchedDatasourceUid(
  sources: DatasourceItem[],
  driver: string,
  matchers: LabelPair[]
): string {
  if (
    !driver.startsWith("METRICS_") &&
    driver !== "LOGS_LOKI" &&
    driver !== "LOGS_VLOGS"
  ) {
    return ""
  }
  return (
    sources.find((item) => {
      if (item.driver !== driver || item.status !== "ACTIVE") {
        return false
      }
      const labels = datasourceMatchLabels(item)
      return labels !== null && labelMatchersMatch(matchers, labels)
    })?.uid ?? ""
  )
}

export function AlertRulePage() {
  const { t } = useTranslation()
  const session = useSession()
  const groupsQuery = useListQuery()
  const [params, setParams] = useSearchParams()
  const groupUid = params.get("groupUid") || ""
  const ruleUid = params.get("ruleUid") || ""
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
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [groupDetail, setGroupDetail] = useState<AlertRuleGroupItem | null>(
    null
  )
  const [attempt, setAttempt] = useState(0)
  const [groupOpen, setGroupOpen] = useState(false)
  const [groupName, setGroupName] = useState("")
  const [editingGroup, setEditingGroup] = useState<AlertRuleGroupItem | null>(
    null
  )
  const [ruleOpen, setRuleOpen] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRuleItem | null>(null)
  const [form, setForm] = useState<RuleDraft>(emptyRule)
  const [formStep, setFormStep] = useState(0)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AlertRuleItem | null>(null)
  const [confirm, setConfirm] = useState<{
    kind: "group" | "rule"
    action: "disable" | "delete"
    uid: string
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const [shownAction, setShownAction] = useState<"disable" | "delete">(
    "disable"
  )
  const canWrite = session.can("alert_rule:write")

  function askConfirm(
    kind: "group" | "rule",
    action: "disable" | "delete",
    uid: string
  ) {
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
          setLoadError(failureMessage(err, ""))
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
  }, [
    groupsQuery.page,
    groupsQuery.pageSize,
    groupsQuery.name,
    attempt,
    session.namespaceUid,
  ])

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
  }, [
    groupUid,
    rulePage,
    ruleName,
    groupsQuery.pageSize,
    attempt,
    session.namespaceUid,
  ])

  // 实时告警详情的规则名称链接带上 ruleUid，进入本页后直接打开规则详情。
  useEffect(() => {
    if (!ruleUid) {
      return
    }
    let active = true
    async function load() {
      try {
        const result = await getAlertRule(ruleUid)
        if (!active) {
          return
        }
        setDetail(result.rule)
      } catch {
        // 规则已删除或无权访问时保持列表，不打开空详情。
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [ruleUid, session.namespaceUid])

  useEffect(() => {
    let active = true
    async function load() {
      try {
        const [levelResult, sourceResult] = await Promise.all([
          listAlertDictionaries({ page: 1, pageSize: 100, type: "ALERT" }),
          listDatasources({ page: 1, pageSize: 100 }),
        ])
        if (!active) {
          return
        }
        setLevels(levelResult.items ?? [])
        setSources(sourceResult.items ?? [])
      } catch {
        if (active) {
          setLevels([])
          setSources([])
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
          }
        : emptyRule
    )
    setError("")
    setFieldErrors({})
    setFormStep(0)
    setRuleOpen(true)
  }

  async function onSaveGroup(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(groupName, t("form.required"))],
      ["name", maxLength(groupName, 64, t("form.nameTooLong"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
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
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  function validateRuleStep(step: number): FieldErrors {
    if (step === 0) {
      return collectFieldErrors([
        ["name", requireTrimmed(form.name, t("form.required"))],
        ["name", maxLength(form.name, 64, t("form.nameTooLong"))],
        ["driver", requireTrimmed(form.driver, t("form.required"))],
      ])
    }
    if (step === 1) {
      return collectFieldErrors([
        ["expr", requireTrimmed(form.expr, t("form.required"))],
      ])
    }
    if (step === 2) {
      if (form.levels.length === 0) {
        return { levels: t("form.required") }
      }
      const checks: Array<[string, string | null]> = []
      form.levels.forEach((level, index) => {
        checks.push(
          [
            `level.${index}.levelUid`,
            requireSelected(level.levelUid, t("form.required")),
          ],
          [
            `level.${index}.condition`,
            requireSelected(level.condition, t("form.required")),
          ],
          [
            `level.${index}.threshold`,
            requireTrimmed(level.threshold, t("form.required")),
          ],
          [
            `level.${index}.sample`,
            requireSelected(level.sample, t("form.required")),
          ],
          [
            `level.${index}.forSeconds`,
            !Number.isFinite(level.forSeconds) || level.forSeconds < 0
              ? t("form.required")
              : null,
          ]
        )
      })
      return collectFieldErrors(checks)
    }
    return {}
  }

  async function onSaveRule(event: React.FormEvent) {
    event.preventDefault()
    if (!groupUid) {
      return
    }
    const stepErrors = [0, 1, 2].map(validateRuleStep)
    const errors = stepErrors.reduce<FieldErrors>(
      (all, current) => ({ ...all, ...current }),
      {}
    )
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      setError("")
      const firstInvalidStep = stepErrors.findIndex(hasFieldErrors)
      if (firstInvalidStep >= 0) {
        setFormStep(firstInvalidStep)
      }
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
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
        ? {
            key: "disable",
            label: t("action.disable"),
            destructive: true,
            onSelect: () => askConfirm("rule", "disable", item.uid),
          }
        : {
            key: "enable",
            label: t("action.enable"),
            onSelect: () =>
              void updateAlertRuleStatus(item.uid, "ACTIVE").then(() =>
                setAttempt((value) => value + 1)
              ),
          },
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => askConfirm("rule", "delete", item.uid),
      },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("alertRule.title")}
        description={t("alertRule.description")}
      />
      <div className="grid min-h-0 min-w-0 flex-1 gap-6 lg:grid-cols-[minmax(0,18rem)_minmax(0,1fr)]">
        <div className="flex min-h-0 min-w-0 flex-col gap-4">
          <form
            className="flex gap-2"
            onSubmit={(event) => {
              event.preventDefault()
              groupsQuery.update({ name: groupDraft })
            }}
          >
            <Input
              value={groupDraft}
              placeholder={t("alertRule.groupPlaceholder")}
              onChange={(event) => setGroupDraft(event.target.value)}
            />
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
                setFieldErrors({})
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
            onPrev={() =>
              groupsQuery.update({ page: Math.max(1, groupsQuery.page - 1) })
            }
            onNext={() => groupsQuery.update({ page: groupsQuery.page + 1 })}
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
                <Button
                  variant="outline"
                  onClick={() => setAttempt((value) => value + 1)}
                >
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!loading && !failed && groups.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("alertRule.groupEmpty")}
              </p>
            ) : null}
            <div className="grid min-h-0 min-w-0 flex-1 content-start gap-2 overflow-x-hidden overflow-y-auto">
              {groups.map((item) => (
                <div key={item.uid} className="flex min-w-0 items-center gap-1">
                  <Button
                    variant={item.uid === groupUid ? "secondary" : "ghost"}
                    className="h-auto min-w-0 flex-1 justify-start overflow-hidden px-2 py-2"
                    onClick={() => setParam("groupUid", item.uid)}
                  >
                    <span className="flex w-full min-w-0 items-center gap-2">
                      <span
                        className="min-w-0 flex-1 truncate text-left"
                        title={item.name}
                      >
                        {item.name}
                      </span>
                      <StatusBadge value={item.status} />
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
                      <DropdownMenuItem onSelect={() => setGroupDetail(item)}>
                        {t("action.detail")}
                      </DropdownMenuItem>
                      {canWrite ? (
                        <DropdownMenuItem
                          onSelect={() => {
                            setEditingGroup(item)
                            setGroupName(item.name)
                            setError("")
                            setFieldErrors({})
                            setGroupOpen(true)
                          }}
                        >
                          {t("action.edit")}
                        </DropdownMenuItem>
                      ) : null}
                      {canWrite && item.status !== "ACTIVE" ? (
                        <DropdownMenuItem
                          onSelect={() => {
                            void updateAlertRuleGroupStatus(
                              item.uid,
                              "ACTIVE"
                            ).then(() => setAttempt((value) => value + 1))
                          }}
                        >
                          {t("action.enable")}
                        </DropdownMenuItem>
                      ) : null}
                      {canWrite ? <DropdownMenuSeparator /> : null}
                      {canWrite && item.status === "ACTIVE" ? (
                        <DropdownMenuItem
                          variant="destructive"
                          onSelect={() =>
                            askConfirm("group", "disable", item.uid)
                          }
                        >
                          {t("action.disable")}
                        </DropdownMenuItem>
                      ) : null}
                      {canWrite ? (
                        <DropdownMenuItem
                          variant="destructive"
                          onSelect={() =>
                            askConfirm("group", "delete", item.uid)
                          }
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
        </div>
        <div className="flex min-h-0 flex-col gap-4">
          {!groupUid ? (
            <p className="text-sm text-muted-foreground">
              {t("alertRule.pickGroup")}
            </p>
          ) : null}
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
                <Input
                  value={ruleDraft}
                  placeholder={t("alertRule.namePlaceholder")}
                  onChange={(event) => setRuleDraft(event.target.value)}
                />
                <Button type="submit" variant="outline">
                  {t("action.search")}
                </Button>
              </form>
              {canWrite ? (
                <Button onClick={() => openRule()}>{t("action.create")}</Button>
              ) : null}
            </div>
          ) : null}
          <ListFrame
            total={ruleTotal}
            page={rulePage}
            pageSize={groupsQuery.pageSize}
            showPager={Boolean(groupUid) && rules.length > 0}
            onPrev={() =>
              setParam("rulePage", String(Math.max(1, rulePage - 1)))
            }
            onNext={() => setParam("rulePage", String(rulePage + 1))}
          >
            {groupUid && rules.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("alertRule.empty")}
              </p>
            ) : null}
            {rules.length > 0 ? (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("alertRule.name")}</TableHead>
                    <TableHead>{t("datasource.expr")}</TableHead>
                    <TableHead>{t("datasource.driver")}</TableHead>
                    <TableHead>{t("namespace.status")}</TableHead>
                    <TableHead className="text-center">{t("action.actions")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((item) => (
                    <TableRow key={item.uid}>
                      <TableCell>{item.name}</TableCell>
                      <TableCell className="max-w-[28rem] min-w-[16rem] align-top">
                        <RuleExpr
                          expr={item.expr}
                          levels={item.levels}
                          maxHeightClass="max-h-24"
                        />
                      </TableCell>
                      <TableCell><StatusBadge value={item.driver} /></TableCell>
                      <TableCell><StatusBadge value={item.status} /></TableCell>
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
            <DialogTitle>
              {editingGroup ? t("action.edit") : t("alertRule.createGroup")}
            </DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onSaveGroup}>
            <FormField
              label={t("alertRule.group")}
              htmlFor="group-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="group-name"
                value={groupName}
                maxLength={64}
                placeholder={t("alertRule.groupPlaceholder")}
                onChange={(event) => setGroupName(event.target.value)}
                required
              />
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
            <DialogTitle>
              {editingRule ? t("action.edit") : t("action.create")}
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
              const errors = validateRuleStep(formStep)
              if (hasFieldErrors(errors)) {
                setFieldErrors(errors)
                setError("")
                return
              }
              setError("")
              setFieldErrors({})
              setFormStep((value) => Math.min(2, value + 1))
            }}
            onSubmit={onSaveRule}
          >
            {formStep === 0 ? (
              <>
                <FormField
                  label={t("alertRule.name")}
                  htmlFor="rule-name"
                  className="sm:col-span-full"
                  required
                  error={fieldErrors.name}
                >
                  <Input
                    id="rule-name"
                    value={form.name}
                    maxLength={64}
                    placeholder={t("alertRule.namePlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, name: event.target.value })
                    }
                    required
                  />
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
                      {drivers.map((item) => (
                        <SelectItem key={item} value={item}>
                          {enumLabel(t, item)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <div className="grid gap-2 sm:col-span-full">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-medium">
                      {t("alertRule.datasourceLabels")}
                    </span>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        setForm({
                          ...form,
                          datasourceLabels: [
                            ...form.datasourceLabels,
                            { key: "", value: "" },
                          ],
                        })
                      }
                    >
                      {t("alertRule.addDatasourceLabel")}
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t("alertRule.datasourceLabelsHint")}
                  </p>
                  {form.datasourceLabels.map((pair, index) => (
                    <div key={index} className="grid gap-2 sm:grid-cols-2">
                      <Input
                        value={pair.key}
                        placeholder={t(
                          "alertRule.datasourceLabelKeyPlaceholder"
                        )}
                        onChange={(event) => {
                          const datasourceLabels = [...form.datasourceLabels]
                          datasourceLabels[index] = {
                            ...pair,
                            key: event.target.value,
                          }
                          setForm({ ...form, datasourceLabels })
                        }}
                      />
                      <Input
                        value={pair.value}
                        placeholder={t(
                          "alertRule.datasourceLabelValuePlaceholder"
                        )}
                        onChange={(event) => {
                          const datasourceLabels = [...form.datasourceLabels]
                          datasourceLabels[index] = {
                            ...pair,
                            value: event.target.value,
                          }
                          setForm({ ...form, datasourceLabels })
                        }}
                      />
                    </div>
                  ))}
                  <p className="text-xs text-muted-foreground">
                    {t("alertRule.datasourceMatchPreview", {
                      count: matchDatasourceCount(
                        sources,
                        form.driver,
                        form.datasourceLabels
                      ),
                    })}
                  </p>
                </div>
                <FormField
                  label={t("alertRule.summary")}
                  htmlFor="rule-summary"
                  className="sm:col-span-full"
                >
                  <Textarea
                    id="rule-summary"
                    value={form.summary}
                    placeholder={t("alertRule.summaryPlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, summary: event.target.value })
                    }
                  />
                </FormField>
                <FormField
                  label={t("alertRule.descriptionField")}
                  htmlFor="rule-description"
                  className="sm:col-span-full"
                >
                  <Textarea
                    id="rule-description"
                    value={form.description}
                    placeholder={t("alertRule.descriptionPlaceholder")}
                    onChange={(event) =>
                      setForm({ ...form, description: event.target.value })
                    }
                  />
                </FormField>
              </>
            ) : null}
            {formStep === 1 ? (
              <>
                <FormField
                  label={t("datasource.expr")}
                  htmlFor="rule-expr"
                  hint={
                    form.driver === "LOGS_LOKI"
                      ? t("datasource.lokiMetricExprHint")
                      : form.driver === "LOGS_VLOGS"
                        ? t("datasource.vlogsMetricExprHint")
                        : form.driver === "LOGS_ELASTICSEARCH"
                          ? t("datasource.esExprHint")
                          : form.driver === "METRICS_VICTORIA_METRICS"
                            ? t("datasource.vmExprHint")
                            : t("datasource.exprHint")
                  }
                  className="sm:col-span-full"
                  error={fieldErrors.expr}
                >
                  {form.driver === "LOGS_LOKI" ? (
                    <LogQLEditor
                      value={form.expr}
                      datasourceUid={matchedDatasourceUid(
                        sources,
                        form.driver,
                        form.datasourceLabels
                      )}
                      onChange={(value) => setForm({ ...form, expr: value })}
                    />
                  ) : form.driver === "LOGS_VLOGS" ? (
                    <LogsQLEditor
                      value={form.expr}
                      datasourceUid={matchedDatasourceUid(
                        sources,
                        form.driver,
                        form.datasourceLabels
                      )}
                      onChange={(value) => setForm({ ...form, expr: value })}
                    />
                  ) : (
                    <PromQLEditor
                      value={form.expr}
                      variant={
                        form.driver === "LOGS_ELASTICSEARCH" ? "json" : "promql"
                      }
                      datasourceUid={matchedDatasourceUid(
                        sources,
                        form.driver,
                        form.datasourceLabels
                      )}
                      onChange={(value) => setForm({ ...form, expr: value })}
                    />
                  )}
                </FormField>
                <FormField
                  label={t("alertRule.labels")}
                  htmlFor="rule-labels"
                  hint={t("alertRule.labelsHint")}
                  className="sm:col-span-full"
                >
                  <Input
                    id="rule-labels"
                    value={form.labels}
                    placeholder='{"job":"api"}'
                    onChange={(event) =>
                      setForm({ ...form, labels: event.target.value })
                    }
                  />
                </FormField>
              </>
            ) : null}
            {formStep === 2 ? (
              <>
                <FormField
                  label={t("alertRule.levels")}
                  hint={t("alertRule.levelsHint")}
                  className="sm:col-span-full"
                  required
                  error={fieldErrors.levels}
                >
                  <div className="grid gap-3">
                    <p className="text-xs text-muted-foreground">
                      {t("alertRule.levelsHint")}
                    </p>
                    {form.levels.map((level, index) => (
                      <div
                        key={index}
                        className="grid gap-3 rounded-lg border p-3"
                      >
                        <div className="flex items-start gap-2">
                          <div className="min-w-0 flex-1">
                            <FormField
                              label={t("alertRule.levelLabel")}
                              required
                              error={fieldErrors[`level.${index}.levelUid`]}
                            >
                              <Select
                                value={level.levelUid || "NONE"}
                                onValueChange={(value) => {
                                  const next = [...form.levels]
                                  next[index] = {
                                    ...level,
                                    levelUid: value === "NONE" ? "" : value,
                                  }
                                  setForm({ ...form, levels: next })
                                }}
                              >
                                <SelectTrigger
                                  className="w-full"
                                  aria-label={t("alertRule.levelLabel")}
                                >
                                  <SelectValue
                                    placeholder={t(
                                      "alertRule.levelPlaceholder"
                                    )}
                                  />
                                </SelectTrigger>
                                <SelectContent>
                                  <SelectItem value="NONE">
                                    {t("alertRule.levelPlaceholder")}
                                  </SelectItem>
                                  {levels.map((item) => (
                                    <SelectItem key={item.uid} value={item.uid}>
                                      {item.name}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                            </FormField>
                          </div>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="mt-6 shrink-0"
                            aria-label={t("alertRule.removeLevel")}
                            onClick={() =>
                              setForm({
                                ...form,
                                levels: form.levels.filter(
                                  (_, item) => item !== index
                                ),
                              })
                            }
                          >
                            <Trash2 />
                          </Button>
                        </div>
                        <div className="grid gap-3 sm:grid-cols-4">
                          <FormField
                            label={t("alertRule.condition")}
                            required
                            error={fieldErrors[`level.${index}.condition`]}
                          >
                            <Select
                              value={level.condition}
                              onValueChange={(value) => {
                                const next = [...form.levels]
                                next[index] = { ...level, condition: value }
                                setForm({ ...form, levels: next })
                              }}
                            >
                              <SelectTrigger
                                className="w-full"
                                aria-label={t("alertRule.condition")}
                              >
                                <SelectValue
                                  placeholder={t("alertRule.condition")}
                                />
                              </SelectTrigger>
                              <SelectContent>
                                {conditions.map((item) => (
                                  <SelectItem key={item} value={item}>
                                    {enumLabel(t, item)}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </FormField>
                          <FormField
                            label={t("alertRule.threshold")}
                            required
                            error={fieldErrors[`level.${index}.threshold`]}
                          >
                            <Input
                              value={level.threshold}
                              placeholder={t("alertRule.threshold")}
                              onChange={(event) => {
                                const next = [...form.levels]
                                next[index] = {
                                  ...level,
                                  threshold: event.target.value,
                                }
                                setForm({ ...form, levels: next })
                              }}
                            />
                          </FormField>
                          <FormField
                            label={t("alertRule.sample")}
                            required
                            error={fieldErrors[`level.${index}.sample`]}
                          >
                            <Select
                              value={level.sample}
                              onValueChange={(value) => {
                                const next = [...form.levels]
                                next[index] = { ...level, sample: value }
                                setForm({ ...form, levels: next })
                              }}
                            >
                              <SelectTrigger
                                className="w-full"
                                aria-label={t("alertRule.sample")}
                              >
                                <SelectValue
                                  placeholder={t("alertRule.sample")}
                                />
                              </SelectTrigger>
                              <SelectContent>
                                {samples.map((item) => (
                                  <SelectItem key={item} value={item}>
                                    {enumLabel(t, item)}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </FormField>
                          <FormField
                            label={t("alertRule.forSeconds")}
                            required
                            error={fieldErrors[`level.${index}.forSeconds`]}
                          >
                            <div className="flex items-center gap-2">
                              <Input
                                value={String(level.forSeconds)}
                                placeholder={t("alertRule.forSeconds")}
                                onChange={(event) => {
                                  const next = [...form.levels]
                                  next[index] = {
                                    ...level,
                                    forSeconds: Number(
                                      event.target.value || "0"
                                    ),
                                  }
                                  setForm({ ...form, levels: next })
                                }}
                              />
                              <span className="shrink-0 text-sm text-muted-foreground">
                                {t("alertRule.forSecondsUnit")}
                              </span>
                            </div>
                          </FormField>
                        </div>
                      </div>
                    ))}
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() =>
                        setForm({
                          ...form,
                          levels: [...form.levels, { ...emptyLevel }],
                        })
                      }
                    >
                      {t("alertRule.addLevel")}
                    </Button>
                  </div>
                </FormField>
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
            if (ruleUid) {
              setParam("ruleUid", "")
            }
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                {
                  label: t("datasource.driver"),
                  value: <StatusBadge value={detail.driver} />,
                },
                {
                  label: t("namespace.status"),
                  value: <StatusBadge value={detail.status} />,
                },
                {
                  label: t("datasource.expr"),
                  value: <RuleExpr expr={detail.expr} levels={detail.levels} />,
                },
                {
                  label: t("alertRule.datasourceLabels"),
                  value: detail.datasourceLabels || "{}",
                },
                { label: t("alertRule.summary"), value: detail.summary },
                {
                  label: t("alertRule.forSeconds"),
                  value: String(detail.levels[0]?.forSeconds ?? 0),
                },
              ]
            : []
        }
      />
      <DetailSheet
        open={groupDetail !== null}
        title={groupDetail?.name ?? ""}
        description={groupDetail ? <StatusBadge value={groupDetail.status} /> : ""}
        onOpenChange={(value) => {
          if (!value) {
            setGroupDetail(null)
          }
        }}
        fields={
          groupDetail
            ? [{ label: t("detail.uid"), value: groupDetail.uid }]
            : []
        }
      />
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

/**
 * 告警规则的表达式展示：不加边框或底色容器，先展示原始表达式，
 * 再逐行展示级别条件。列表和详情共用，避免两处排版不一致。
 */
function RuleExpr({
  expr,
  levels,
  maxHeightClass = "max-h-64",
}: {
  expr: string
  levels: AlertRuleLevelItem[]
  maxHeightClass?: string
}) {
  const { t } = useTranslation()
  return (
    <div className="grid min-w-0 gap-2">
      <pre
        className={`overflow-x-hidden overflow-y-auto font-mono text-xs break-words whitespace-pre-wrap ${maxHeightClass}`}
      >
        {expr || "—"}
      </pre>
      {levels.length > 0 ? (
        <div className="grid min-w-0 gap-0.5 font-mono text-xs text-muted-foreground">
          {levels.map((level, index) => (
            <span key={level.uid || index} className="min-w-0 break-words">
              {t("alertEvent.conditionSummary", {
                level: level.levelName || level.levelUid,
                sample: enumLabel(t, level.sample),
                condition: enumLabel(t, level.condition),
                threshold: level.threshold || "—",
                forSeconds: level.forSeconds,
              })}
            </span>
          ))}
        </div>
      ) : null}
    </div>
  )
}
