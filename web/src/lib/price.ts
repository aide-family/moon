/** 支持的定价币种 → 展示符号。与后端 bo.Currencies() 保持一致。 */
export const CURRENCY_SYMBOLS: Record<string, string> = {
  CNY: "¥",
  USD: "$",
  EUR: "€",
  JPY: "¥",
}

/** 当前支持的币种编码，顺序与后端一致。 */
export const CURRENCIES = ["CNY", "USD", "EUR", "JPY"] as const

export type CurrencyCode = (typeof CURRENCIES)[number]

/** 返回币种符号；未知币种回落到编码本身。 */
export function currencySymbol(code?: string): string {
  const value = code || "CNY"
  return CURRENCY_SYMBOLS[value] ?? value
}

/** 把最小货币单位（分）格式化成带币种符号的金额，例如 ¥12.00 或 $9.90。 */
export function formatPriceCents(cents?: number, currency = "CNY"): string {
  return `${currencySymbol(currency)}${((cents ?? 0) / 100).toFixed(2)}`
}
