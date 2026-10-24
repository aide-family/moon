// 历史告警的日期语义只有一处实现：选择器和 URL 里都是本地自然日。
// 查询时下界补到当日 00:00:00（包含），上界补到当日 23:59:59.999999999（排除），
// 这样上界落在所选当天而不是次日零点，同时覆盖当天所有带小数秒的事件。

import { addDays, parseDateValue } from "@/lib/date-range"

// 单次查询允许的最大天数（含首尾）。历史明细和分析都按这个上限选择日期，
// 超过 30 天的区间会让归档周表扫描和浏览器侧聚合同时失控。
export const HISTORY_RANGE_MAX_DAYS = 30

/** 本地自然日的毫秒区间 [start, end)，供分析视图按日分桶时定界。 */
export function localDayWindow(
  value: string
): { start: number; end: number } | undefined {
  const date = parseDateValue(value)
  if (!date) {
    return undefined
  }
  return { start: date.getTime(), end: addDays(date, 1).getTime() }
}
