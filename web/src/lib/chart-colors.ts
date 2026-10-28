/**
 * 图表色只返回 CSS 变量名，不返回 hex。
 *
 * 这样切换品牌/明暗时变量即时生效，图表不需要重渲染，也不需要把主题状态
 * 传进每个图表面板。语义层负责保证每个 --chart-N 对当前底色 ≥3:1。
 */
const CHART_TOKENS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
  "var(--chart-6)",
  "var(--chart-7)",
  "var(--chart-8)",
]

/** 按序列序号循环取色，图表和图例共用同一套颜色。 */
export function chartColor(index: number): string {
  return CHART_TOKENS[index % CHART_TOKENS.length]
}
