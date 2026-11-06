import { GripVertical, Plus, Trash2 } from "lucide-react"
import { Fragment, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  createAlertEscalationPolicy,
  updateAlertEscalationPolicy,
} from "@/api/alert-escalation"
import type {
  AlertEscalationPolicyItem,
  AlertEscalationStepInput,
  SaveAlertEscalationPolicyParams,
} from "@/api/alert-escalation/types"
import { FormField } from "@/components/form-field"
import { MultiSelect, type MultiSelectOption } from "@/components/multi-select"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { usePointerReorder } from "@/hooks/use-flip-reorder"
import { cn } from "@/lib/class-names"
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

// 与后端 proto / biz 的校验上限保持一致；前端只做即时提示，服务端仍是最终判定。
const NAME_MAX = 64
const DESCRIPTION_MAX = 1024
const MAX_STEPS = 20
const MAX_GROUPS_PER_STEP = 20
const MAX_DELAY_MINUTES = 10080

/** 步骤草稿：延迟用字符串保存，允许输入过程中为空。 */
interface StepDraft {
  /** 仅用于 React key；不提交。 */
  key: string
  /** 编辑已有步骤时由服务端返回；空表示新增。 */
  uid?: string
  delay: string
  groupUids: string[]
}

interface LabelPairErrors {
  key?: string
  value?: string
}

let stepKeySeed = 0

function nextStepKey(): string {
  stepKeySeed += 1
  return `escalation-step-${stepKeySeed}`
}

function emptyStep(delay: number): StepDraft {
  return { key: nextStepKey(), delay: String(delay), groupUids: [] }
}

function toStepDraft(
  step: AlertEscalationPolicyItem["steps"][number]
): StepDraft {
  return {
    key: nextStepKey(),
    uid: step.uid,
    delay: String(step.delayMinutes),
    groupUids: step.notificationGroupUids ?? [],
  }
}

/** 整数分钟且在闭区间内；空串、小数、越界都算非法。 */
function parseDelay(value: string): number | null {
  const trimmed = value.trim()
  if (!/^\d+$/.test(trimmed)) {
    return null
  }
  const parsed = Number(trimmed)
  return parsed <= MAX_DELAY_MINUTES ? parsed : null
}

interface AlertEscalationPolicyDialogProps {
  open: boolean
  /** null 表示新建；否则编辑该策略。 */
  policy: AlertEscalationPolicyItem | null
  levels: MultiSelectOption[]
  datasources: MultiSelectOption[]
  rules: MultiSelectOption[]
  groups: MultiSelectOption[]
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

/**
 * 新建 / 编辑升级策略。父组件每次打开换 key 重新挂载，初始值直接来自当前编辑对象。
 * 升级步骤顺序即执行顺序，延迟是距事件触发时间的累计分钟数。
 */
export function AlertEscalationPolicyDialog({
  open,
  policy,
  levels,
  datasources,
  rules,
  groups,
  onOpenChange,
  onSaved,
}: AlertEscalationPolicyDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => policy?.name ?? "")
  const [description, setDescription] = useState(
    () => policy?.description ?? ""
  )
  const [labels, setLabels] = useState<LabelPair[]>(() =>
    parseLabelPairs(policy?.labels ?? "{}")
  )
  const [labelPairErrors, setLabelPairErrors] = useState<LabelPairErrors[]>([])
  const [levelUids, setLevelUids] = useState<string[]>(
    () => policy?.levelUids ?? []
  )
  const [datasourceUids, setDatasourceUids] = useState<string[]>(
    () => policy?.datasourceUids ?? []
  )
  const [ruleUids, setRuleUids] = useState<string[]>(
    () => policy?.ruleUids ?? []
  )
  const [steps, setSteps] = useState<StepDraft[]>(() =>
    policy
      ? [...policy.steps]
          .sort((left, right) => left.stepOrder - right.stepOrder)
          .map(toStepDraft)
      : [emptyStep(5)]
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)
  /** 步骤列表尾部的 ref：新增后把「添加步骤」留在视野内，方便连续添加。 */
  const stepsTailRef = useRef<HTMLDivElement>(null)
  const keepTailVisible = useRef(false)
  // 与实时告警自定义列一致的拖动排序：拖动源脱离列表流，目标位置只显示空占位。
  const {
    containerRef: stepsRef,
    dragKey,
    dropTarget,
    dropPosition,
    dragSize,
    startDrag,
  } = usePointerReorder<HTMLDivElement>(
    steps.map((step) => step.key).join("|"),
    "y",
    reorderSteps
  )

  useEffect(() => {
    if (!keepTailVisible.current) {
      return
    }
    keepTailVisible.current = false
    stepsTailRef.current?.scrollIntoView({
      block: "nearest",
      behavior: "smooth",
    })
  }, [steps.length])

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  function patchStep(index: number, patch: Partial<StepDraft>) {
    setSteps((current) =>
      current.map((step, position) =>
        position === index ? { ...step, ...patch } : step
      )
    )
  }

  function reorderSteps(
    sourceKey: string,
    targetKey: string,
    position: "before" | "after"
  ) {
    setSteps((current) => {
      const from = current.findIndex((step) => step.key === sourceKey)
      const target = current.findIndex((step) => step.key === targetKey)
      if (from < 0 || target < 0 || sourceKey === targetKey) {
        return current
      }
      const next = [...current]
      const [moved] = next.splice(from, 1)
      let insertionIndex = target + (position === "after" ? 1 : 0)
      if (from < insertionIndex) {
        insertionIndex -= 1
      }
      next.splice(insertionIndex, 0, moved)
      return next
    })
  }

  function removeStep(index: number) {
    setSteps((current) => current.filter((_, position) => position !== index))
  }

  function addStep() {
    keepTailVisible.current = true
    setSteps((current) => {
      const last = parseDelay(current[current.length - 1]?.delay ?? "")
      const nextDelay = last === null ? 5 : last + 5
      return [...current, emptyStep(nextDelay)]
    })
  }

  /** 标签行逐项校验：键不能有值无键、不能重复，值必须是合法正则。 */
  function validateLabelPairs(pairs: LabelPair[]): LabelPairErrors[] {
    const seen = new Set<string>()
    return pairs.map((pair) => {
      const key = pair.key.trim()
      const value = pair.value.trim()
      const next: LabelPairErrors = {}
      if (key && value) {
        try {
          new RegExp(`^(?:${value})$`)
        } catch {
          next.value = t("escalation.errors.labelValueInvalid")
        }
      }
      if (key) {
        if (seen.has(key)) {
          next.key = t("escalation.errors.labelDuplicateKey")
        }
        seen.add(key)
      } else if (value) {
        next.key = t("escalation.errors.labelKeyRequired")
      }
      return next
    })
  }

  function validate(): { errors: FieldErrors; pairErrors: LabelPairErrors[] } {
    const pairErrors = validateLabelPairs(labels)
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("escalation.errors.nameRequired"))],
      ["name", maxLength(name, NAME_MAX, t("escalation.errors.nameTooLong"))],
      [
        "description",
        maxLength(
          description,
          DESCRIPTION_MAX,
          t("escalation.errors.descriptionTooLong")
        ),
      ],
      [
        "steps",
        steps.length === 0
          ? t("escalation.errors.stepsRequired")
          : steps.length > MAX_STEPS
            ? t("escalation.errors.stepsTooMany")
            : null,
      ],
    ])
    if (pairErrors.some((item) => item.key || item.value)) {
      errors.labels = t("escalation.errors.labelsInvalid")
    }
    const delays: number[] = []
    let allDelaysValid = true
    steps.forEach((step, index) => {
      const delay = parseDelay(step.delay)
      if (delay === null) {
        allDelaysValid = false
        errors[`step.${index}.delay`] = t("escalation.errors.delayInvalid")
      } else {
        delays.push(delay)
      }
      if (step.groupUids.length === 0) {
        errors[`step.${index}.groups`] = t("escalation.errors.groupsRequired")
      } else if (step.groupUids.length > MAX_GROUPS_PER_STEP) {
        errors[`step.${index}.groups`] = t("escalation.errors.groupsTooMany")
      }
    })
    if (!errors.steps && allDelaysValid && delays.length > 0) {
      const increasing = delays.every(
        (delay, index) => index === 0 || delay > delays[index - 1]
      )
      if (!increasing) {
        errors.steps = t("escalation.errors.delaysNotIncreasing")
      } else if (delays[delays.length - 1] === 0) {
        errors.steps = t("escalation.errors.lastStepDelay")
      }
    }
    return { errors, pairErrors }
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const { errors, pairErrors } = validate()
    setLabelPairErrors(pairErrors)
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    const payload: SaveAlertEscalationPolicyParams = {
      name: name.trim(),
      description: description.trim(),
      labels: serializeLabelPairs(labels),
      levelUids,
      datasourceUids,
      ruleUids,
      steps: steps.map<AlertEscalationStepInput>((step) => ({
        ...(step.uid ? { uid: step.uid } : {}),
        delayMinutes: Number(step.delay.trim()),
        notificationGroupUids: step.groupUids,
      })),
    }
    setPending(true)
    setFormError("")
    setFieldErrors({})
    try {
      if (policy) {
        await updateAlertEscalationPolicy({ ...payload, uid: policy.uid })
      } else {
        await createAlertEscalationPolicy(payload)
      }
      onSaved()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setFormError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {policy ? t("escalation.editTitle") : t("escalation.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSave}
        >
          <DialogBody className="grid gap-4">
            <FormField
              label={t("escalation.name")}
              htmlFor="escalation-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="escalation-name"
                value={name}
                maxLength={NAME_MAX}
                placeholder={t("escalation.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("escalation.descriptionLabel")}
              htmlFor="escalation-description"
              error={fieldErrors.description}
            >
              <Textarea
                id="escalation-description"
                value={description}
                className="min-h-16"
                placeholder={t("escalation.descriptionPlaceholder")}
                onChange={(event) => setDescription(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("escalation.labels")}
              hint={t("escalation.labelsHint")}
              error={fieldErrors.labels}
            >
              <div className="grid gap-2">
                {labels.map((pair, index) => (
                  <div key={index} className="grid gap-1">
                    <div className="flex items-center gap-2">
                      <Input
                        value={pair.key}
                        placeholder={t("escalation.labelKey")}
                        aria-label={t("escalation.labelKey")}
                        onChange={(event) => {
                          const next = [...labels]
                          next[index] = { ...pair, key: event.target.value }
                          setLabels(next)
                          setLabelPairErrors([])
                        }}
                      />
                      <Input
                        value={pair.value}
                        placeholder={t("escalation.labelValue")}
                        aria-label={t("escalation.labelValue")}
                        onChange={(event) => {
                          const next = [...labels]
                          next[index] = { ...pair, value: event.target.value }
                          setLabels(next)
                          setLabelPairErrors([])
                        }}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() =>
                          setLabels(
                            labels.filter((_, current) => current !== index)
                          )
                        }
                      >
                        {t("action.delete")}
                      </Button>
                    </div>
                    {labelPairErrors[index]?.key ? (
                      <p role="alert" className="text-xs text-destructive">
                        {labelPairErrors[index].key}
                      </p>
                    ) : null}
                    {labelPairErrors[index]?.value ? (
                      <p role="alert" className="text-xs text-destructive">
                        {labelPairErrors[index].value}
                      </p>
                    ) : null}
                  </div>
                ))}
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setLabels([...labels, { key: "", value: "" }])}
                >
                  {t("escalation.addLabel")}
                </Button>
              </div>
            </FormField>
            <FormField
              label={t("escalation.levels")}
              htmlFor="escalation-levels"
              hint={t("escalation.levelsHint")}
            >
              <MultiSelect
                id="escalation-levels"
                values={levelUids}
                options={levels}
                onChange={setLevelUids}
                placeholder={t("multiSelect.placeholder")}
                searchPlaceholder={t("multiSelect.searchPlaceholder")}
                emptyText={t("escalation.noLevels")}
                allLabel={t("multiSelect.all")}
                clearLabel={t("multiSelect.clear")}
                selectedLabel={selectionSummary}
              />
            </FormField>
            <FormField
              label={t("escalation.datasources")}
              htmlFor="escalation-datasources"
              hint={t("escalation.datasourcesHint")}
            >
              <MultiSelect
                id="escalation-datasources"
                values={datasourceUids}
                options={datasources}
                onChange={setDatasourceUids}
                placeholder={t("multiSelect.placeholder")}
                searchPlaceholder={t("multiSelect.searchPlaceholder")}
                emptyText={t("escalation.noDatasources")}
                allLabel={t("multiSelect.all")}
                clearLabel={t("multiSelect.clear")}
                selectedLabel={selectionSummary}
              />
            </FormField>
            <FormField
              label={t("escalation.rules")}
              htmlFor="escalation-rules"
              hint={t("escalation.rulesHint")}
            >
              <MultiSelect
                id="escalation-rules"
                values={ruleUids}
                options={rules}
                onChange={setRuleUids}
                placeholder={t("multiSelect.placeholder")}
                searchPlaceholder={t("multiSelect.searchPlaceholder")}
                emptyText={t("escalation.noRules")}
                allLabel={t("multiSelect.all")}
                clearLabel={t("multiSelect.clear")}
                selectedLabel={selectionSummary}
              />
            </FormField>
            <div
              ref={stepsRef}
              className={cn(
                "relative grid content-start gap-3",
                dragKey && "select-none"
              )}
            >
              <div className="grid gap-1">
                <p className="text-sm font-medium">{t("escalation.steps")}</p>
                <p className="text-xs text-muted-foreground">
                  {t("escalation.stepsHint")}
                </p>
              </div>
              {fieldErrors.steps ? (
                <p role="alert" className="text-sm text-destructive">
                  {fieldErrors.steps}
                </p>
              ) : null}
              {steps.map((step, index) => {
                const row = (
                  <div
                    key={step.key}
                    data-flip-id={dragKey === step.key ? undefined : step.key}
                    data-sort-key={step.key}
                    className={cn(
                      "grid gap-3 rounded-lg border border-border/70 bg-muted/20 p-3",
                      dropTarget === step.key &&
                        dragKey !== step.key &&
                        "border-primary/60 ring-1 ring-primary/30",
                      dragKey === step.key &&
                        "pointer-events-none absolute opacity-0"
                    )}
                  >
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <div className="flex min-w-0 items-center gap-1.5">
                        <GripVertical
                          className="size-4 shrink-0 cursor-grab text-muted-foreground active:cursor-grabbing"
                          aria-hidden
                          style={{ touchAction: "none" }}
                          onPointerDown={(event) => startDrag(step.key, event)}
                        />
                        <span className="text-xs font-medium text-muted-foreground">
                          {t("escalation.stepIndex", { index: index + 1 })}
                        </span>
                      </div>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon-sm"
                        aria-label={t("escalation.removeStep")}
                        title={t("escalation.removeStep")}
                        disabled={steps.length <= 1}
                        onClick={() => removeStep(index)}
                      >
                        <Trash2 aria-hidden="true" />
                      </Button>
                    </div>
                    <div className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
                      <FormField
                        label={t("escalation.delay")}
                        htmlFor={`escalation-delay-${step.key}`}
                        required
                        error={fieldErrors[`step.${index}.delay`]}
                      >
                        <Input
                          id={`escalation-delay-${step.key}`}
                          type="number"
                          inputMode="numeric"
                          min={0}
                          max={MAX_DELAY_MINUTES}
                          value={step.delay}
                          onChange={(event) =>
                            patchStep(index, { delay: event.target.value })
                          }
                        />
                      </FormField>
                      <FormField
                        label={t("escalation.groups")}
                        htmlFor={`escalation-groups-${step.key}`}
                        required
                        error={fieldErrors[`step.${index}.groups`]}
                      >
                        <MultiSelect
                          id={`escalation-groups-${step.key}`}
                          values={step.groupUids}
                          options={groups}
                          onChange={(values) =>
                            patchStep(index, { groupUids: values })
                          }
                          placeholder={t("multiSelect.placeholder")}
                          searchPlaceholder={t("multiSelect.searchPlaceholder")}
                          emptyText={t("escalation.noGroups")}
                          allLabel={t("multiSelect.all")}
                          clearLabel={t("multiSelect.clear")}
                          selectedLabel={selectionSummary}
                          invalid={Boolean(fieldErrors[`step.${index}.groups`])}
                        />
                      </FormField>
                    </div>
                  </div>
                )
                const showDropSlot =
                  dropTarget === step.key && dragKey !== step.key
                const dropSlot = showDropSlot ? (
                  <div
                    key={`drop-slot:${step.key}`}
                    data-drop-slot
                    data-drop-target={dropTarget}
                    data-drop-position={dropPosition}
                    aria-hidden="true"
                    className="w-full rounded-lg border border-dashed border-primary/60 bg-primary/5"
                    style={{ height: dragSize.height || undefined }}
                  />
                ) : null

                return (
                  <Fragment key={step.key}>
                    {showDropSlot && dropPosition === "before"
                      ? dropSlot
                      : null}
                    {row}
                    {showDropSlot && dropPosition === "after" ? dropSlot : null}
                  </Fragment>
                )
              })}
              {/* 添加按钮放在步骤列表末尾：新增的步骤就在它上方，不用滚回顶部再滚回来。 */}
              <div ref={stepsTailRef} className="pt-1">
                <Button
                  type="button"
                  variant="outline"
                  className="w-full"
                  disabled={pending || steps.length >= MAX_STEPS}
                  onClick={addStep}
                >
                  <Plus aria-hidden="true" data-icon="inline-start" />
                  {t("escalation.addStep")}
                </Button>
              </div>
            </div>
          </DialogBody>
          {formError ? (
            <p className="text-sm text-destructive">{formError}</p>
          ) : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? t("action.saving") : t("action.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
