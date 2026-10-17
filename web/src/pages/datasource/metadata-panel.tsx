import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getDatasourceMetric, listDatasourceMetrics } from "@/api/datasource"
import { Button } from "@/components/ui/button"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { enumLabel } from "@/lib/enum-label"
import { failureText } from "@/lib/form-error"

interface MetricMeta {
  type: string
  help: string
  unit: string
}

function parseMetricMeta(body: string, name: string): MetricMeta | null {
  try {
    const payload = JSON.parse(body) as { data?: Record<string, Array<{ type?: string; help?: string; unit?: string }>> }
    const item = payload.data?.[name]?.[0]
    if (!item) {
      return null
    }
    return { type: item.type ?? "", help: item.help ?? "", unit: item.unit ?? "" }
  } catch {
    return null
  }
}

export function MetadataPanel({ uid }: { uid: string }) {
  const { t } = useTranslation()
  const [names, setNames] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState("")
  const [selected, setSelected] = useState("")
  const [meta, setMeta] = useState<MetricMeta | null>(null)
  const [detailError, setDetailError] = useState("")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed("")
      setSelected("")
      setMeta(null)
      try {
        const result = await listDatasourceMetrics(uid)
        if (!active) {
          return
        }
        setNames(result.names ?? [])
      } catch (err) {
        if (active) {
          setNames([])
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
  }, [uid, t])

  async function openMetric(name: string) {
    setSelected(name)
    setMeta(null)
    setDetailError("")
    try {
      const result = await getDatasourceMetric(uid, name)
      setMeta(parseMetricMeta(result.body, name))
      if (!parseMetricMeta(result.body, name)) {
        setDetailError(result.body)
      }
    } catch (err) {
      setDetailError(failureText(err, t("state.failed")))
    }
  }

  if (loading) {
    return <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
  }
  if (failed) {
    return <p className="text-sm text-destructive">{failed}</p>
  }
  if (names.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("datasource.metadataEmpty")}</p>
  }

  return (
    <div className="grid min-h-0 gap-4">
      <div className="max-h-80 overflow-auto rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("datasource.metricName")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {names.map((name) => (
              <TableRow key={name} data-state={name === selected ? "selected" : undefined}>
                <TableCell>
                  <Button variant="link" className="h-auto px-0" onClick={() => void openMetric(name)}>
                    {name}
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      {selected && meta ? (
        <dl className="grid gap-2 text-sm">
          <div>
            <dt className="text-muted-foreground">{t("datasource.metricType")}</dt>
            <dd>{enumLabel(t, meta.type) || meta.type || "—"}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">{t("datasource.metricHelp")}</dt>
            <dd>{meta.help || "—"}</dd>
          </div>
          <div>
            <dt className="text-muted-foreground">{t("datasource.metricUnit")}</dt>
            <dd>{meta.unit || "—"}</dd>
          </div>
        </dl>
      ) : null}
      {detailError ? <pre className="overflow-auto rounded-lg border p-3 text-sm">{detailError}</pre> : null}
    </div>
  )
}
