import { useMemo, type ReactNode } from "react"

import { SiteBody } from "@/components/site-body"
import { SiteToc } from "@/components/site-toc"
import { cn } from "@/lib/class-names"
import { extractHeadings, slugifyHeading } from "@/lib/markdown-headings"

interface SiteArticleProps {
  title: string
  /** 标题下方的时间、语言等元信息，文案与顺序由调用方决定。 */
  meta?: ReactNode
  coverImageUrl?: string
  bodyMd: string
  emptyText: string
  /** 标题前的返回入口；放在主栏里，才能和正文一起对齐到目录栅格。 */
  beforeTitle?: ReactNode
  /** 元信息与正文之间的附加提示（例如语言回退提示）。 */
  children?: ReactNode
}

/**
 * 官网文档页的固定版式：标题、元信息、封面、附加提示、Markdown 正文与右侧目录。
 *
 * `/p/:slug`（已发布快照）与 `/site-preview/:uid`（后台草稿）共用这一份版式，
 * 两边的区别只在数据来源与分析行文案。
 *
 * 桌面（lg+，1024px 起）是「正文 68ch + 14rem 粘性目录」两栏；lg 以下只保留单栏正文，
 * 避免横向挤压。1024px 下正文仍有 68ch（≥ 60ch），不会因为目录提前出现而变窄。
 * 正文没有 h2/h3 时目录整列不渲染，单栏保持居中的 3xl 阅读宽度。
 */
export function SiteArticle({
  title,
  meta,
  coverImageUrl,
  bodyMd,
  emptyText,
  beforeTitle,
  children,
}: SiteArticleProps) {
  const headings = useMemo(() => {
    const items = extractHeadings(bodyMd, { levels: [1, 2, 3] })
    // 正文第一个 h1 通常就是页面标题的重复：渲染时它被降级成 h2，
    // 目录里也不该再为目标标题列一条同名项。
    const first = items[0]
    if (
      first &&
      first.level === 1 &&
      slugifyHeading(first.text) === slugifyHeading(title)
    ) {
      return items.slice(1)
    }
    return items
  }, [bodyMd, title])
  const hasToc = headings.length > 0

  return (
    <div
      className={cn(
        "mx-auto w-full",
        hasToc &&
          "lg:grid lg:grid-cols-[minmax(0,68ch)_14rem] lg:justify-center lg:gap-8 xl:gap-12"
      )}
    >
      <div
        className={cn(
          "min-w-0",
          hasToc
            ? "mx-auto w-full max-w-3xl lg:mx-0 lg:max-w-[68ch]"
            : "mx-auto w-full max-w-3xl"
        )}
      >
        {beforeTitle}
        <h1 className="font-heading text-3xl font-semibold tracking-tight break-words sm:text-4xl">
          {title}
        </h1>
        {meta}
        {coverImageUrl ? (
          <img
            src={coverImageUrl}
            alt=""
            className="mt-6 w-full rounded-2xl border border-border/60 object-cover"
          />
        ) : null}
        {children}
        <SiteBody
          content={bodyMd}
          emptyText={emptyText}
          className="mt-8"
          headingShift={1}
          headingLevels={[1, 2, 3]}
        />
      </div>
      {hasToc ? <SiteToc headings={headings} /> : null}
    </div>
  )
}
