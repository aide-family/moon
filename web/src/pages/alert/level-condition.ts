import { useEffect, useState } from "react"
import type { TFunction } from "i18next"

import { getAlertRule } from "@/api/alert-rule"
import type { AlertRuleLevelItem } from "@/api/alert-rule/types"
import { enumLabel } from "@/lib/enum-label"

/**
 * 级别条件只存在于规则上：按事件的规则和级别取命中的那一条。
 * 规则已删除或当前用户没有规则读权限时返回 null，只展示表达式。
 */
export function useEventLevelCondition(
  ruleUid: string,
  levelUid: string,
  enabled = true
): AlertRuleLevelItem | null {
  const [condition, setCondition] = useState<{
    ruleUid: string
    levelUid: string
    item: AlertRuleLevelItem | null
  } | null>(null)

  useEffect(() => {
    if (!enabled || !ruleUid || !levelUid) {
      // 不在这里清空状态：下面的 guard 只在规则和级别都匹配时才返回值。
      return
    }
    let alive = true
    async function load() {
      try {
        const result = await getAlertRule(ruleUid)
        if (!alive) {
          return
        }
        setCondition({
          ruleUid,
          levelUid,
          item:
            result.rule.levels.find((item) => item.levelUid === levelUid) ??
            null,
        })
      } catch {
        // 取不到条件不影响表达式和图表展示。
        if (alive) {
          setCondition({ ruleUid, levelUid, item: null })
        }
      }
    }
    void load()
    return () => {
      alive = false
    }
  }, [enabled, levelUid, ruleUid])

  // 只在取回的条件确实属于当前事件的规则和级别时才返回，切换事件不会带出上一条。
  if (
    !condition ||
    condition.ruleUid !== ruleUid ||
    condition.levelUid !== levelUid
  ) {
    return null
  }
  return condition.item
}

/** 把级别条件格式化成「级别：最大值大于 90 · 60 秒」。 */
export function conditionSummary(
  t: TFunction,
  condition: AlertRuleLevelItem,
  levelName: string
) {
  return t("alertEvent.conditionSummary", {
    level: levelName || condition.levelName || condition.levelUid,
    sample: enumLabel(t, condition.sample),
    condition: enumLabel(t, condition.condition),
    threshold: condition.threshold || "—",
    forSeconds: condition.forSeconds,
  })
}
