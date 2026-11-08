import { Info, Plus, Trash2 } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { selectDatasources } from "@/api/datasource"
import {
  createInspectionPlan,
  updateInspectionPlan,
} from "@/api/inspection"
import type {
  InspectionItemDisplay,
  InspectionItemInput,
  InspectionItemKind,
  InspectionOutputFormat,
  InspectionPlan,
  InspectionPlanInput,
  InspectionTemplate,
} from "@/api/inspection/types"
import { FormField } from "@/components/form-field"
import { MultiSelect, type MultiSelectOption } from "@/components/multi-select"
import { StepForm } from "@/components/step-form"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
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
import { Textarea } from "@/components/ui/textarea"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { readFormError } from "@/lib/form-error"

const ITEM_KINDS: InspectionItemKind[] = [
  "METRICS_QUERY",
  "LOGS_QUERY",
  "PROBE_TARGET",
  "ALERT_EVENTS",
  "NOTE",
]
const DISPLAYS: InspectionItemDisplay[] = [
  "METRIC",
  "TABLE",
  "TREND",
  "LOGS",
  "TEXT",
]
const FORMATS: InspectionOutputFormat[] = ["HTML", "MARKDOWN"]

const CRON_PRESETS = [
  { value: "0 9 * * *", zh: "每天 09:00", en: "Daily 09:00" },
  { value: "0 8 * * *", zh: "每天 08:00", en: "Daily 08:00" },
  { value: "0 9 * * 1", zh: "每周一 09:00", en: "Monday 09:00" },
  { value: "0 9 1 * *", zh: "每月 1 日 09:00", en: "Monthly 1st 09:00" },
]
const CUSTOM_CRON = "custom"

interface DraftItem extends InspectionItemInput {
  key: string
}

interface InspectionPlanDialogProps {
  open: boolean
  plan: InspectionPlan | null
  templates: InspectionTemplate[]
  groups: MultiSelectOption[]
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

function emptyItem(): DraftItem {
  return {
    key: `item-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
    kind: "METRICS_QUERY",
    title: "",
    datasourceUid: "",
    query: "",
    windowSeconds: 3600,
    stepSeconds: 60,
    display: "TABLE",
    params: "{}",
    enabled: true,
  }
}

function draftFromPlan(plan: InspectionPlan | null): DraftItem[] {
  if (!plan || plan.items.length === 0) {
    return [emptyItem()]
  }
  return plan.items.map((item) => ({
    key: item.uid || `item-${Math.random().toString(36).slice(2, 8)}`,
    kind: item.kind,
    title: item.title,
    datasourceUid: item.datasourceUid,
    query: item.query,
    windowSeconds: item.windowSeconds,
    stepSeconds: item.stepSeconds,
    display: item.display,
    params: item.params || "{}",
    enabled: item.enabled,
  }))
}

/** 新建 / 编辑巡检计划：基础信息 → 检查项 → 报告 → 调度与通知。 */
export function InspectionPlanDialog({
  open,
  plan,
  templates,
  groups,
  onOpenChange,
  onSaved,
}: InspectionPlanDialogProps) {
  const { t, i18n } = useTranslation()
  const zh = i18n.language.startsWith("zh")
  const [step, setStep] = useState(0)
  const [name, setName] = useState(() => plan?.name ?? "")
  const [description, setDescription] = useState(() => plan?.description ?? "")
  const [timeoutSeconds, setTimeoutSeconds] = useState(
    () => plan?.timeoutSeconds || 300
  )
  const [items, setItems] = useState<DraftItem[]>(() => draftFromPlan(plan))
  const [expanded, setExpanded] = useState(0)
  const [templateUid, setTemplateUid] = useState(() => plan?.templateUid ?? "")
  const [formats, setFormats] = useState<InspectionOutputFormat[]>(
    () => plan?.outputFormats ?? ["HTML"]
  )
  const [titleTemplate, setTitleTemplate] = useState(
    () => plan?.titleTemplate ?? ""
  )
  const [cron, setCron] = useState(() => plan?.cron ?? CRON_PRESETS[0].value)
  const [cronPreset, setCronPreset] = useState(() => {
    const found = CRON_PRESETS.find((preset) => preset.value === plan?.cron)
    return found ? found.value : plan ? CUSTOM_CRON : CRON_PRESETS[0].value
  })
  const [timezone, setTimezone] = useState(
    () => plan?.timezone ?? "Asia/Shanghai"
  )
  const [groupUids, setGroupUids] = useState<string[]>(
    () => plan?.notificationGroupUids ?? []
  )
  const [notifyPolicy, setNotifyPolicy] = useState<"ALWAYS" | "ON_ABNORMAL">(
    () => plan?.notifyPolicy ?? "ON_ABNORMAL"
  )
  const [metricSources, setMetricSources] = useState<MultiSelectOption[]>([])
  const [logSources, setLogSources] = useState<MultiSelectOption[]>([])
  const [pending, setPending] = useState(false)
  const [formError, setFormError] = useState("")

  useEffect(() => {
    let active = true
    Promise.all([
      selectDatasources({ type: "METRICS", limit: 100 }),
      selectDatasources({ type: "LOGS", limit: 100 }),
    ])
      .then(([metrics, logs]) => {
        if (!active) {
          return
        }
        setMetricSources(metrics.items.map((item) => ({ value: item.uid, label: item.name })))
        setLogSources(logs.items.map((item) => ({ value: item.uid, label: item.name })))
      })
      .catch(() => {
        if (active) {
          setMetricSources([])
          setLogSources([])
        }
      })
    return () => {
      active = false
    }
  }, [])

  const effectiveTemplateUid = templateUid || templates[0]?.uid || ""

  function updateItem(index: number, patch: Partial<DraftItem>) {
    setItems((current) =>
      current.map((item, itemIndex) =>
        itemIndex === index ? { ...item, ...patch } : item
      )
    )
  }

  function selectPreset(value: string) {
    setCronPreset(value)
    if (value !== CUSTOM_CRON) {
      setCron(value)
    }
  }

  function validate(): string {
    if (!name.trim()) {
      return t("inspection.nameRequired")
    }
    if (items.length === 0) {
      return t("inspection.noItems")
    }
    if (!effectiveTemplateUid) {
      return t("inspection.template")
    }
    if (formats.length === 0) {
      return t("inspection.outputFormats")
    }
    return ""
  }

  async function save(event: React.FormEvent) {
    event.preventDefault()
    const error = validate()
    if (error) {
      setFormError(error)
      return
    }
    const payload: InspectionPlanInput = {
      name: name.trim(),
      description: description.trim(),
      cron,
      timezone: timezone.trim(),
      timeoutSeconds,
      templateUid: effectiveTemplateUid,
      outputFormats: formats,
      titleTemplate: titleTemplate.trim(),
      notificationGroupUids: groupUids,
      notifyPolicy,
      labels: plan?.labels ?? "{}",
      remark: plan?.remark ?? "",
      items: items.map((item, index) => ({
        kind: item.kind,
        title: item.title.trim() || `${t("inspection.items")} ${index + 1}`,
        datasourceUid: item.datasourceUid,
        query: item.query,
        windowSeconds: item.windowSeconds,
        stepSeconds: item.stepSeconds,
        display: item.display,
        params: item.params || "{}",
        enabled: item.enabled,
      })),
    }
    setPending(true)
    setFormError("")
    try {
      if (plan) {
        await updateInspectionPlan(plan.uid, payload)
      } else {
        await createInspectionPlan(payload)
      }
      onSaved()
    } catch (err) {
      setFormError(readFormError(err, t("state.failed")).form)
    } finally {
      setPending(false)
    }
  }

  function kindLabel(kind: InspectionItemKind): string {
    switch (kind) {
      case "METRICS_QUERY":
        return t("inspection.kindMetrics")
      case "LOGS_QUERY":
        return t("inspection.kindLogs")
      case "PROBE_TARGET":
        return t("inspection.kindProbe")
      case "ALERT_EVENTS":
        return t("inspection.kindAlert")
      default:
        return t("inspection.kindNote")
    }
  }

  function displayLabel(display: InspectionItemDisplay): string {
    switch (display) {
      case "METRIC":
        return t("inspection.displayMetric")
      case "TREND":
        return t("inspection.displayTrend")
      case "LOGS":
        return t("inspection.displayLogs")
      case "TEXT":
        return t("inspection.displayText")
      default:
        return t("inspection.displayTable")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {plan ? t("inspection.editPlan") : t("inspection.newPlan")}
          </DialogTitle>
        </DialogHeader>
        <StepForm
          step={step}
          stepCount={4}
          pending={pending}
          error={formError}
          onBack={() => {
            setFormError("")
            setStep((value) => Math.max(0, value - 1))
          }}
          onNext={() => {
            if (step === 0 && !name.trim()) {
              setFormError(t("inspection.nameRequired"))
              return
            }
            if (step === 1 && items.length === 0) {
              setFormError(t("inspection.noItems"))
              return
            }
            if (step === 2 && !effectiveTemplateUid) {
              setFormError(t("inspection.template"))
              return
            }
            setFormError("")
            setStep((value) => Math.min(3, value + 1))
          }}
          onSubmit={save}
        >
          {step === 0 ? (
            <>
              <FormField
                label={t("inspection.planName")}
                htmlFor="inspection-plan-name"
                required
              >
                <Input
                  id="inspection-plan-name"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder={zh ? "例如：核心业务每日巡检" : "e.g. Core business daily"}
                />
              </FormField>
              <FormField label={t("inspection.description")} htmlFor="inspection-plan-desc">
                <Input
                  id="inspection-plan-desc"
                  value={description}
                  onChange={(event) => setDescription(event.target.value)}
                />
              </FormField>
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField
                  label={t("inspection.timeout")}
                  htmlFor="inspection-plan-timeout"
                >
                  <Input
                    id="inspection-plan-timeout"
                    type="number"
                    min={30}
                    max={1800}
                    value={timeoutSeconds}
                    onChange={(event) =>
                      setTimeoutSeconds(Number(event.target.value) || 300)
                    }
                  />
                </FormField>
                <FormField label={t("inspection.timezone")} htmlFor="inspection-plan-tz">
                  <Input
                    id="inspection-plan-tz"
                    value={timezone}
                    onChange={(event) => setTimezone(event.target.value)}
                  />
                </FormField>
              </div>
            </>
          ) : null}

          {step === 1 ? (
            <div className="grid gap-3">
              {items.map((item, index) => {
                const queryKinds = item.kind === "METRICS_QUERY" || item.kind === "LOGS_QUERY"
                const sources = item.kind === "LOGS_QUERY" ? logSources : metricSources
                const open = expanded === index
                return (
                  <div
                    key={item.key}
                    className={`overflow-hidden rounded-lg border ${open ? "border-primary/40" : "border-border"}`}
                  >
                    <button
                      type="button"
                      className="flex w-full items-center gap-2 px-3 py-2 text-left hover:bg-muted/40"
                      aria-expanded={open}
                      onClick={() => setExpanded(open ? -1 : index)}
                    >
                      <span className="font-mono text-xs text-muted-foreground">
                        {String(index + 1).padStart(2, "0")}
                      </span>
                      <span className="min-w-0 flex-1 truncate text-sm font-medium">
                        {item.title || t("inspection.items")}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {kindLabel(item.kind)}
                      </span>
                    </button>
                    {open ? (
                      <div className="grid gap-3 border-t border-border px-3 py-3">
                        <div className="grid gap-3 sm:grid-cols-2">
                          <FormField label={t("inspection.itemTitle")}>
                            <Input
                              value={item.title}
                              onChange={(event) =>
                                updateItem(index, { title: event.target.value })
                              }
                            />
                          </FormField>
                          <FormField label={t("inspection.kind")}>
                            <Select
                              value={item.kind}
                              onValueChange={(value) =>
                                updateItem(index, {
                                  kind: value as InspectionItemKind,
                                  display: value === "NOTE" ? "TEXT" : item.display,
                                })
                              }
                            >
                              <SelectTrigger>
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                {ITEM_KINDS.map((kind) => (
                                  <SelectItem key={kind} value={kind}>
                                    {kindLabel(kind)}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </FormField>
                        </div>
                        {queryKinds ? (
                          <div className="grid gap-3 sm:grid-cols-2">
                            <FormField label={t("inspection.datasource")}>
                              <Select
                                value={item.datasourceUid}
                                onValueChange={(value) =>
                                  updateItem(index, { datasourceUid: value })
                                }
                              >
                                <SelectTrigger>
                                  <SelectValue placeholder={t("inspection.datasource")} />
                                </SelectTrigger>
                                <SelectContent>
                                  {sources.map((source) => (
                                    <SelectItem key={source.value} value={source.value}>
                                      {source.label}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                            </FormField>
                            <FormField label={t("inspection.window")}>
                              <Input
                                type="number"
                                min={60}
                                value={item.windowSeconds}
                                onChange={(event) =>
                                  updateItem(index, {
                                    windowSeconds: Number(event.target.value) || 3600,
                                  })
                                }
                              />
                            </FormField>
                          </div>
                        ) : null}
                        <FormField
                          label={
                            item.kind === "NOTE"
                              ? t("inspection.noteContent")
                              : t("inspection.query")
                          }
                        >
                          <Textarea
                            rows={3}
                            className="font-mono text-xs"
                            value={item.query}
                            onChange={(event) =>
                              updateItem(index, { query: event.target.value })
                            }
                          />
                        </FormField>
                        <div className="grid gap-3 sm:grid-cols-2">
                          <FormField label={t("inspection.display")}>
                            <Select
                              value={item.display}
                              onValueChange={(value) =>
                                updateItem(index, {
                                  display: value as InspectionItemDisplay,
                                })
                              }
                            >
                              <SelectTrigger>
                                <SelectValue />
                              </SelectTrigger>
                              <SelectContent>
                                {DISPLAYS.map((display) => (
                                  <SelectItem key={display} value={display}>
                                    {displayLabel(display)}
                                  </SelectItem>
                                ))}
                              </SelectContent>
                            </Select>
                          </FormField>
                          <div className="flex items-end justify-between gap-2 pb-2">
                            <label className="flex items-center gap-2 text-sm">
                              <Checkbox
                                checked={item.enabled}
                                onCheckedChange={(value) =>
                                  updateItem(index, { enabled: value === true })
                                }
                              />
                              {t("inspection.enabled")}
                            </label>
                            <Button
                              type="button"
                              size="sm"
                              variant="ghost"
                              className="text-destructive"
                              disabled={items.length <= 1}
                              onClick={() =>
                                setItems((current) =>
                                  current.filter((_, currentIndex) => currentIndex !== index)
                                )
                              }
                            >
                              <Trash2 data-icon="inline-start" />
                              {t("inspection.remove")}
                            </Button>
                          </div>
                        </div>
                      </div>
                    ) : null}
                  </div>
                )
              })}
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setItems((current) => [...current, emptyItem()])
                  setExpanded(items.length)
                }}
              >
                <Plus data-icon="inline-start" />
                {t("inspection.addItem")}
              </Button>
            </div>
          ) : null}

          {step === 2 ? (
            <>
              <FormField label={t("inspection.template")} required>
                <Select value={effectiveTemplateUid} onValueChange={setTemplateUid}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {templates.map((template) => (
                      <SelectItem key={template.uid} value={template.uid}>
                        {template.name}
                        {template.isBuiltin ? ` · ${t("inspection.builtin")}` : ""}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
              <FormField label={t("inspection.outputFormats")} required>
                <div className="flex flex-wrap gap-4">
                  {FORMATS.map((format) => (
                    <label key={format} className="flex items-center gap-2 text-sm">
                      <Checkbox
                        checked={formats.includes(format)}
                        onCheckedChange={(value) =>
                          setFormats((current) =>
                            value === true
                              ? [...new Set([...current, format])]
                              : current.filter((item) => item !== format)
                          )
                        }
                      />
                      {format}
                    </label>
                  ))}
                </div>
              </FormField>
              <FormField
                label={t("inspection.titleTemplate")}
                hint={zh ? "留空使用默认标题。" : "Leave empty for the default title."}
              >
                <Input
                  className="font-mono text-xs"
                  value={titleTemplate}
                  onChange={(event) => setTitleTemplate(event.target.value)}
                  placeholder={`{{.Plan.Name}} · {{.Run.FinishedAt | date "2006-01-02 15:04"}}`}
                />
              </FormField>
            </>
          ) : null}

          {step === 3 ? (
            <>
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField label={t("inspection.cronPreset")}>
                  <Select value={cronPreset} onValueChange={selectPreset}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {CRON_PRESETS.map((preset) => (
                        <SelectItem key={preset.value} value={preset.value}>
                          {zh ? preset.zh : preset.en}
                        </SelectItem>
                      ))}
                      <SelectItem value={CUSTOM_CRON}>
                        {t("inspection.customCron")}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField label={t("inspection.timezone")}>
                  <Input
                    value={timezone}
                    onChange={(event) => setTimezone(event.target.value)}
                  />
                </FormField>
              </div>
              <FormField
                label={t("inspection.cron")}
                actions={
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <button type="button" aria-label={t("inspection.cronHint")}>
                        <Info className="size-3.5 text-muted-foreground" />
                      </button>
                    </TooltipTrigger>
                    <TooltipContent className="max-w-xs">
                      {t("inspection.cronHint")}
                    </TooltipContent>
                  </Tooltip>
                }
                hint={t("inspection.cronHint")}
              >
                <Input
                  className={`font-mono text-xs ${cronPreset !== CUSTOM_CRON ? "bg-muted text-muted-foreground" : ""}`}
                  value={cron}
                  readOnly={cronPreset !== CUSTOM_CRON}
                  onChange={(event) => setCron(event.target.value)}
                />
              </FormField>
              <FormField label={t("inspection.groups")}>
                <MultiSelect
                  values={groupUids}
                  options={groups}
                  onChange={setGroupUids}
                  placeholder={t("inspection.groups")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("multiSelect.placeholder")}
                  allLabel={t("multiSelect.all")}
                />
              </FormField>
              <FormField label={t("inspection.notifyPolicy")}>
                <Select
                  value={notifyPolicy}
                  onValueChange={(value) =>
                    setNotifyPolicy(value as "ALWAYS" | "ON_ABNORMAL")
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="ALWAYS">{t("inspection.notifyAlways")}</SelectItem>
                    <SelectItem value="ON_ABNORMAL">
                      {t("inspection.notifyAbnormal")}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </FormField>
            </>
          ) : null}
        </StepForm>
      </DialogContent>
    </Dialog>
  )
}
