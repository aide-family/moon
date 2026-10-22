import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  toTimeSeriesGraphSeries,
  type TimeSeriesApiResponse,
} from "@/pages/datasource/time-series"
import { cn } from "@/lib/class-names"

const CHART_COLORS = [
  "#1677ff",
  "#52c41a",
  "#faad14",
  "#eb2f96",
  "#13c2c2",
  "#722ed1",
  "#fa541c",
  "#2f54eb",
]

export interface MetricGraphDomain {
  start: number
  end: number
}

interface MetricGraphProps {
  response: TimeSeriesApiResponse | null
  emptyText?: string
  // 固定横轴范围，单位秒；事件图表用它保证时间轴从事件开始到今天。
  domain?: MetricGraphDomain
  // 只有一条数据时不显示图例，把宽度留给图表。
  showLegend?: boolean
}

function formatTooltipTime(ts: number): string {
  const date = new Date(ts * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function formatAxisTime(ts: number, span: number): string {
  const date = new Date(ts * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  if (span >= 24 * 3600) {
    return `${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
  }
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

function findNearestPoint(
  points: { time: number; value: number }[],
  targetTime: number
) {
  if (points.length === 0) {
    return null
  }
  let nearest = points[0]
  let minDiff = Math.abs(points[0].time - targetTime)
  for (let index = 1; index < points.length; index += 1) {
    const diff = Math.abs(points[index].time - targetTime)
    if (diff < minDiff) {
      minDiff = diff
      nearest = points[index]
    }
  }
  return nearest
}

function seriesColor(index: number): string {
  return CHART_COLORS[index % CHART_COLORS.length]
}

export function MetricGraph({
  response,
  emptyText,
  domain,
  showLegend = true,
}: MetricGraphProps) {
  const { t } = useTranslation()
  const containerRef = useRef<HTMLDivElement>(null)
  const [size, setSize] = useState({ width: 480, height: 280 })
  const [selected, setSelected] = useState<Set<string>>(() => new Set())
  const [hover, setHover] = useState<{
    time: number
    plotX: number
    items: { name: string; value: number; color: string }[]
  } | null>(null)

  const series = useMemo(
    () => (response ? toTimeSeriesGraphSeries(response) : []),
    [response]
  )
  const visibleSeries = useMemo(
    () =>
      selected.size === 0
        ? series
        : series.filter((item) => selected.has(item.name)),
    [series, selected]
  )
  const width = size.width
  const height = Math.max(size.height, 160)
  const padding = { top: 24, right: 20, bottom: 44, left: 56 }
  const plotW = Math.max(width - padding.left - padding.right, 1)
  const plotH = height - padding.top - padding.bottom

  useEffect(() => {
    const node = containerRef.current
    if (!node) {
      return
    }
    const observer = new ResizeObserver((entries) => {
      const box = entries[0]?.contentRect
      if (!box) {
        return
      }
      setSize({
        width: Math.max(box.width, 240),
        height: Math.max(box.height, 160),
      })
    })
    observer.observe(node)
    return () => observer.disconnect()
  }, [])

  const chartMeta = useMemo(() => {
    const allPoints = visibleSeries.flatMap((item) => item.points)
    if (allPoints.length === 0) {
      return null
    }
    const minTime = domain
      ? domain.start
      : Math.min(...allPoints.map((point) => point.time))
    const maxTime = domain
      ? domain.end
      : Math.max(...allPoints.map((point) => point.time))
    const minVal = Math.min(...allPoints.map((point) => point.value))
    const maxVal = Math.max(...allPoints.map((point) => point.value))
    const valSpan = maxVal - minVal || 1
    const timeSpan = maxTime - minTime || 1
    return {
      minTime,
      maxTime,
      minVal,
      valSpan,
      timeSpan,
      x: (time: number) => padding.left + ((time - minTime) / timeSpan) * plotW,
      y: (value: number) =>
        padding.top + plotH - ((value - minVal) / valSpan) * plotH,
    }
  }, [visibleSeries, domain, plotW, plotH, padding.left, padding.top])

  function toggleSeries(name: string) {
    setSelected((current) => {
      const next = new Set(current)
      if (next.has(name)) {
        next.delete(name)
      } else {
        next.add(name)
      }
      return next
    })
  }

  function seriesShown(name: string) {
    return selected.size === 0 || selected.has(name)
  }

  if (!response || series.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {emptyText || t("datasource.queryGraphEmpty")}
      </p>
    )
  }

  const yTicks = chartMeta
    ? Array.from({ length: 6 }, (_, index) => {
        const ratio = index / 5
        return {
          value: chartMeta.minVal + chartMeta.valSpan * (1 - ratio),
          py: padding.top + plotH * ratio,
        }
      })
    : []
  const xLabels = chartMeta
    ? Array.from({ length: 5 }, (_, index) => {
        const ratio = index / 4
        return {
          time: chartMeta.minTime + chartMeta.timeSpan * ratio,
          x: padding.left + plotW * ratio,
        }
      })
    : []

  return (
    <div className="flex min-h-0 min-w-0 flex-1 gap-3">
      <div
        ref={containerRef}
        className="relative min-h-0 min-w-0 flex-1 overflow-hidden rounded-lg border bg-muted/20"
      >
        {hover && chartMeta ? (
          <div
            className="pointer-events-none absolute z-10 max-w-[min(100%,20rem)] rounded-md border bg-background px-3 py-2 text-xs shadow-sm"
            style={{
              left: Math.min(Math.max(hover.plotX - 120, 8), width - 260),
              top: 8,
            }}
          >
            <div className="mb-1.5 font-medium">
              {formatTooltipTime(Math.round(hover.time))}
            </div>
            <div className="flex max-h-40 flex-col gap-1 overflow-auto">
              {hover.items.map((item) => (
                <div
                  key={item.name}
                  className="flex min-w-0 items-center gap-2"
                >
                  <span
                    className="size-2 shrink-0 rounded-full"
                    style={{ backgroundColor: item.color }}
                  />
                  <span className="min-w-0 flex-1 truncate text-muted-foreground">
                    {item.name}
                  </span>
                  <span className="shrink-0 font-mono">{item.value}</span>
                </div>
              ))}
            </div>
          </div>
        ) : null}
        {chartMeta ? (
          <svg
            viewBox={`0 0 ${width} ${height}`}
            className="block w-full select-none"
            role="img"
            aria-label={t("datasource.queryGraphAria")}
          >
            {yTicks.map((tick) => (
              <g key={tick.py}>
                <line
                  x1={padding.left}
                  y1={tick.py}
                  x2={width - padding.right}
                  y2={tick.py}
                  className="stroke-border"
                  strokeDasharray="4 4"
                />
                <text
                  x={padding.left - 8}
                  y={tick.py + 4}
                  textAnchor="end"
                  fontSize={11}
                  className="fill-muted-foreground"
                >
                  {tick.value.toPrecision(4)}
                </text>
              </g>
            ))}
            {xLabels.map((label) => (
              <text
                key={label.x}
                x={label.x}
                y={height - 12}
                textAnchor="middle"
                fontSize={11}
                className="fill-muted-foreground"
              >
                {formatAxisTime(label.time, chartMeta.timeSpan)}
              </text>
            ))}
            <line
              x1={padding.left}
              y1={padding.top + plotH}
              x2={width - padding.right}
              y2={padding.top + plotH}
              className="stroke-border"
            />
            <line
              x1={padding.left}
              y1={padding.top}
              x2={padding.left}
              y2={padding.top + plotH}
              className="stroke-border"
            />
            {hover ? (
              <line
                x1={hover.plotX}
                y1={padding.top}
                x2={hover.plotX}
                y2={padding.top + plotH}
                className="stroke-muted-foreground"
                strokeDasharray="4 2"
              />
            ) : null}
            {series.map((item, index) => {
              if (!seriesShown(item.name) || item.points.length === 0) {
                return null
              }
              const path = item.points
                .map((point, pointIndex) => {
                  const command = pointIndex === 0 ? "M" : "L"
                  return `${command}${chartMeta.x(point.time)} ${chartMeta.y(point.value)}`
                })
                .join(" ")
              return (
                <path
                  key={item.name}
                  d={path}
                  fill="none"
                  stroke={seriesColor(index)}
                  strokeWidth={1.5}
                />
              )
            })}
            <rect
              x={padding.left}
              y={padding.top}
              width={plotW}
              height={plotH}
              fill="transparent"
              onMouseMove={(event) => {
                const svg = event.currentTarget.ownerSVGElement
                if (!svg) {
                  return
                }
                const rect = svg.getBoundingClientRect()
                const scaleX = width / rect.width
                const relX = (event.clientX - rect.left) * scaleX - padding.left
                if (relX < 0 || relX > plotW) {
                  setHover(null)
                  return
                }
                const targetTime =
                  chartMeta.minTime + (relX / plotW) * chartMeta.timeSpan
                const items = series
                  .map((item, index) => {
                    if (!seriesShown(item.name)) {
                      return null
                    }
                    const nearest = findNearestPoint(item.points, targetTime)
                    if (!nearest) {
                      return null
                    }
                    return {
                      name: item.name,
                      value: nearest.value,
                      color: seriesColor(index),
                    }
                  })
                  .filter(
                    (item): item is NonNullable<typeof item> => item != null
                  )
                setHover({
                  time: targetTime,
                  plotX: padding.left + relX,
                  items,
                })
              }}
              onMouseLeave={() => setHover(null)}
            />
          </svg>
        ) : (
          <div className="flex h-full min-h-40 items-center justify-center px-4">
            <p className="text-sm text-muted-foreground">
              {emptyText || t("datasource.queryGraphEmpty")}
            </p>
          </div>
        )}
      </div>
      {showLegend ? (
        <TooltipProvider delayDuration={1500} skipDelayDuration={0}>
          <div className="flex h-full w-48 shrink-0 flex-col overflow-hidden rounded-lg border">
            <p className="shrink-0 px-2 py-1.5 text-xs font-medium text-muted-foreground">
              {t("datasource.legend")}
            </p>
            <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-y-auto p-2 pt-0">
              {series.map((item, index) => {
                const color = seriesColor(index)
                const dimmed = selected.size > 0 && !selected.has(item.name)
                return (
                  <Tooltip key={item.name}>
                    <TooltipTrigger asChild>
                      <button
                        type="button"
                        className={cn(
                          "flex w-full items-center gap-2 rounded-md px-1.5 py-1 text-left text-xs hover:bg-muted/60",
                          dimmed && "opacity-40"
                        )}
                        onClick={() => toggleSeries(item.name)}
                      >
                        <span
                          className="size-2.5 shrink-0 rounded-sm"
                          style={{ backgroundColor: color }}
                        />
                        <span className="min-w-0 flex-1 truncate font-mono">
                          {item.name}
                        </span>
                      </button>
                    </TooltipTrigger>
                    <TooltipContent>{item.name}</TooltipContent>
                  </Tooltip>
                )
              })}
            </div>
          </div>
        </TooltipProvider>
      ) : null}
    </div>
  )
}
