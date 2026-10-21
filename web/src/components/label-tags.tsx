import { useMemo } from "react"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

interface LabelTagsProps {
  labels: string
  className?: string
}

interface LabelEntry {
  key: string
  value: string
}

/**
 * 把标签 JSON 显示成 key=value 的 tag。值过长时在 tag 内换行，
 * 不撑出容器宽度；无法解析的内容按原样显示，避免详情变成空白。
 */
export function LabelTags({ labels, className }: LabelTagsProps) {
  const entries = useMemo(() => parseLabels(labels), [labels])

  if (entries === null) {
    return <span className="break-all font-mono text-xs">{labels || "—"}</span>
  }
  if (entries.length === 0) {
    return <span className="text-sm text-muted-foreground">—</span>
  }
  return (
    <div className={cn("flex min-w-0 max-w-full flex-wrap gap-1.5", className)}>
      {entries.map((entry) => (
        <Badge
          key={entry.key}
          variant="secondary"
          className="h-auto max-w-full min-w-0 justify-start rounded-md px-2 py-1 text-left font-mono text-xs leading-tight whitespace-normal"
        >
          <span className="min-w-0 break-all">
            <span className="text-muted-foreground">{entry.key}</span>
            <span>=</span>
            {entry.value}
          </span>
        </Badge>
      ))}
    </div>
  )
}

function parseLabels(labels: string): LabelEntry[] | null {
  try {
    const parsed: unknown = JSON.parse(labels || "{}")
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return null
    }
    return Object.entries(parsed as Record<string, unknown>)
      .filter(([key]) => key !== "")
      .map(([key, value]) => ({
        key,
        value: typeof value === "string" ? value : JSON.stringify(value),
      }))
      .sort((left, right) => left.key.localeCompare(right.key))
  } catch {
    return null
  }
}
