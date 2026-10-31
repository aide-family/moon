import { useEffect, useRef, useState } from "react"
import { ChevronRight, Info, Loader2, Search } from "lucide-react"
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/class-names"
import {
  CAPABILITY_PERMISSIONS,
  capabilityDescription,
  capabilityLabel,
} from "@/lib/feature-catalog"
import { formErrorMessage } from "@/lib/form-error"
import {
  OWNER_ONLY,
  actionLabel,
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

interface MenuGroup {
  key: string
  title: string
  /** 菜单权限名 menu:<能力>；租户设置和未归类分组没有菜单权限。 */
  menuName?: string
  children: PermissionItem[]
  /** 菜单勾选后自动带上的读权限。 */
  defaultReads: Set<string>
  description?: string
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

function isDefaultRead(name: string): boolean {
  return name.endsWith(":read") || name.endsWith(":logs")
}

/** 「租户设置」是固定菜单：namespace:* 不属于任何套餐能力，但仍要有归类的菜单。 */
const NAMESPACE_MENU = "namespace"
const NAMESPACE_READ = "namespace:read"

/**
 * 角色权限配置器：功能权限跟随菜单。
 * - 一级是菜单（能力）权限，勾选菜单默认带上读权限；
 * - 二级是菜单下的子权限（写、删除、执行等），菜单未勾选时不可选；
 * - `namespace:*` 归到固定的「租户设置」菜单；套餐外的旧授权单独展示、默认保留。
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
  const { t, i18n } = useTranslation()
  const available = new Set(catalog.map((item) => item.name))
  const availableItem = new Map(catalog.map((item) => [item.name, item]))
  const menuItems = catalog.filter((item) => item.resource === "menu")
  const menuCodes = menuItems
    .map((item) => item.action)
    .filter((code) => CAPABILITY_PERMISSIONS[code] !== undefined)

  // 每个子权限只挂在一个菜单下：多个能力共享同一组接口权限时归到目录里第一个能力。
  const ownerByPermission = new Map<string, string>()
  for (const code of menuCodes) {
    for (const permission of CAPABILITY_PERMISSIONS[code] ?? []) {
      if (!ownerByPermission.has(permission)) {
        ownerByPermission.set(permission, code)
      }
    }
  }

  const namespaceItems = catalog.filter((item) => item.resource === "namespace")
  const groups: MenuGroup[] = menuCodes.map((code) => {
    const label = capabilityLabel(code)
    return {
      key: `cap-${code}`,
      title: `${label.zh} / ${label.en}`,
      menuName: `menu:${code}`,
      children: [...ownerByPermission.entries()]
        .filter(([permission, owner]) => owner === code && available.has(permission))
        .map(([permission]) => availableItem.get(permission)!)
        .filter(Boolean),
      defaultReads: new Set(
        (CAPABILITY_PERMISSIONS[code] ?? []).filter(
          (permission) => available.has(permission) && isDefaultRead(permission)
        )
      ),
      description: capabilityDescription(code, i18n.language),
    }
  })
  if (namespaceItems.length > 0) {
    groups.push({
      key: `fixed-${NAMESPACE_MENU}`,
      title: t("role.grant.namespaceMenu"),
      children: namespaceItems,
      defaultReads: available.has(NAMESPACE_READ)
        ? new Set([NAMESPACE_READ])
        : new Set(),
    })
  }
  const unclassified = catalog.filter(
    (item) =>
      item.resource !== "menu" &&
      item.resource !== "namespace" &&
      !ownerByPermission.has(item.name)
  )
  if (unclassified.length > 0) {
    groups.push({
      key: "unclassified",
      title: t("role.grant.unclassified"),
      children: unclassified,
      defaultReads: new Set(),
    })
  }

  // 每次打开都由页面用 key 重新挂载，初始值直接取当前角色。
  const held = role?.permissions ?? []
  const preservedNames = held.filter((name) => !available.has(name))
  const initialMenus = new Set(
    held.filter((name) => name.startsWith("menu:") && available.has(name))
  )
  const initialPermissions = new Set(
    held.filter((name) => !name.startsWith("menu:"))
  )
  const [baseline] = useState(
    () => new Set([...normalizeRolePermissions(held, available), ...preservedNames])
  )
  const [selectedMenus, setSelectedMenus] = useState(initialMenus)
  const [selectedPermissions, setSelectedPermissions] = useState(
    initialPermissions
  )
  const [selectedPreserved, setSelectedPreserved] = useState(
    new Set(preservedNames)
  )
  const [query, setQuery] = useState("")
  const [onlySelected, setOnlySelected] = useState(false)
  const [collapsed, setCollapsed] = useState<Set<string>>(new Set())
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [confirmLeave, setConfirmLeave] = useState(false)
  const [undoSnapshot, setUndoSnapshot] = useState<{
    menus: Set<string>
    permissions: Set<string>
    preserved: Set<string>
  } | null>(null)
  const undoTimer = useRef<number | undefined>(undefined)

  useEffect(() => () => window.clearTimeout(undoTimer.current), [])

  const selectedPayload = buildPayload({
    menus: selectedMenus,
    permissions: selectedPermissions,
    preserved: selectedPreserved,
    ownerByPermission,
    available,
  })
  const dirty = !sameSet(baseline, selectedPayload)
  const added = [...selectedPayload].filter((name) => !baseline.has(name)).length
  const removed = [...baseline].filter((name) => !selectedPayload.has(name)).length
  const ownerOnly = OWNER_ONLY

  const filtered = groups
    .map((group) => {
      const menuItem = group.menuName
        ? availableItem.get(group.menuName)
        : undefined
      const menuMatches =
        menuItem !== undefined && matchesPermission(menuItem, query)
      const menuSelected = group.menuName
        ? selectedMenus.has(group.menuName)
        : true
      const children = group.children.filter(
        (item) =>
          (query.trim() === "" || menuMatches || matchesPermission(item, query)) &&
          (!onlySelected ||
            menuSelected ||
            selectedPermissions.has(item.name) ||
            group.defaultReads.has(item.name))
      )
      const visible =
        query.trim() === ""
          ? !onlySelected || menuSelected ||
            group.children.some((item) => selectedPermissions.has(item.name))
          : menuMatches ||
            group.children.some((item) => matchesPermission(item, query))
      return { ...group, items: children, visible }
    })
    .filter((group) => group.visible)

  const preservedItems = [...selectedPreserved].filter((name) =>
    query.trim() === ""
      ? true
      : name.toLowerCase().includes(query.trim().toLowerCase())
  )
  const visibleMenus = filtered
    .map((group) => group.menuName)
    .filter((name): name is string => name !== undefined && !ownerOnly.has(name))
  const allVisibleMenusSelected =
    visibleMenus.length > 0 &&
    visibleMenus.every((name) => selectedMenus.has(name))

  function toggleMenu(name: string, checked: boolean) {
    setSelectedMenus((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(name)
      } else {
        next.delete(name)
      }
      return next
    })
  }

  function togglePermission(name: string, checked: boolean) {
    setSelectedPermissions((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(name)
      } else {
        next.delete(name)
      }
      return next
    })
  }

  function togglePreserved(name: string, checked: boolean) {
    setSelectedPreserved((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(name)
      } else {
        next.delete(name)
      }
      return next
    })
  }

  function toggleCollapse(key: string) {
    if (query.trim() !== "" || onlySelected) {
      return
    }
    setCollapsed((current) => {
      const next = new Set(current)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  function selectVisibleMenus() {
    setSelectedMenus((current) => {
      const next = new Set(current)
      for (const name of visibleMenus) {
        next.add(name)
      }
      return next
    })
    setSelectedPermissions((current) => {
      const next = new Set(current)
      for (const group of filtered) {
        for (const item of group.items) {
          if (!ownerOnly.has(item.name)) {
            next.add(item.name)
          }
        }
      }
      return next
    })
  }

  function clearAll() {
    setUndoSnapshot({
      menus: new Set(selectedMenus),
      permissions: new Set(selectedPermissions),
      preserved: new Set(selectedPreserved),
    })
    setSelectedMenus(new Set())
    setSelectedPermissions(new Set())
    setSelectedPreserved(new Set())
    window.clearTimeout(undoTimer.current)
    undoTimer.current = window.setTimeout(() => setUndoSnapshot(null), 5000)
  }

  function undoClear() {
    if (!undoSnapshot) {
      return
    }
    setSelectedMenus(new Set(undoSnapshot.menus))
    setSelectedPermissions(new Set(undoSnapshot.permissions))
    setSelectedPreserved(new Set(undoSnapshot.preserved))
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
      await onSave([...selectedPayload])
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

  const selectedCount = [...selectedPayload].filter(
    (name) => !ownerOnly.has(name)
  ).length
  const totalAssignable = [...available].filter(
    (name) => !ownerOnly.has(name)
  ).length

  return (
    <>
      <Sheet
        open={open}
        onOpenChange={(next) => (next ? onOpenChange(true) : requestClose())}
      >
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
                {t("role.grant.selected", {
                  selected: selectedCount,
                  total: totalAssignable,
                })}
              </span>
            </div>
            {role?.remark?.trim() ? (
              <p className="line-clamp-2 text-xs text-muted-foreground">
                {role.remark}
              </p>
            ) : null}
            <p className="text-xs text-muted-foreground">
              {t("role.grant.menuHint")}
            </p>
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
              onClick={() =>
                setCollapsed(new Set(groups.map((group) => group.key)))
              }
            >
              {t("role.grant.collapseAll")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={
                pending || allVisibleMenusSelected || visibleMenus.length === 0
              }
              onClick={selectVisibleMenus}
            >
              {query.trim()
                ? t("role.grant.selectMatches", { count: visibleMenus.length })
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
                disabled={pending || selectedPayload.size === 0}
                onClick={clearAll}
              >
                {t("role.grant.clear")}
              </Button>
            )}
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto p-4">
            {catalogLoading ? (
              <p className="text-sm text-muted-foreground">
                {t("state.loading")}
              </p>
            ) : null}
            {!catalogLoading && catalogFailed ? (
              <div className="flex items-center gap-2">
                <p className="text-sm text-destructive">
                  {t("role.grant.loadFailed")}
                </p>
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={onRetryCatalog}
                >
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
                {filtered.map((group) => {
                  const menuSelected = group.menuName
                    ? selectedMenus.has(group.menuName)
                    : true
                  const expanded =
                    query.trim() !== "" ||
                    onlySelected ||
                    !collapsed.has(group.key)
                  const groupSelected = group.children.filter(
                    (item) =>
                      group.defaultReads.has(item.name) ||
                      (menuSelected && selectedPermissions.has(item.name))
                  ).length
                  return (
                    <Collapsible
                      key={group.key}
                      open={expanded}
                      onOpenChange={() => toggleCollapse(group.key)}
                      className="rounded-md border"
                    >
                      <div className="flex min-h-11 items-center gap-2 px-2">
                        <CollapsibleTrigger asChild>
                          <button
                            type="button"
                            className={cn(
                              "flex min-w-0 flex-1 items-center gap-2 py-2 text-left",
                              (query.trim() !== "" || onlySelected) &&
                                "cursor-default"
                            )}
                            aria-label={group.title}
                          >
                            <ChevronRight
                              className={cn(
                                "size-4 shrink-0 text-muted-foreground transition-transform duration-150 motion-reduce:transition-none",
                                expanded && "rotate-90"
                              )}
                            />
                            <span className="truncate text-sm font-medium">
                              {group.title}
                            </span>
                            {group.description ? (
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <button
                                    type="button"
                                    aria-label={group.description}
                                    className="shrink-0 rounded p-0.5 text-muted-foreground hover:text-foreground"
                                    onClick={(event) =>
                                      event.stopPropagation()
                                    }
                                  >
                                    <Info className="size-3.5" />
                                  </button>
                                </TooltipTrigger>
                                <TooltipContent className="max-w-xs">
                                  {group.description}
                                </TooltipContent>
                              </Tooltip>
                            ) : null}
                            <span className="ml-auto text-xs text-muted-foreground">
                              {t("role.grant.groupCount", {
                                selected: groupSelected,
                                total: group.children.length,
                              })}
                            </span>
                          </button>
                        </CollapsibleTrigger>
                        {group.menuName ? (
                          <Checkbox
                            checked={menuSelected}
                            disabled={pending}
                            aria-label={t("role.grant.selectMenu", {
                              menu: group.title,
                            })}
                            onCheckedChange={(checked) =>
                              toggleMenu(group.menuName!, checked === true)
                            }
                          />
                        ) : null}
                      </div>
                      <CollapsibleContent>
                        <div className="grid gap-1 border-t p-2 sm:grid-cols-2">
                          {group.items.length === 0 ? (
                            <p className="text-xs text-muted-foreground">
                              {t("role.grant.noChildren")}
                            </p>
                          ) : null}
                          {group.items.map((item) => {
                            const readLocked =
                              group.defaultReads.has(item.name) && menuSelected
                            const locked = ownerOnly.has(item.name) || readLocked
                            const checked = locked
                              ? true
                              : menuSelected &&
                                selectedPermissions.has(item.name)
                            const action = permissionActionLabel(item)
                            return (
                              <div
                                key={item.name}
                                className={cn(
                                  "flex min-h-11 items-center gap-2 rounded-sm border-l-2 border-transparent px-2 py-1",
                                  checked &&
                                    "border-primary bg-primary/5 dark:bg-primary/10",
                                  locked && "opacity-80"
                                )}
                              >
                                <Checkbox
                                  id={`perm-${item.name}`}
                                  checked={checked}
                                  disabled={pending || locked || !menuSelected}
                                  onCheckedChange={(value) =>
                                    togglePermission(item.name, value === true)
                                  }
                                />
                                <Label
                                  htmlFor={`perm-${item.name}`}
                                  className={cn(
                                    "flex min-w-0 flex-1 items-center gap-2",
                                    locked || !menuSelected
                                      ? "cursor-default"
                                      : "cursor-pointer"
                                  )}
                                >
                                  <span className="text-sm">{action.zh}</span>
                                  <span className="text-xs text-muted-foreground">
                                    {action.en}
                                  </span>
                                  {readLocked ? (
                                    <span className="rounded-4xl border px-1.5 text-[10px] whitespace-nowrap text-muted-foreground">
                                      {t("role.grant.readDefault")}
                                    </span>
                                  ) : null}
                                </Label>
                                {ownerOnly.has(item.name) ? (
                                  <span className="rounded-4xl border px-1.5 text-[11px] whitespace-nowrap text-muted-foreground">
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
                {selectedPreserved.size > 0 && preservedItems.length > 0 ? (
                  <div className="rounded-md border">
                    <div className="flex min-h-11 items-center gap-2 px-2">
                      <span className="flex min-w-0 flex-1 items-center gap-2 py-2">
                        <ChevronRight className="size-4 shrink-0 rotate-90 text-muted-foreground" />
                        <span className="truncate text-sm font-medium">
                          {t("role.grant.preserved")}
                        </span>
                        <span className="ml-auto text-xs text-muted-foreground">
                          {selectedPreserved.size}
                        </span>
                      </span>
                    </div>
                    <div className="grid gap-1 border-t p-2 sm:grid-cols-2">
                      {preservedItems.map((name) => {
                        const label = presetLabel(name)
                        return (
                          <div
                            key={name}
                            className="flex min-h-11 items-center gap-2 rounded-sm border-l-2 border-transparent px-2 py-1"
                          >
                            <Checkbox
                              id={`perm-${name}`}
                              checked={selectedPreserved.has(name)}
                              disabled={pending}
                              onCheckedChange={(value) =>
                                togglePreserved(name, value === true)
                              }
                            />
                            <Label
                              htmlFor={`perm-${name}`}
                              className="flex min-w-0 flex-1 items-center gap-2"
                            >
                              <span className="text-sm">{label.zh}</span>
                              <span className="text-xs text-muted-foreground">
                                {label.en}
                              </span>
                            </Label>
                            <span
                              aria-hidden="true"
                              className="hidden font-mono text-[11px] text-muted-foreground/70 sm:inline"
                            >
                              {name}
                            </span>
                          </div>
                        )
                      })}
                    </div>
                  </div>
                ) : null}
              </div>
            ) : null}
          </div>

          <SheetFooter className="flex-row items-center justify-between gap-2 border-t p-4">
            <p className="text-xs text-muted-foreground">
              {t("role.grant.ownerOnly")}
            </p>
            <div className="flex gap-2">
              {error ? <p className="text-sm text-destructive">{error}</p> : null}
              {pending ? (
                <Loader2 className="size-4 animate-spin text-muted-foreground" />
              ) : null}
              <Button type="button" variant="outline" onClick={requestClose}>
                {t("action.cancel")}
              </Button>
              <Button
                type="button"
                disabled={!dirty || pending}
                onClick={() => void save()}
              >
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

  /** 套餐外权限的展示文案：菜单权限按能力标签，其它按 resource/action 拆分。 */
  function presetLabel(name: string) {
    if (name.startsWith("menu:")) {
      return capabilityLabel(name.slice(5))
    }
    const parts = name.split(":")
    return actionLabel(parts[parts.length - 1] ?? name)
  }
}

/** 打开时的权限集合：有菜单的补默认读权限，其它权限原样保留。 */
function normalizeRolePermissions(
  permissions: string[],
  available: Set<string>
): string[] {
  const out = new Set<string>()
  for (const name of permissions) {
    if (available.has(name)) {
      out.add(name)
    }
  }
  for (const name of [...out]) {
    if (!name.startsWith("menu:")) {
      continue
    }
    const capability = name.slice(5)
    for (const permission of CAPABILITY_PERMISSIONS[capability] ?? []) {
      if (available.has(permission) && isDefaultRead(permission)) {
        out.add(permission)
      }
    }
  }
  if (available.has(NAMESPACE_READ)) {
    out.add(NAMESPACE_READ)
  }
  return [...out]
}

interface BuildPayloadOptions {
  menus: Set<string>
  permissions: Set<string>
  preserved: Set<string>
  ownerByPermission: Map<string, string>
  available: Set<string>
}

/** 由当前勾选推导最终提交的权限：菜单 + 默认读 + 已选子权限 + 固定菜单 + 保留项。 */
function buildPayload({
  menus,
  permissions,
  preserved,
  ownerByPermission,
  available,
}: BuildPayloadOptions): Set<string> {
  const out = new Set<string>()
  for (const menu of menus) {
    if (!menu.startsWith("menu:") || !available.has(menu)) {
      continue
    }
    out.add(menu)
    const capability = menu.slice(5)
    for (const permission of CAPABILITY_PERMISSIONS[capability] ?? []) {
      if (available.has(permission) && isDefaultRead(permission)) {
        out.add(permission)
      }
    }
  }
  if (available.has(NAMESPACE_READ)) {
    out.add(NAMESPACE_READ)
  }
  for (const permission of permissions) {
    if (!available.has(permission)) {
      continue
    }
    if (permission.startsWith("namespace:")) {
      out.add(permission)
      continue
    }
    const owner = ownerByPermission.get(permission)
    if (!owner || menus.has(`menu:${owner}`)) {
      out.add(permission)
    }
  }
  for (const permission of preserved) {
    out.add(permission)
  }
  return out
}
