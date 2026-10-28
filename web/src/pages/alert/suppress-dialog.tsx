import { Check } from "lucide-react"
import { useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"

import type { AlertEventItem } from "@/api/alert-event/types"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
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
import { cn } from "@/lib/class-names"

// 默认抑制时长梯度，秒。范围和后端 buf.validate 保持一致。
const PRESETS = [
  5 * 60,
  15 * 60,
  30 * 60,
  60 * 60,
  2 * 60 * 60,
  6 * 60 * 60,
  12 * 60 * 60,
  24 * 60 * 60,
]
const DEFAULT_SECONDS = 30 * 60
const MIN_SECONDS = 60
const MAX_SECONDS = 30 * 24 * 60 * 60
const UNIT_SECONDS = { MINUTE: 60, HOUR: 3600, DAY: 86400 } as const
type CustomUnit = keyof typeof UNIT_SECONDS
// 永久抑制不设到期时间，和「自定义」一样是时长选择的一种。
const PERMANENT = "permanent"
type Preset = number | "custom" | typeof PERMANENT

export interface SuppressInput {
  // 抑制时长，秒。永久抑制时为 0。
  seconds: number
  permanent: boolean
}

interface AlertEventSuppressDialogProps {
  event: AlertEventItem | null
  open: boolean
  pending: boolean
  error: string
  onConfirm: (input: SuppressInput) => void
  onOpenChange: (open: boolean) => void
}

/**
 * 实时告警抑制对话框：给一组常用时长，也允许按时长 + 单位自定义，或永久抑制。
 * 提交前在前端限制在后端允许的 1 分钟到 30 天之间，避免空值和越界请求。
 */
export function AlertEventSuppressDialog({
  event,
  open,
  pending,
  error,
  onConfirm,
  onOpenChange,
}: AlertEventSuppressDialogProps) {
  const { t } = useTranslation()
  const [preset, setPreset] = useState<Preset>(DEFAULT_SECONDS)
  const [customValue, setCustomValue] = useState("30")
  const [customUnit, setCustomUnit] = useState<CustomUnit>("MINUTE")
  // 对话框每次打开都重新挂载，初始值就是本次的默认梯度。
  const [now] = useState(() => Date.now())

  const permanent = preset === PERMANENT
  const seconds =
    preset === "custom"
      ? Number(customValue || "0") * UNIT_SECONDS[customUnit]
      : typeof preset === "number"
        ? preset
        : 0
  const valid =
    permanent ||
    (Number.isFinite(seconds) &&
      Number.isInteger(seconds) &&
      seconds >= MIN_SECONDS &&
      seconds <= MAX_SECONDS)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("alertEvent.suppressTitle")}</DialogTitle>
          <DialogDescription className="truncate">
            {event?.summary || event?.fingerprint}
          </DialogDescription>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <div className="grid gap-2">
            <span className="text-sm text-muted-foreground">
              {t("alertEvent.suppressPreset")}
            </span>
            <div className="flex flex-wrap gap-2">
              {PRESETS.map((value) => (
                <PresetButton
                  key={value}
                  selected={preset === value}
                  className="font-mono"
                  label={durationLabel(t, value)}
                  onClick={() => setPreset(value)}
                />
              ))}
              <PresetButton
                selected={preset === "custom"}
                label={t("alertEvent.suppressCustom")}
                onClick={() => setPreset("custom")}
              />
              <PresetButton
                selected={permanent}
                label={t("alertEvent.suppressPermanent")}
                onClick={() => setPreset(PERMANENT)}
              />
            </div>
          </div>
          {preset === "custom" ? (
            <div className="grid grid-cols-[1fr_8rem] gap-2">
              <Input
                type="number"
                min={1}
                step={1}
                inputMode="numeric"
                value={customValue}
                placeholder={t("alertEvent.suppressCustomPlaceholder")}
                aria-label={t("alertEvent.suppressDuration")}
                onChange={(event) => setCustomValue(event.target.value)}
              />
              <Select
                value={customUnit}
                onValueChange={(value) => setCustomUnit(value as CustomUnit)}
              >
                <SelectTrigger
                  className="w-full"
                  aria-label={t("alertEvent.suppressUnit")}
                >
                  <SelectValue placeholder={t("alertEvent.suppressUnit")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="MINUTE">
                    {t("alertEvent.suppressUnitMinute")}
                  </SelectItem>
                  <SelectItem value="HOUR">
                    {t("alertEvent.suppressUnitHour")}
                  </SelectItem>
                  <SelectItem value="DAY">
                    {t("alertEvent.suppressUnitDay")}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          ) : null}
          <div className="grid gap-1 text-xs text-muted-foreground">
            {permanent ? (
              <span>{t("alertEvent.suppressPermanentHint")}</span>
            ) : (
              <>
                <span>{t("alertEvent.suppressRangeHint")}</span>
                {valid ? (
                  <span>
                    {t("alertEvent.suppressUntilHint", {
                      until: new Date(now + seconds * 1000).toLocaleString(),
                    })}
                  </span>
                ) : (
                  <span className="text-destructive">
                    {t("alertEvent.suppressInvalid")}
                  </span>
                )}
              </>
            )}
          </div>
        </DialogBody>
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            {t("action.cancel")}
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={pending || !valid}
            onClick={() => onConfirm({ seconds, permanent })}
          >
            {t("action.suppress")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

/**
 * 时长梯度按钮。选中态用主色实心 + 勾选图标 + 外圈，未选中态弱化文字，
 * 让当前选择在浅色和深色主题下都能一眼分辨。
 */
function PresetButton({
  selected,
  label,
  className,
  onClick,
}: {
  selected: boolean
  label: string
  className?: string
  onClick: () => void
}) {
  return (
    <Button
      type="button"
      size="sm"
      variant={selected ? "default" : "outline"}
      className={cn(
        "gap-1.5",
        selected
          ? "ring-2 ring-primary/40 ring-offset-1 ring-offset-background"
          : "text-muted-foreground",
        className
      )}
      aria-pressed={selected}
      onClick={onClick}
    >
      {selected ? <Check aria-hidden /> : null}
      {label}
    </Button>
  )
}

// 把秒数显示成梯度按钮上的短文案：天、小时或分钟。
function durationLabel(t: TFunction, seconds: number) {
  if (seconds % UNIT_SECONDS.DAY === 0) {
    return t("alertEvent.suppressDurationDays", {
      count: seconds / UNIT_SECONDS.DAY,
    })
  }
  if (seconds % UNIT_SECONDS.HOUR === 0) {
    return t("alertEvent.suppressDurationHours", {
      count: seconds / UNIT_SECONDS.HOUR,
    })
  }
  return t("alertEvent.suppressDurationMinutes", {
    count: seconds / UNIT_SECONDS.MINUTE,
  })
}
