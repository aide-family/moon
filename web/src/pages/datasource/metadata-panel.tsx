import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { listDatasourceMetrics } from "@/api/datasource"
import type { DatasourceMetricItem } from "@/api/datasource/types"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { failureText } from "@/lib/form-error"

export function MetadataPanel({ uid }: { uid: string }) {
  const { t } = useTranslation()
  const [metrics, setMetrics] = useState<DatasourceMetricItem[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState("")
  const [attempt, setAttempt] = useState(0)

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
        <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
          {t("action.retry")}
        </Button>
      </div>
    )
  }
  if (metrics.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("datasource.metadataEmpty")}</p>
  }

  return (
    <TooltipProvider delayDuration={1500} skipDelayDuration={0}>
    <div className="flex min-h-0 flex-1 flex-col">
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
              <TableCell className="max-w-[16rem] truncate font-mono text-xs">{item.name}</TableCell>
              <TableCell className="max-w-[24rem] text-muted-foreground">
                {item.help ? (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <span className="block truncate">{item.help}</span>
                    </TooltipTrigger>
                    <TooltipContent className="max-w-sm">{item.help}</TooltipContent>
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
