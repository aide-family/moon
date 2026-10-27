import { Inbox, Loader2, TriangleAlert } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/class-names"

interface ClusterResourceStateProps {
  state: "loading" | "error" | "empty"
  title: string
  /** 一句话说明「为什么空」或「下一步做什么」，可选。 */
  hint?: string
  onRetry?: () => void
}

/**
 * 资源列表的统一状态块：加载、错误、空结果。
 * 三种状态都占满列表视口并居中，避免空态只留一行文字贴在表格顶部。
 */
export function ClusterResourceState({
  state,
  title,
  hint,
  onRetry,
}: ClusterResourceStateProps) {
  const { t } = useTranslation()
  const Icon = state === "error" ? TriangleAlert : Inbox

  return (
    <div
      className="flex min-h-40 min-w-0 flex-1 flex-col items-center justify-center gap-2 px-6 py-10 text-center"
      role={state === "error" ? "alert" : "status"}
    >
      <span
        aria-hidden="true"
        className={cn(
          "grid size-9 place-items-center rounded-full",
          state === "error"
            ? "bg-destructive/10 text-destructive"
            : "bg-muted/70 text-muted-foreground"
        )}
      >
        {state === "loading" ? (
          <Loader2 className="size-4 animate-spin" />
        ) : (
          <Icon className="size-4" />
        )}
      </span>
      <p
        className={cn(
          "max-w-md text-sm break-words",
          state === "error" ? "text-destructive" : "text-foreground"
        )}
      >
        {title}
      </p>
      {hint ? (
        <p className="max-w-md text-xs leading-5 text-muted-foreground">
          {hint}
        </p>
      ) : null}
      {state === "error" && onRetry ? (
        <Button variant="outline" size="sm" onClick={onRetry}>
          {t("action.retry")}
        </Button>
      ) : null}
    </div>
  )
}
