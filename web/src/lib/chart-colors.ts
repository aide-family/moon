export const CHART_COLORS = [
  "#1677ff",
  "#52c41a",
  "#faad14",
  "#eb2f96",
  "#13c2c2",
  "#722ed1",
  "#fa541c",
  "#2f54eb",
]

/** 按序列序号循环取色，图表和图例共用同一套颜色。 */
export function chartColor(index: number): string {
  return CHART_COLORS[index % CHART_COLORS.length]
}
