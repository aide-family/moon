import { useTranslation } from "react-i18next"

import type { InspectionPlan } from "@/api/inspection/types"
import type { MultiSelectOption } from "@/components/multi-select"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { StatusBadge } from "@/components/status-badge"
import { formatDateTime } from "@/lib/date-time"
import {
  inspectionDisplayLabel,
  inspectionKindLabel,
} from "@/pages/inspection-plans/inspection-item-editor"

function prettyJSON(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2)
  } catch {
    return raw
  }
}

/** 巡检任务详情：只读展示计划配置、调度通知与每个检查项的查询/阈值参数。 */
export function InspectionPlanDetail({
  open,
  plan,
  templateName,
  groups,
  onOpenChange,
}: {
  open: boolean
  plan: InspectionPlan | null
  templateName: string
  groups: MultiSelectOption[]
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  if (!plan) {
    return null
  }
  const groupNames = (plan.notificationGroupUids ?? []).map(
    (uid) => groups.find((group) => group.value === uid)?.label ?? uid
  )
  const sections: DetailSection[] = [
    {
      title: t("inspection.stepBasic"),
      fields: [
        { label: t("inspection.planName"), value: plan.name },
        { label: t("inspection.description"), value: plan.description || "—" },
        {
          label: t("inspection.runStatus"),
          value: (
            <StatusBadge
              value={plan.status}
              label={
                plan.status === "ACTIVE"
                  ? t("inspection.statusActive")
                  : t("inspection.statusPaused")
              }
              tone={plan.status === "ACTIVE" ? "success" : "neutral"}
            />
          ),
        },
        {
          label: t("inspection.timeout"),
          value: `${plan.timeoutSeconds}s`,
        },
        { label: t("inspection.template"), value: templateName || "—" },
        { label: t("inspection.createdAt"), value: formatDateTime(plan.createdAt) || "—" },
        { label: t("inspection.labels"), value: plan.labels || "{}", span: 2 },
        { label: t("inspection.remark"), value: plan.remark || "—", span: 2 },
        { label: t("inspection.cron"), value: plan.cron },
        { label: t("inspection.timezone"), value: plan.timezone },
        {
          label: t("inspection.nextRunLabel"),
          value:
            plan.status === "ACTIVE" ? formatDateTime(plan.nextRunAt) || "—" : "—",
        },
        {
          label: t("inspection.lastRun"),
          value: formatDateTime(plan.lastRunAt) || "—",
        },
        {
          label: t("inspection.notifyPolicy"),
          value:
            plan.notifyPolicy === "ALWAYS"
              ? t("inspection.notifyAlways")
              : t("inspection.notifyAbnormal"),
        },
        {
          label: t("inspection.groups"),
          value: groupNames.length > 0 ? groupNames.join(" · ") : "—",
          span: 2,
        },
      ],
    },
    {
      title: t("inspection.detailItems"),
      content:
        (plan.items ?? []).length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("inspection.detailNoItems")}</p>
        ) : (
          <div className="grid gap-3">
            {(plan.items ?? []).map((item, index) => (
              <div key={item.uid || index} className="grid gap-2 rounded-lg border border-border p-3">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-mono text-xs text-muted-foreground">
                    {String(index + 1).padStart(2, "0")}
                  </span>
                  <span className="min-w-0 flex-1 text-sm font-medium break-words">
                    {item.title}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {inspectionKindLabel(t, item.kind)}
                  </span>
                  <StatusBadge
                    value={item.enabled ? "ACTIVE" : "PAUSED"}
                    label={
                      item.enabled
                        ? t("inspection.enabled")
                        : t("inspection.statusPaused")
                    }
                    tone={item.enabled ? "success" : "neutral"}
                  />
                </div>
                <p className="text-xs text-muted-foreground">
                  {t("inspection.window")} {item.windowSeconds}s · {t("inspection.step")}{" "}
                  {item.stepSeconds}s · {t("inspection.display")}{" "}
                  {inspectionDisplayLabel(t, item.display)}
                </p>
                {item.query ? (
                  <pre className="max-h-48 overflow-auto rounded-md bg-muted/50 p-2 font-mono text-[11px] whitespace-pre-wrap break-words">
                    {item.query}
                  </pre>
                ) : null}
                {item.params && item.params !== "{}" ? (
                  <pre className="max-h-56 overflow-auto rounded-md bg-muted/50 p-2 font-mono text-[11px] whitespace-pre-wrap break-words">
                    {prettyJSON(item.params)}
                  </pre>
                ) : null}
              </div>
            ))}
          </div>
        ),
    },
  ]

  return (
    <DetailSheet
      open={open}
      onOpenChange={onOpenChange}
      size="wide"
      title={plan.name}
      description={
        <>
          <StatusBadge
            value={plan.status}
            label={
              plan.status === "ACTIVE"
                ? t("inspection.statusActive")
                : t("inspection.statusPaused")
            }
            tone={plan.status === "ACTIVE" ? "success" : "neutral"}
          />
          <span className="font-mono text-xs">{plan.uid}</span>
        </>
      }
      sections={sections}
    />
  )
}
