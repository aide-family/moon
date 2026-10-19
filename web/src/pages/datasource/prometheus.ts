export type PrometheusLabelSet = Record<string, string>

export interface PrometheusApiResponse {
  status: string
  data?: {
    resultType?: string
    result?: unknown
  }
  error?: string
  errorType?: string
}

interface PrometheusVectorResult {
  metric: PrometheusLabelSet
  value: [number, string]
}

interface PrometheusMatrixResult {
  metric: PrometheusLabelSet
  values: [number, string][]
}

export function formatSeriesName(metric: PrometheusLabelSet): string {
  const entries = Object.entries(metric ?? {})
  if (entries.length === 0) {
    return "{}"
  }
  return `{${entries.map(([key, value]) => `${key}="${value}"`).join(", ")}}`
}

export interface PrometheusTableRow {
  key: string
  series: string
  timestamp: number
  value: string
  labels: PrometheusLabelSet
}

function isVectorResult(
  item: PrometheusVectorResult | PrometheusMatrixResult,
): item is PrometheusVectorResult {
  return "value" in item
}

export function parsePrometheusBody(body: string): PrometheusApiResponse | null {
  try {
    return JSON.parse(body) as PrometheusApiResponse
  } catch {
    return null
  }
}

export function toTableRows(response: PrometheusApiResponse): PrometheusTableRow[] {
  if (response.status !== "success" || !response.data) {
    return []
  }

  const { resultType, result } = response.data

  if (resultType === "scalar" || resultType === "string") {
    const pair = result as [number, string]
    return [
      {
        key: "scalar",
        series: resultType,
        timestamp: pair[0],
        value: pair[1],
        labels: {},
      },
    ]
  }

  const items = (result ?? []) as (PrometheusVectorResult | PrometheusMatrixResult)[]
  const rows: PrometheusTableRow[] = []

  items.forEach((item, index) => {
    const series = formatSeriesName(item.metric)
    if (isVectorResult(item)) {
      rows.push({
        key: `${series}-${index}`,
        series,
        timestamp: item.value[0],
        value: item.value[1],
        labels: item.metric,
      })
      return
    }

    const values = item.values ?? []
    const latest = values[values.length - 1]
    if (!latest) {
      return
    }
    rows.push({
      key: `${series}-${index}`,
      series,
      timestamp: latest[0],
      value: latest[1],
      labels: item.metric,
    })
  })

  return rows
}

export interface PrometheusGraphSeries {
  name: string
  points: { time: number; value: number }[]
}

export function toGraphSeries(response: PrometheusApiResponse): PrometheusGraphSeries[] {
  if (response.status !== "success" || response.data?.resultType !== "matrix") {
    return []
  }

  return ((response.data.result ?? []) as PrometheusMatrixResult[]).map((item, index) => ({
    name: formatSeriesName(item.metric) || `series-${index}`,
    points: (item.values ?? [])
      .map(([time, value]) => ({
        time,
        value: Number.parseFloat(value),
      }))
      .filter((point) => Number.isFinite(point.value)),
  }))
}

export function collectLabelKeys(rows: PrometheusTableRow[]): string[] {
  const keys = new Set<string>()
  rows.forEach((row) => {
    Object.keys(row.labels).forEach((key) => keys.add(key))
  })
  return Array.from(keys).sort()
}

export function extractPrometheusResultData(response: PrometheusApiResponse): unknown {
  if (response.status === "error") {
    return { error: response.error, errorType: response.errorType }
  }
  return response.data?.result ?? response
}
