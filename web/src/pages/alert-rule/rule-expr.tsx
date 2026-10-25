import { useTranslation } from "react-i18next"

import type { AlertRuleLevelItem } from "@/api/alert-rule/types"
import { enumLabel } from "@/lib/enum-label"

/**
 * 告警规则的表达式展示：不加边框或底色容器，先展示原始表达式，
 * 再逐行展示级别条件。列表和详情共用，避免两处排版不一致。
 */
export function AlertRuleExpr({
  expr,
  levels,
  maxHeightClass = "max-h-64",
}: {
  expr: string
  levels: AlertRuleLevelItem[]
  maxHeightClass?: string
}) {
  const { t } = useTranslation()
  return (
    <div className="grid min-w-0 gap-2">
      <pre
        className={`overflow-x-hidden overflow-y-auto font-mono text-xs break-words whitespace-pre-wrap ${maxHeightClass}`}
      >
        {expr || "—"}
      </pre>
      {levels.length > 0 ? (
        <div className="grid min-w-0 gap-0.5 font-mono text-xs text-muted-foreground">
          {levels.map((level, index) => (
            <span key={level.uid || index} className="min-w-0 break-words">
              {t("alertEvent.conditionSummary", {
                level: level.levelName || level.levelUid,
                sample: enumLabel(t, level.sample),
                condition: enumLabel(t, level.condition),
                threshold: level.threshold || "—",
                forSeconds: level.forSeconds,
              })}
            </span>
          ))}
        </div>
      ) : null}
    </div>
  )
}
