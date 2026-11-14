import { Info, Plus } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { listAlertDictionaries } from "@/api/alert-dictionary"
import { listDatasources } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import {
  createInspectionPlan,
  updateInspectionPlan,
} from "@/api/inspection"
import { listProbeTargetGroups } from "@/api/probe-target"
import type {
  InspectionItemKind,
  InspectionPlan,
  InspectionPlanInput,
  InspectionTemplate,
} from "@/api/inspection/types"
import {
  INSPECTION_ITEM_KINDS,
  InspectionItemEditor,
  defaultInspectionDisplay,
  defaultInspectionParams,
  inspectionKindLabel,
  type DraftItem,
} from "@/pages/inspection-plans/inspection-item-editor"
import { FormField } from "@/components/form-field"
import { MultiSelect, type MultiSelectOption } from "@/components/multi-select"
import { StatusBadge } from "@/components/status-badge"
import { StepForm } from "@/components/step-form"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { readFormError } from "@/lib/form-error"


const CRON_PRESETS = [
  { value: "0 9 * * *", zh: "每天 09:00", en: "Daily 09:00" },
  { value: "0 8 * * *", zh: "每天 08:00", en: "Daily 08:00" },
  { value: "0 9 * * 1", zh: "每周一 09:00", en: "Monday 09:00" },
  { value: "0 9 1 * *", zh: "每月 1 日 09:00", en: "Monthly 1st 09:00" },
]
const CUSTOM_CRON = "custom"

interface InspectionPlanDialogProps {
  open: boolean
  plan: InspectionPlan | null
  templates: InspectionTemplate[]
  groups: MultiSelectOption[]
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

function draftFromPlan(plan: InspectionPlan | null): DraftItem[] {
  if (!plan || plan.items.length === 0) {
    return []
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
  const [expanded, setExpanded] = useState(plan ? 0 : -1)
  const [choosingType, setChoosingType] = useState(false)
  const [focusKey, setFocusKey] = useState<string | null>(null)
  const nextItemKey = useRef(0)
  const [templateUid, setTemplateUid] = useState(() => plan?.templateUid ?? "")
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
  const [metricSources, setMetricSources] = useState<DatasourceItem[]>([])
  const [logSources, setLogSources] = useState<DatasourceItem[]>([])
  const [probeGroups, setProbeGroups] = useState<MultiSelectOption[]>([])
  const [alertLevels, setAlertLevels] = useState<MultiSelectOption[]>([])
  const [savedItems, setSavedItems] = useState<DraftItem[]>(() => draftFromPlan(plan))
  const [dirtyIndex, setDirtyIndex] = useState<number | null>(null)
  const [confirmAction, setConfirmAction] = useState<(() => void) | null>(null)
  const [pending, setPending] = useState(false)
  const [formError, setFormError] = useState("")

  useEffect(() => {
    let active = true
    Promise.all([
      listDatasources({ page: 1, pageSize: 100, type: "METRICS" }),
      listDatasources({ page: 1, pageSize: 100, type: "LOGS" }),
      listProbeTargetGroups({ page: 1, pageSize: 100 }),
      listAlertDictionaries({ page: 1, pageSize: 100, type: "ALERT" }),
    ])
      .then(([metrics, logs, groups, levels]) => {
        if (!active) {
          return
        }
        setMetricSources(metrics.items.filter((item) => item.status === "ACTIVE"))
        setLogSources(logs.items.filter((item) => item.status === "ACTIVE"))
        setProbeGroups(groups.items.map((item) => ({ value: item.uid, label: item.name })))
        setAlertLevels(levels.items.map((item) => ({ value: item.uid, label: item.name })))
      })
      .catch(() => {
        if (active) {
          setMetricSources([])
          setLogSources([])
          setProbeGroups([])
          setAlertLevels([])
        }
      })
    return () => {
      active = false
    }
  }, [])

  const effectiveTemplateUid = templateUid || templates[0]?.uid || ""
  const selectedTemplate = templates.find((item) => item.uid === effectiveTemplateUid)

  function createDraftItem(kind: InspectionItemKind): DraftItem {
    const source =
      kind === "METRICS_QUERY"
        ? metricSources[0]?.uid ?? ""
        : kind === "LOGS_QUERY"
          ? logSources[0]?.uid ?? ""
          : ""
    const params =
      kind === "ALERT_EVENTS" && alertLevels[0]
        ? JSON.stringify({
            status: "FIRING",
            level_uids: [],
            labels: "",
            limit: 50,
            warn_levels: [alertLevels[0].label],
          })
        : defaultInspectionParams(kind)
    return {
      key: `item-${kind}-${nextItemKey.current++}`,
      kind,
      title: "",
      datasourceUid: source,
      query: "",
      windowSeconds: 86400,
      stepSeconds: 60,
      display: defaultInspectionDisplay(kind),
      params,
      enabled: true,
    }
  }

  function addItem(kind: InspectionItemKind) {
    const created = createDraftItem(kind)
    setItems((current) => [...current, created])
    setSavedItems((current) => [...current, JSON.parse(JSON.stringify(created))])
    setExpanded(items.length)
    setDirtyIndex(null)
    setFocusKey(created.key)
    setChoosingType(false)
  }

  function kindHint(kind: InspectionItemKind): string {
    switch (kind) {
      case "METRICS_QUERY":
        return t("inspection.kindMetricsHint")
      case "LOGS_QUERY":
        return t("inspection.kindLogsHint")
      case "PROBE_TARGET":
        return t("inspection.kindProbeHint")
      case "ALERT_EVENTS":
        return t("inspection.kindAlertHint")
      default:
        return t("inspection.kindNoteHint")
    }
  }

  function updateItem(index: number, patch: Partial<DraftItem>) {
    setItems((current) =>
      current.map((item, itemIndex) =>
        itemIndex === index ? { ...item, ...patch } : item
      )
    )
    setDirtyIndex(index)
  }

  function saveItem(index: number) {
    const item = items[index]
    if (!item.title.trim()) {
      toast.error(t("inspection.itemTitle"))
      return
    }
    if ((item.kind === "METRICS_QUERY" || item.kind === "LOGS_QUERY") && (!item.datasourceUid || !item.query.trim())) {
      toast.error(t("inspection.query"))
      return
    }
    if (item.kind === "PROBE_TARGET") {
      const params = JSON.parse(item.params || "{}") as { group_uids?: string[] }
      if (!params.group_uids || params.group_uids.length === 0) {
        toast.error(t("inspection.probeGroups"))
        return
      }
    }
    setSavedItems((current) =>
      current.map((item, itemIndex) =>
        itemIndex === index ? JSON.parse(JSON.stringify(items[index])) : item
      )
    )
    setDirtyIndex(null)
    toast.success(t("inspection.itemSavedToast"))
  }

  function restoreItem(index: number) {
    const snapshot = savedItems[index]
    if (!snapshot) {
      return
    }
    setItems((current) =>
      current.map((item, itemIndex) =>
        itemIndex === index ? JSON.parse(JSON.stringify(snapshot)) : item
      )
    )
    setDirtyIndex(null)
  }

  function openItem(index: number) {
    const apply = () => {
      setExpanded(expanded === index ? -1 : index)
      setDirtyIndex(null)
      setFocusKey(null)
    }
    if (dirtyIndex !== null && (expanded === index || dirtyIndex !== index)) {
      setConfirmAction(() => apply)
      return
    }
    apply()
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
      outputFormats: selectedTemplate ? [selectedTemplate.format] : [],
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

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-5xl">
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
                  placeholder={t("inspection.descriptionPlaceholder")}
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
                    placeholder={t("inspection.timeoutPlaceholder")}
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
                    placeholder={t("inspection.timezonePlaceholder")}
                  />
                </FormField>
              </div>
            </>
          ) : null}

          {step === 1 ? (
            <div className="grid gap-3">
              {items.map((item, index) => {
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
                      onClick={() => openItem(index)}
                    >
                      <span className="font-mono text-xs text-muted-foreground">
                        {String(index + 1).padStart(2, "0")}
                      </span>
                      <span className="min-w-0 flex-1 truncate text-sm font-medium">
                        {item.title || t("inspection.items")}
                      </span>
                      {dirtyIndex === index ? (
                        <span className="text-xs text-warning">{t("inspection.itemDirty")}</span>
                      ) : null}
                      <span className="text-xs text-muted-foreground">{inspectionKindLabel(t, item.kind)}</span>
                    </button>
                    {open ? (
                      <div className="border-t border-border p-3">
                        <InspectionItemEditor
                          key={`${item.key}-${item.kind}`}
                          item={item}
                          sources={{ metrics: metricSources, logs: logSources }}
                          probeGroups={probeGroups}
                          alertLevels={alertLevels}
                          dirty={dirtyIndex === index}
                          autoFocus={focusKey === item.key}
                          onChange={(patch) => updateItem(index, patch)}
                          onSave={() => saveItem(index)}
                          onRestore={() => restoreItem(index)}
                          onRemove={() => {
                            setItems((current) => current.filter((_, currentIndex) => currentIndex !== index))
                            setSavedItems((current) => current.filter((_, currentIndex) => currentIndex !== index))
                            setExpanded(-1)
                            setDirtyIndex(null)
                            setFocusKey(null)
                          }}
                        />
                      </div>
                    ) : null}
                  </div>
                )
              })}
              {choosingType ? (
                <div className="grid gap-3 rounded-lg border border-primary/40 bg-muted/20 p-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="grid gap-1">
                      <p className="text-sm font-medium">{t("inspection.chooseItemType")}</p>
                      <p className="text-xs text-muted-foreground">{t("inspection.chooseItemTypeHint")}</p>
                    </div>
                    <Button
                      type="button"
                      size="sm"
                      variant="ghost"
                      onClick={() => setChoosingType(false)}
                    >
                      {t("action.cancel")}
                    </Button>
                  </div>
                  <div className="grid gap-2 sm:grid-cols-2">
                    {INSPECTION_ITEM_KINDS.map((kind) => (
                      <button
                        key={kind}
                        type="button"
                        className="grid gap-1 rounded-lg border border-border bg-background px-3 py-2 text-left transition hover:border-primary/50 hover:bg-muted/40 focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none"
                        onClick={() => addItem(kind)}
                      >
                        <span className="text-sm font-medium">{inspectionKindLabel(t, kind)}</span>
                        <span className="text-xs leading-relaxed text-muted-foreground">
                          {kindHint(kind)}
                        </span>
                      </button>
                    ))}
                  </div>
                </div>
              ) : (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setChoosingType(true)}
                >
                  <Plus data-icon="inline-start" />
                  {t("inspection.addItem")}
                </Button>
              )}
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
                        {" · "}
                        {template.format === "MARKDOWN" ? "Markdown" : "HTML"}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
              <FormField label={t("inspection.templateFormat")}>
                <div className="flex items-center gap-2 text-sm">
                  <StatusBadge
                    value={selectedTemplate?.format ?? "HTML"}
                    label={selectedTemplate?.format === "MARKDOWN" ? "Markdown" : "HTML"}
                    tone={selectedTemplate?.format === "MARKDOWN" ? "violet" : "info"}
                  />
                  <span className="text-xs text-muted-foreground">
                    {t("inspection.templateFormatHint")}
                  </span>
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
                    placeholder={t("inspection.timezonePlaceholder")}
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
                  placeholder={t("inspection.cronPlaceholder")}
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
      <AlertDialog open={confirmAction !== null} onOpenChange={(next) => { if (!next) setConfirmAction(null) }}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t("inspection.itemDirty")}</AlertDialogTitle>
            <AlertDialogDescription>{t("inspection.discardChanges")}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t("action.cancel")}</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                const action = confirmAction
                setConfirmAction(null)
                action?.()
              }}
            >
              {t("inspection.discardChangesAction")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Dialog>
  )
}
