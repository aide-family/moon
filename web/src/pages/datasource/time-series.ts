export interface TimeSeriesApiResponse {
  status: string
  data?: {
    resultType?: string
    result?: unknown
  }
  error?: string
  errorType?: string
}

export interface TimeSeriesGraphSeries {
  name: string
  points: { time: number; value: number }[]
}

interface MatrixResult {
  metric?: Record<string, string>
  values?: [number, string][]
}

export function formatTimeSeriesName(metric: Record<string, string> = {}): string {
  const entries = Object.entries(metric)
  if (entries.length === 0) {
    return "{}"
  }
  return `{${entries.map(([key, value]) => `${key}="${value}"`).join(", ")}}`
}

/** Converts a numeric matrix to graph series, independent of Prometheus or Loki transport. */
export function toTimeSeriesGraphSeries(
  response: TimeSeriesApiResponse,
): TimeSeriesGraphSeries[] {
  if (response.status !== "success" || response.data?.resultType !== "matrix") {
    return []
  }

  const result = response.data.result
  if (!Array.isArray(result)) {
    return []
  }
  return (result as MatrixResult[]).map((item, index) => ({
    name: formatTimeSeriesName(item.metric) || `series-${index}`,
    points: (item.values ?? [])
      .map(([time, value]) => ({ time: Number(time), value: Number.parseFloat(value) }))
      .filter((point) => Number.isFinite(point.time) && Number.isFinite(point.value)),
  }))
}

/** Parse the shared matrix/vector envelope returned by metric and LogQL queries. */
export function parseTimeSeriesBody(body: string): TimeSeriesApiResponse | null {
  try {
    return JSON.parse(body) as TimeSeriesApiResponse
  } catch {
    return null
  }
}
