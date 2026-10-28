import type { BrandPalette } from "@/lib/brand-palette.generated"
import { cn } from "@/lib/class-names"

/**
 * 品牌预览色卡：上浅下深两段，一眼看出这套品牌色在两种模式下长什么样。
 *
 * 刻意使用品牌自己的字面色值（generated 注册表），不引用当前主题变量——
 * 一旦引用，7 张卡片会全部跟着当前主题一起变，就失去“预览”的意义。
 */
export function BrandSwatch({
  palette,
  className,
}: {
  palette: BrandPalette
  className?: string
}) {
  return (
    <span
      className={cn("block overflow-hidden rounded-md border border-border/60", className)}
    >
      {[palette.previewLight, palette.previewDark].map((tone, index) => (
        <span
          key={index}
          className="flex h-4 items-center gap-1 px-1.5"
          style={{ backgroundColor: tone.bg }}
        >
          <span className="h-2 flex-1 rounded-[2px]" style={{ backgroundColor: tone.card }} />
          <span className="size-2 rounded-[2px]" style={{ backgroundColor: tone.fill }} />
          <span className="h-2 w-3 rounded-[2px]" style={{ backgroundColor: tone.tint }} />
        </span>
      ))}
    </span>
  )
}
