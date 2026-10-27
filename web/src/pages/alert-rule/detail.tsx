import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import type { AlertRuleGroupItem, AlertRuleItem } from "@/api/alert-rule/types"
import type { DatasourceItem } from "@/api/datasource/types"
import { CodeBlock } from "@/components/code-block"
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
          <CodeBlock
            variant="plain"
            code={detail.summary}
            language="text"
            wrap
          />
        ),
        span: 2,
      },
      {
        label: t("alertRule.descriptionField"),
        value: (
          <div className="grid gap-2">
            <CodeBlock
              variant="plain"
              code={detail.description}
              language="text"
              wrap
            />
            <p className="text-xs text-muted-foreground">
              {t("alertRule.detail.templateHint")}
            </p>
          </div>
        ),
        span: 2,
      }
    )

    // 自定义变量在编辑时和摘要/描述同一步，详情里也放在它们后面、级别条件之前。
    const variables = detail.variables ?? []
    if (variables.length > 0) {
      fields.push({
        label: t("alertRule.variables"),
        value: (
          <div className="grid min-w-0 gap-2">
            {variables.map((variable, index) => (
              <div
                key={`${variable.name}-${index}`}
                className="grid min-w-0 gap-1 rounded-md border p-2"
              >
                <div className="flex min-w-0 flex-wrap items-baseline gap-x-2 gap-y-1">
                  <code className="font-mono text-xs text-foreground">
                    {"$" + variable.name}
                  </code>
                  <span className="min-w-0 text-xs break-words text-muted-foreground">
                    {variable.description || "—"}
                  </span>
                </div>
                <CodeBlock
                  variant="plain"
                  code={variable.expr}
                  language="text"
                  wrap
                />
              </div>
            ))}
          </div>
        ),
        span: 2,
      })
    }

    fields.push({
      label: t("alertRule.levels"),
      value: (
        <AlertRuleExpr
          expr={detail.expr}
          levels={detail.levels}
          maxHeightClass="max-h-96"
        />
      ),
      span: 2,
    })

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
