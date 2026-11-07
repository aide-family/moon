import { useEffect, useRef } from "react"

import { cn } from "@/lib/class-names"

interface DeployConsoleProps {
  /** 终端标题栏文案，通常带上机器名。 */
  title: string
  /** 服务端追加的安装日志。 */
  log: string
  /** 还在安装中：末尾显示闪烁光标。 */
  running: boolean
  /** 日志已被服务端截断。 */
  truncated: boolean
  /** 还没有日志时的占位文案。 */
  empty: string
  truncatedNotice: string
}

/**
 * 安装日志按终端的样子呈现：等宽字体、命令与步骤分行着色、新日志贴到底部。
 * 日志本身很长，滚动只发生在终端内部；DialogBody 保持不滚动，头尾按钮始终可见。
 */

/** 按前缀上色：==> 步骤、!!! 失败、$ 命令，其余是命令输出。 */
function lineTone(line: string): string {
  if (line.startsWith("!!!")) {
    return "text-danger"
  }
  if (line.startsWith("==>")) {
    return "text-info"
  }
  if (line.startsWith("$ ")) {
    return "text-muted-foreground"
  }
  return ""
}

export function DeployConsole({
  title,
  log,
  running,
  truncated,
  empty,
  truncatedNotice,
}: DeployConsoleProps) {
  const viewport = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = viewport.current
    if (!el) {
      return
    }
    // 终端语义是只往下追加：日志变化时贴到底部。
    el.scrollTop = el.scrollHeight
  }, [log, running])

  const lines = log ? log.replace(/\n+$/, "").split("\n") : []

  return (
    <div className="min-w-0 overflow-hidden rounded-lg border border-border/70 bg-[var(--code-bg)]">
      <div className="flex items-center gap-1.5 border-b border-border/70 px-3 py-1.5">
        <span aria-hidden className="size-2 rounded-full bg-danger" />
        <span aria-hidden className="size-2 rounded-full bg-warning" />
        <span aria-hidden className="size-2 rounded-full bg-success" />
        <span className="ml-2 min-w-0 truncate font-mono text-[11px] text-muted-foreground">
          {title}
        </span>
      </div>
      <div
        ref={viewport}
        data-testid="deploy-console"
        className="max-h-[46vh] min-h-28 overflow-auto p-3 font-mono text-xs leading-5 break-all whitespace-pre-wrap"
      >
        {lines.length === 0 ? (
          <p className="text-muted-foreground">{empty}</p>
        ) : (
          lines.map((line, index) => (
            <div key={index} className={cn("min-h-5", lineTone(line))}>
              {line}
            </div>
          ))
        )}
        {running ? (
          <span
            aria-hidden
            className="mt-1 inline-block h-3.5 w-1.5 animate-pulse bg-muted-foreground align-middle"
          />
        ) : null}
      </div>
      {truncated ? (
        <p className="border-t border-border/70 px-3 py-1.5 text-[11px] text-muted-foreground">
          {truncatedNotice}
        </p>
      ) : null}
    </div>
  )
}
