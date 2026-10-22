import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

interface ServerErrorPageProps {
  // 后端返回的错误摘要，只用于辅助定位，不替代 502 的标准文案。
  detail?: string
  onRetry?: () => void
}

/**
 * 后端不可用时的标准 502 页面。页面级会话加载失败会直接覆盖整个应用，
 * 不再显示零散的“加载失败 + 重试”提示。
 */
export function ServerErrorPage({ detail, onRetry }: ServerErrorPageProps) {
  const { t } = useTranslation()

  return (
    <main className="flex min-h-svh flex-col items-center justify-center gap-6 bg-background px-6 py-16 text-center">
      <div className="flex items-center gap-4">
        <span className="font-heading text-6xl font-semibold tracking-tight text-primary">
          502
        </span>
        <span className="h-12 w-px bg-border" aria-hidden />
        <div className="text-left">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Bad Gateway
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {t("error.badGatewayTitle")}
          </p>
        </div>
      </div>
      <p className="max-w-md text-sm leading-6 text-muted-foreground">
        {t("error.badGatewayDescription")}
      </p>
      {detail ? (
        <code className="max-w-xl break-all rounded-md border bg-muted/50 px-3 py-2 text-left font-mono text-xs text-muted-foreground">
          {detail}
        </code>
      ) : null}
      {onRetry ? (
        <Button variant="outline" onClick={onRetry}>
          {t("action.retry")}
        </Button>
      ) : null}
    </main>
  )
}
