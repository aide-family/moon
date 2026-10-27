import { useTranslation } from "react-i18next"

import error502Illustration from "@/assets/undraw/error-502.svg"
import error503Illustration from "@/assets/undraw/error-503.svg"
import { ErrorPage } from "@/components/error-page"
import { Button } from "@/components/ui/button"

interface ServerErrorPageProps {
  /** 503 表示服务维护/不可用；其余后端不可用场景沿用 502。 */
  status?: 502 | 503
  // 后端返回的错误摘要，只用于辅助定位，不替代标准文案。
  detail?: string
  onRetry?: () => void
  // 默认整页展示（会话加载失败）；在应用外壳内容区展示时传 false。
  fullPage?: boolean
}

/**
 * 后端不可用时的标准错误页。页面级会话加载失败会直接覆盖整个应用，
 * 不再显示零散的“加载失败 + 重试”提示。
 */
export function ServerErrorPage({
  status = 502,
  detail,
  onRetry,
  fullPage = true,
}: ServerErrorPageProps) {
  const { t } = useTranslation()
  const unavailable = status === 503

  return (
    <ErrorPage
      fullPage={fullPage}
      status={status}
      label={unavailable ? "Service Unavailable" : "Bad Gateway"}
      title={t(
        unavailable ? "error.serviceUnavailableTitle" : "error.badGatewayTitle"
      )}
      description={t(
        unavailable
          ? "error.serviceUnavailableDescription"
          : "error.badGatewayDescription"
      )}
      image={unavailable ? error503Illustration : error502Illustration}
      detail={detail}
    >
      {onRetry ? (
        <Button variant="outline" onClick={onRetry}>
          {t("action.retry")}
        </Button>
      ) : null}
    </ErrorPage>
  )
}
