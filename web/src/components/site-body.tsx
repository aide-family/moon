import { Check, Copy } from "lucide-react"
import {
  isValidElement,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react"
import { useTranslation } from "react-i18next"
import ReactMarkdown, { type Components } from "react-markdown"
import remarkGfm from "remark-gfm"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/class-names"
import { extractHeadings, slugifyHeading } from "@/lib/markdown-headings"

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

/** 把 React 子树拍平成纯文本，用于复制按钮与标题 id 兜底。 */
function textOf(node: ReactNode): string {
  if (node === null || node === undefined || typeof node === "boolean") {
    return ""
  }
  if (typeof node === "string" || typeof node === "number") {
    return String(node)
  }
  if (Array.isArray(node)) {
    return node.map(textOf).join("")
  }
  if (isValidElement<{ children?: ReactNode }>(node)) {
    return textOf(node.props.children)
  }
  return ""
}

/** react-markdown 传给标题组件的 hast 节点里我们只关心源码行号。 */
interface HeadingNode {
  position?: { start?: { line?: number } }
}

/**
 * 标题 id：优先按源码行号反查与目录同一份提取结果；setext 标题等没进目录的
 * 情况退回同一个 slug 规则，保证锚点不悬空。
 */
function headingId(
  node: HeadingNode | undefined,
  lineIds: Map<number, string>,
  children: ReactNode
): string | undefined {
  const line = node?.position?.start?.line
  if (line !== undefined) {
    const known = lineIds.get(line)
    if (known) {
      return known
    }
  }
  return slugifyHeading(textOf(children)) || undefined
}

interface CodeElementProps {
  className?: string
  children?: ReactNode
}

/** 复制按钮：clipboard 不可用（非安全上下文 / 权限被拒）时不假装成功。 */
function CopyCodeButton({ code }: { code: string }) {
  const { t, i18n } = useTranslation()
  const [copied, setCopied] = useState(false)
  const timer = useRef<number | null>(null)
  const english = (i18n.resolvedLanguage ?? i18n.language ?? "").startsWith(
    "en"
  )
  const copyLabel = t("site.copyCode", {
    defaultValue: english ? "Copy code" : "复制代码",
  })
  const copiedLabel = t("site.copiedCode", {
    defaultValue: english ? "Copied" : "已复制",
  })

  useEffect(
    () => () => {
      if (timer.current !== null) {
        window.clearTimeout(timer.current)
      }
    },
    []
  )

  async function copy() {
    if (!navigator.clipboard) {
      return
    }
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      if (timer.current !== null) {
        window.clearTimeout(timer.current)
      }
      timer.current = window.setTimeout(() => setCopied(false), 2000)
    } catch {
      // 权限被拒或非安全上下文：保持原文，不把失败显示成已复制。
    }
  }

  const label = copied ? copiedLabel : copyLabel
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-xs"
      aria-label={label}
      title={label}
      onClick={copy}
    >
      {copied ? <Check aria-hidden="true" /> : <Copy aria-hidden="true" />}
    </Button>
  )
}

/**
 * GitHub 风格的代码块外壳：语义底色 + 描边 + 圆角 + 语言标签栏 + 横向滚动。
 *
 * 语言名来自 fenced code 的 `language-xxx` 类；没有语言信息时不渲染标签栏。
 */
function CodeBlock({ children }: { children?: ReactNode }) {
  const codeElement = isValidElement<CodeElementProps>(children)
    ? children
    : null
  const className = codeElement?.props.className ?? ""
  const language = /(?:^|\s)language-([^\s]+)/.exec(className)?.[1] ?? ""
  const content = codeElement?.props.children ?? children
  // 复制时去掉 fenced code 末尾换行，粘贴进编辑器不会多出空行。
  const source = textOf(content).replace(/\n$/, "")

  return (
    <div className="overflow-hidden rounded-[12px] border border-border bg-muted">
      {language ? (
        <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-1.5">
          <span className="font-mono text-xs text-muted-foreground">
            {language}
          </span>
          <CopyCodeButton code={source} />
        </div>
      ) : null}
      <pre className="overflow-x-auto p-4 font-mono text-[13px] leading-[1.6] text-foreground">
        <code>{content}</code>
      </pre>
    </div>
  )
}

/**
 * Markdown 元素的排版：只使用语义 token，不引入 @tailwindcss/typography。
 *
 * 容器用 grid row-gap 控制块间距；标题、列表、引用、代码与表格通过后代选择器
 * 统一上色和定尺寸。代码块的外壳由 CodeBlock 负责，行内代码再单独与代码块区分开。
 */
const markdownClassName = cn(
  "grid min-w-0 gap-4 text-sm leading-7 break-words text-foreground/90 sm:text-base",
  // 标题阶梯；h2/h3 是目录锚点，滚动时给吸顶页头留出偏移。
  "[&_h1]:font-heading [&_h1]:text-2xl [&_h1]:font-semibold [&_h1]:tracking-tight [&_h1]:text-foreground",
  "[&_h2]:scroll-mt-24 [&_h2]:font-heading [&_h2]:text-xl [&_h2]:font-semibold [&_h2]:tracking-tight [&_h2]:text-foreground",
  "[&_h3]:scroll-mt-24 [&_h3]:font-heading [&_h3]:text-lg [&_h3]:font-semibold [&_h3]:text-foreground",
  "[&_h4]:text-base [&_h4]:font-semibold [&_h4]:text-foreground",
  "[&_h5]:text-sm [&_h5]:font-semibold [&_h5]:text-foreground",
  "[&_h6]:text-sm [&_h6]:font-semibold [&_h6]:text-muted-foreground",
  // 列表：Tailwind preflight 会清掉标记，这里恢复。
  "[&_ul]:list-disc [&_ul]:pl-6",
  "[&_ol]:list-decimal [&_ol]:pl-6",
  "[&_li]:pl-1",
  "[&_li+li]:mt-1",
  "[&_li>ol]:mt-1 [&_li>ul]:mt-1",
  // 引用与分隔线
  "[&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-4 [&_blockquote]:text-muted-foreground",
  "[&_blockquote]:space-y-2",
  "[&_hr]:border-border/60",
  // 行内代码：与代码块区分，底色/圆角/内边距都走语义 token。
  "[&_:not(pre)>code]:rounded-[4px] [&_:not(pre)>code]:bg-muted [&_:not(pre)>code]:px-1.5 [&_:not(pre)>code]:py-0.5 [&_:not(pre)>code]:font-mono [&_:not(pre)>code]:text-[0.9em] [&_:not(pre)>code]:text-foreground",
  // GFM 表格：补齐 GitHub 表头观感（语义底色 + 单元格下边框）。
  "[&_th]:border-b [&_th]:border-border [&_th]:bg-muted [&_th]:px-3 [&_th]:py-2 [&_th]:text-left [&_th]:font-semibold [&_th]:text-foreground",
  "[&_td]:border-b [&_td]:border-border/60 [&_td]:px-3 [&_td]:py-2 [&_td]:align-top",
  "[&_tr:last-child_td]:border-b-0",
  // GFM 任务列表复选框
  "[&_li>input]:mr-1.5 [&_li>input]:align-middle",
  // 链接与强调
  "[&_a]:font-medium [&_a]:text-primary-ink [&_a]:underline [&_a]:underline-offset-4 [&_a]:hover:text-primary-hover",
  "[&_strong]:font-semibold [&_strong]:text-foreground",
  "[&_del]:text-muted-foreground",
  "[&_sub]:text-[0.75em] [&_sup]:text-[0.75em]"
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
    return <CodeBlock>{children}</CodeBlock>
  },
  code({ className, children }) {
    // 保留 language-xxx，CodeBlock 靠它显示语言标签；行内代码没有这个类。
    return <code className={cn("font-mono", className)}>{children}</code>
  },
  table({ children }) {
    return (
      <div className="overflow-x-auto rounded-xl border border-border">
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
  // 标题 id 与 SiteArticle 的目录共用同一份提取结果（按源码行号对齐）。
  const lineIds = useMemo(() => {
    const map = new Map<number, string>()
    for (const heading of extractHeadings(content)) {
      map.set(heading.line, heading.id)
    }
    return map
  }, [content])
  const components = useMemo<Components>(
    () => ({
      ...markdownComponents,
      h2({ children, node }) {
        return <h2 id={headingId(node, lineIds, children)}>{children}</h2>
      },
      h3({ children, node }) {
        return <h3 id={headingId(node, lineIds, children)}>{children}</h3>
      },
    }),
    [lineIds]
  )

  if (content.trim() === "") {
    return (
      <p className={cn("text-sm text-muted-foreground", className)}>
        {emptyText}
      </p>
    )
  }
  return (
    <div className={cn(markdownClassName, className)}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {content}
      </ReactMarkdown>
    </div>
  )
}
