import {
  ArrowUpRight,
  Building2,
  Check,
  ChevronsUpDown,
  Loader2,
  Search,
  SearchX,
  Settings2,
  X,
} from "lucide-react"
import {
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent,
  type ReactNode,
} from "react"
import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import type { NamespaceItem } from "@/api/namespace/types"
import { useSession } from "@/auth/session"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/class-names"

type TenantGroup = {
  key: "current" | "others" | "disabled" | "results"
  label?: string
  items: NamespaceItem[]
}

// 租户没有 logo 时用名称首字生成稳定色块，避免每个租户看起来一模一样。
const avatarTones = [
  "bg-sky-500/10 text-sky-700 ring-sky-500/20 dark:text-sky-300",
  "bg-violet-500/10 text-violet-700 ring-violet-500/20 dark:text-violet-300",
  "bg-emerald-500/10 text-emerald-700 ring-emerald-500/20 dark:text-emerald-300",
  "bg-amber-500/10 text-amber-700 ring-amber-500/20 dark:text-amber-300",
  "bg-rose-500/10 text-rose-700 ring-rose-500/20 dark:text-rose-300",
  "bg-cyan-500/10 text-cyan-700 ring-cyan-500/20 dark:text-cyan-300",
]

function avatarTone(uid: string): string {
  let hash = 0
  for (const letter of uid) {
    hash = (hash * 31 + letter.charCodeAt(0)) >>> 0
  }
  return avatarTones[hash % avatarTones.length]
}

function tenantInitials(name: string): string {
  const trimmed = name.trim()
  if (!trimmed) {
    return "?"
  }
  if (/[\u3400-\u9fff]/.test(trimmed)) {
    return Array.from(trimmed)[0] ?? "?"
  }
  const parts = trimmed.split(/\s+/)
  if (parts.length > 1) {
    return `${parts[0][0] ?? ""}${parts[1][0] ?? ""}`.toUpperCase()
  }
  return Array.from(trimmed).slice(0, 2).join("").toUpperCase()
}

function optionId(uid: string): string {
  return `tenant-option-${uid}`
}

function Kbd({ children }: { children: ReactNode }) {
  return (
    <kbd className="inline-flex h-4 min-w-4 items-center justify-center rounded border border-border bg-muted/60 px-1 font-mono text-[10px] font-normal text-muted-foreground">
      {children}
    </kbd>
  )
}

function TenantAvatar({
  item,
  size = "md",
  dotClassName,
}: {
  item: NamespaceItem | null
  size?: "sm" | "md"
  dotClassName?: string
}) {
  const logo = item?.logo?.trim()
  const disabled = item?.status && item.status !== "ACTIVE"

  return (
    <span
      aria-hidden
      className={cn(
        "relative inline-flex shrink-0 items-center justify-center overflow-hidden rounded-lg ring-1 ring-inset",
        size === "sm" ? "size-7 text-[10px]" : "size-8 text-[11px]",
        logo
          ? "bg-muted ring-border/60"
          : item
            ? avatarTone(item.uid)
            : "bg-muted text-muted-foreground ring-border/60"
      )}
    >
      {logo ? (
        <img src={logo} alt="" className="size-full object-cover" />
      ) : item ? (
        <span className="font-semibold">{tenantInitials(item.name)}</span>
      ) : (
        <Building2 className="size-3.5" />
      )}
      {disabled ? (
        <span
          className={cn(
            "absolute right-0 bottom-0 size-2.5 rounded-full border-2 border-popover bg-amber-500",
            dotClassName
          )}
        />
      ) : null}
    </span>
  )
}

export function TenantSwitcher() {
  const { t } = useTranslation()
  const session = useSession()
  const { clearSwitchError, switchError } = session
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState("")
  const [activeIndex, setActiveIndex] = useState(0)
  const searchRef = useRef<HTMLInputElement>(null)
  const previousSwitchingToRef = useRef("")
  const listId = useId()

  const current =
    session.namespaces.find((item) => item.uid === session.namespaceUid) ?? null
  const switchingTarget = session.switchingTo
    ? (session.namespaces.find((item) => item.uid === session.switchingTo) ??
      null)
    : null
  const displayItem = switchingTarget ?? current
  const displayName =
    session.loading && !displayItem
      ? t("state.loading")
      : (displayItem?.name ?? t("namespace.none"))
  const normalizedQuery = query.trim().toLocaleLowerCase()

  const filtered = normalizedQuery
    ? session.namespaces.filter((item) => {
        const haystack = [
          item.name,
          item.ownerName,
          item.ownerEmail,
          item.remark,
        ]
          .filter(Boolean)
          .join(" ")
          .toLocaleLowerCase()
        return normalizedQuery
          .split(/\s+/)
          .every((token) => haystack.includes(token))
      })
    : session.namespaces

  const groups: TenantGroup[] = []
  if (normalizedQuery) {
    if (filtered.length > 0) {
      groups.push({ key: "results", items: filtered })
    }
  } else {
    const currentItems = current ? [current] : []
    const others = session.namespaces.filter(
      (item) => item.uid !== current?.uid
    )
    const activeItems = others.filter((item) => item.status === "ACTIVE")
    const disabledItems = others.filter((item) => item.status !== "ACTIVE")
    if (currentItems.length > 0) {
      groups.push({
        key: "current",
        label: t("namespace.switcher.currentSection"),
        items: currentItems,
      })
    }
    if (activeItems.length > 0) {
      groups.push({
        key: "others",
        label: t("namespace.switcher.otherSection"),
        items: activeItems,
      })
    }
    if (disabledItems.length > 0) {
      groups.push({
        key: "disabled",
        label: t("namespace.switcher.disabledSection"),
        items: disabledItems,
      })
    }
  }

  const flatItems = groups.flatMap((group) => group.items)
  const indexByUid = new Map(flatItems.map((item, index) => [item.uid, index]))
  const safeActiveIndex =
    flatItems.length === 0
      ? -1
      : Math.min(Math.max(activeIndex, 0), flatItems.length - 1)
  const activeItem =
    safeActiveIndex >= 0 ? flatItems[safeActiveIndex] : undefined
  const activeUid = activeItem?.uid

  useEffect(() => {
    if (!open || !activeUid) {
      return
    }
    document
      .getElementById(optionId(activeUid))
      ?.scrollIntoView({ block: "nearest" })
  }, [activeUid, open, query])

  useEffect(() => {
    const previous = previousSwitchingToRef.current
    previousSwitchingToRef.current = session.switchingTo
    if (!previous || session.switchingTo) {
      return
    }
    if (session.namespaceUid !== previous) {
      return
    }
    const name =
      session.namespaces.find((item) => item.uid === previous)?.name ?? previous
    toast.success(t("namespace.switcher.switched", { name }))
  }, [session.namespaceUid, session.namespaces, session.switchingTo, t])

  useEffect(() => {
    if (!switchError) {
      return
    }
    toast.error(t("namespace.switcher.switchFailed"), {
      description: switchError,
    })
    clearSwitchError()
  }, [clearSwitchError, switchError, t])

  function handleOpenChange(next: boolean) {
    setOpen(next)
    if (next) {
      setQuery("")
      setActiveIndex(0)
    }
  }

  function changeQuery(value: string) {
    setQuery(value)
    setActiveIndex(0)
  }

  function choose(item: NamespaceItem) {
    if (session.switching || session.loading) {
      return
    }
    handleOpenChange(false)
    if (item.uid === session.namespaceUid) {
      return
    }
    session.setNamespace(item.uid)
  }

  function onSearchKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    // 中文输入法组合期间不要抢占方向键和回车。
    if (event.nativeEvent.isComposing) {
      return
    }
    if (event.key === "ArrowDown") {
      if (flatItems.length === 0) {
        return
      }
      event.preventDefault()
      setActiveIndex(Math.min(safeActiveIndex + 1, flatItems.length - 1))
      return
    }
    if (event.key === "ArrowUp") {
      if (flatItems.length === 0) {
        return
      }
      event.preventDefault()
      setActiveIndex(Math.max(safeActiveIndex - 1, 0))
      return
    }
    if (event.key === "Enter" && activeItem) {
      event.preventDefault()
      choose(activeItem)
    }
  }

  function onTriggerKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault()
      handleOpenChange(true)
    }
  }

  const busy = session.loading || session.switching

  return (
    <>
      <Popover open={open} onOpenChange={handleOpenChange}>
        <PopoverTrigger asChild>
          <Button
            variant="ghost"
            className="h-8 max-w-[min(14rem,38vw)] justify-start gap-2 overflow-hidden px-1.5 font-normal"
            aria-label={t("namespace.switcher.trigger", { name: displayName })}
            aria-busy={busy}
            title={displayItem?.name}
            onKeyDown={onTriggerKeyDown}
          >
            <TenantAvatar
              item={displayItem}
              size="sm"
              dotClassName="border-background"
            />
            <span className="min-w-0 truncate text-sm font-medium">
              {displayName}
            </span>
            {busy ? (
              <Loader2
                className="size-3.5 animate-spin text-muted-foreground"
                aria-hidden
              />
            ) : (
              <ChevronsUpDown
                className="size-3.5 text-muted-foreground"
                aria-hidden
              />
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent
          align="start"
          sideOffset={6}
          collisionPadding={8}
          className="w-[min(22rem,calc(100vw-1rem))] overflow-hidden p-0"
          aria-label={t("namespace.switch")}
        >
          <div className="border-b border-border/60 px-2 pt-2 pb-2">
            <div className="flex items-center justify-between gap-2 px-1 pb-2">
              <span className="text-xs font-medium text-muted-foreground">
                {t("namespace.switch")}
              </span>
              <span className="font-mono text-[11px] text-muted-foreground">
                {t("namespace.switcher.count", {
                  count: session.namespaces.length,
                })}
              </span>
            </div>
            {session.namespaces.length > 0 ? (
              <div className="relative">
                <Search
                  className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground"
                  aria-hidden
                />
                <input
                  ref={searchRef}
                  type="text"
                  role="combobox"
                  aria-expanded
                  aria-autocomplete="list"
                  aria-label={t("namespace.switcher.search")}
                  aria-controls={filtered.length > 0 ? listId : undefined}
                  aria-activedescendant={
                    activeUid ? optionId(activeUid) : undefined
                  }
                  autoComplete="off"
                  spellCheck={false}
                  placeholder={t("namespace.switcher.searchPlaceholder")}
                  value={query}
                  onChange={(event) => changeQuery(event.target.value)}
                  onKeyDown={onSearchKeyDown}
                  className="h-8 w-full rounded-lg border border-input bg-transparent pr-8 pl-8 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring md:text-sm dark:bg-input/30"
                />
                {query ? (
                  <button
                    type="button"
                    aria-label={t("namespace.switcher.clearSearch")}
                    className="absolute top-1/2 right-1.5 flex size-6 -translate-y-1/2 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-hidden"
                    onClick={() => {
                      changeQuery("")
                      searchRef.current?.focus()
                    }}
                  >
                    <X className="size-3.5" />
                  </button>
                ) : null}
              </div>
            ) : null}
          </div>

          {session.switching && switchingTarget ? (
            <div className="flex items-center gap-2 border-b border-primary/15 bg-primary/5 px-3 py-2 text-xs text-primary-ink">
              <Loader2 className="size-3.5 animate-spin" aria-hidden />
              {t("namespace.switcher.switching", {
                name: switchingTarget.name,
              })}
            </div>
          ) : null}

          <div
            className="max-h-[min(20rem,calc(100vh-12rem))] overflow-y-auto overscroll-contain p-1"
            aria-busy={busy}
          >
            {session.loading && session.namespaces.length === 0 ? (
              <div className="space-y-1 p-1">
                {[0, 1, 2].map((index) => (
                  <div
                    key={index}
                    className="flex items-center gap-2.5 rounded-md px-2 py-1.5"
                  >
                    <Skeleton className="size-8 rounded-lg" />
                    <div className="flex-1 space-y-1.5">
                      <Skeleton className="h-3 w-1/2" />
                      <Skeleton className="h-2.5 w-1/3" />
                    </div>
                  </div>
                ))}
              </div>
            ) : session.namespaces.length === 0 ? (
              <div className="flex flex-col items-center gap-2 px-5 py-8 text-center">
                <span className="flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground">
                  <Building2 className="size-5" aria-hidden />
                </span>
                <p className="text-sm font-medium">
                  {t("namespace.switcher.emptyTitle")}
                </p>
                <p className="text-xs leading-relaxed text-muted-foreground">
                  {t("namespace.switcher.emptyHint")}
                </p>
                <Button asChild size="sm" className="mt-1">
                  <Link
                    to="/namespaces"
                    onClick={() => handleOpenChange(false)}
                  >
                    {t("namespace.create")}
                  </Link>
                </Button>
              </div>
            ) : filtered.length === 0 ? (
              <div className="flex flex-col items-center gap-2 px-5 py-7 text-center">
                <span className="flex size-10 items-center justify-center rounded-full bg-muted text-muted-foreground">
                  <SearchX className="size-5" aria-hidden />
                </span>
                <p className="text-sm font-medium">
                  {t("namespace.switcher.noResults")}
                </p>
                <p className="text-xs leading-relaxed text-muted-foreground">
                  {t("namespace.switcher.noResultsHint", {
                    query: query.trim(),
                  })}
                </p>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => {
                    changeQuery("")
                    searchRef.current?.focus()
                  }}
                >
                  {t("namespace.switcher.clearSearch")}
                </Button>
              </div>
            ) : (
              <div
                role="listbox"
                id={listId}
                aria-label={t("namespace.switcher.listLabel")}
              >
                {groups.map((group) => (
                  <div
                    key={group.key}
                    role="group"
                    aria-labelledby={
                      group.label ? `${listId}-${group.key}` : undefined
                    }
                  >
                    {group.label ? (
                      <div
                        id={`${listId}-${group.key}`}
                        className="sticky top-0 z-10 bg-popover/95 px-2 pt-2 pb-1 text-[11px] font-medium text-muted-foreground backdrop-blur-sm"
                      >
                        {group.label}
                      </div>
                    ) : null}
                    {group.items.map((item) => {
                      const index = indexByUid.get(item.uid) ?? -1
                      const isCurrent = item.uid === session.namespaceUid
                      const isActive = index === safeActiveIndex
                      const owner =
                        item.ownerName || item.ownerEmail || item.ownerUid
                      return (
                        <div
                          key={item.uid}
                          id={optionId(item.uid)}
                          role="option"
                          aria-selected={isCurrent}
                          data-active={isActive}
                          title={item.remark || item.name}
                          className={cn(
                            "group/tenant relative flex min-h-11 cursor-default items-center gap-2.5 rounded-md px-2 py-1.5 transition-colors select-none",
                            isActive && "bg-accent text-accent-foreground"
                          )}
                          onPointerMove={() => setActiveIndex(index)}
                          onClick={() => choose(item)}
                        >
                          <TenantAvatar item={item} />
                          <div className="min-w-0 flex-1">
                            <div className="flex min-w-0 items-center gap-1.5">
                              <span className="truncate text-sm font-medium">
                                {item.name}
                              </span>
                              {item.status !== "ACTIVE" ? (
                                <StatusBadge
                                  value={item.status}
                                  tone="warning"
                                  className="shrink-0"
                                />
                              ) : null}
                            </div>
                            {owner ? (
                              <p className="truncate text-xs text-muted-foreground">
                                {t("namespace.switcher.ownerLabel")} · {owner}
                              </p>
                            ) : null}
                          </div>
                          {isCurrent ? (
                            <Check
                              className="size-4 shrink-0 text-primary-ink"
                              aria-hidden
                            />
                          ) : null}
                        </div>
                      )
                    })}
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="flex items-center justify-between gap-2 border-t border-border/60 p-1.5">
            <Link
              to="/namespaces"
              onClick={() => handleOpenChange(false)}
              className="flex h-8 items-center gap-2 rounded-md px-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-hidden"
            >
              <Settings2 className="size-4" aria-hidden />
              {t("namespace.switcher.manage")}
              <ArrowUpRight className="ml-auto size-3.5" aria-hidden />
            </Link>
            {session.namespaces.length > 0 ? (
              <span className="hidden items-center gap-1 pr-1 text-[10px] text-muted-foreground sm:flex">
                <Kbd>↑</Kbd>
                <Kbd>↓</Kbd>
                <span>{t("namespace.switcher.hintNavigate")}</span>
                <Kbd>↵</Kbd>
                <span>{t("namespace.switcher.hintSelect")}</span>
              </span>
            ) : null}
          </div>

          <span className="sr-only" role="status" aria-live="polite">
            {normalizedQuery
              ? t("namespace.switcher.resultCount", { count: filtered.length })
              : ""}
          </span>
        </PopoverContent>
      </Popover>
      {/* 面板关闭后切换仍在进行，独立 live region 负责播报。 */}
      <span className="sr-only" role="status" aria-live="polite">
        {session.switching && switchingTarget
          ? t("namespace.switcher.switching", {
              name: switchingTarget.name,
            })
          : ""}
      </span>
    </>
  )
}
