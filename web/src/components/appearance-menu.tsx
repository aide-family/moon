import { Check, Moon, Sun, SunMoon } from "lucide-react"
import * as React from "react"
import { useTranslation } from "react-i18next"

import { BrandSwatch } from "@/components/brand-swatch"
import { useTheme } from "@/components/theme-provider"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import type { ThemeMode } from "@/lib/appearance"
import { BRAND_IDS, BRAND_PALETTES, type BrandId } from "@/lib/brand-palette.generated"
import { cn } from "@/lib/class-names"

const MODES: { value: ThemeMode; icon: typeof Sun; labelKey: string }[] = [
  { value: "light", icon: Sun, labelKey: "appearance.light" },
  { value: "dark", icon: Moon, labelKey: "appearance.dark" },
  { value: "system", icon: SunMoon, labelKey: "appearance.system" },
]

const BRAND_COLUMNS = 3

/**
 * 外观切换器：明暗模式 + 品牌主题。
 *
 * 交互约定：
 * - 点击即生效（不点“确定”），Popover 保持打开，方便左右对比 7 套品牌色。
 * - 选中态不只用颜色表达：对勾 + 字重 + ring + aria-checked 三重冗余。
 * - 两个分组都是 radiogroup：方向键移动并选中，Home/End 到首尾。
 * - 切换结果通过 aria-live 播报，读屏用户也能确认生效。
 */
export function AppearanceMenu() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme, brand, setBrand } = useTheme()
  const [announcement, setAnnouncement] = React.useState("")
  const modeRefs = React.useRef<Array<HTMLButtonElement | null>>([])
  const brandRefs = React.useRef<Array<HTMLButtonElement | null>>([])

  const zh = i18n.language.startsWith("zh")
  const labelOf = (id: BrandId) => (zh ? BRAND_PALETTES[id].labelZh : BRAND_PALETTES[id].labelEn)
  const TriggerIcon = MODES.find((mode) => mode.value === theme)?.icon ?? SunMoon

  function pickMode(next: ThemeMode) {
    if (next === theme) {
      return
    }

    setTheme(next)
    setAnnouncement(t("appearance.switchedMode", { mode: t(`appearance.${next}`) }))
  }

  function pickBrand(next: BrandId) {
    if (next === brand) {
      return
    }

    setBrand(next)
    setAnnouncement(t("appearance.switchedBrand", { brand: labelOf(next) }))
  }

  function moveInGroup(
    group: "mode" | "brand",
    current: number,
    delta: number,
    absolute?: "first" | "last"
  ) {
    const length = group === "mode" ? MODES.length : BRAND_IDS.length
    const target =
      absolute === "first"
        ? 0
        : absolute === "last"
          ? length - 1
          : (current + delta + length) % length

    if (group === "mode") {
      pickMode(MODES[target].value)
      modeRefs.current[target]?.focus()
      return
    }

    pickBrand(BRAND_IDS[target])
    brandRefs.current[target]?.focus()
  }

  function onGroupKeyDown(event: React.KeyboardEvent, group: "mode" | "brand", index: number) {
    const isGrid = group === "brand"
    let handled = true

    switch (event.key) {
      case "ArrowRight":
        moveInGroup(group, index, 1)
        break
      case "ArrowLeft":
        moveInGroup(group, index, -1)
        break
      case "ArrowDown":
        if (!isGrid) {
          handled = false
          break
        }
        moveInGroup(group, index, BRAND_COLUMNS)
        break
      case "ArrowUp":
        if (!isGrid) {
          handled = false
          break
        }
        moveInGroup(group, index, -BRAND_COLUMNS)
        break
      case "Home":
        moveInGroup(group, index, 0, "first")
        break
      case "End":
        moveInGroup(group, index, 0, "last")
        break
      default:
        handled = false
    }

    if (handled) {
      event.preventDefault()
    }
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="relative shrink-0"
          aria-label={t("appearance.label")}
          data-slot="appearance-trigger"
        >
          <TriggerIcon />
          {/* 右下角品牌点：颜色只是附加信息，模式由图标本身体现。
              用 --primary-ink 而不是 --brand-500：后者在柠檬黄浅色下几乎是白点，
              而 ink 对页面底色有 ≥4.5:1 的保证。 */}
          <span
            aria-hidden
            className="absolute right-1 bottom-1 size-1.5 rounded-full ring-2 ring-background"
            style={{ backgroundColor: "var(--primary-ink)" }}
          />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        align="end"
        className="max-h-[var(--radix-popover-content-available-height)] w-[19rem] overflow-y-auto p-3"
      >
        <p className="px-1 pb-2 text-xs font-medium text-muted-foreground">
          {t("appearance.mode")}
        </p>
        <div
          role="radiogroup"
          aria-label={t("appearance.mode")}
          className="grid grid-cols-3 gap-1 rounded-lg bg-muted p-1"
        >
          {MODES.map((mode, index) => {
            const active = theme === mode.value
            return (
              <button
                key={mode.value}
                type="button"
                role="radio"
                aria-checked={active}
                tabIndex={active ? 0 : -1}
                ref={(node) => {
                  modeRefs.current[index] = node
                }}
                onClick={() => pickMode(mode.value)}
                onKeyDown={(event) => onGroupKeyDown(event, "mode", index)}
                className={cn(
                  "relative flex h-8 items-center justify-center gap-1.5 rounded-md text-xs font-medium",
                  "focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                  active
                    ? "bg-background text-foreground shadow-sm"
                    : "text-muted-foreground hover:text-foreground"
                )}
              >
                <mode.icon className="size-3.5" aria-hidden />
                {t(mode.labelKey)}
                {active ? (
                  <span aria-hidden className="absolute inset-x-2 bottom-0.5 h-0.5 rounded-full bg-primary" />
                ) : null}
              </button>
            )
          })}
        </div>

        <Separator className="my-3" />

        <p className="px-1 pb-2 text-xs font-medium text-muted-foreground">
          {t("appearance.brand")}
        </p>
        <div
          role="radiogroup"
          aria-label={t("appearance.brand")}
          className="grid grid-cols-3 gap-2"
        >
          {BRAND_IDS.map((id, index) => {
            const active = brand === id
            return (
              <button
                key={id}
                type="button"
                role="radio"
                aria-checked={active}
                tabIndex={active ? 0 : -1}
                ref={(node) => {
                  brandRefs.current[index] = node
                }}
                onClick={() => pickBrand(id)}
                onKeyDown={(event) => onGroupKeyDown(event, "brand", index)}
                className={cn(
                  "group relative rounded-lg p-1 text-left",
                  "focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                  active && "ring-2 ring-primary ring-offset-2 ring-offset-background"
                )}
              >
                <BrandSwatch palette={BRAND_PALETTES[id]} />
                <span
                  className={cn(
                    "mt-1 flex items-center gap-1 px-0.5 text-[11px] leading-4",
                    active ? "font-medium text-foreground" : "text-muted-foreground"
                  )}
                >
                  {labelOf(id)}
                  {active ? <Check className="size-3 text-primary-ink" aria-hidden /> : null}
                </span>
                {announcement && active ? (
                  <span aria-hidden className="brand-confirm pointer-events-none absolute inset-0 rounded-lg" />
                ) : null}
              </button>
            )
          })}
        </div>

        <p aria-live="polite" className="sr-only">
          {announcement}
        </p>
      </PopoverContent>
    </Popover>
  )
}
