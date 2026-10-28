import { EllipsisVertical, Trash2 } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { useSearchParams } from "react-router-dom"
import { toast } from "sonner"

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
  AlertRuleVariable,
} from "@/api/alert-rule/types"
import { listDatasources } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { ApiError } from "@/api/http"
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
import { SearchableSelect } from "@/components/searchable-select"
import { ListFrame } from "@/components/list-frame"
import { StatusBadge } from "@/components/status-badge"
import { StepForm } from "@/components/step-form"
import { TemplateEditor } from "@/components/template-editor"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
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
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
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
import { parseLabelPairs, serializeLabelPairs, type LabelPair } from "@/lib/labels"
import { loadAllPagedItems } from "@/lib/select-pages"
import { useListQuery } from "@/hooks/use-list-query"

import { AlertRuleDetailSheet } from "./detail"
import {
  datasourceMatchKeys,
  datasourceMatchValues,
  matchDatasourceCount,
  matchedDatasourceUid,
  ruleLabelKeys,
} from "./matching"
import { AlertRuleExpr } from "./rule-expr"

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

// 自定义变量的前端约束与后端 alertrule.prepare 保持一致：
// 名字是 ASCII 模板标识符且在规则内唯一，数量/长度有上限，保留名不能再定义。
const maxRuleVariables = 20
const ruleVariableNameMax = 32
const ruleVariableExprMax = 4096
const ruleVariableDescriptionMax = 200
const ruleVariableNamePattern = /^[A-Za-z_][A-Za-z0-9_]{0,31}$/
const reservedRuleVariableNames = new Set([
  "labels",
  "value",
  "externalURL",
  "externalLabels",
])

// 后端 validateRuleVariables 的固定错误原文 → 变量区的 i18n key。原文变化时
// 退回通用提示，不会把英文原文直接显示在中文界面里。
const ruleVariableErrorKeys: Record<string, string> = {
  "too many variables": "alertRule.variableCountExceeded",
  "invalid variable name": "alertRule.variableNameInvalid",
  "reserved variable name": "alertRule.variableNameReserved",
  "duplicate variable name": "alertRule.variableNameDuplicate",
  "empty variable expr": "alertRule.variableExprRequired",
  "variable expr is too long": "alertRule.variableExprTooLong",
  "variable description is too long": "alertRule.variableDescriptionTooLong",
  "invalid variable template": "alertRule.variableExprInvalid",
}

const emptyVariable: AlertRuleVariable = { name: "", expr: "", description: "" }

// 变量查询语句的语言跟随规则驱动：不同数据源的示例不能混用，链路规则尚未
// 参与评测，因此不展示任何查询示例。返回 i18n key 的后缀。
function variableDriverKey(driver: string): "Metrics" | "Loki" | "Vlogs" | "Es" | "Trace" {
  switch (driver) {
    case "LOGS_LOKI":
      return "Loki"
    case "LOGS_VLOGS":
      return "Vlogs"
    case "LOGS_ELASTICSEARCH":
      return "Es"
    case "TRACE_JAEGER":
      return "Trace"
    default:
      return "Metrics"
  }
}

interface RuleDraft {
  name: string
  driver: string
  expr: string
  labels: string
  datasourceLabels: LabelPair[]
  summary: string
  description: string
  variables: AlertRuleVariable[]
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
  variables: [],
  levels: [{ ...emptyLevel }],
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
  const [failedRuleUid, setFailedRuleUid] = useState("")
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
  // 深链规则打不开时（已删除、无权限）仍然按默认规则组展示列表。
  const deepLinkFailed = ruleUid !== "" && failedRuleUid === ruleUid
  // 表达式编辑器和模板补全共用同一个数据源选择；规则标签会并入评测标签。
  const matchedUid = matchedDatasourceUid(
    sources,
    form.driver,
    form.datasourceLabels
  )
  const labelKeys = useMemo(() => ruleLabelKeys(form.labels), [form.labels])
  // 摘要/描述模板补全用：已声明且格式合法的变量名（空行和非法名字不参与补全）。
  const variableNames = useMemo(
    () =>
      form.variables
        .map((variable) => variable.name.trim())
        .filter((name) => ruleVariableNamePattern.test(name)),
    [form.variables]
  )
  // 表达式编辑器为空时按驱动给出对应语言的示例，避免占位文案和查询语言不匹配。
  const exprPlaceholder =
    form.driver === "LOGS_LOKI"
      ? t("alertRule.exprPlaceholderLoki")
      : form.driver === "LOGS_VLOGS"
        ? t("alertRule.exprPlaceholderVlogs")
        : form.driver === "LOGS_ELASTICSEARCH"
          ? t("alertRule.exprPlaceholderEs")
          : t("alertRule.exprPlaceholderMetrics")
  // 自定义变量的示例、占位和说明按驱动选择，避免在 Loki/ES 规则里出现 PromQL。
  const variableDriver = variableDriverKey(form.driver)
  const variableExprPlaceholder = t(
    `alertRule.variableExprPlaceholder${variableDriver}`
  )
  const variableExprHint = t(`alertRule.variableExprHint${variableDriver}`)
  const variableExample =
    variableDriver === "Trace"
      ? ""
      : t(`alertRule.variableExample${variableDriver}`)
  // 「数据源类型」就是规则驱动：匹配键 = 系统键 + 同驱动启用数据源的自定义标签键。
  // 具体的值在渲染每个匹配标签时按当前键推导。
  const datasourceMatchKeyOptions = useMemo(
    () =>
      datasourceMatchKeys(sources, form.driver).map((key) => ({
        value: key,
        label: key,
      })),
    [sources, form.driver]
  )

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

  // 实时告警详情的规则名称链接带上 ruleUid，进入本页后直接打开规则详情，
  // 并把左侧选中对齐到规则所属分组；打不开时交给默认选中逻辑选第一组。
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
        setFailedRuleUid("")
        setParams(
          (prev) => {
            if (prev.get("groupUid")) {
              return prev
            }
            const next = new URLSearchParams(prev)
            next.set("groupUid", result.rule.groupUid)
            return next
          },
          { replace: true }
        )
      } catch {
        // 规则已删除或无权访问时保持列表，仍按默认规则组展示。
        if (active) {
          setFailedRuleUid(ruleUid)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [ruleUid, session.namespaceUid, setParams])

  // 未指定规则组时默认选中第一组：规则列表、新建规则都依赖 groupUid。
  // 用 replace 写 URL，避免把默认选中记进浏览器历史。
  useEffect(() => {
    if (loading || failed || groupUid || groups.length === 0) {
      return
    }
    if (ruleUid && !deepLinkFailed) {
      return
    }
    const next = new URLSearchParams(params)
    next.set("groupUid", groups[0].uid)
    setParams(next, { replace: true })
  }, [
    loading,
    failed,
    groupUid,
    groups,
    ruleUid,
    deepLinkFailed,
    params,
    setParams,
  ])

  useEffect(() => {
    let active = true
    async function load() {
      try {
        const [levelResult, sourceResult] = await Promise.all([
          loadAllPagedItems((page, pageSize) =>
            listAlertDictionaries({ page, pageSize, type: "ALERT" })
          ),
          loadAllPagedItems((page, pageSize) =>
            listDatasources({ page, pageSize })
          ),
        ])
        if (!active) {
          return
        }
        setLevels(levelResult)
        setSources(sourceResult)
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

  function setParam(key: string, value: string, replace = false) {
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev)
        if (value) {
          next.set(key, value)
        } else {
          next.delete(key)
        }
        return next
      },
      { replace }
    )
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
            variables: (item.variables ?? []).map((variable) => ({
              name: variable.name,
              expr: variable.expr,
              description: variable.description ?? "",
            })),
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

  function updateDatasourceLabel(index: number, patch: Partial<LabelPair>) {
    const datasourceLabels = form.datasourceLabels.map((pair, current) =>
      current === index ? { ...pair, ...patch } : pair
    )
    setForm({ ...form, datasourceLabels })
  }

  function updateVariable(index: number, patch: Partial<AlertRuleVariable>) {
    setForm({
      ...form,
      variables: form.variables.map((variable, current) =>
        current === index ? { ...variable, ...patch } : variable
      ),
    })
  }

  // 打开详情时先用列表快照立即出内容，再用单条接口刷新；切换规则时不覆盖新详情。
  async function openDetail(item: AlertRuleItem) {
    setDetail(item)
    try {
      const result = await getAlertRule(item.uid)
      setDetail((current) => (current?.uid === item.uid ? result.rule : current))
    } catch {
      // 单条刷新失败时保留列表快照，避免抽屉变成空白。
    }
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
        ["expr", requireTrimmed(form.expr, t("form.required"))],
      ])
    }
    if (step === 1) {
      const checks: Array<[string, string | null]> = [
        [
          "summary",
          maxLength(form.summary, 2000, t("alertRule.templateTooLong")),
        ],
        [
          "description",
          maxLength(form.description, 2000, t("alertRule.templateTooLong")),
        ],
      ]
      if (form.variables.length > maxRuleVariables) {
        checks.push([
          "variables",
          t("alertRule.variableCountExceeded", { max: maxRuleVariables }),
        ])
      }
      // 重名只标后出现的那一行；name 为空时交给 required 报错。
      const seenNames = new Set<string>()
      form.variables.forEach((variable, index) => {
        const name = variable.name.trim()
        const invalidName = name !== "" && !ruleVariableNamePattern.test(name)
        const reservedName =
          name !== "" && reservedRuleVariableNames.has(name)
        const duplicateName = name !== "" && seenNames.has(name)
        if (name !== "" && !duplicateName) {
          seenNames.add(name)
        }
        checks.push(
          [
            `variable.${index}.name`,
            requireTrimmed(variable.name, t("form.required")),
          ],
          [
            `variable.${index}.name`,
            invalidName ? t("alertRule.variableNameInvalid") : null,
          ],
          [
            `variable.${index}.name`,
            reservedName ? t("alertRule.variableNameReserved") : null,
          ],
          [
            `variable.${index}.name`,
            duplicateName ? t("alertRule.variableNameDuplicate") : null,
          ],
          [
            `variable.${index}.expr`,
            requireTrimmed(variable.expr, t("form.required")),
          ],
          [
            `variable.${index}.expr`,
            maxLength(
              variable.expr,
              ruleVariableExprMax,
              t("alertRule.variableExprTooLong", { max: ruleVariableExprMax })
            ),
          ],
          [
            `variable.${index}.description`,
            maxLength(
              variable.description,
              ruleVariableDescriptionMax,
              t("alertRule.variableDescriptionTooLong", {
                max: ruleVariableDescriptionMax,
              })
            ),
          ]
        )
      })
      return collectFieldErrors(checks)
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

  // 后端的变量校验（名字、保留名、重名、数量、语句长度、模板语法）以 400 错误
  // 原文返回，字段路径不固定是 variables[0].name 这种形状。统一识别后落到第二步
  // 的「自定义变量」区域，避免用户只看到一条全局报错却不知道改哪。
  function variableFailureMessage(err: unknown): string {
    if (!(err instanceof ApiError) || err.status !== 400) {
      return ""
    }
    const key = ruleVariableErrorKeys[err.raw.trim()]
    if (key) {
      const vars: Record<string, number> = {}
      if (key === "alertRule.variableCountExceeded") {
        vars.max = maxRuleVariables
      } else if (key === "alertRule.variableExprTooLong") {
        vars.max = ruleVariableExprMax
      } else if (key === "alertRule.variableDescriptionTooLong") {
        vars.max = ruleVariableDescriptionMax
      }
      return t(key, vars)
    }
    const hasVariableField = Object.keys(err.fields).some((key) =>
      key.toLowerCase().includes("variable")
    )
    if (hasVariableField || err.raw.toLowerCase().includes("variable")) {
      return t("alertRule.variablesInvalid")
    }
    return ""
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
        variables: form.variables.map((variable) => ({
          name: variable.name.trim(),
          expr: variable.expr.trim(),
          description: variable.description.trim(),
        })),
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
      // 服务端模板语法校验只返回整体错误，映射到第二步对应字段。
      const templateField = failure.form.includes("summary template")
        ? "summary"
        : failure.form.includes("description template")
          ? "description"
          : ""
      if (templateField) {
        setFieldErrors({ [templateField]: t("alertRule.templateInvalid") })
        setError("")
        setFormStep(1)
        return
      }
      const variableError = variableFailureMessage(err)
      if (variableError) {
        setFieldErrors({ variables: variableError })
        setError("")
        setFormStep(1)
        return
      }
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
                    <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                    <TableActionsHead>{t("action.actions")}</TableActionsHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {rules.map((item) => (
                    <TableRow key={item.uid}>
                      <TableCell>{item.name}</TableCell>
                      <TableCell className="max-w-[28rem] min-w-[16rem] align-top">
                        <AlertRuleExpr
                          expr={item.expr}
                          levels={item.levels}
                          maxHeightClass="max-h-24"
                        />
                      </TableCell>
                      <TableCell><StatusBadge value={item.driver} /></TableCell>
                      <TableCell><StatusBadge value={item.status} /></TableCell>
                      <TableCell>
                        <RowActions nowrap
                          detailLabel={t("action.detail")}
                          moreLabel={t("action.more")}
                          onDetail={() => void openDetail(item)}
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
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onSaveGroup}>
            <DialogBody className="grid gap-4">
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
            </DialogBody>
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
                    onValueChange={(value) => {
                      if (value === form.driver) {
                        return
                      }
                      // 切换驱动后旧匹配键/值可能不再被任何数据源上报，继续保留会保存出永不命中的条件。
                      const clearsMatchers = form.datasourceLabels.length > 0
                      setForm({
                        ...form,
                        driver: value,
                        datasourceLabels: clearsMatchers
                          ? []
                          : form.datasourceLabels,
                      })
                      if (clearsMatchers) {
                        toast(t("alertRule.datasourceLabelsCleared"))
                      }
                    }}
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
                  {form.datasourceLabels.map((pair, index) => {
                    // 编辑旧规则时，历史键/值可能已不在当前候选里；保留为可选项，保存时不会丢数据。
                    const keyOptions =
                      pair.key === "" ||
                      datasourceMatchKeyOptions.some(
                        (option) => option.value === pair.key
                      )
                        ? datasourceMatchKeyOptions
                        : [
                            ...datasourceMatchKeyOptions,
                            { value: pair.key, label: pair.key },
                          ]
                    const values = datasourceMatchValues(
                      sources,
                      form.driver,
                      pair.key
                    )
                    const valueOptions = values.map((value) => ({
                      value,
                      label: value,
                    }))
                    if (pair.value !== "" && !values.includes(pair.value)) {
                      valueOptions.push({
                        value: pair.value,
                        label: pair.value,
                      })
                    }
                    return (
                      <div key={index} className="grid gap-2 sm:grid-cols-2">
                        <SearchableSelect
                          ariaLabel={t("alertRule.datasourceLabelKeyAria")}
                          value={pair.key}
                          options={keyOptions}
                          placeholder={t(
                            "alertRule.datasourceLabelKeyPlaceholder"
                          )}
                          searchPlaceholder={t(
                            "alertRule.datasourceLabelKeySearch"
                          )}
                          emptyText={t("alertRule.datasourceLabelKeyEmpty")}
                          clearLabel={t("alertRule.datasourceLabelClear")}
                          onChange={(key) => {
                            // 重新选中同一个键不算变更，避免丢掉编辑态里保留的值。
                            if (key === pair.key) {
                              return
                            }
                            const nextValues = datasourceMatchValues(
                              sources,
                              form.driver,
                              key
                            )
                            updateDatasourceLabel(index, {
                              key,
                              value: nextValues.includes(pair.value)
                                ? pair.value
                                : "",
                            })
                          }}
                        />
                        <SearchableSelect
                          ariaLabel={t("alertRule.datasourceLabelValueAria")}
                          value={pair.value}
                          options={valueOptions}
                          allowCustom
                          customLabel={(value) =>
                            t("alertRule.datasourceLabelCustomValue", { value })
                          }
                          customHint={t("alertRule.datasourceLabelCustomHint")}
                          placeholder={t(
                            "alertRule.datasourceLabelValuePlaceholder"
                          )}
                          searchPlaceholder={t(
                            "alertRule.datasourceLabelValueSearch"
                          )}
                          emptyText={t("alertRule.datasourceLabelValueEmpty")}
                          clearLabel={t("alertRule.datasourceLabelClear")}
                          onChange={(value) =>
                            updateDatasourceLabel(index, { value })
                          }
                        />
                      </div>
                    )
                  })}
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
                      datasourceUid={matchedUid}
                      placeholder={exprPlaceholder}
                      onChange={(value) => setForm({ ...form, expr: value })}
                    />
                  ) : form.driver === "LOGS_VLOGS" ? (
                    <LogsQLEditor
                      value={form.expr}
                      datasourceUid={matchedUid}
                      placeholder={exprPlaceholder}
                      onChange={(value) => setForm({ ...form, expr: value })}
                    />
                  ) : (
                    <PromQLEditor
                      value={form.expr}
                      variant={
                        form.driver === "LOGS_ELASTICSEARCH" ? "json" : "promql"
                      }
                      datasourceUid={matchedUid}
                      placeholder={exprPlaceholder}
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
            {formStep === 1 ? (
              <>
                <FormField
                  label={t("alertRule.summary")}
                  hint={t("alertRule.summaryHint")}
                  className="sm:col-span-full"
                  error={fieldErrors.summary}
                >
                  <TemplateEditor
                    value={form.summary}
                    singleLine
                    scope="alert"
                    datasourceUid={matchedUid}
                    datasourceDriver={form.driver}
                    expression={form.expr}
                    labelKeys={labelKeys}
                    variableNames={variableNames}
                    placeholder={t("alertRule.summaryPlaceholder")}
                    ariaLabel={t("alertRule.summary")}
                    onChange={(value) => setForm({ ...form, summary: value })}
                  />
                </FormField>
                <FormField
                  label={t("alertRule.descriptionField")}
                  hint={t("alertRule.descriptionHint")}
                  className="sm:col-span-full"
                  error={fieldErrors.description}
                >
                  <TemplateEditor
                    value={form.description}
                    height={160}
                    scope="alert"
                    datasourceUid={matchedUid}
                    datasourceDriver={form.driver}
                    expression={form.expr}
                    labelKeys={labelKeys}
                    variableNames={variableNames}
                    placeholder={t("alertRule.descriptionPlaceholder")}
                    ariaLabel={t("alertRule.descriptionField")}
                    onChange={(value) =>
                      setForm({ ...form, description: value })
                    }
                  />
                </FormField>
                <div className="grid min-w-0 gap-2 sm:col-span-full">
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-medium">
                      {t("alertRule.variables")}
                    </span>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      disabled={form.variables.length >= maxRuleVariables}
                      onClick={() =>
                        setForm({
                          ...form,
                          variables: [
                            ...form.variables,
                            { ...emptyVariable },
                          ],
                        })
                      }
                    >
                      {t("alertRule.addVariable")}
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {t("alertRule.variablesHint")}
                  </p>
                  {variableExample ? (
                    <p className="font-mono text-xs break-all text-muted-foreground">
                      {variableExample}
                    </p>
                  ) : null}
                  {fieldErrors.variables ? (
                    <p role="alert" className="text-sm text-destructive">
                      {fieldErrors.variables}
                    </p>
                  ) : null}
                  {form.variables.length === 0 ? (
                    <p className="text-xs text-muted-foreground">
                      {t("alertRule.variablesEmpty")}
                    </p>
                  ) : null}
                  {form.variables.map((variable, index) => (
                    <div
                      key={index}
                      className="grid min-w-0 gap-3 rounded-lg border p-3"
                    >
                      <div className="flex items-start gap-2">
                        <div className="grid min-w-0 flex-1 gap-3 sm:grid-cols-2">
                          <FormField
                            label={t("alertRule.variableName")}
                            htmlFor={`rule-variable-name-${index}`}
                            hint={t("alertRule.variableNameHint")}
                            required
                            error={fieldErrors[`variable.${index}.name`]}
                          >
                            <Input
                              id={`rule-variable-name-${index}`}
                              value={variable.name}
                              maxLength={ruleVariableNameMax}
                              placeholder={t(
                                "alertRule.variableNamePlaceholder"
                              )}
                              onChange={(event) =>
                                updateVariable(index, {
                                  name: event.target.value,
                                })
                              }
                            />
                          </FormField>
                          <FormField
                            label={t("alertRule.variableDescription")}
                            htmlFor={`rule-variable-description-${index}`}
                            error={
                              fieldErrors[`variable.${index}.description`]
                            }
                          >
                            <Input
                              id={`rule-variable-description-${index}`}
                              value={variable.description}
                              maxLength={ruleVariableDescriptionMax}
                              placeholder={t(
                                "alertRule.variableDescriptionPlaceholder"
                              )}
                              onChange={(event) =>
                                updateVariable(index, {
                                  description: event.target.value,
                                })
                              }
                            />
                          </FormField>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="mt-6 shrink-0"
                          aria-label={t("alertRule.removeVariable")}
                          onClick={() =>
                            setForm({
                              ...form,
                              variables: form.variables.filter(
                                (_, current) => current !== index
                              ),
                            })
                          }
                        >
                          <Trash2 />
                        </Button>
                      </div>
                      <FormField
                        label={t("alertRule.variableExpr")}
                        hint={variableExprHint}
                        required
                        error={fieldErrors[`variable.${index}.expr`]}
                      >
                        <TemplateEditor
                          value={variable.expr}
                          height={96}
                          scope="alert"
                          datasourceUid={matchedUid}
                          datasourceDriver={form.driver}
                          expression={form.expr}
                          labelKeys={labelKeys}
                          placeholder={variableExprPlaceholder}
                          ariaLabel={t("alertRule.variableExpr")}
                          onChange={(value) =>
                            updateVariable(index, { expr: value })
                          }
                        />
                      </FormField>
                    </div>
                  ))}
                </div>
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
      <AlertRuleDetailSheet
        detail={detail}
        groups={groups}
        sources={sources}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
            if (ruleUid) {
              setParam("ruleUid", "")
            }
          }
        }}
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
