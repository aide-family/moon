import { CircleHelp } from "lucide-react"
import { useTranslation } from "react-i18next"

import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/class-names"

interface RoleRemarkHintProps {
  remark?: string
  // 无障碍名称，例如「值班运维 的备注」。
  label: string
  className?: string
}

/**
 * 角色备注的 ? 提示。分配角色时先看这里，再决定勾不勾。
 * 备注为空也保留图标位并给出「暂无备注」，避免同名角色左右错位；
 * 触发器是普通 button，不嵌在 label 里，点它不会切换勾选。
 */
export function RoleRemarkHint({ remark, label, className }: RoleRemarkHintProps) {
  const { t } = useTranslation()
  return (
    <TooltipProvider delayDuration={300}>
      <Tooltip>
        <TooltipTrigger asChild>
          <button
            type="button"
            aria-label={label}
            onClick={(event) => event.preventDefault()}
            className={cn(
              "inline-flex size-5 shrink-0 items-center justify-center rounded-sm text-muted-foreground transition-colors hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
              className,
            )}
          >
            <CircleHelp className="size-4" />
          </button>
        </TooltipTrigger>
        <TooltipContent side="right" align="start" className="max-w-xs">
          <div className="grid max-h-60 gap-1 overflow-y-auto whitespace-pre-wrap break-words">
            <span className="font-medium">{t("role.remark")}</span>
            <span>{remark?.trim() ? remark : t("role.remarkEmpty")}</span>
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
