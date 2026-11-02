import { useState } from "react"
import { useTranslation } from "react-i18next"

import { updateCluster } from "@/api/cluster"
import type { ClusterItem } from "@/api/cluster"
import { listDatasources } from "@/api/datasource"
import type { DatasourceItem } from "@/api/datasource/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { failureMessage, formErrorMessage } from "@/lib/form-error"
import { loadAllPagedItems } from "@/lib/select-pages"
import { useEffect } from "react"

/** Radix Select 不允许空字符串选项，用哨兵值表示「解绑」。 */
const BIND_NONE = "NONE"

// 只有 Prometheus 兼容驱动能执行概况的四条 PromQL；其他指标驱动不提供绑定。
const METRICS_DRIVERS = ["METRICS_PROMETHEUS", "METRICS_VICTORIA_METRICS"]

interface MetricsBindDialogProps {
  open: boolean
  cluster: ClusterItem | null
  /** 绑定成功后刷新集群元数据与指标汇总。 */
  onSaved: () => void
  onOpenChange: (open: boolean) => void
}

/**
 * 指标数据源绑定：只列出同租户的 Prometheus / VictoriaMetrics 数据源，
 * 禁用或已删除的数据源不进入候选（后端仍会兜底校验）。
 * 保存时只提交 metricsDatasourceUid，kubeconfig 留空保持原值。
 */
export function MetricsBindDialog({
  open,
  cluster,
  onSaved,
  onOpenChange,
}: MetricsBindDialogProps) {
  const { t } = useTranslation()
  const [options, setOptions] = useState<DatasourceItem[]>([])
  const [loading, setLoading] = useState(true)
  const [loadFailed, setLoadFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const [value, setValue] = useState(() => cluster?.metricsDatasourceUid ?? "")
  const [pending, setPending] = useState(false)
  const [formError, setFormError] = useState("")

  useEffect(() => {
    if (!open) {
      return
    }
    let active = true
    async function load() {
      setLoading(true)
      setLoadFailed(false)
      try {
        const items = await loadAllPagedItems((page, pageSize) =>
          listDatasources({ page, pageSize, type: "METRICS" })
        )
        if (!active) {
          return
        }
        setOptions(
          items.filter(
            (item) =>
              METRICS_DRIVERS.includes(item.driver) &&
              (item.status === "ACTIVE" ||
                item.uid === cluster?.metricsDatasourceUid)
          )
        )
      } catch (err) {
        if (!active) {
          return
        }
        setOptions([])
        setLoadFailed(true)
        setLoadError(failureMessage(err, t("kube.metricsBindLoadFailed")))
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
  }, [open, attempt, cluster?.metricsDatasourceUid, t])

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (!cluster || pending) {
      return
    }
    setPending(true)
    setFormError("")
    try {
      await updateCluster({
        uid: cluster.uid,
        name: cluster.name,
        kubeconfig: "",
        remark: cluster.remark,
        metricsDatasourceUid: value === BIND_NONE ? "" : value,
      })
      onSaved()
      onOpenChange(false)
    } catch (err) {
      setFormError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("kube.metricsBindTitle")}</DialogTitle>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSubmit}
        >
          <DialogBody className="grid gap-4">
            <p className="text-xs leading-relaxed text-muted-foreground">
              {t("kube.metricsBindHint")}
            </p>
            {loading ? (
              <p className="text-sm text-muted-foreground">
                {t("state.loading")}
              </p>
            ) : null}
            {!loading && loadFailed ? (
              <div className="flex flex-wrap items-center gap-2">
                <p className="text-sm text-destructive">
                  {loadError || t("kube.metricsBindLoadFailed")}
                </p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setAttempt((current) => current + 1)}
                >
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!loading && !loadFailed ? (
              <FormField
                label={t("kube.metricsBindDatasource")}
                htmlFor="cluster-metrics-datasource"
              >
                <Select value={value || BIND_NONE} onValueChange={setValue}>
                  <SelectTrigger
                    id="cluster-metrics-datasource"
                    className="w-full"
                  >
                    <SelectValue
                      placeholder={t("kube.metricsBindPlaceholder")}
                    />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={BIND_NONE}>
                      {t("kube.metricsBindNone")}
                    </SelectItem>
                    {options.map((item) => (
                      <SelectItem key={item.uid} value={item.uid}>
                        {item.name}
                        {item.status !== "ACTIVE"
                          ? ` (${t("enum.DISABLED")})`
                          : ""}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            ) : null}
            {!loading && !loadFailed && options.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                {t("kube.metricsBindNoOptions")}
              </p>
            ) : null}
          </DialogBody>
          {formError ? (
            <p className="text-sm text-destructive">{formError}</p>
          ) : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              disabled={pending}
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending || loading || loadFailed}>
              {pending ? t("action.saving") : t("action.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
