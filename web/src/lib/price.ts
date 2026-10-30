/** 把人民币「分」格式化成两位小数的「元」金额，供价格展示复用。 */
export function formatPriceCents(cents?: number): string {
  return ((cents ?? 0) / 100).toFixed(2)
}
