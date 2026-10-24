import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { PermissionItem } from "@/api/role/types"
import {
  actionTitle,
  groupPermissions,
  itemsFromNames,
  permissionTitleEn,
  permissionTitleZh,
} from "@/lib/permission-catalog"

interface PermissionSummaryProps {
  permissions: string[]
  catalog: PermissionItem[]
}

/**
 * 只读的权限分组。详情页看这里就够了：模块头给出「成员 · Members」，
 * 行内只需要动作（中英并排），完整的中英文给 Tooltip，code 也留在 Tooltip 里排障用。
 */
export function PermissionSummary({ permissions, catalog }: PermissionSummaryProps) {
  const { t } = useTranslation()
  const byName = new Map(catalog.map((item) => [item.name, item]))
  const granted = permissions.map((name) => byName.get(name) ?? itemsFromNames([name])[0])
  const groups = groupPermissions(granted)
  const totals = new Map(
    groupPermissions(catalog).map((group) => [group.resource, group.items.length]),
  )

  if (granted.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">{t("role.grant.noSelected")}</p>
    )
  }

  return (
    <div className="grid max-h-[24rem] gap-4 overflow-y-auto pr-1">
      {groups.map((group) => (
        <section key={group.resource} className="grid gap-2">
          <div className="flex items-center justify-between gap-2">
            <h4 className="text-sm font-medium">{group.title}</h4>
            <span className="text-xs text-muted-foreground">
              {totals.has(group.resource)
                ? t("role.grant.groupCount", {
                    selected: group.items.length,
                    total: totals.get(group.resource),
                  })
                : t("role.permissionCount", { count: group.items.length })}
            </span>
          </div>
          <div className="flex flex-wrap gap-1.5">
            {group.items.map((item) => (
              <Tooltip key={item.name}>
                <TooltipTrigger asChild>
                  <Badge
                    variant="outline"
                    className="cursor-default border-border bg-muted/50 font-normal"
                  >
                    {actionTitle(item.action)}
                  </Badge>
                </TooltipTrigger>
                <TooltipContent side="top" className="max-w-xs">
                  <div className="grid gap-0.5 whitespace-pre-wrap">
                    <span className="font-medium">
                      {permissionTitleZh(item)} · {permissionTitleEn(item)}
                    </span>
                    <span className="font-mono text-[11px] opacity-80">{item.name}</span>
                  </div>
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}
