import { ArrowDown, ArrowUp, Trash2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"

import {
  createOnCallRotation,
  updateOnCallRotation,
} from "@/api/oncall-rotation"
import type {
  OnCallRotationItem,
  OnCallRotationMemberInput,
  SaveOnCallRotationParams,
} from "@/api/oncall-rotation/types"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import {
  collectFieldErrors,
  hasFieldErrors,
  maxLength,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { readFormError } from "@/lib/form-error"

// 与后端 proto / biz 的校验上限保持一致；前端只做即时提示，服务端仍是最终判定。
const NAME_MAX = 64
const REMARK_MAX = 1024
const MAX_MEMBERS = 50
const DEFAULT_TIMEZONE =
  Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai"

// 常见 IANA 时区，仅作为输入建议；后端仍会校验时区是否合法。
const TIMEZONE_OPTIONS = [
  "UTC",
  "Asia/Shanghai",
  "Asia/Tokyo",
  "Asia/Singapore",
  "Asia/Hong_Kong",
  "Asia/Seoul",
  "Asia/Kolkata",
  "Asia/Dubai",
  "Europe/London",
  "Europe/Paris",
  "Europe/Berlin",
  "Europe/Moscow",
  "America/New_York",
  "America/Chicago",
  "America/Denver",
  "America/Los_Angeles",
  "America/Sao_Paulo",
  "Australia/Sydney",
]

/** RFC3339 → datetime-local 输入值（本地时区，精确到分钟）。 */
function toDateTimeInput(value: string): string {
  if (!value) {
    return ""
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return ""
  }
  const pad = (part: number) => String(part).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate()
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

/** datetime-local 输入值按本地时区解释，提交前转成 RFC3339。 */
function toRFC3339(value: string): string {
  if (!value) {
    return ""
  }
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? "" : date.toISOString()
}

function isValidTimezone(name: string): boolean {
  const trimmed = name.trim()
  if (!trimmed) {
    return false
  }
  try {
    new Intl.DateTimeFormat(undefined, { timeZone: trimmed })
    return true
  } catch {
    return false
  }
}

interface RotationDialogProps {
  open: boolean
  /** null 表示新建；否则编辑该轮换。 */
  rotation: OnCallRotationItem | null
  /** 可选的租户成员（已过滤到 JOINED）。 */
  members: MultiSelectOption[]
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

/**
 * 新建 / 编辑值班轮换。父组件每次打开换 key 重新挂载，初始值直接来自当前编辑对象。
 * 成员顺序即轮转顺序，服务端按数组下标重写 position。
 */
export function RotationDialog({
  open,
  rotation,
  members,
  onOpenChange,
  onSaved,
}: RotationDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => rotation?.name ?? "")
  const [remark, setRemark] = useState(() => rotation?.remark ?? "")
  const [timezone, setTimezone] = useState(
    () => rotation?.timezone || DEFAULT_TIMEZONE
  )
  const [period, setPeriod] = useState(() => rotation?.period || "WEEKLY")
  const [startsAt, setStartsAt] = useState(() =>
    rotation?.startsAt
      ? toDateTimeInput(rotation.startsAt)
      : toDateTimeInput(new Date().toISOString())
  )
  const [handoffHour, setHandoffHour] = useState(() =>
    String(rotation?.handoffHour ?? 0)
  )
  const [memberUids, setMemberUids] = useState<string[]>(() =>
    rotation
      ? [...rotation.members]
          .sort((left, right) => left.position - right.position)
          .map((member) => member.memberUid)
      : []
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)

  function selectionSummary(selected: number, total: number): string {
    return selected >= total && total > 0
      ? t("multiSelect.allSelected", { count: selected })
      : t("multiSelect.summary", { count: selected, total })
  }

  function memberLabel(uid: string): string {
    return members.find((member) => member.value === uid)?.label ?? uid
  }

  function memberDescription(uid: string): string {
    return members.find((member) => member.value === uid)?.description ?? ""
  }

  function moveMember(index: number, offset: -1 | 1) {
    setMemberUids((current) => {
      const target = index + offset
      if (target < 0 || target >= current.length) {
        return current
      }
      const next = [...current]
      ;[next[index], next[target]] = [next[target], next[index]]
      return next
    })
  }

  function removeMember(index: number) {
    setMemberUids((current) =>
      current.filter((_, position) => position !== index)
    )
  }

  function validate(): FieldErrors {
    const handoff = handoffHour.trim()
    return collectFieldErrors([
      ["name", requireTrimmed(name, t("rotation.errors.nameRequired"))],
      ["name", maxLength(name, NAME_MAX, t("rotation.errors.nameTooLong"))],
      [
        "remark",
        maxLength(remark, REMARK_MAX, t("rotation.errors.remarkTooLong")),
      ],
      [
        "timezone",
        requireTrimmed(timezone, t("rotation.errors.timezoneRequired")),
      ],
      [
        "timezone",
        timezone.trim() && !isValidTimezone(timezone)
          ? t("rotation.errors.timezoneInvalid")
          : null,
      ],
      [
        "startsAt",
        requireTrimmed(startsAt, t("rotation.errors.startsAtRequired")),
      ],
      [
        "handoffHour",
        /^\d+$/.test(handoff) && Number(handoff) >= 0 && Number(handoff) <= 23
          ? null
          : t("rotation.errors.handoffHourInvalid"),
      ],
      [
        "members",
        memberUids.length === 0 ? t("rotation.errors.membersRequired") : null,
      ],
      [
        "members",
        memberUids.length > MAX_MEMBERS
          ? t("rotation.errors.membersTooMany")
          : null,
      ],
    ])
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = validate()
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    const payload: SaveOnCallRotationParams = {
      name: name.trim(),
      remark: remark.trim(),
      timezone: timezone.trim(),
      period,
      startsAt: toRFC3339(startsAt),
      handoffHour: Number(handoffHour.trim()),
      members: memberUids.map<OnCallRotationMemberInput>((memberUid) => ({
        memberUid,
      })),
    }
    setPending(true)
    setFormError("")
    setFieldErrors({})
    try {
      if (rotation) {
        await updateOnCallRotation({ ...payload, uid: rotation.uid })
      } else {
        await createOnCallRotation(payload)
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
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {rotation ? t("rotation.editTitle") : t("rotation.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSave}
        >
          <DialogBody className="grid gap-4 sm:grid-cols-2">
            <FormField
              className="sm:col-span-full"
              label={t("rotation.name")}
              htmlFor="rotation-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="rotation-name"
                value={name}
                maxLength={NAME_MAX}
                placeholder={t("rotation.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
              />
            </FormField>
            <FormField
              className="sm:col-span-full"
              label={t("rotation.remark")}
              htmlFor="rotation-remark"
              error={fieldErrors.remark}
            >
              <Textarea
                id="rotation-remark"
                value={remark}
                className="min-h-16"
                placeholder={t("rotation.remarkPlaceholder")}
                onChange={(event) => setRemark(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("rotation.timezone")}
              htmlFor="rotation-timezone"
              hint={t("rotation.timezoneHint")}
              required
              error={fieldErrors.timezone}
            >
              <Input
                id="rotation-timezone"
                list="rotation-timezones"
                value={timezone}
                placeholder={t("rotation.timezonePlaceholder")}
                onChange={(event) => setTimezone(event.target.value)}
              />
              <datalist id="rotation-timezones">
                {TIMEZONE_OPTIONS.map((zone) => (
                  <option key={zone} value={zone} />
                ))}
              </datalist>
            </FormField>
            <FormField
              label={t("rotation.period")}
              htmlFor="rotation-period"
              required
            >
              <Select
                value={period}
                onValueChange={(value) => setPeriod(value)}
              >
                <SelectTrigger id="rotation-period" className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="DAILY">
                    {t("rotation.periodDaily")}
                  </SelectItem>
                  <SelectItem value="WEEKLY">
                    {t("rotation.periodWeekly")}
                  </SelectItem>
                </SelectContent>
              </Select>
            </FormField>
            <FormField
              label={t("rotation.startsAt")}
              htmlFor="rotation-starts-at"
              hint={t("rotation.startsAtHint")}
              required
              error={fieldErrors.startsAt}
            >
              <Input
                id="rotation-starts-at"
                type="datetime-local"
                value={startsAt}
                onChange={(event) => setStartsAt(event.target.value)}
              />
            </FormField>
            <FormField
              label={t("rotation.handoffHour")}
              htmlFor="rotation-handoff-hour"
              hint={t("rotation.handoffHourHint")}
              required
              error={fieldErrors.handoffHour}
            >
              <Input
                id="rotation-handoff-hour"
                type="number"
                min={0}
                max={23}
                inputMode="numeric"
                value={handoffHour}
                onChange={(event) => setHandoffHour(event.target.value)}
              />
            </FormField>
            <FormField
              className="sm:col-span-full"
              label={t("rotation.members")}
              htmlFor="rotation-members"
              hint={t("rotation.membersHint")}
              required
              error={fieldErrors.members}
            >
              <div className="grid gap-2">
                <MultiSelect
                  id="rotation-members"
                  values={memberUids}
                  options={members}
                  onChange={setMemberUids}
                  placeholder={t("multiSelect.placeholder")}
                  searchPlaceholder={t("multiSelect.searchPlaceholder")}
                  emptyText={t("rotation.noMembers")}
                  allLabel={t("multiSelect.all")}
                  clearLabel={t("multiSelect.clear")}
                  selectedLabel={selectionSummary}
                  invalid={Boolean(fieldErrors.members)}
                />
                {memberUids.length === 0 ? (
                  <p className="text-xs text-muted-foreground">
                    {t("rotation.membersEmpty")}
                  </p>
                ) : (
                  <ol className="grid gap-2">
                    {memberUids.map((uid, index) => (
                      <li
                        key={uid}
                        className="flex flex-wrap items-center gap-2 rounded-md border border-border/70 bg-muted/20 px-3 py-2"
                      >
                        <span className="font-mono text-xs text-muted-foreground tabular-nums">
                          {index + 1}
                        </span>
                        <span className="min-w-0 flex-1 text-sm font-medium break-words">
                          {memberLabel(uid)}
                          {memberDescription(uid) ? (
                            <span className="ml-2 text-xs font-normal text-muted-foreground">
                              {memberDescription(uid)}
                            </span>
                          ) : null}
                        </span>
                        <div className="flex shrink-0 items-center gap-1">
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("rotation.moveUp")}
                            title={t("rotation.moveUp")}
                            disabled={index === 0}
                            onClick={() => moveMember(index, -1)}
                          >
                            <ArrowUp aria-hidden="true" />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("rotation.moveDown")}
                            title={t("rotation.moveDown")}
                            disabled={index === memberUids.length - 1}
                            onClick={() => moveMember(index, 1)}
                          >
                            <ArrowDown aria-hidden="true" />
                          </Button>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-label={t("rotation.removeMember")}
                            title={t("rotation.removeMember")}
                            onClick={() => removeMember(index)}
                          >
                            <Trash2 aria-hidden="true" />
                          </Button>
                        </div>
                      </li>
                    ))}
                  </ol>
                )}
              </div>
            </FormField>
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
