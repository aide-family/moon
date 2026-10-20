import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceMetadata, listDatasourceMetrics } from "@/api/datasource"
import type {
  DatasourceItem,
  DatasourceMetadataItem,
  DatasourceMetricItem,
} from "@/api/datasource/types"
import { BrandIcon } from "@/components/brand-icon"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { enumLabel } from "@/lib/enum-label"
import { failureText } from "@/lib/form-error"
import { cn } from "@/lib/utils"

function statusLabel(t: (key: string) => string, value: string): string {
  switch (value) {
    case "up":
      return t("datasource.statusUp")
    case "green":
      return t("datasource.statusGreen")
    case "yellow":
      return t("datasource.statusYellow")
    case "red":
      return t("datasource.statusRed")
    default:
      return value || t("datasource.statusUnknown")
  }
}

function statusClasses(value: string): string {
  switch (value) {
    case "up":
    case "green":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"
    case "yellow":
      return "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-400"
    case "red":
      return "border-red-500/30 bg-red-500/10 text-red-700 dark:text-red-400"
    default:
      return "border-border bg-muted text-muted-foreground"
  }
}

function statusDotClasses(value: string): string {
  switch (value) {
    case "up":
    case "green":
      return "bg-emerald-500"
    case "yellow":
      return "bg-amber-500"
    case "red":
      return "bg-red-500"
    default:
      return "bg-muted-foreground"
  }
}

function metadataFieldLabel(
  t: (key: string, options?: Record<string, string>) => string,
  key: string
): string {
  return t(`datasource.metadataField.${key}`, { defaultValue: key })
}

function useDriverMetadata(uid: string) {
  const { t } = useTranslation()
  const [metadata, setMetadata] = useState<Awaited<
    ReturnType<typeof getDatasourceMetadata>
  > | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState("")
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed("")
      try {
        const result = await getDatasourceMetadata(uid)
        if (active) {
          setMetadata(result)
        }
      } catch (err) {
        if (active) {
          setMetadata(null)
          setFailed(failureText(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [uid, t, attempt])

  return {
    metadata,
    loading,
    failed,
    retry: () => setAttempt((value) => value + 1),
  }
}

function MetadataFieldGrid({
  items,
  exclude,
}: {
  items: DatasourceMetadataItem[]
  exclude: string[]
}) {
  const { t } = useTranslation()
  const visible = items.filter((item) => !exclude.includes(item.key))
  if (visible.length === 0) {
    return null
  }
  return (
    <dl className="grid gap-x-6 gap-y-3 rounded-lg border bg-card/60 p-4 sm:grid-cols-2 xl:grid-cols-3">
      {visible.map((item) => (
        <div key={item.key} className="min-w-0">
          <dt className="text-xs text-muted-foreground">
            {metadataFieldLabel(t, item.key)}
          </dt>
          <dd
            className={cn(
              "mt-1 font-mono text-sm break-all",
              item.key === "revision" && "text-xs"
            )}
          >
            {item.value || "—"}
          </dd>
        </div>
      ))}
    </dl>
  )
}

function MetadataHeader({
  datasource,
  status,
}: {
  datasource: DatasourceItem
  status: string
}) {
  const { t } = useTranslation()
  return (
    <div className="flex flex-wrap items-center gap-3 rounded-lg border bg-muted/20 p-4">
      <BrandIcon
        driver={datasource.driver}
        className="size-10 rounded-lg border bg-background p-1.5"
      />
      <div className="min-w-0 flex-1">
        <p className="font-medium">{enumLabel(t, datasource.driver)}</p>
        <p className="truncate font-mono text-xs text-muted-foreground">
          {datasource.endpoint}
        </p>
      </div>
      {status ? (
        <span
          className={cn(
            "inline-flex shrink-0 items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-medium",
            statusClasses(status)
          )}
        >
          <span
            className={cn("size-1.5 rounded-full", statusDotClasses(status))}
            aria-hidden
          />
          {statusLabel(t, status)}
        </span>
      ) : null}
    </div>
  )
}

function ElasticsearchNodes({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const { metadata, loading, failed, retry } = useDriverMetadata(datasource.uid)
  const status =
    metadata?.items.find((item) => item.key === "status")?.value ?? ""

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
  }
  if (failed) {
    return (
      <div className="flex items-center gap-2">
        <p className="text-sm text-destructive">{failed}</p>
        <Button variant="outline" onClick={retry}>
          {t("action.retry")}
        </Button>
      </div>
    )
  }
  if (!metadata) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.metadataEmpty")}
      </p>
    )
  }

  const nodes = metadata.nodes ?? []
  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
      <MetadataHeader datasource={datasource} status={status} />
      <MetadataFieldGrid items={metadata.items ?? []} exclude={["status"]} />
      <div className="flex min-h-0 flex-1 flex-col gap-2">
        <h3 className="shrink-0 text-sm font-medium">
          {t("datasource.nodeAllocation")}
        </h3>
        {nodes.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            {t("datasource.nodeAllocationEmpty")}
          </p>
        ) : (
          <div className="min-h-0 flex-1 overflow-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.nodeName")}
                  </TableHead>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.nodeAddress")}
                  </TableHead>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.nodeRole")}
                  </TableHead>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.nodeShards")}
                  </TableHead>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.diskUsed")}
                  </TableHead>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.diskAvail")}
                  </TableHead>
                  <TableHead className="sticky top-0 z-10 bg-background">
                    {t("datasource.diskPercent")}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {nodes.map((node, index) => (
                  <TableRow key={`${node.name}-${index}`}>
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {node.name || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs whitespace-nowrap text-muted-foreground">
                      {node.address || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {node.role || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {node.shards || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {node.diskUsed || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {node.diskAvail || "—"}
                    </TableCell>
                    <TableCell className="font-mono text-xs whitespace-nowrap">
                      {node.diskPercent ? `${node.diskPercent}%` : "—"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        )}
      </div>
    </div>
  )
}

function LogsMetadata({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  const { metadata, loading, failed, retry } = useDriverMetadata(datasource.uid)
  const status =
    metadata?.items.find((item) => item.key === "status")?.value ?? ""

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
  }
  if (failed) {
    return (
      <div className="flex items-center gap-2">
        <p className="text-sm text-destructive">{failed}</p>
        <Button variant="outline" onClick={retry}>
          {t("action.retry")}
        </Button>
      </div>
    )
  }
  if (!metadata) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.metadataEmpty")}
      </p>
    )
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto">
      <MetadataHeader datasource={datasource} status={status} />
      <MetadataFieldGrid items={metadata.items ?? []} exclude={["status"]} />
    </div>
  )
}

function MetricsMetadata({ uid, driver }: { uid: string; driver: string }) {
  const { t } = useTranslation()
  const [metrics, setMetrics] = useState<DatasourceMetricItem[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState("")
  const [attempt, setAttempt] = useState(0)
  const isVictoriaMetrics = driver === "METRICS_VICTORIA_METRICS"

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed("")
      try {
        const result = await listDatasourceMetrics(uid)
        if (!active) {
          return
        }
        setMetrics(result.metrics ?? [])
      } catch (err) {
        if (active) {
          setMetrics([])
          setFailed(failureText(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [uid, t, attempt])

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
  }
  if (failed) {
    return (
      <div className="flex items-center gap-2">
        <p className="text-sm text-destructive">{failed}</p>
        <Button
          variant="outline"
          onClick={() => setAttempt((value) => value + 1)}
        >
          {t("action.retry")}
        </Button>
      </div>
    )
  }
  if (metrics.length === 0) {
    return (
      <div className="grid gap-2">
        {isVictoriaMetrics ? (
          <p className="text-sm text-muted-foreground">
            {t("datasource.vmMetadataNote")}
          </p>
        ) : null}
        <p className="text-sm text-muted-foreground">
          {t("datasource.metadataEmpty")}
        </p>
      </div>
    )
  }

  return (
    <TooltipProvider delayDuration={1500} skipDelayDuration={0}>
      <div className="flex min-h-0 flex-1 flex-col">
        {isVictoriaMetrics ? (
          <p className="shrink-0 pb-2 text-sm text-muted-foreground">
            {t("datasource.vmMetadataNote")}
          </p>
        ) : null}
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("datasource.metricName")}</TableHead>
              <TableHead>{t("datasource.metricHelp")}</TableHead>
              <TableHead>{t("datasource.metricType")}</TableHead>
              <TableHead>{t("datasource.metricUnit")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {metrics.map((item) => (
              <TableRow key={item.name}>
                <TableCell className="max-w-[16rem] truncate font-mono text-xs">
                  {item.name}
                </TableCell>
                <TableCell className="max-w-[24rem] text-muted-foreground">
                  {item.help ? (
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span className="block truncate">{item.help}</span>
                      </TooltipTrigger>
                      <TooltipContent className="max-w-sm">
                        {item.help}
                      </TooltipContent>
                    </Tooltip>
                  ) : (
                    "—"
                  )}
                </TableCell>
                <TableCell>{item.type || "—"}</TableCell>
                <TableCell>{item.unit || "—"}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </TooltipProvider>
  )
}

export function MetadataPanel({ datasource }: { datasource: DatasourceItem }) {
  const { t } = useTranslation()
  if (datasource.type === "METRICS") {
    return <MetricsMetadata uid={datasource.uid} driver={datasource.driver} />
  }
  if (datasource.type === "TRACE") {
    return (
      <p className="text-sm text-muted-foreground">
        {t("datasource.metadataEmpty")}
      </p>
    )
  }
  if (datasource.driver === "LOGS_ELASTICSEARCH") {
    return <ElasticsearchNodes datasource={datasource} />
  }
  return <LogsMetadata datasource={datasource} />
}
