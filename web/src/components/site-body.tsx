import { cn } from "@/lib/class-names"

/** 官网正文当前按纯文本渲染：保留段落与换行，不解析 Markdown。 */
function paragraphsOf(content: string): string[] {
  return content
    .replace(/\r\n?/g, "\n")
    .split(/\n{2,}/)
    .map((paragraph) => paragraph.replace(/^\n+|\n+$/g, ""))
    .filter((paragraph) => paragraph.trim() !== "")
}

/** 首页正文摘要：按字符截断后去掉半个词，避免破坏段落。 */
export function excerptOf(content: string, max = 480): string {
  const text = content.replace(/\r\n?/g, "\n").trim()
  if (text.length <= max) {
    return text
  }
  const clipped = text.slice(0, max)
  const lastBreak = Math.max(
    clipped.lastIndexOf("\n"),
    clipped.lastIndexOf(" ")
  )
  const base = lastBreak > max * 0.6 ? clipped.slice(0, lastBreak) : clipped
  return `${base.trimEnd()}…`
}

interface SiteBodyProps {
  content: string
  emptyText: string
  className?: string
}

/**
 * 公开官网与后台草稿预览共用的正文渲染。
 *
 * 项目的 package.json 没有 Markdown 渲染依赖，这里不新增依赖：段落与换行按纯文本
 * 保留（whitespace-pre-wrap），避免把 `**粗体**` 之类的源码当富文本误解析。
 */
export function SiteBody({ content, emptyText, className }: SiteBodyProps) {
  const paragraphs = paragraphsOf(content)
  if (paragraphs.length === 0) {
    return (
      <p className={cn("text-sm text-muted-foreground", className)}>
        {emptyText}
      </p>
    )
  }
  return (
    <div className={cn("grid gap-4", className)}>
      {paragraphs.map((paragraph, index) => (
        <p
          key={`${index}-${paragraph.slice(0, 24)}`}
          className="text-sm leading-7 break-words whitespace-pre-wrap text-foreground/90 sm:text-base"
        >
          {paragraph}
        </p>
      ))}
    </div>
  )
}
