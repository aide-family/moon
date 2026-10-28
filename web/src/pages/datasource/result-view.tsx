import { ChevronRight } from "lucide-react"
import { Fragment, useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { chartColor } from "@/lib/chart-colors"
import { cn } from "@/lib/class-names"
import {
  formatChartValue,
  resultChartLayout,
  truncateLabel,
  type ResultLabelPair,
  type ResultStatGroup,
} from "@/pages/datasource/result-data"

export interface ResultLogRow {
  key: string
  time: string
  line: string
  labels: ResultLabelPair[]
}

function fieldText(field: string): string {
  return field
}

// LogResultTable 只保留时间与日志正文，其余标签折叠到行展开区，
// 避免每条日志的 app/env 等字段把表格撑出大量列。
export function LogResultTable({ rows }: { rows: ResultLogRow[] }) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState<Set<number>>(() => new Set())
  const stickyHead =
    "sticky top-0 z-20 bg-background shadow-[inset_0_-1px_0_0_hsl(var(--border))]"

  if (rows.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.logsEmpty")}
      </p>
    )
  }

  function toggleRow(index: number) {
    setExpanded((current) => {
      const next = new Set(current)
      if (next.has(index)) {
        next.delete(index)
      } else {
        next.add(index)
      }
      return next
    })
  }

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-auto rounded-lg border">
      <Table className="w-full">
        <TableHeader>
          <TableRow>
            <TableHead className={`${stickyHead} w-48 whitespace-nowrap`}>
              {t("datasource.resultTime")}
            </TableHead>
            <TableHead className={`${stickyHead} whitespace-nowrap`}>
              {t("datasource.logsLine")}
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row, index) => {
            const open = expanded.has(index)
            return (
              <Fragment key={`${row.key}-${index}`}>
                <TableRow
                  aria-expanded={open}
                  tabIndex={0}
                  className="cursor-pointer align-top"
                  onClick={() => toggleRow(index)}
                  onKeyDown={(event) => {
                    if (event.key !== "Enter" && event.key !== " ") {
                      return
                    }
                    event.preventDefault()
                    toggleRow(index)
                  }}
                >
                  <TableCell className="align-top font-mono text-xs whitespace-nowrap text-muted-foreground">
                    {row.time || "—"}
                  </TableCell>
                  <TableCell className="align-top font-mono text-xs">
                    <div className="flex items-start gap-1.5">
                      <ChevronRight
                        className={cn(
                          "mt-0.5 size-3.5 shrink-0 text-muted-foreground transition-transform",
                          open && "rotate-90"
                        )}
                      />
                      <pre className="min-w-0 break-all whitespace-pre-wrap">
                        {row.line || "—"}
                      </pre>
                    </div>
                  </TableCell>
                </TableRow>
                {open ? (
                  <TableRow className="bg-muted/30 hover:bg-muted/30">
                    <TableCell />
                    <TableCell className="py-2">
                      {row.labels.length > 0 ? (
                        <div className="flex flex-wrap gap-x-4 gap-y-1 font-mono text-xs">
                          {row.labels.map((pair, labelIndex) => (
                            <span
                              key={`${pair.key}-${labelIndex}`}
                              className="max-w-full break-all text-muted-foreground"
                            >
                              <span className="text-foreground">
                                {pair.key}
                              </span>
                              ={pair.value || "—"}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-xs text-muted-foreground">—</span>
                      )}
                    </TableCell>
                  </TableRow>
                ) : null}
              </Fragment>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}

// StatChart 用与指标图一致的 SVG 画法渲染数值结果：每行一个分组，
// 每个数值字段一根柱子，多个字段时分组内并排。
export function StatChart({
  groups,
  fields,
  fieldLabel = fieldText,
}: {
  groups: ResultStatGroup[]
  fields: string[]
  fieldLabel?: (field: string) => string
}) {
  const { t } = useTranslation()
  const containerRef = useRef<HTMLDivElement>(null)
  // 图表随容器拉伸：宽高都跟随卡片剩余空间，窗口缩放和结果行数变化后
  // 都由 ResizeObserver 重新计算。
  const [size, setSize] = useState({ width: 640, height: 320 })

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
      const width = Math.max(box.width, 240)
      const height = Math.max(box.height, 0)
      setSize((current) =>
        current.width === width && current.height === height
          ? current
          : { width, height }
      )
    })
    observer.observe(node)
    return () => observer.disconnect()
  }, [])

  const labels = useMemo(() => groups.map((group) => group.label), [groups])
  const layout = useMemo(
    () => resultChartLayout(size.width, size.height, labels),
    [size, labels]
  )

  if (groups.length === 0 || fields.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.queryEmpty")}
      </p>
    )
  }

  const { padding, plotWidth, plotHeight } = layout
  const plotBottom = padding.top + plotHeight
  const labelY = plotBottom + layout.labelOffset
  const values = groups
    .flatMap((group) => group.values.map((item) => item.value))
    .filter((value) => Number.isFinite(value))
  const minValue = Math.min(0, ...values)
  const maxValue = Math.max(0, ...values)
  const span = maxValue - minValue || 1
  const y = (value: number) =>
    padding.top + plotHeight - ((value - minValue) / span) * plotHeight
  const groupWidth = plotWidth / Math.max(groups.length, 1)
  const barWidth = Math.max(2, (groupWidth * 0.68) / Math.max(fields.length, 1))
  // 高度很小时减少刻度数量，避免 y 轴数字挤在一起。
  const yTickCount = Math.min(5, Math.max(2, Math.floor(plotHeight / 30) + 1))
  const yTicks = Array.from({ length: yTickCount }, (_, index) => {
    const value = minValue + (span * index) / (yTickCount - 1)
    return { value, py: y(value) }
  })

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-auto rounded-lg border bg-muted/20">
      {fields.length > 1 ? (
        <div className="flex shrink-0 flex-wrap items-center gap-3 border-b px-3 py-1.5 text-xs">
          {fields.map((field, index) => (
            <span
              key={field}
              className="flex items-center gap-1.5 text-muted-foreground"
            >
              <span
                className="size-2.5 shrink-0 rounded-sm"
                style={{ backgroundColor: chartColor(index) }}
              />
              <span className="font-mono">{fieldLabel(field)}</span>
            </span>
          ))}
        </div>
      ) : null}
      <div ref={containerRef} className="min-h-0 min-w-0 flex-1 p-2">
        <svg
          viewBox={`0 0 ${layout.width} ${layout.height}`}
          width={layout.width}
          height={layout.height}
          preserveAspectRatio="none"
          style={{ width: "100%", height: layout.height }}
          className="block select-none"
          role="img"
          aria-label={t("datasource.statsChartAria")}
        >
          {yTicks.map((tick) => (
            <g key={tick.py}>
              <line
                x1={padding.left}
                y1={tick.py}
                x2={layout.width - padding.right}
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
                {formatChartValue(tick.value)}
              </text>
            </g>
          ))}
          {groups.map((group, groupIndex) => {
            const groupX = padding.left + groupIndex * groupWidth
            const barsWidth = barWidth * group.values.length
            const barStart = groupX + (groupWidth - barsWidth) / 2
            const centerX = groupX + groupWidth / 2
            return (
              <g key={group.key}>
                {group.values.map((item, fieldIndex) => {
                  if (!Number.isFinite(item.value)) {
                    return null
                  }
                  const valueY = y(item.value)
                  const baseY = y(0)
                  const barY = Math.min(valueY, baseY)
                  const barH = Math.max(Math.abs(baseY - valueY), 1)
                  const barGap = Math.min(barWidth * 0.2, 4)
                  return (
                    <rect
                      key={item.field}
                      x={barStart + fieldIndex * barWidth + barGap / 2}
                      y={barY}
                      width={Math.max(barWidth - barGap, 1)}
                      height={barH}
                      rx={2}
                      style={{ fill: chartColor(fieldIndex) }}
                    >
                      <title>
                        {`${group.label} · ${fieldLabel(item.field)}=${
                          item.raw || "—"
                        }`}
                      </title>
                    </rect>
                  )
                })}
                {groupIndex % layout.labelStep === 0 ? (
                  <text
                    x={centerX}
                    y={labelY}
                    textAnchor={layout.rotateLabels ? "end" : "middle"}
                    fontSize={11}
                    className="fill-muted-foreground"
                    transform={
                      layout.rotateLabels
                        ? `rotate(-${layout.labelAngle} ${centerX} ${labelY})`
                        : undefined
                    }
                  >
                    {truncateLabel(group.label, layout.labelChars)}
                  </text>
                ) : null}
              </g>
            )
          })}
          <line
            x1={padding.left}
            y1={plotBottom}
            x2={layout.width - padding.right}
            y2={plotBottom}
            className="stroke-border"
          />
          <line
            x1={padding.left}
            y1={padding.top}
            x2={padding.left}
            y2={plotBottom}
            className="stroke-border"
          />
        </svg>
      </div>
    </div>
  )
}

export function SingleValueCard({
  value,
  fieldLabel,
  context,
}: {
  value: string
  fieldLabel: string
  context: string
}) {
  return (
    <div className="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 overflow-auto rounded-lg border bg-muted/20 p-6">
      <p className="text-xs text-muted-foreground">{fieldLabel}</p>
      <p className="font-mono text-4xl font-semibold tabular-nums">
        {value || "—"}
      </p>
      {context ? (
        <p className="max-w-full truncate text-xs text-muted-foreground">
          {context}
        </p>
      ) : null}
    </div>
  )
}

// StatResult 只有一个数值时显示单值；多个值（多行或多数值字段）画柱状图。
export function StatResult({
  groups,
  fields,
  fieldLabel = fieldText,
}: {
  groups: ResultStatGroup[]
  fields: string[]
  fieldLabel?: (field: string) => string
}) {
  if (groups.length === 1 && fields.length === 1) {
    const group = groups[0]
    const item = group.values[0]
    if (item && Number.isFinite(item.value)) {
      return (
        <SingleValueCard
          value={item.raw}
          fieldLabel={fieldLabel(item.field)}
          context={/^#\d+$/.test(group.label) ? "" : group.label}
        />
      )
    }
  }
  return <StatChart groups={groups} fields={fields} fieldLabel={fieldLabel} />
}
