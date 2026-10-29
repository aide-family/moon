import { useTranslation } from "react-i18next"

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { cn } from "@/lib/class-names"

import {
  findClusterNavEntry,
  readClusterNavLabel,
  type ClusterNavCounts,
  type ClusterNavGroup,
  type ClusterSection,
} from "./sections"

interface ClusterResourceNavProps {
  section: ClusterSection
  kind: string
  /** 已按权限过滤的导航分组；无权限的资源视图不会出现在这里。 */
  groups: ClusterNavGroup[]
  /** 导航项 id → 当前命名空间下的数量；接口未返回时不显示数字。 */
  counts: ClusterNavCounts
  onSelect: (section: ClusterSection, kind?: string) => void
}

/**
 * 左侧资源管理导航：按负载 / 服务与路由 / 配置管理 / 存储 分组。
 * 分组由调用方按权限过滤后传入；每项右侧显示当前命名空间下的资源数量。
 */
export function ClusterResourceNav({
  section,
  kind,
  groups,
  counts,
  onSelect,
}: ClusterResourceNavProps) {
  const { t } = useTranslation()
  const activeId = findClusterNavEntry(section, kind)?.id

  return (
    <nav
      aria-label={t("kube.resourceNav")}
      data-slot="cluster-resource-nav"
      className="flex flex-col gap-1"
    >
      {groups.map((group) => (
        <div key={group.id} className="min-w-0">
          {group.title ? (
            <p className="px-2 pt-3 pb-1 text-[10px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
              {readClusterNavLabel(t, group.title)}
            </p>
          ) : null}
          <ul className="grid gap-0.5">
            {group.entries.map((entry) => {
              const active = entry.id === activeId
              const count = counts[entry.id]
              const Icon = entry.icon
              return (
                <li key={entry.id}>
                  <button
                    type="button"
                    aria-current={active ? "page" : undefined}
                    onClick={() => onSelect(entry.section, entry.kind)}
                    className={cn(
                      "group flex min-h-9 w-full min-w-0 items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
                      active
                        ? "bg-primary/10 font-medium text-foreground ring-1 ring-primary/15 ring-inset"
                        : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
                    )}
                  >
                    <Icon
                      aria-hidden="true"
                      className={cn(
                        "size-4 shrink-0",
                        active ? "text-primary-ink" : "text-muted-foreground/80"
                      )}
                    />
                    <span className="min-w-0 flex-1 truncate">
                      {readClusterNavLabel(t, entry.label)}
                    </span>
                    {count !== undefined ? (
                      <span className="shrink-0 font-mono text-[11px] text-muted-foreground tabular-nums">
                        {count}
                      </span>
                    ) : null}
                  </button>
                </li>
              )
            })}
          </ul>
        </div>
      ))}
    </nav>
  )
}

interface ClusterResourcePickerProps {
  section: ClusterSection
  kind: string
  /** 已按权限过滤的导航分组，与左侧导航复用同一份数据。 */
  groups: ClusterNavGroup[]
  onSelect: (section: ClusterSection, kind?: string) => void
  className?: string
}

/** 窄屏下的资源切换：分组导航收起，用一个分组下拉替代。 */
export function ClusterResourcePicker({
  section,
  kind,
  groups,
  onSelect,
  className,
}: ClusterResourcePickerProps) {
  const { t } = useTranslation()
  const active = findClusterNavEntry(section, kind)
  // 当前 section 无权限时页面会回退到概况，下拉也要显示概况而不是空值。
  const activeVisible =
    active !== null &&
    groups.some((group) =>
      group.entries.some((entry) => entry.id === active.id)
    )
  const fallbackId = groups[0]?.entries[0]?.id ?? "overview"
  const value = activeVisible && active ? active.id : fallbackId

  return (
    <Select
      value={value}
      onValueChange={(next) => {
        for (const group of groups) {
          const entry = group.entries.find((item) => item.id === next)
          if (entry) {
            onSelect(entry.section, entry.kind)
            return
          }
        }
      }}
    >
      <SelectTrigger
        className={cn("w-44 shrink-0", className)}
        aria-label={t("kube.resourceNav")}
      >
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {groups.map((group) => (
          <SelectGroup key={group.id}>
            {group.title ? (
              <SelectLabel>{readClusterNavLabel(t, group.title)}</SelectLabel>
            ) : null}
            {group.entries.map((entry) => (
              <SelectItem key={entry.id} value={entry.id}>
                {readClusterNavLabel(t, entry.label)}
              </SelectItem>
            ))}
          </SelectGroup>
        ))}
      </SelectContent>
    </Select>
  )
}
