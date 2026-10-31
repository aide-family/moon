import { useTranslation } from "react-i18next"

import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import type { PermissionItem, PermissionMenu } from "@/api/role/types"
import {
  groupPermissions,
  itemsFromNames,
  permissionActionTitle,
  permissionTitleEn,
  permissionTitleZh,
} from "@/lib/permission-catalog"
import { menuLabel, moduleDefinition } from "@/lib/feature-catalog"

interface PermissionSummaryProps {
  permissions: string[]
  catalog: PermissionItem[]
  menus: PermissionMenu[]
}

/**
 * 只读的权限概览：按菜单目录展示「分组 → 菜单 → 子权限」。
 * 后端返回菜单目录时不再按资源平铺；旧接口没有 menus 时回落到资源分组。
 */
export function PermissionSummary({
  permissions,
  catalog,
  menus,
}: PermissionSummaryProps) {
  const { t } = useTranslation()
  const granted = new Set(permissions)
  const byName = new Map(catalog.map((item) => [item.name, item]))

  if (permissions.length === 0) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("role.grant.noSelected")}
      </p>
    )
  }

  if (menus.length === 0) {
    return <LegacySummary permissions={permissions} catalog={catalog} />
  }

  const groups = new Map<string, PermissionMenu[]>()
  for (const menu of menus) {
    if (!granted.has(`menu:${menu.code}`)) {
      continue
    }
    const list = groups.get(menu.group)
    if (list) {
      list.push(menu)
    } else {
      groups.set(menu.group, [menu])
    }
  }
  const known = new Set(catalog.map((item) => item.name))
  const extras = permissions.filter((name) => !known.has(name))

  return (
    <div className="grid max-h-[24rem] gap-3 overflow-y-auto pr-1">
      {[...groups.entries()].map(([group, groupMenus]) => (
        <section key={group} className="grid gap-1.5">
          <h4 className="text-sm font-medium">
            {groupTitle(group)}
          </h4>
          {groupMenus.map((menu) => {
            const label = menuLabel(menu.code, menu.name)
            const items = menu.permissions.filter((item) =>
              granted.has(item.name)
            )
            return (
              <div key={menu.code} className="grid gap-1 pl-2">
                <span className="text-xs font-medium text-muted-foreground">
                  {label.zh} · {label.en}
                  {items.length > 0
                    ? ` (${t("role.permissionCount", { count: items.length })})`
                    : ""}
                </span>
                <div className="flex flex-wrap gap-1.5">
                  {items.length === 0 ? (
                    <span className="text-xs text-muted-foreground">
                      {t("role.grant.readDefault")}
                    </span>
                  ) : (
                    items.map((item) => (
                      <PermissionBadge key={item.name} item={item} />
                    ))
                  )}
                </div>
              </div>
            )
          })}
        </section>
      ))}
      {extras.length > 0 ? (
        <section className="grid gap-1.5">
          <h4 className="text-sm font-medium">{t("role.grant.preserved")}</h4>
          <div className="flex flex-wrap gap-1.5">
            {extras.map((name) => (
              <PermissionBadge
                key={name}
                item={byName.get(name) ?? itemsFromNames([name])[0]}
              />
            ))}
          </div>
        </section>
      ) : null}
    </div>
  )

  function groupTitle(group: string): string {
    const label =
      moduleDefinition(group)?.label ??
      (group === "tenant"
        ? { zh: "租户与成员", en: "Tenant & members" }
        : { zh: group, en: group })
    return `${label.zh} · ${label.en}`
  }
}

function PermissionBadge({ item }: { item: PermissionItem }) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Badge
          variant="outline"
          className="cursor-default border-border bg-muted/50 font-normal"
        >
          {permissionActionTitle(item)}
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
  )
}

/** 旧接口没有菜单目录时的只读概览，保持原有资源分组。 */
function LegacySummary({
  permissions,
  catalog,
}: {
  permissions: string[]
  catalog: PermissionItem[]
}) {
  const { t } = useTranslation()
  const byName = new Map(catalog.map((item) => [item.name, item]))
  const granted = permissions.map(
    (name) => byName.get(name) ?? itemsFromNames([name])[0]
  )
  const groups = groupPermissions(granted)
  return (
    <div className="grid max-h-[24rem] gap-4 overflow-y-auto pr-1">
      {groups.map((group) => (
        <section key={group.resource} className="grid gap-2">
          <div className="flex items-center justify-between gap-2">
            <h4 className="text-sm font-medium">{group.title}</h4>
            <span className="text-xs text-muted-foreground">
              {t("role.permissionCount", { count: group.items.length })}
            </span>
          </div>
          <div className="flex flex-wrap gap-1.5">
            {group.items.map((item) => (
              <PermissionBadge key={item.name} item={item} />
            ))}
          </div>
        </section>
      ))}
    </div>
  )
}
