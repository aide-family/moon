import { useEffect, useRef, useState } from "react"
import { ChevronRight, Loader2, Search } from "lucide-react"
import { useTranslation } from "react-i18next"

import type {
  PermissionItem,
  PermissionMenu,
  RoleItem,
} from "@/api/role/types"
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
import { menuLabel, moduleDefinition } from "@/lib/feature-catalog"
import { formErrorMessage } from "@/lib/form-error"
import {
  OWNER_ONLY,
  matchesPermission,
  permissionActionLabel,
} from "@/lib/permission-catalog"

interface PermissionEditorProps {
  open: boolean
  role: RoleItem | null
  /** 后端返回的完整权限目录，用于显示没有菜单归属的旧授权。 */
  catalog: PermissionItem[]
  /** 后端返回的菜单目录。授权页按「分组 → 菜单 → 子权限」渲染。 */
  menus: PermissionMenu[]
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

const TENANT_GROUP = { zh: "租户与成员", en: "Tenant & members" }

/**
 * 角色权限配置器：功能权限按菜单归类，布局为「分组 → 菜单 → 子权限」。
 * - 勾选菜单默认包含该菜单的读权限（锁定不可取消）；
 * - 菜单下的写/删/执行子权限须先勾选菜单；
 * - 多个菜单共享的权限在分组顶部只出现一次，避免同一个权限出现多份勾选框；
 * - `namespace:*` 归入固定「租户设置」菜单，套餐外旧授权单独展示并默认保留。
 */
export function PermissionEditor({
  open,
  role,
  catalog,
  menus,
  catalogLoading,
  catalogFailed,
  onRetryCatalog,
  onOpenChange,
  onSave,
}: PermissionEditorProps) {
  const { t, i18n } = useTranslation()
  const available = new Set(catalog.map((item) => item.name))
  const menuByCode = new Map(menus.map((menu) => [menu.code, menu]))
  const held = role?.permissions ?? []
  const heldMenus = new Set(
    held
      .filter((name) => name.startsWith("menu:"))
      .map((name) => name.slice(5))
      .filter((code) => menuByCode.has(code))
  )
  const heldPermissions = new Set(
    held.filter((name) => !name.startsWith("menu:") && available.has(name))
  )
  const preservedNames = held.filter((name) => !available.has(name))
  // 旧授权可能只有子权限、没有菜单权限；打开时按子权限补上声明的菜单，
  // 与后端 normalizeRoleGrant 的规则保持一致，避免保存时静默丢掉授权。
  const initialMenus = new Set(heldMenus)
  for (const name of heldPermissions) {
    const declaring = menus.filter((menu) =>
      menu.permissions.some((item) => item.name === name)
    )
    if (
      declaring.length > 0 &&
      !declaring.some((menu) => initialMenus.has(menu.code))
    ) {
      initialMenus.add(declaring[0].code)
    }
  }

  const [baseline] = useState(
    () => new Set(buildPayload(initialMenus, heldPermissions, new Set(preservedNames), menus, available))
  )
  const [selectedMenus, setSelectedMenus] = useState(initialMenus)
  const [selectedPermissions, setSelectedPermissions] = useState(heldPermissions)
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

  const selectedPayload = buildPayload(
    selectedMenus,
    selectedPermissions,
    selectedPreserved,
    menus,
    available
  )
  const dirty = !sameSet(baseline, selectedPayload)
  const added = [...selectedPayload].filter((name) => !baseline.has(name)).length
  const removed = [...baseline].filter((name) => !selectedPayload.has(name)).length
  const ownerOnly = OWNER_ONLY
  const totalAssignable = [...available].filter(
    (name) => !ownerOnly.has(name)
  ).length

  // 按后端菜单顺序分组；共享权限（被同组多个菜单声明）只在分组顶部出现一次。
  const groups = buildGroups(menus)

  const filteredGroups = groups
    .map((group) => {
      const groupMatches = group.title.toLowerCase().includes(query.trim().toLowerCase())
      const visibleMenus = group.menus.filter((menu) => {
        const label = menuLabelText(menu)
        const menuMatches =
          groupMatches ||
          label.toLowerCase().includes(query.trim().toLowerCase()) ||
          menu.permissions.some(
            (item) =>
              query.trim() === "" ||
              matchesPermission(item, query) ||
              permissionActionLabel(item).zh.includes(query.trim())
          )
        const selected =
          selectedMenus.has(menu.code) ||
          menu.permissions.some((item) => selectedPermissions.has(item.name))
        return menuMatches && (!onlySelected || selected)
      })
      const sharedVisible = group.shared.filter(
        (item) =>
          !onlySelected ||
          group.menus.some((menu) => selectedMenus.has(menu.code)) ||
          selectedPermissions.has(item.name)
      )
      return {
        ...group,
        menus: visibleMenus,
        shared: sharedVisible,
        visible: visibleMenus.length > 0 || sharedVisible.length > 0,
      }
    })
    .filter((group) => group.visible)

  const preservedItems = [...selectedPreserved].filter((name) =>
    query.trim() === ""
      ? true
      : name.toLowerCase().includes(query.trim().toLowerCase())
  )
  const visibleMenuCodes = filteredGroups.flatMap((group) =>
    group.menus.map((menu) => menu.code).filter((code) => !ownerOnly.has(code))
  )
  const allVisibleSelected =
    visibleMenuCodes.length > 0 &&
    visibleMenuCodes.every((code) => selectedMenus.has(code))

  function menuLabelText(menu: PermissionMenu): string {
    const label = menuLabel(menu.code, menu.name)
    return i18n.language.startsWith("zh") ? label.zh : label.en
  }

  function toggleMenu(code: string, checked: boolean) {
    setSelectedMenus((current) => {
      const next = new Set(current)
      if (checked) {
        next.add(code)
      } else {
        next.delete(code)
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

  function readLocked(menu: PermissionMenu, name: string): boolean {
    return menu.defaultReads.includes(name) && selectedMenus.has(menu.code)
  }

  function toggleCollapse(group: string) {
    if (query.trim() !== "" || onlySelected) {
      return
    }
    setCollapsed((current) => {
      const next = new Set(current)
      if (next.has(group)) {
        next.delete(group)
      } else {
        next.add(group)
      }
      return next
    })
  }

  function selectAll() {
    setSelectedMenus((current) => {
      const next = new Set(current)
      for (const code of visibleMenuCodes) {
        next.add(code)
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

  function groupTitle(group: string): string {
    const label =
      moduleDefinition(group)?.label ??
      (group === "tenant" ? TENANT_GROUP : { zh: group, en: group })
    return `${label.zh} · ${label.en}`
  }

  function menuCount(menu: PermissionMenu): string {
    if (!selectedMenus.has(menu.code)) {
      return t("role.grant.menuOff")
    }
    return t("role.permissionCount", {
      count: menu.permissions.length,
    })
  }

  const selectedCount = [...selectedPayload].filter(
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
                setCollapsed(new Set(groups.map((group) => group.group)))
              }
            >
              {t("role.grant.collapseAll")}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={pending || allVisibleSelected || visibleMenuCodes.length === 0}
              onClick={selectAll}
            >
              {query.trim()
                ? t("role.grant.selectMatches", { count: visibleMenuCodes.length })
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
            {!catalogLoading && !catalogFailed && filteredGroups.length === 0 ? (
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
              <div className="grid gap-3">
                {filteredGroups.map((group) => {
                  const expanded =
                    query.trim() !== "" ||
                    onlySelected ||
                    !collapsed.has(group.group)
                  return (
                    <Collapsible
                      key={group.group}
                      open={expanded}
                      onOpenChange={() => toggleCollapse(group.group)}
                      className="rounded-md border"
                    >
                      <CollapsibleTrigger asChild>
                        <button
                          type="button"
                          className={cn(
                            "flex min-h-11 w-full items-center gap-2 px-3 py-2 text-left",
                            (query.trim() !== "" || onlySelected) &&
                              "cursor-default"
                          )}
                          aria-label={groupTitle(group.group)}
                        >
                          <ChevronRight
                            className={cn(
                              "size-4 shrink-0 text-muted-foreground transition-transform duration-150 motion-reduce:transition-none",
                              expanded && "rotate-90"
                            )}
                          />
                          <span className="text-sm font-semibold">
                            {groupTitle(group.group)}
                          </span>
                          <span className="ml-auto text-xs text-muted-foreground">
                            {t("role.grant.menuCount", {
                              count: group.menus.length,
                            })}
                          </span>
                        </button>
                      </CollapsibleTrigger>
                      <CollapsibleContent>
                        <div className="grid gap-3 border-t p-3">
                          {group.shared.length > 0 ? (
                            <div className="grid gap-1.5 rounded-md border border-dashed border-border bg-muted/20 p-2">
                              <span className="text-xs font-medium text-muted-foreground">
                                {t("role.grant.sharedPermissions")}
                              </span>
                              <div className="grid gap-1 sm:grid-cols-2">
                                {group.shared.map((item) => {
                                  const locked =
                                    isReadAction(item.name) &&
                                    group.menus.some((menu) =>
                                      selectedMenus.has(menu.code)
                                    )
                                  const checked =
                                    group.menus.some((menu) =>
                                      menu.defaultReads.includes(item.name) &&
                                      selectedMenus.has(menu.code)
                                    ) || selectedPermissions.has(item.name)
                                  const anyMenuSelected = group.menus.some(
                                    (menu) => selectedMenus.has(menu.code)
                                  )
                                  const action = permissionActionLabel(item)
                                  return (
                                    <div
                                      key={item.name}
                                      className={cn(
                                        "flex min-h-9 items-center gap-2 rounded-sm px-2 py-1",
                                        checked &&
                                          "bg-primary/5 dark:bg-primary/10"
                                      )}
                                    >
                                      <Checkbox
                                        id={`shared-${item.name}`}
                                        checked={checked}
                                        disabled={
                                          pending || locked || !anyMenuSelected
                                        }
                                        onCheckedChange={(value) =>
                                          togglePermission(
                                            item.name,
                                            value === true
                                          )
                                        }
                                      />
                                      <Label
                                        htmlFor={`shared-${item.name}`}
                                        className="flex min-w-0 flex-1 items-center gap-2"
                                      >
                                        <span className="text-sm">{action.zh}</span>
                                        <span className="text-xs text-muted-foreground">
                                          {action.en}
                                        </span>
                                        {locked ? (
                                          <span className="rounded-4xl border px-1.5 text-[10px] whitespace-nowrap text-muted-foreground">
                                            {t("role.grant.readDefault")}
                                          </span>
                                        ) : null}
                                      </Label>
                                      <span
                                        aria-hidden="true"
                                        className="hidden font-mono text-[11px] text-muted-foreground/60 sm:inline"
                                      >
                                        {item.name}
                                      </span>
                                    </div>
                                  )
                                })}
                              </div>
                            </div>
                          ) : null}
                          {group.menus.map((menu) => {
                            const label = menuLabel(menu.code, menu.name)
                            const selected = selectedMenus.has(menu.code)
                            const specific = menu.permissions.filter(
                              (item) =>
                                !group.shared.some(
                                  (shared) => shared.name === item.name
                                )
                            )
                            return (
                              <div
                                key={menu.code}
                                className={cn(
                                  "grid gap-1.5 rounded-md border p-2",
                                  selected && "border-primary/40 bg-primary/5"
                                )}
                              >
                                <div className="flex items-center gap-2">
                                  <Checkbox
                                    id={`menu-${menu.code}`}
                                    checked={selected}
                                    disabled={pending}
                                    aria-label={t("role.grant.selectMenu", {
                                      menu: label.zh,
                                    })}
                                    onCheckedChange={(checked) =>
                                      toggleMenu(menu.code, checked === true)
                                    }
                                  />
                                  <Label
                                    htmlFor={`menu-${menu.code}`}
                                    className="flex min-w-0 flex-1 items-center gap-2"
                                  >
                                    <span className="text-sm font-medium">
                                      {label.zh}
                                    </span>
                                    <span className="text-xs text-muted-foreground">
                                      {label.en}
                                    </span>
                                    {menu.always ? (
                                      <span className="rounded-4xl border px-1.5 text-[10px] whitespace-nowrap text-muted-foreground">
                                        {t("role.grant.alwaysMenu")}
                                      </span>
                                    ) : null}
                                  </Label>
                                  <span className="text-xs text-muted-foreground tabular-nums">
                                    {menuCount(menu)}
                                  </span>
                                </div>
                                <div className="grid gap-1 pl-6 sm:grid-cols-2">
                                  {specific.length === 0 ? (
                                    <span className="text-xs text-muted-foreground">
                                      {t("role.grant.onlyShared")}
                                    </span>
                                  ) : null}
                                  {specific.map((item) => {
                                    const locked = readLocked(menu, item.name)
                                    const checked = locked
                                      ? true
                                      : selected &&
                                        selectedPermissions.has(item.name)
                                    const action = permissionActionLabel(item)
                                    return (
                                      <div
                                        key={item.name}
                                        className={cn(
                                          "flex min-h-9 items-center gap-2 rounded-sm px-2 py-1",
                                          checked &&
                                            "bg-primary/5 dark:bg-primary/10"
                                        )}
                                      >
                                        <Checkbox
                                          id={`perm-${menu.code}-${item.name}`}
                                          checked={checked}
                                          disabled={
                                            pending || locked || !selected
                                          }
                                          onCheckedChange={(value) =>
                                            togglePermission(
                                              item.name,
                                              value === true
                                            )
                                          }
                                        />
                                        <Label
                                          htmlFor={`perm-${menu.code}-${item.name}`}
                                          className="flex min-w-0 flex-1 items-center gap-2"
                                        >
                                          <span className="text-sm">
                                            {action.zh}
                                          </span>
                                          <span className="text-xs text-muted-foreground">
                                            {action.en}
                                          </span>
                                          {locked ? (
                                            <span className="rounded-4xl border px-1.5 text-[10px] whitespace-nowrap text-muted-foreground">
                                              {t("role.grant.readDefault")}
                                            </span>
                                          ) : null}
                                        </Label>
                                        {ownerOnly.has(item.name) ? (
                                          <span className="rounded-4xl border px-1.5 text-[11px] whitespace-nowrap text-muted-foreground">
                                            {t("role.grant.ownerOnlyBadge")}
                                          </span>
                                        ) : null}
                                      </div>
                                    )
                                  })}
                                </div>
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
                    <div className="flex min-h-11 items-center gap-2 px-3 py-2">
                      <ChevronRight className="size-4 shrink-0 rotate-90 text-muted-foreground" />
                      <span className="text-sm font-medium">
                        {t("role.grant.preserved")}
                      </span>
                      <span className="ml-auto text-xs text-muted-foreground">
                        {selectedPreserved.size}
                      </span>
                    </div>
                    <div className="grid gap-1 border-t p-2 sm:grid-cols-2">
                      {preservedItems.map((name) => {
                        const label = presetLabel(name)
                        return (
                          <div
                            key={name}
                            className="flex min-h-9 items-center gap-2 rounded-sm px-2 py-1"
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
                              className="hidden font-mono text-[11px] text-muted-foreground/60 sm:inline"
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

  function presetLabel(name: string) {
    if (name.startsWith("menu:")) {
      const label = menuLabel(name.slice(5))
      return label
    }
    const parts = name.split(":")
    return permissionActionLabel({
      resource: parts.slice(0, -1).join(":"),
      action: parts[parts.length - 1] ?? name,
    })
  }
}

type MenuGroupView = {
  group: string
  title: string
  menus: PermissionMenu[]
  shared: PermissionItem[]
}

/** 按菜单分组，并把同组内被多个菜单共享的权限提取出来只展示一次。 */
function buildGroups(menus: PermissionMenu[]): MenuGroupView[] {
  const groups = new Map<string, PermissionMenu[]>()
  for (const menu of menus) {
    const list = groups.get(menu.group)
    if (list) {
      list.push(menu)
    } else {
      groups.set(menu.group, [menu])
    }
  }
  return [...groups.entries()].map(([group, groupMenus]) => {
    const counts = new Map<string, PermissionItem>()
    const occurrences = new Map<string, number>()
    for (const menu of groupMenus) {
      for (const item of menu.permissions) {
        counts.set(item.name, item)
        occurrences.set(item.name, (occurrences.get(item.name) ?? 0) + 1)
      }
    }
    const shared = [...counts.values()].filter(
      (item) => (occurrences.get(item.name) ?? 0) > 1
    )
    return {
      group,
      title: group,
      menus: groupMenus,
      shared,
    }
  })
}

function isReadAction(name: string): boolean {
  return name.endsWith(":read") || name.endsWith(":logs")
}

/** 由勾选推导提交内容：菜单 + 菜单默认读 + 已选子权限 + 保留项。 */
function buildPayload(
  menus: Set<string>,
  permissions: Set<string>,
  preserved: Set<string>,
  menuList: PermissionMenu[],
  available: Set<string>
): Set<string> {
  const out = new Set<string>()
  const menuByCode = new Map(menuList.map((menu) => [menu.code, menu]))
  const declaring = new Map<string, PermissionMenu[]>()
  for (const menu of menuList) {
    for (const item of menu.permissions) {
      const list = declaring.get(item.name)
      if (list) {
        list.push(menu)
      } else {
        declaring.set(item.name, [menu])
      }
    }
  }
  for (const code of menus) {
    const menu = menuByCode.get(code)
    if (!menu) {
      continue
    }
    out.add(`menu:${code}`)
    for (const read of menu.defaultReads) {
      if (available.has(read)) {
        out.add(read)
      }
    }
  }
  for (const permission of permissions) {
    if (!available.has(permission)) {
      continue
    }
    const declaringMenus = declaring.get(permission) ?? []
    if (
      declaringMenus.length === 0 ||
      declaringMenus.some((menu) => menus.has(menu.code))
    ) {
      out.add(permission)
    }
  }
  for (const permission of preserved) {
    out.add(permission)
  }
  return out
}
