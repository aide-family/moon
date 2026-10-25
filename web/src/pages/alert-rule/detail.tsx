import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import type { AlertRuleGroupItem, AlertRuleItem } from "@/api/alert-rule/types"
import type { DatasourceItem } from "@/api/datasource/types"
import { DetailSheet } from "@/components/detail-sheet"
import { LabelTags } from "@/components/label-tags"
import { StatusBadge } from "@/components/status-badge"
import { parseLabelPairs } from "@/lib/labels"

import { matchedDatasources } from "./matching"
import { AlertRuleExpr } from "./rule-expr"

interface DetailField {
  label: string
  value: ReactNode
  span?: 1 | 2
}

interface AlertRuleDetailSheetProps {
  detail: AlertRuleItem | null
  groups: AlertRuleGroupItem[]
  sources: DatasourceItem[]
  onOpenChange: (open: boolean) => void
}

/**
 * 告警规则详情抽屉：数据源范围、规则标签、摘要/描述模板和级别条件都来自
 * 已加载的规则项，规则组名称和匹配数据源由页面传入的列表补充，避免详情再发请求。
 */
export function AlertRuleDetailSheet({
  detail,
  groups,
  sources,
  onOpenChange,
}: AlertRuleDetailSheetProps) {
  const { t } = useTranslation()
  const fields: DetailField[] = []

  if (detail) {
    const groupName =
      groups.find((item) => item.uid === detail.groupUid)?.name ||
      detail.groupUid
    const matchers = parseLabelPairs(detail.datasourceLabels)
    const hasMatchers = matchers.some((pair) => pair.key.trim() !== "")
    const matched = matchedDatasources(sources, detail.driver, matchers)

    fields.push(
      { label: t("alertRule.group"), value: groupName },
      { label: t("detail.uid"), value: detail.uid },
      {
        label: t("datasource.driver"),
        value: <StatusBadge value={detail.driver} />,
      },
      {
        label: t("namespace.status"),
        value: <StatusBadge value={detail.status} />,
      },
      {
        label: t("alertRule.datasourceLabels"),
        value: hasMatchers ? (
          <LabelTags labels={detail.datasourceLabels} />
        ) : (
          t("alertRule.detail.allDatasources")
        ),
        span: 2,
      },
      {
        label: t("alertRule.detail.matchedDatasources"),
        value:
          matched.length > 0 ? (
            <span className="grid gap-1">
              <span>
                {t("alertRule.detail.matchedCount", { count: matched.length })}
              </span>
              <span className="break-words text-muted-foreground">
                {matched.map((item) => item.name).join("、")}
              </span>
            </span>
          ) : (
            <span className="text-muted-foreground">
              {t("alertRule.detail.noDatasource")}
            </span>
          ),
        span: 2,
      },
      {
        label: t("alertRule.labels"),
        value: <LabelTags labels={detail.labels} />,
        span: 2,
      },
      {
        label: t("alertRule.summary"),
        value: (
          <pre className="overflow-x-hidden overflow-y-auto font-mono text-xs break-words whitespace-pre-wrap">
            {detail.summary || "—"}
          </pre>
        ),
        span: 2,
      },
      {
        label: t("alertRule.descriptionField"),
        value: (
          <div className="grid gap-2">
            <pre className="overflow-x-hidden overflow-y-auto font-mono text-xs break-words whitespace-pre-wrap">
              {detail.description || "—"}
            </pre>
            <p className="text-xs text-muted-foreground">
              {t("alertRule.detail.templateHint")}
            </p>
          </div>
        ),
        span: 2,
      },
      {
        label: t("alertRule.levels"),
        value: (
          <AlertRuleExpr
            expr={detail.expr}
            levels={detail.levels}
            maxHeightClass="max-h-96"
          />
        ),
        span: 2,
      }
    )

    if (detail.datasourceUid) {
      fields.push({
        label: t("alertRule.detail.legacyDatasource"),
        value: detail.datasourceUid,
        span: 2,
      })
    }
  }

  return (
    <DetailSheet
      open={detail !== null}
      title={detail?.name ?? ""}
      description={
        detail ? <StatusBadge value={detail.status} /> : undefined
      }
      wide
      fields={fields}
      onOpenChange={onOpenChange}
    />
  )
}
