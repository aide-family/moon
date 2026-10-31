import { Info, Lock } from "lucide-react"
import { useId } from "react"
import { useTranslation } from "react-i18next"

import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/class-names"
import {
  capabilityDescription,
  capabilityLabelText,
  capabilitySecondaryText,
  groupCapabilityCodes,
} from "@/lib/feature-catalog"

interface FeaturePickerProps {
  /** 可选能力 code；通常来自后端目录，接口失败时由调用方用本地目录兜底。 */
  codes: string[]
  /** 已勾选的能力 code。 */
  value: string[]
  onChange: (next: string[]) => void
  disabled?: boolean
  /**
   * 锁定能力：来自套餐，不能取消勾选（只增不减）。
   * 勾选状态按「已选 ∪ 锁定」计算，回调里也始终包含它们。
   */
  lockedCodes?: string[]
  /** 锁定行右侧的说明文案（由调用方本地化）。 */
  lockedLabel?: string
}

/**
 * 两级能力勾选器：模块（第一级，带三态全选）→ 能力（第二级，中英并标）。
 * 只负责选择，不包含滚动容器——放在 DialogBody 里由弹窗外框滚动。
 */
export function FeaturePicker({
  codes,
  value,
  onChange,
  disabled = false,
  lockedCodes,
  lockedLabel,
}: FeaturePickerProps) {
  const { t, i18n } = useTranslation()
  const id = useId()
  const locked = new Set(lockedCodes ?? [])
  const selected = new Set([...value, ...locked])
  const groups = groupCapabilityCodes(codes)

  function toggle(code: string, checked: boolean) {
    if (locked.has(code) && !checked) {
      return
    }
    const next = new Set(selected)
    if (checked) {
      next.add(code)
    } else {
      next.delete(code)
    }
    onChange([...next])
  }

  function toggleGroup(groupCodes: string[], checked: boolean) {
    const next = new Set(selected)
    for (const code of groupCodes) {
      if (!checked && locked.has(code)) {
        continue
      }
      if (checked) {
        next.add(code)
      } else {
        next.delete(code)
      }
    }
    onChange([...next])
  }

  if (groups.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("state.empty")}</p>
  }

  return (
    <div className="grid gap-2">
      {groups.map((group) => {
        const groupSelected = group.codes.filter((code) =>
          selected.has(code)
        ).length
        const allSelected =
          group.codes.length > 0 && groupSelected === group.codes.length
        const someSelected = groupSelected > 0 && !allSelected
        return (
          <div key={group.code} className="rounded-md border">
            {/* 组头复选框是行的最后一个元素：Checkbox 用 ::after 扩大点击热区，
                右侧再多留 12px，热区才不会被 DialogBody 的 overflow-auto 算成横向溢出。 */}
            <div className="flex min-h-10 items-center gap-2 border-b py-1.5 pr-3 pl-2">
              <span className="min-w-0 flex-1 text-sm font-medium break-words">
                {group.title}
              </span>
              <span className="text-xs text-muted-foreground tabular-nums">
                {groupSelected}/{group.codes.length}
              </span>
              <Checkbox
                checked={
                  allSelected ? true : someSelected ? "indeterminate" : false
                }
                disabled={disabled}
                aria-label={t("packageAdmin.selectGroup", {
                  group: group.title,
                })}
                onCheckedChange={(checked) =>
                  toggleGroup(group.codes, checked === true)
                }
              />
            </div>
            <div className="grid gap-1 p-2 sm:grid-cols-2">
              {group.codes.map((code) => {
                const isLocked = locked.has(code)
                const checked = selected.has(code)
                const description = capabilityDescription(code, i18n.language)
                return (
                  <div
                    key={code}
                    className={cn(
                      "flex min-h-11 items-start gap-2 rounded-sm border-l-2 border-transparent px-2 py-1.5",
                      checked &&
                        "border-primary bg-primary/5 dark:bg-primary/10"
                    )}
                  >
                    <Checkbox
                      id={`${id}-${code}`}
                      checked={checked}
                      disabled={disabled || isLocked}
                      className="mt-0.5"
                      onCheckedChange={(next) => toggle(code, next === true)}
                    />
                    <Label
                      htmlFor={`${id}-${code}`}
                      className={cn(
                        "grid min-w-0 flex-1 gap-0.5 font-normal",
                        disabled || isLocked
                          ? "cursor-default"
                          : "cursor-pointer"
                      )}
                    >
                      <span className="flex min-w-0 flex-wrap items-baseline gap-x-1.5">
                        <span className="text-sm">
                          {capabilityLabelText(code, i18n.language)}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {capabilitySecondaryText(code, i18n.language)}
                        </span>
                      </span>
                    </Label>
                    {description ? (
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <button
                            type="button"
                            aria-label={t("packageAdmin.featureTip", {
                              feature: capabilityLabelText(
                                code,
                                i18n.language
                              ),
                            })}
                            className="mt-0.5 shrink-0 rounded p-0.5 text-muted-foreground hover:text-foreground"
                          >
                            <Info aria-hidden="true" className="size-3.5" />
                          </button>
                        </TooltipTrigger>
                        <TooltipContent className="max-w-xs">
                          <p className="text-xs">{description}</p>
                        </TooltipContent>
                      </Tooltip>
                    ) : null}
                    {isLocked && lockedLabel ? (
                      <span className="mt-0.5 inline-flex shrink-0 items-center gap-1 rounded-4xl border border-border px-1.5 text-[11px] whitespace-nowrap text-muted-foreground">
                        <Lock aria-hidden="true" className="size-3" />
                        {lockedLabel}
                      </span>
                    ) : null}
                  </div>
                )
              })}
            </div>
          </div>
        )
      })}
    </div>
  )
}
