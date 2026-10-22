export interface ResultLabelPair {
  key: string
  value: string
}

export interface ResultStatValue {
  field: string
  raw: string
  value: number
}

export interface ResultStatGroup {
  key: string
  label: string
  values: ResultStatValue[]
}

/** 日志和数值结果共用的时间格式：本地时区，毫秒精度。 */
export function formatResultTimestamp(timestamp: number): string {
  if (!timestamp) {
    return "—"
  }
  const date = new Date(timestamp)
  const pad = (value: number, size = 2) => String(value).padStart(size, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.${pad(date.getMilliseconds(), 3)}`
}

export function formatChartValue(value: number): string {
  if (!Number.isFinite(value)) {
    return "—"
  }
  return Number.isInteger(value)
    ? String(value)
    : String(Number(value.toPrecision(6)))
}

export function truncateLabel(label: string, max: number): string {
  return label.length > max ? `${label.slice(0, max - 1)}…` : label
}

export interface ResultChartPadding {
  top: number
  right: number
  bottom: number
  left: number
}

export interface ResultChartLayout {
  width: number
  height: number
  padding: ResultChartPadding
  plotWidth: number
  plotHeight: number
  rotateLabels: boolean
  labelAngle: number
  labelChars: number
  /** 横轴标签抽稀步长；结果行数很多时只画每第 N 个标签。 */
  labelStep: number
  /** 标签文字基线相对绘图区底部的偏移。 */
  labelOffset: number
}

// 轴标签固定 11px；字符宽度按数字、冒号、短横线的平均宽度估算，
// 不依赖 DOM 测量，首帧就能算出稳定的边距。
const CHART_FONT_SIZE = 11
const CHART_CHAR_WIDTH = 6.4
const CHART_LABEL_GAP = 8
const CHART_LABEL_PADDING = 4
const CHART_LABEL_DESCENT = 4
// 旋转标签保留“2026-10-23 01:05”的长度（日期 + 时分），时间桶才读得出；
// 完整标签仍在柱子的 title 里。
const CHART_ROTATED_LABEL_CHARS = 17
const CHART_ROTATE_DEGREES = 30
const CHART_ROTATE_THRESHOLD = 5
const CHART_MIN_LABEL_CHARS = 8
const CHART_MAX_LABEL_CHARS = 32
const CHART_MIN_PLOT_HEIGHT = 64
const CHART_MIN_WIDTH = 240

/**
 * 数值结果柱状图的几何布局：高度跟随容器铺满可用空间；横轴旋转标签按
 * 实际字数和角度预留底部边距，并为首个标签补足左侧留白；标签过密时
 * 抽稀，避免被 SVG 边缘裁切或相互重叠。
 */
export function resultChartLayout(
  width: number,
  height: number,
  labels: string[]
): ResultChartLayout {
  const chartWidth = Math.max(width, CHART_MIN_WIDTH)
  const rotateLabels = labels.length > CHART_ROTATE_THRESHOLD
  const labelAngle = rotateLabels ? CHART_ROTATE_DEGREES : 0
  const radians = (labelAngle * Math.PI) / 180
  const padding: ResultChartPadding = {
    top: 16,
    right: 16,
    bottom: 0,
    left: 56,
  }
  const count = Math.max(labels.length, 1)
  const longestLabel = labels.reduce(
    (max, label) => Math.max(max, label.length),
    1
  )
  const groupWidthAt = (left: number) =>
    Math.max(chartWidth - left - padding.right, 1) / count

  let labelChars: number
  let labelStep = 1
  if (rotateLabels) {
    labelChars = Math.min(longestLabel, CHART_ROTATED_LABEL_CHARS)
    const rotatedWidth = labelChars * CHART_CHAR_WIDTH
    // 首个标签的锚点在首组中心，旋转后向左下延伸，需要额外左侧留白。
    // 加 2px 余量，抵消左侧留白本身带来的分组宽度变化。
    const gutter = Math.ceil(
      rotatedWidth * Math.cos(radians) +
        CHART_LABEL_PADDING +
        2 -
        groupWidthAt(padding.left) / 2
    )
    padding.left = Math.max(
      padding.left,
      Math.min(gutter, Math.round(chartWidth * 0.45))
    )
    // 图表过窄时留白被截断，再按实际空间收窄标签，保证仍然完整可见。
    const firstCenter = padding.left + groupWidthAt(padding.left) / 2
    labelChars = Math.min(
      labelChars,
      Math.floor(
        Math.max(firstCenter - CHART_LABEL_PADDING, 1) /
          Math.cos(radians) /
          CHART_CHAR_WIDTH
      )
    )
    // 相邻标签的水平投影不能超过最终分组间距，放不下就抽稀标签。
    const requiredSpacing =
      CHART_LABEL_GAP + labelChars * CHART_CHAR_WIDTH * Math.cos(radians)
    labelStep = Math.max(
      1,
      Math.ceil(requiredSpacing / groupWidthAt(padding.left))
    )
  } else {
    // 标签水平居中，分组不多时可以放宽，但不能宽到互相压住。
    labelChars = Math.min(
      Math.max(
        Math.floor(groupWidthAt(padding.left) / (CHART_CHAR_WIDTH + 0.6)),
        CHART_MIN_LABEL_CHARS
      ),
      CHART_MAX_LABEL_CHARS
    )
  }

  const plotWidth = Math.max(chartWidth - padding.left - padding.right, 1)
  // 旋转后标签从锚点向左下延伸，垂直占用 = 文本宽度 × sin(角度)。
  const labelWidth = Math.min(longestLabel, labelChars) * CHART_CHAR_WIDTH
  const paddingBottom = Math.round(
    CHART_LABEL_GAP +
      (rotateLabels ? labelWidth * Math.sin(radians) : CHART_FONT_SIZE) +
      CHART_LABEL_DESCENT +
      CHART_LABEL_PADDING
  )
  padding.bottom = paddingBottom
  // 容器再矮也要给绘图区和标签留出最小高度，超出部分由卡片滚动。
  const chartHeight = Math.max(
    height,
    padding.top + paddingBottom + CHART_MIN_PLOT_HEIGHT
  )
  return {
    width: chartWidth,
    height: chartHeight,
    padding,
    plotWidth,
    plotHeight: Math.max(chartHeight - padding.top - paddingBottom, 1),
    rotateLabels,
    labelAngle,
    labelChars,
    labelStep,
    labelOffset: rotateLabels
      ? CHART_LABEL_GAP
      : CHART_LABEL_GAP + CHART_FONT_SIZE,
  }
}
