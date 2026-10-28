import { Check, Copy } from "lucide-react"
import { useEffect, useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { CopyIconButton } from "@/components/copy-button"
import { cn } from "@/lib/class-names"

/** 空值占位：显示破折号，同时给屏幕阅读器一个明确的「未设置」。 */
export function DetailEmpty() {
  const { t } = useTranslation()
  return (
    <span className="text-muted-foreground">
      <span aria-hidden="true">—</span>
      <span className="sr-only">{t("detail.empty")}</span>
    </span>
  )
}

interface DetailCopyTextProps {
  value: string
  /** 字段名，用于复制按钮的可访问标签，例如「命名空间」。 */
  label: string
  /** 标识符默认用等宽字体；版本号、类型名等可以关掉。 */
  mono?: boolean
  /** 弱化显示：用于附注类标识（例如镜像名）。 */
  muted?: boolean
  className?: string
}

/** 可复制的标识：IP、节点名、镜像、路径、命名空间等。 */
export function DetailCopyText({
  value,
  label,
  mono = true,
  muted = false,
  className,
}: DetailCopyTextProps) {
  const { t } = useTranslation()
  if (!value) {
    return <DetailEmpty />
  }
  return (
    <span
      className={cn(
        "group/detail-copy inline-flex max-w-full min-w-0 items-start gap-1",
        className
      )}
    >
      <span
        className={cn(
          "min-w-0 break-all",
          mono ? "font-mono text-xs leading-5" : "text-sm leading-relaxed",
          muted && "text-muted-foreground"
        )}
      >
        {value}
      </span>
      <CopyIconButton
        value={value}
        label={t("detail.copyField", { label })}
        copiedLabel={t("action.copied")}
        errorLabel={t("action.copyFailed")}
        className="-my-1 size-6 opacity-50 hover:opacity-100 focus-visible:opacity-100"
      />
    </span>
  )
}

interface DetailTagListProps {
  items: string[]
  /** 键名、主机名、镜像名等默认等宽。 */
  mono?: boolean
  /** 点击标签复制其内容；用于键名、主机、路径这类需要复制到别处的值。 */
  copyable?: boolean
  /** 复制按钮的可访问标签前缀，例如「键名」。 */
  copyLabel?: string
  className?: string
}

/** 标签组：角色、访问模式、键名、主机、镜像等可枚举的短值。 */
export function DetailTagList({
  items,
  mono = true,
  copyable = false,
  copyLabel,
  className,
}: DetailTagListProps) {
  if (items.length === 0) {
    return <DetailEmpty />
  }
  return (
    <ul className={cn("flex max-w-full min-w-0 flex-wrap gap-1", className)}>
      {items.map((item, index) => (
        <li key={`${item}-${index}`} className="max-w-full min-w-0">
          {copyable ? (
            <CopyTag value={item} label={copyLabel} mono={mono} />
          ) : (
            <span
              className={cn(
                "inline-flex max-w-full rounded-md border border-border/60 bg-muted/40 px-2 py-0.5 text-[11px] text-foreground",
                mono && "font-mono"
              )}
            >
              <span className="min-w-0 break-all">{item}</span>
            </span>
          )}
        </li>
      ))}
    </ul>
  )
}

function CopyTag({
  value,
  label,
  mono,
}: {
  value: string
  label?: string
  mono: boolean
}) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!copied) {
      return
    }
    const timer = window.setTimeout(() => setCopied(false), 1600)
    return () => window.clearTimeout(timer)
  }, [copied])

  async function copy() {
    try {
      await navigator.clipboard.writeText(value)
      setCopied(true)
    } catch {
      toast.error(t("action.copyFailed"))
    }
  }

  return (
    <button
      type="button"
      className={cn(
        "group/tag inline-flex max-w-full items-center gap-1 rounded-md border border-border/60 bg-muted/40 px-1.5 py-0.5 text-[11px] text-foreground transition-colors hover:border-border hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
        mono && "font-mono"
      )}
      aria-label={t("detail.copyField", { label: label ?? value })}
      title={t("detail.copyField", { label: label ?? value })}
      data-copied={copied}
      onClick={() => void copy()}
    >
      <span className="min-w-0 break-all">{value}</span>
      {copied ? (
        <Check
          aria-hidden="true"
          className="size-3 shrink-0 text-emerald-600 dark:text-emerald-400"
        />
      ) : (
        <Copy
          aria-hidden="true"
          className="size-3 shrink-0 text-muted-foreground/60 transition-opacity group-hover/tag:text-muted-foreground"
        />
      )}
    </button>
  )
}

export interface DetailItem {
  key: string
  /** 主标识，例如容器名、端口映射、转发规则。 */
  title: ReactNode
  /** 次级说明，例如镜像地址；统一按等宽弱化排版。 */
  subtitle?: ReactNode
  /** 右侧状态位，例如就绪/重启次数。 */
  status?: ReactNode
  /** 底部元信息，用 · 分隔的短句。 */
  meta?: ReactNode
}

/** 结构化子项列表：Pod 容器、Service 端口、Ingress 转发规则。 */
export function DetailItemList({
  items,
  className,
}: {
  items: DetailItem[]
  className?: string
}) {
  if (items.length === 0) {
    return <DetailEmpty />
  }
  return (
    <ul className={cn("grid min-w-0 gap-1.5", className)}>
      {items.map((item) => (
        <li
          key={item.key}
          className="grid min-w-0 gap-1 rounded-lg border border-border/60 bg-muted/20 px-2.5 py-2"
        >
          <div className="flex min-w-0 flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
            <div className="min-w-0 text-sm leading-snug">{item.title}</div>
            {item.status ? (
              <div className="shrink-0 text-xs text-muted-foreground">
                {item.status}
              </div>
            ) : null}
          </div>
          {item.subtitle ? (
            <div className="min-w-0 font-mono text-[11px] leading-4 break-all text-muted-foreground">
              {item.subtitle}
            </div>
          ) : null}
          {item.meta ? (
            <div className="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-0.5 text-[11px] text-muted-foreground">
              {item.meta}
            </div>
          ) : null}
        </li>
      ))}
    </ul>
  )
}

/** 布尔字段的紧凑表达：true 用主色点，false 保持弱化，避免两行纯文本。 */
export function DetailFlag({
  value,
  trueLabel,
  falseLabel,
}: {
  value: boolean
  trueLabel: string
  falseLabel: string
}) {
  return (
    <span className="inline-flex items-center gap-1.5 text-sm">
      <span
        aria-hidden="true"
        className={cn(
          "size-1.5 rounded-full",
          value ? "bg-warning" : "bg-muted-foreground/40"
        )}
      />
      {value ? trueLabel : falseLabel}
    </span>
  )
}
