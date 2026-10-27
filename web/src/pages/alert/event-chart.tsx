import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { getAlertEventChart } from "@/api/alert-event"
import type { AlertEventChart, AlertEventItem } from "@/api/alert-event/types"
import type { AlertRuleLevelItem } from "@/api/alert-rule/types"
import { CodeBlock } from "@/components/code-block"
import { MetricGraph } from "@/components/metric-graph"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { formErrorMessage } from "@/lib/form-error"
import {
  conditionSummary,
  useEventLevelCondition,
} from "@/pages/alert/level-condition"
import type { TimeSeriesApiResponse } from "@/pages/datasource/time-series"
import { parseTimeSeriesBody } from "@/pages/datasource/time-series"

interface AlertEventChartDialogProps {
  event: AlertEventItem | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function formatTime(value: number | string): string {
  const seconds = Number(value)
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "—"
  }
  return new Date(seconds * 1000).toLocaleString()
}

/**
 * 只保留与事件标签一致的序列。事件标签由样本标签、数据源标签和展示用元数据组成，
 * 因此判断序列时要求序列的每个标签都出现在事件标签里且值相同；多出来的元数据键
 * 不影响匹配。命中为空时返回空结果，不再回退展示表达式的全部序列。
 */
function filterSeriesByLabels(
  response: TimeSeriesApiResponse | null,
  labels: Record<string, string>
): TimeSeriesApiResponse | null {
  if (
    !response ||
    response.status !== "success" ||
    !response.data?.result ||
    Object.keys(labels).length === 0
  ) {
    return response
  }
  const result = response.data.result as { metric?: Record<string, string> }[]
  if (!Array.isArray(result)) {
    return response
  }
  const matched = result.filter((item) => {
    const metric = item.metric ?? {}
    return Object.entries(metric).every(([key, value]) => labels[key] === value)
  })
  return { ...response, data: { ...response.data, result: matched } }
}

export function AlertEventChartDialog({
  event,
  open,
  onOpenChange,
}: AlertEventChartDialogProps) {
  const { t } = useTranslation()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[80vh] flex-col gap-4 overflow-hidden sm:max-w-5xl">
        <DialogHeader>
          <DialogTitle>{t("alertEvent.eventChart")}</DialogTitle>
          <DialogDescription className="truncate">
            {event?.summary || event?.fingerprint}
          </DialogDescription>
        </DialogHeader>
        {open && event ? (
          <AlertEventChartBody key={event.uid} event={event} />
        ) : null}
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("action.close")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function AlertEventChartBody({
  event,
  compact = false,
  levelCondition,
}: {
  event: AlertEventItem
  compact?: boolean
  // 详情抽屉已经按事件取回命中的级别条件；不传则图表自己去取。
  levelCondition?: AlertRuleLevelItem | null
}) {
  const { t } = useTranslation()
  const [chart, setChart] = useState<AlertEventChart | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [tab, setTab] = useState("graph")
  const fetchedCondition = useEventLevelCondition(
    event.ruleUid,
    event.levelUid,
    levelCondition === undefined
  )
  const condition =
    levelCondition === undefined ? fetchedCondition : levelCondition

  useEffect(() => {
    let alive = true
    getAlertEventChart(event.uid)
      .then((result) => {
        if (alive) {
          setChart(result)
        }
      })
      .catch((err: unknown) => {
        if (alive) {
          setError(formErrorMessage(err, t("state.failed")))
        }
      })
      .finally(() => {
        if (alive) {
          setLoading(false)
        }
      })
    return () => {
      alive = false
    }
  }, [event.uid, t])

  const response = useMemo(
    () => (chart ? parseTimeSeriesBody(chart.body) : null),
    [chart]
  )
  const eventLabels = useMemo(() => {
    try {
      const parsed = JSON.parse(event.labels || "{}") as Record<string, string>
      return parsed && typeof parsed === "object" ? parsed : {}
    } catch {
      return {}
    }
  }, [event.labels])
  // 区间查询会返回表达式命中的全部序列；图表只保留事件标签对应的那一条。
  const chartResponse = useMemo(
    () => filterSeriesByLabels(response, eventLabels),
    [eventLabels, response]
  )
  const start = Number(chart?.start ?? 0)
  const end = Number(chart?.end ?? 0)
  const step = Number(chart?.step ?? 0)
  const raw = useMemo(() => {
    if (!chart) {
      return ""
    }
    const query = chartResponse
      ? chartResponse.status === "success"
        ? (chartResponse.data?.result ?? chartResponse)
        : { error: chartResponse.error, errorType: chartResponse.errorType }
      : chart.body
    return JSON.stringify(
      {
        event: {
          summary: event.summary,
          description: event.description,
          fingerprint: event.fingerprint,
          labels: event.labels,
          sampleValue: event.sampleValue,
          startsAt: event.startsAt,
        },
        query: { expr: chart.expr, start, end, step },
        result: query,
      },
      null,
      2
    )
  }, [chart, chartResponse, event, start, end, step])

  return (
    <>
      {chart ? (
        <div className="grid min-w-0 gap-1 text-sm">
          <div className="flex min-w-0 flex-wrap gap-x-6 gap-y-1 text-xs text-muted-foreground">
            <span className="min-w-0 truncate">
              {t("alertEvent.datasource")}:{" "}
              <span className="text-foreground">
                {chart.datasourceName || event.datasourceUid}
              </span>
            </span>
            <span className="min-w-0 truncate">
              {t("alertEvent.chartRange")}: {formatTime(start)} —{" "}
              {formatTime(end)} · {t("alertEvent.chartStep", { step })}
            </span>
          </div>
          {/* 实时详情里的图表也要能直接看到规则表达式，不再因为 compact 而隐藏。 */}
          <p className="min-w-0 font-mono text-xs break-words whitespace-pre-wrap text-muted-foreground">
            <span className="font-sans">{t("alertEvent.expr")}: </span>
            <span className="text-foreground">{chart.expr || "—"}</span>
          </p>
          {/* 表达式和命中的级别条件搭配展示：规则已删除或取不到条件时只留表达式。 */}
          {condition ? (
            <p className="min-w-0 font-mono text-xs break-words whitespace-pre-wrap text-muted-foreground">
              <span className="mr-2 font-sans">
                {t("alertEvent.levelCondition")}
              </span>
              {conditionSummary(t, condition, event.levelName)}
            </p>
          ) : null}
        </div>
      ) : null}
      {loading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!loading && error ? (
        <p className="text-sm text-destructive">{error}</p>
      ) : null}
      {!loading && !error && chart ? (
        <Tabs
          value={tab}
          onValueChange={setTab}
          className="flex min-h-0 flex-1 flex-col"
        >
          <TabsList className="shrink-0 self-start">
            <TabsTrigger value="graph">
              {t("alertEvent.chartGraph")}
            </TabsTrigger>
            <TabsTrigger value="raw">{t("alertEvent.chartRaw")}</TabsTrigger>
          </TabsList>
          <TabsContent value="graph" className="flex min-h-0 flex-1 flex-col">
            {response?.status === "error" ? (
              <p className="text-sm text-destructive">
                {response.error || t("state.failed")}
              </p>
            ) : (
              <div
                className={
                  compact
                    ? "flex h-56 min-h-0 flex-col"
                    : "flex min-h-[22rem] min-w-0 flex-1 flex-col"
                }
              >
                <MetricGraph
                  response={chartResponse}
                  domain={{ start, end }}
                  emptyText={t("alertEvent.chartNoMatch")}
                  showLegend={false}
                />
              </div>
            )}
          </TabsContent>
          <TabsContent value="raw" className="flex min-h-0 flex-1 flex-col">
            <CodeBlock
              code={raw}
              language="json"
              label={t("alertEvent.chartRaw")}
              copyable
              fill={!compact}
              maxHeightClass={compact ? "max-h-56" : undefined}
            />
          </TabsContent>
        </Tabs>
      ) : null}
    </>
  )
}
