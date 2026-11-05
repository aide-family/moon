import ReactMarkdown, { type Components } from "react-markdown"
import remarkGfm from "remark-gfm"

import { cn } from "@/lib/class-names"

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

/**
 * Markdown 元素的排版：只使用语义 token，不引入 @tailwindcss/typography。
 *
 * 容器用 grid row-gap 控制块间距；标题、列表、引用、代码与表格通过后代选择器
 * 统一上色和定尺寸，行内代码再单独与代码块区分开。
 */
const markdownClassName = cn(
  "grid min-w-0 gap-4 text-sm leading-7 break-words text-foreground/90 sm:text-base",
  // 标题阶梯
  "[&_h1]:font-heading [&_h1]:text-2xl [&_h1]:font-semibold [&_h1]:tracking-tight [&_h1]:text-foreground",
  "[&_h2]:font-heading [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:tracking-tight [&_h2]:text-foreground",
  "[&_h3]:font-heading [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:text-foreground",
  "[&_h4]:text-base [&_h4]:font-semibold [&_h4]:text-foreground",
  "[&_h5]:text-sm [&_h5]:font-semibold [&_h5]:text-foreground",
  "[&_h6]:text-sm [&_h6]:font-semibold [&_h6]:text-muted-foreground",
  // 列表：Tailwind preflight 会清掉标记，这里恢复。
  "[&_ul]:list-disc [&_ul]:pl-6",
  "[&_ol]:list-decimal [&_ol]:pl-6",
  "[&_li]:pl-1",
  "[&_li+li]:mt-1",
  "[&_li>ul]:mt-1 [&_li>ol]:mt-1",
  // 引用与分隔线
  "[&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-4 [&_blockquote]:text-muted-foreground",
  "[&_blockquote]:space-y-2",
  "[&_hr]:border-border/60",
  // 行内代码：代码块里的 code 由 pre 自己排版，不再加底色。
  "[&_:not(pre)>code]:rounded-md [&_:not(pre)>code]:bg-muted [&_:not(pre)>code]:px-1.5 [&_:not(pre)>code]:py-0.5 [&_:not(pre)>code]:font-mono [&_:not(pre)>code]:text-[0.9em] [&_:not(pre)>code]:text-foreground",
  // GFM 表格
  "[&_th]:border-b [&_th]:border-border/60 [&_th]:px-3 [&_th]:py-2 [&_th]:text-left [&_th]:font-semibold [&_th]:text-foreground",
  "[&_td]:border-b [&_td]:border-border/40 [&_td]:px-3 [&_td]:py-2 [&_td]:align-top",
  "[&_tr:last-child_td]:border-b-0",
  // GFM 任务列表复选框
  "[&_li>input]:mr-1.5 [&_li>input]:align-middle",
  // 链接与强调
  "[&_a]:font-medium [&_a]:text-primary-ink [&_a]:underline [&_a]:underline-offset-4 [&_a]:hover:text-primary-hover",
  "[&_strong]:font-semibold [&_strong]:text-foreground",
  "[&_del]:text-muted-foreground",
  "[&_sup]:text-[0.75em] [&_sub]:text-[0.75em]"
)

/**
 * 链接一律新窗口打开；Markdown 里的图片懒加载并沿用官网的圆角描边。
 * 不启用 rehype-raw：正文里的 HTML 不会被当成标签执行。
 */
const markdownComponents: Components = {
  a({ href, title, children }) {
    return (
      <a href={href} title={title} target="_blank" rel="noreferrer">
        {children}
      </a>
    )
  },
  img({ src, alt, title }) {
    return (
      <img
        src={src}
        alt={alt ?? ""}
        title={title}
        loading="lazy"
        decoding="async"
        className="max-w-full rounded-2xl border border-border/60"
      />
    )
  },
  pre({ children }) {
    return (
      <pre className="overflow-x-auto rounded-xl border border-border/60 bg-muted/40 p-4 font-mono text-xs leading-6">
        {children}
      </pre>
    )
  },
  code({ children }) {
    return <code className="font-mono">{children}</code>
  },
  table({ children }) {
    return (
      <div className="overflow-x-auto rounded-xl border border-border/60">
        <table className="w-full border-collapse text-left text-sm">
          {children}
        </table>
      </div>
    )
  },
}

interface SiteBodyProps {
  content: string
  emptyText: string
  className?: string
}

/**
 * 公开官网与后台草稿预览共用的正文渲染。
 *
 * 正文按 GitHub Flavored Markdown 解析（标题、列表、引用、代码块、表格、任务列表）；
 * 不开启 raw HTML，避免后台正文注入脚本。样式全部使用语义 token。
 */
export function SiteBody({ content, emptyText, className }: SiteBodyProps) {
  if (content.trim() === "") {
    return (
      <p className={cn("text-sm text-muted-foreground", className)}>
        {emptyText}
      </p>
    )
  }
  return (
    <div className={cn(markdownClassName, className)}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
        {content}
      </ReactMarkdown>
    </div>
  )
}
