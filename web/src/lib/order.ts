import type { TFunction } from "i18next"

/** 订单状态文案。三种状态由后端枚举固定，未知值原样展示。 */
export function orderStatusLabel(t: TFunction, status: string): string {
  switch (status) {
    case "PENDING":
      return t("order.pending")
    case "PAID":
      return t("order.paid")
    case "CANCELLED":
      return t("order.cancelled")
    default:
      return status
  }
}

/** 订单状态标签颜色：待支付用 warning，已支付用 success，已取消用 neutral。 */
export function orderStatusTone(status: string): "warning" | "success" | "neutral" {
  switch (status) {
    case "PENDING":
      return "warning"
    case "PAID":
      return "success"
    default:
      return "neutral"
  }
}

/** 订单类型文案：新建租户 / 套餐升级 / 套餐续期。 */
export function orderKindLabel(t: TFunction, kind: string): string {
  switch (kind) {
    case "UPGRADE":
      return t("order.kindUpgrade")
    case "RENEW":
      return t("order.kindRenew")
    default:
      return t("order.kindCreate")
  }
}
