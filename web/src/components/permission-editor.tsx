import { useEffect, useRef, useState } from "react"
import { ChevronRight, Loader2, Search } from "lucide-react"
import { useTranslation } from "react-i18next"

import type { PermissionItem, RoleItem } from "@/api/role/types"
import { ConfirmAction } from "@/components/confirm-action"
import { RoleTypeTag } from "@/components/role-type-tag"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { cn } from "@/lib/class-names"
import { formErrorMessage } from "@/lib/form-error"
import {
  OWNER_ONLY,
  groupPermissions,
  matchesPermission,
  permissionActionLabel,
} from "@/lib/permission-catalog"

interface PermissionEditorProps {
  open: boolean
  role: RoleItem | null
  catalog: PermissionItem[]
  catalogLoading: boolean
  catalogFailed: boolean
  onRetryCatalog: () => void
  onOpenChange: (open: boolean) => void
  // 保存由页面负责：成功后页面需要 reload，失败则抛出错误，编辑器保留当前勾选。
  onSave: (permissions: string[]) => Promise<void>
}

function sameSet(left: Set<string>, right: Set<string>): boolean {
  if (left.size !== right.size) {
    return false
  }
  for (const value of left) {
    if (!right.has(value)) {
      return false
    }
  }
  return true
}

/**
 * 权限配置器。替代过去「22 个裸编码勾选框平铺」的弹窗：
 * 分组折叠、中英并标、搜索、组内三态、全选/清空、脏值保护和 ⌘S。
 */
export function PermissionEditor({
  open,
  role,
  catalog,
  catalogLoading,
  catalogFailed,
  onRetryCatalog,
  onOpenChange,
  onSave,
}: PermissionEditorProps) {
  const { t } = useTranslation()
  // 每次打开都由页面用 key 重新挂载，所以初始值直接取当前角色，不需要在 effect 里回灌。
  const [baseline] = useState(() => new Set(role?.permissions ?? []))
  const [selected, setSelected] = useState(() => new Set(role?.permissions ?? []))
  const [query, setQuery] = useState("")
  const [onlySelected, setOnlySelected] = useState(false)
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [confirmLeave, setConfirmLeave] = useState(false)
  const [undoSnapshot, setUndoSnapshot] = useState<Set<string> | null>(null)
  const undoTimer = useRef<number | undefined>(undefined)

  useEffect(() => () => window.clearTimeout(undoTimer.current), [])

  // 目录里的 22 项都展示出来（包括只看得到、授予不了的两项），
  // 但「可分配」只有 20 项：namespace:delete/transfer 属于负责人，后端会拒绝授予。
  const ASSIGNABLE = catalog.filter((item) => !OWNER_ONLY.has(item.name))
  const groups = groupPermissions(catalog)
  const total = ASSIGNABLE.length
  const dirty = !sameSet(baseline, selected)
  const added = [...selected].filter((name) => !baseline.has(name)).length
  const removed = [...baseline].filter((name) => !selected.has(name)).length

  const filtered = groups
    .map((group) => {
      const matched = group.items.filter((item) => matchesPermission(item, query))
      const items = onlySelected
        ? matched.filter((item) => selected.has(item.name))
        : matched
      return { ...group, items }
    })
    .filter((group) => group.items.length > 0)

  // 只对「可分配」的可见项计数与全选，避免把负责人专属项算进去。
  const visibleNames = filtered.flatMap((group) =>
    group.items.filter((item) => !OWNER_ONLY.has(item.name)).map((item) => item.name),
  )
  const visibleAllSelected =
    visibleNames.length > 0 && visibleNames.every((name) => selected.has(name))

  function togglePermission(name: string, checked: boolean) {
    setSelected((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(name)
      } else {
        next.delete(name)
      }
      return next
    })
  }

  function toggleGroup(names: string[], checked: boolean) {
    setSelected((current) => {
      const next = new Set(current)
      for (const name of names) {
        if (checked) {
          next.add(name)
        } else {
          next.delete(name)
        }
      }
      return next
    })
  }

  function toggleCollapse(resource: string) {
    if (query.trim() !== "" || onlySelected) {
      return
    }
    setCollapsed((current) => {
      const next = new Set(current)
      if (next.has(resource)) {
        next.delete(resource)
      } else {
        next.add(resource)
      }
      return next
    })
  }

  function selectVisible() {
    setSelected((current) => {
      const next = new Set(current)
      for (const name of visibleNames) {
        next.add(name)
      }
      return next
    })
  }

  function clearAll() {
    setUndoSnapshot(new Set(selected))
    setSelected(new Set())
    window.clearTimeout(undoTimer.current)
    undoTimer.current = window.setTimeout(() => setUndoSnapshot(null), 5000)
  }

  function undoClear() {
    if (!undoSnapshot) {
      return
    }
    setSelected(new Set(undoSnapshot))
    setUndoSnapshot(null)
    window.clearTimeout(undoTimer.current)
  }

  async function save() {
    if (!dirty || pending) {
      return
    }
    setPending(true)
    setError("")
    try {
      await onSave([...selected])
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  useEffect(() => {
    if (!open) {
      return
    }
    function onKeyDown(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "s") {
        event.preventDefault()
        void save()
      }
    }
    window.addEventListener("keydown", onKeyDown)
    return () => window.removeEventListener("keydown", onKeyDown)
  })

  function requestClose() {
    if (dirty && !pending) {
      setConfirmLeave(true)
      return
    }
    onOpenChange(false)
  }

  const selectedCount = [...selected].filter((name) =>
    ASSIGNABLE.some((item) => item.name === name),
  ).length

  return (
    <>
      <Sheet open={open} onOpenChange={(next) => (next ? onOpenChange(true) : requestClose())}>
        <SheetContent
          className="flex flex-col gap-0 p-0 data-[side=right]:w-full data-[side=right]:sm:max-w-3xl!"
          aria-describedby={undefined}
        >
          <SheetHeader className="gap-1 border-b p-4 pr-12">
            <SheetTitle>{t("role.grant.title")}</SheetTitle>
            <SheetDescription className="sr-only">
              {role?.remark?.trim() || t("role.grant.title")}
            </SheetDescription>
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-sm font-medium">{role?.name ?? ""}</span>
              {role ? <RoleTypeTag isSystem={role.isSystem} /> : null}
              <span className="ml-auto text-xs text-muted-foreground">
                {t("role.grant.selected", { selected: selectedCount, total })}
              </span>
            </div>
            {role?.remark?.trim() ? (
              <p className="line-clamp-2 text-xs text-muted-foreground">{role.remark}</p>
            ) : null}
            {dirty ? (
              <p className="text-xs text-muted-foreground">
                {t("role.grant.diff", { added, removed })}
              </p>
            ) : null}
          </SheetHeader>

          <div className="flex flex-wrap items-center gap-2 border-b p-4">
            <div className="relative w-full min-w-0 sm:flex-1">
              <Search className="pointer-events-none absolute top-1/2 left-2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder={t("role.grant.search")}
                className="pl-8"
                aria-label={t("role.grant.search")}
              />
            </div>
            <Button
              type="button"
              size="sm"
              variant={onlySelected ? "secondary" : "outline"}
              aria-pressed={onlySelected}
              onClick={() => setOnlySelected((value) => !value)}
            >
              {t("role.grant.onlySelected")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => setCollapsed(new Set())}
            >
              {t("role.grant.expandAll")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() => setCollapsed(new Set(groups.map((group) => group.resource)))}
            >
              {t("role.grant.collapseAll")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={pending || visibleAllSelected || visibleNames.length === 0}
              onClick={selectVisible}
            >
              {query.trim()
                ? t("role.grant.selectMatches", { count: visibleNames.length })
                : t("role.grant.selectAll")}
            </Button>
            {undoSnapshot ? (
              <Button type="button" size="sm" variant="outline" onClick={undoClear}>
                {t("role.grant.undo")}
              </Button>
            ) : (
              <Button
                type="button"
                size="sm"
                variant="outline"
                disabled={pending || selected.size === 0}
                onClick={clearAll}
              >
                {t("role.grant.clear")}
              </Button>
            )}
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto p-4">
            {catalogLoading ? (
              <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
            ) : null}
            {!catalogLoading && catalogFailed ? (
              <div className="flex items-center gap-2">
                <p className="text-sm text-destructive">{t("role.grant.loadFailed")}</p>
                <Button type="button" size="sm" variant="outline" onClick={onRetryCatalog}>
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!catalogLoading && !catalogFailed && filtered.length === 0 ? (
              <div className="grid justify-items-start gap-2">
                <p className="text-sm text-muted-foreground">
                  {onlySelected && query.trim() === ""
                    ? t("role.grant.noSelected")
                    : t("role.grant.noMatch")}
                </p>
                {query.trim() !== "" ? (
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    onClick={() => setQuery("")}
                  >
                    {t("role.grant.clearSearch")}
                  </Button>
                ) : null}
              </div>
            ) : null}
            {!catalogLoading && !catalogFailed ? (
              <div className="grid gap-2">
                {selectedCount === 0 ? (
                  <p className="text-xs text-muted-foreground">{t("role.grant.emptyHint")}</p>
                ) : null}
                {filtered.map((group) => {
                  const expanded = query.trim() !== "" || onlySelected || !collapsed.has(group.resource)
                  // 组头计数与全选只作用于可分配项，负责人专属项只展示、不参与勾选。
                  const selectable = group.items.filter((item) => !OWNER_ONLY.has(item.name))
                  const groupSelected = selectable.filter((item) => selected.has(item.name)).length
                  const allSelected = selectable.length > 0 && groupSelected === selectable.length
                  const someSelected = groupSelected > 0 && !allSelected
                  return (
                    <Collapsible
                      key={group.resource}
                      open={expanded}
                      onOpenChange={() => toggleCollapse(group.resource)}
                      className="rounded-md border"
                    >
                      <div className="flex min-h-11 items-center gap-2 px-2">
                        <CollapsibleTrigger asChild>
                          <button
                            type="button"
                            className={cn(
                              "flex min-w-0 flex-1 items-center gap-2 py-2 text-left",
                              (query.trim() !== "" || onlySelected) && "cursor-default",
                            )}
                            aria-label={group.title}
                          >
                            <ChevronRight
                              className={cn(
                                "size-4 shrink-0 text-muted-foreground transition-transform duration-150 motion-reduce:transition-none",
                                expanded && "rotate-90",
                              )}
                            />
                            <span className="truncate text-sm font-medium">{group.title}</span>
                            <span className="ml-auto text-xs text-muted-foreground">
                              {t("role.grant.groupCount", {
                                selected: groupSelected,
                                total: selectable.length,
                              })}
                            </span>
                          </button>
                        </CollapsibleTrigger>
                        <Checkbox
                          checked={allSelected ? true : someSelected ? "indeterminate" : false}
                          disabled={pending}
                          aria-label={t("role.grant.selectGroup", { module: group.title })}
                          onCheckedChange={(checked) =>
                            toggleGroup(
                              selectable.map((item) => item.name),
                              checked === true,
                            )
                          }
                        />
                      </div>
                      <CollapsibleContent>
                        <div className="grid gap-1 border-t p-2 sm:grid-cols-2">
                          {group.items.map((item) => {
                            const checked = selected.has(item.name)
                            // 菜单组的动作是能力 code，这里按能力目录渲染双语标签。
                            const action = permissionActionLabel(item)
                            const locked = OWNER_ONLY.has(item.name)
                            return (
                              <div
                                key={item.name}
                                className={cn(
                                  "flex min-h-11 items-center gap-2 rounded-sm border-l-2 border-transparent px-2 py-1",
                                  checked && "border-primary bg-primary/5 dark:bg-primary/10",
                                  locked && "opacity-70",
                                )}
                              >
                                <Checkbox
                                  id={`perm-${item.name}`}
                                  checked={checked}
                                  disabled={pending || locked}
                                  aria-describedby={locked ? `perm-locked-${item.name}` : undefined}
                                  onCheckedChange={(value) =>
                                    togglePermission(item.name, value === true)
                                  }
                                />
                                <Label
                                  htmlFor={`perm-${item.name}`}
                                  className={cn(
                                    "flex min-w-0 flex-1 items-center gap-2",
                                    locked ? "cursor-default" : "cursor-pointer",
                                  )}
                                >
                                  <span className="text-sm">{action.zh}</span>
                                  <span className="text-xs text-muted-foreground">{action.en}</span>
                                </Label>
                                {locked ? (
                                  <span
                                    id={`perm-locked-${item.name}`}
                                    className="rounded-4xl border px-1.5 text-[11px] whitespace-nowrap text-muted-foreground"
                                  >
                                    {t("role.grant.ownerOnlyBadge")}
                                  </span>
                                ) : (
                                  <span
                                    aria-hidden="true"
                                    className="hidden font-mono text-[11px] text-muted-foreground/70 sm:inline"
                                  >
                                    {item.name}
                                  </span>
                                )}
                              </div>
                            )
                          })}
                        </div>
                      </CollapsibleContent>
                    </Collapsible>
                  )
                })}
              </div>
            ) : null}
          </div>

          <SheetFooter className="flex-row items-center justify-between gap-2 border-t p-4">
            <p className="text-xs text-muted-foreground">{t("role.grant.ownerOnly")}</p>
            <div className="flex gap-2">
              {error ? <p className="text-sm text-destructive">{error}</p> : null}
              {pending ? <Loader2 className="size-4 animate-spin text-muted-foreground" /> : null}
              <Button type="button" variant="outline" onClick={requestClose}>
                {t("action.cancel")}
              </Button>
              <Button type="button" disabled={!dirty || pending} onClick={() => void save()}>
                {pending ? t("action.saving") : t("action.save")}
              </Button>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>
      <ConfirmAction
        open={confirmLeave}
        title={t("role.grant.unsavedTitle")}
        description={t("role.grant.unsavedDescription")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        onConfirm={() => {
          setConfirmLeave(false)
          onOpenChange(false)
        }}
        onOpenChange={setConfirmLeave}
      />
    </>
  )
}
