import type { ReactNode } from "react"

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { cn } from "@/lib/class-names"

export interface DetailField {
  label: string
  value: ReactNode
  // 占满整行的字段（权限分组这类长内容），默认半行。
  span?: 1 | 2
}

export interface DetailSection {
  // 分组标题，同一抽屉内保持唯一，作为 React key 使用。
  title: string
  // 标题右侧的一句话结论，例如「全部事件」「全天候」。
  description?: ReactNode
  fields?: DetailField[]
  // 字段排版：stacked 标签在上（默认），inline 标签在左、值在右，适合短字段密集展示。
  layout?: "stacked" | "inline"
  // 自定义区块内容；给出后不再渲染 fields，用于通知组时段这类结构化数据。
  content?: ReactNode
}

export type DetailSheetSize = "default" | "wide" | "40vw"

interface DetailSheetProps {
  open: boolean
  title: string
  // 标题下的行内说明，通常放状态标签和 UID，只使用行内元素。
  description?: ReactNode
  // 头部操作按钮，空间不足时整行换行，不挤压标题。
  actions?: ReactNode
  // 顶部关键指标区（无标题、有底色的摘要块）。
  summary?: ReactNode
  // 兼容旧调用：没有分组时的平铺字段。
  fields?: DetailField[]
  sections?: DetailSection[]
  size?: DetailSheetSize
  // 兼容旧调用，等价于 size="wide"。
  wide?: boolean
  // 底部固定操作区。
  footer?: ReactNode
  onOpenChange: (open: boolean) => void
}

// default 沿用 shadcn 抽屉宽度（手机 75%、桌面 384px）；
// wide 是 48rem，放表达式/模板这类宽内容；
// 40vw 是详情抽屉：桌面按视口 40% 取宽，用 26rem/56rem 夹住上下限，手机整屏。
const sizeClassNames: Record<DetailSheetSize, string> = {
  default: "",
  wide: "data-[side=right]:w-full data-[side=right]:sm:max-w-3xl!",
  "40vw":
    "data-[side=right]:w-full data-[side=right]:sm:w-[40vw] data-[side=right]:sm:min-w-[26rem] data-[side=right]:sm:max-w-[56rem]!",
}

/**
 * 详情抽屉：粘性头部（标题 + 状态 + 操作）、单一滚动内容区、可选底部操作区。
 * 头部和底栏固定，只有中间内容滚动，长字段不会把操作按钮顶出视口。
 */
export function DetailSheet({
  open,
  title,
  description,
  actions,
  summary,
  fields,
  sections,
  size,
  wide = false,
  footer,
  onOpenChange,
}: DetailSheetProps) {
  const resolvedSize: DetailSheetSize = size ?? (wide ? "wide" : "default")
  const sectionList = sections ?? []
  const fieldList = fields ?? []

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className={cn("gap-0 p-0", sizeClassNames[resolvedSize])}
      >
        {/* 头部不参与滚动：外层是固定高度的 flex 列，只有下面的内容区滚动，
            所以头部天然「粘」在顶部，不需要 sticky/z-index（否则会盖住右上角关闭按钮）。 */}
        <SheetHeader className="shrink-0 gap-2 border-b bg-popover px-4 py-3 pr-14 sm:px-5">
          <SheetTitle className="font-heading text-base leading-snug break-words">
            {title}
          </SheetTitle>
          {description ? (
            <SheetDescription className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1 text-muted-foreground">
              {description}
            </SheetDescription>
          ) : null}
          {actions ? (
            <div className="flex flex-wrap items-center gap-2 pt-0.5">
              {actions}
            </div>
          ) : null}
        </SheetHeader>
        <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain">
          {/* @container 让字段栅格跟随抽屉实际宽度，而不是浏览器视口宽度。 */}
          <div className="@container grid gap-5 px-4 py-4 sm:px-5">
            {summary ? (
              <div className="rounded-lg border bg-muted/30 p-3">{summary}</div>
            ) : null}
            {sectionList.length > 0 ? (
              sectionList.map((section) => (
                <DetailSectionBlock key={section.title} section={section} />
              ))
            ) : fieldList.length > 0 ? (
              <DetailFields fields={fieldList} />
            ) : null}
          </div>
        </div>
        {footer ? (
          <div className="shrink-0 border-t bg-popover px-4 py-3 sm:px-5">
            {footer}
          </div>
        ) : null}
      </SheetContent>
    </Sheet>
  )
}

function DetailSectionBlock({ section }: { section: DetailSection }) {
  const sectionFields = section.fields ?? []
  return (
    <section className="grid min-w-0 gap-2">
      <div className="flex flex-wrap items-baseline justify-between gap-x-2 gap-y-1 border-b pb-1.5">
        {/* 分组标题用正文色 + 细分隔线，和下面灰一档的字段标签拉开层级。 */}
        <h3 className="font-heading text-xs font-semibold tracking-wide text-foreground">
          {section.title}
        </h3>
        {section.description ? (
          <span className="text-xs text-muted-foreground">
            {section.description}
          </span>
        ) : null}
      </div>
      {section.content ? (
        <div className="min-w-0">{section.content}</div>
      ) : sectionFields.length > 0 ? (
        <DetailFields
          fields={sectionFields}
          layout={section.layout ?? "stacked"}
        />
      ) : null}
    </section>
  )
}

function DetailFields({
  fields,
  layout = "stacked",
}: {
  fields: DetailField[]
  layout?: "stacked" | "inline"
}) {
  const inline = layout === "inline"
  return (
    // 容器宽度够时才把字段并成两列，384px 的默认抽屉仍然单列。
    <dl
      className={cn(
        "grid",
        inline ? "gap-2" : "gap-x-5 gap-y-3 @md:grid-cols-2"
      )}
    >
      {fields.map((field) => (
        <div
          key={field.label}
          className={cn(
            "grid min-w-0",
            inline
              ? "gap-x-3 gap-y-0.5 @md:grid-cols-[7rem_minmax(0,1fr)] @md:items-baseline"
              : "content-start gap-1",
            !inline && field.span === 2 && "@md:col-span-2"
          )}
        >
          <dt className="text-xs text-muted-foreground">{field.label}</dt>
          <dd className="min-w-0 text-sm leading-relaxed break-words">
            {isEmptyValue(field.value) ? (
              <span className="text-muted-foreground">—</span>
            ) : (
              field.value
            )}
          </dd>
        </div>
      ))}
    </dl>
  )
}

/** 空值统一显示占位符，0 和 false 是有效值，不能当成空。 */
function isEmptyValue(value: ReactNode): boolean {
  if (value === null || value === undefined || value === false) {
    return true
  }
  if (typeof value === "string") {
    return value.trim() === ""
  }
  if (typeof value === "number") {
    return false
  }
  if (Array.isArray(value)) {
    return value.length === 0
  }
  return false
}
