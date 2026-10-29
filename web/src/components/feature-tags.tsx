import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/class-names"
import { featureLabelText, featureTitle } from "@/lib/feature-catalog"

interface FeatureTagsProps {
  codes?: string[]
  /** 空列表时的占位文案（由调用方传入已本地化的文案）。 */
  emptyLabel: string
  className?: string
}

/** 功能 code 列表渲染成标签；当前语言优先，title 提供中英对照。 */
export function FeatureTags({
  codes,
  emptyLabel,
  className,
}: FeatureTagsProps) {
  const { i18n } = useTranslation()
  const list = codes ?? []
  if (list.length === 0) {
    return <span className="text-muted-foreground">{emptyLabel}</span>
  }
  return (
    <span className={cn("flex min-w-0 flex-wrap gap-1", className)}>
      {list.map((code) => (
        <Badge
          key={code}
          variant="outline"
          title={featureTitle(code)}
          className="font-normal"
        >
          {featureLabelText(code, i18n.language)}
        </Badge>
      ))}
    </span>
  )
}
