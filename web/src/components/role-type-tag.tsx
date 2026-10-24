import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/class-names"

interface RoleTypeTagProps {
  isSystem: boolean
  // compact 用于行宽紧张的成员分配列表：只写「系统 / 自定义」。
  variant?: "full" | "compact"
  className?: string
}

// 系统角色与自定义角色用独立的两套底色：颜色只表示类型，不表示状态。
// 系统角色用 sky（对比度高于 primary 叠淡底），自定义角色用中性灰但文字走 foreground 级别。
const toneClasses = {
  system: "border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300",
  custom: "border-border bg-muted/60 text-slate-600 dark:text-slate-300",
}

export function RoleTypeTag({ isSystem, variant = "full", className }: RoleTypeTagProps) {
  const { t } = useTranslation()
  const label =
    variant === "compact"
      ? t(isSystem ? "role.typeSystemCompact" : "role.typeCustomCompact")
      : t(isSystem ? "role.typeSystem" : "role.typeCustom")

  return (
    <Badge
      variant="outline"
      className={cn("font-normal", isSystem ? toneClasses.system : toneClasses.custom, className)}
    >
      {label}
    </Badge>
  )
}
