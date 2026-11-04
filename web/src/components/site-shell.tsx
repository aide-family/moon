import {
  Building2,
  ChevronDown,
  LayoutDashboard,
  Loader2,
  LogOut,
  UserRound,
} from "lucide-react"
import {
  createContext,
  useContext,
  useEffect,
  useId,
  useMemo,
  useState,
  type KeyboardEvent,
  type ReactNode,
} from "react"
import { useTranslation } from "react-i18next"
import { Link, useLocation, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import type { NamespaceItem } from "@/api/namespace/types"
import { listSelfNamespaces } from "@/api/self"
import type { UserItem } from "@/api/user/types"
import { AppearanceMenu } from "@/components/appearance-menu"
import { LocaleMenu } from "@/components/locale-menu"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { useSiteConfig } from "@/hooks/use-site-config"
import { cn } from "@/lib/class-names"

/** 公开官网的登录态与「进入后台」入口；页头账户菜单和首页 CTA 共用。 */
interface SiteAccountValue {
  loggedIn: boolean
  user: UserItem | null
  entering: boolean
  enterAdmin: () => void
  loginHref: string
  registerHref: string
}

const SiteAccountContext = createContext<SiteAccountValue | null>(null)

export function useSiteAccount(): SiteAccountValue {
  const value = useContext(SiteAccountContext)
  if (!value) {
    throw new Error("useSiteAccount must be used inside SiteShell")
  }
  return value
}

type TenantDialog = "none" | "pick" | null

function accountName(user: UserItem | null): string {
  return user?.nickname || user?.name || user?.email || ""
}

function initials(name: string): string {
  const trimmed = name.trim()
  if (!trimmed) {
    return "A"
  }
  return Array.from(trimmed)[0]?.toLocaleUpperCase() ?? "A"
}

/** 公开页读取当前账号只为了展示账户菜单，不能用会 401 跳登录的 @/api/http。 */
async function fetchPublicSelfUser(token: string): Promise<UserItem | null> {
  const response = await fetch("/v1/self/info", {
    method: "GET",
    headers: {
      Accept: "application/json",
      Authorization: `Bearer ${token}`,
    },
    cache: "no-store",
  })
  if (response.status === 401) {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    return null
  }
  if (!response.ok) {
    return null
  }
  const data = (await response.json()) as { user?: UserItem }
  return data.user ?? null
}

function TenantOption({
  item,
  id,
  active,
  onHover,
  onSelect,
}: {
  item: NamespaceItem
  id: string
  active: boolean
  onHover: () => void
  onSelect: () => void
}) {
  return (
    <button
      type="button"
      role="option"
      id={id}
      aria-selected={active}
      tabIndex={-1}
      onMouseEnter={onHover}
      onFocus={onHover}
      onClick={onSelect}
      className={cn(
        "flex w-full min-w-0 items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-sm outline-none transition-colors",
        active ? "bg-accent text-accent-foreground" : "hover:bg-muted"
      )}
    >
      <span
        aria-hidden="true"
        className="grid size-7 shrink-0 place-items-center overflow-hidden rounded-md bg-primary/10 text-primary-ink"
      >
        {item.logo ? (
          <img src={item.logo} alt="" className="size-full object-cover" />
        ) : (
          <Building2 className="size-3.5" />
        )}
      </span>
      <span className="min-w-0 flex-1 truncate font-medium">{item.name}</span>
    </button>
  )
}

/**
 * 官网公共外壳：页头（品牌 / 主题 / 语言 / 账户）、内容区、页脚。
 *
 * 不挂 SessionProvider、不读租户权限：公开页只依赖公开接口；登录状态只用于
 * 展示账户菜单和「进入后台」。页脚备案复用 useSiteConfig，ICP 为空时整行不渲染。
 */
export function SiteShell({ children }: { children: ReactNode }) {
  const { t } = useTranslation()
  const { config } = useSiteConfig()
  const navigate = useNavigate()
  const location = useLocation()
  const prefix = useId().replace(/[^a-zA-Z0-9_-]/g, "")
  const [token, setToken] = useState(() => localStorage.getItem("token") ?? "")
  const [user, setUser] = useState<UserItem | null>(null)
  const [entering, setEntering] = useState(false)
  const [tenantDialog, setTenantDialog] = useState<TenantDialog>(null)
  const [tenants, setTenants] = useState<NamespaceItem[]>([])
  const [query, setQuery] = useState("")
  const [highlight, setHighlight] = useState(0)

  const loggedIn = token !== ""
  const currentPath = `${location.pathname}${location.search}${location.hash}`
  const loginHref = `/login?next=${encodeURIComponent(currentPath)}`
  const registerHref = `/register?next=${encodeURIComponent(currentPath)}`
  const siteName = config?.site.siteName.trim() || t("site.brandFallback")
  const logoUrl = config?.site.logoUrl.trim() || "/alioth.svg"
  const displayName = accountName(user) || t("site.account")

  useEffect(() => {
    if (!token) {
      return
    }
    let active = true
    void fetchPublicSelfUser(token)
      .then((next) => {
        if (!active) {
          return
        }
        if (next) {
          setUser(next)
        } else if (!localStorage.getItem("token")) {
          // 401 已清掉过期会话：回到未登录页头，不把官网劫持到登录页。
          setUser(null)
          setToken("")
        }
      })
      .catch(() => {
        // 网络故障时保留 token 判断为已登录，菜单里的跳转仍由后台接口兜底。
      })
    return () => {
      active = false
    }
  }, [token])

  const activeTenants = useMemo(
    () => tenants.filter((tenant) => tenant.status === "ACTIVE"),
    [tenants]
  )
  const filteredTenants = useMemo(() => {
    const keyword = query.trim().toLocaleLowerCase()
    if (!keyword) {
      return activeTenants
    }
    return activeTenants.filter((tenant) =>
      tenant.name.toLocaleLowerCase().includes(keyword)
    )
  }, [activeTenants, query])

  // 键盘移动高亮项时把它滚进 DialogBody 的可视区域，租户很多时也能用方向键选择。
  useEffect(() => {
    const target = filteredTenants[highlight]
    if (!target) {
      return
    }
    document
      .getElementById(`${prefix}-tenant-${target.uid}`)
      ?.scrollIntoView({ block: "nearest" })
  }, [filteredTenants, highlight, prefix])

  async function enterAdmin() {
    if (entering) {
      return
    }
    setEntering(true)
    try {
      const result = await listSelfNamespaces({ silent: true })
      const items = result.items ?? []
      setTenants(items)
      const active = items.filter((tenant) => tenant.status === "ACTIVE")
      if (active.length === 0) {
        setTenantDialog("none")
        return
      }
      if (active.length === 1) {
        selectTenant(active[0].uid)
        return
      }
      setQuery("")
      setHighlight(0)
      setTenantDialog("pick")
    } catch {
      toast.error(t("site.tenant.loadFailed"))
    } finally {
      setEntering(false)
    }
  }

  function selectTenant(uid: string) {
    // 与 tenant-switcher 使用同一个 localStorage 键：AppShell 的 SessionProvider
    // 启动时按它选租户，进入后台后不需要再切一次。
    localStorage.setItem("namespace", uid)
    setTenantDialog(null)
    navigate("/app")
  }

  function logout() {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    // 保留 namespace 与 lastPath：下次登录继续进入离开时的租户与页面。
    setToken("")
    setUser(null)
  }

  function onSearchKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (filteredTenants.length === 0) {
      return
    }
    if (event.key === "ArrowDown") {
      event.preventDefault()
      setHighlight((value) => (value + 1) % filteredTenants.length)
      return
    }
    if (event.key === "ArrowUp") {
      event.preventDefault()
      setHighlight(
        (value) =>
          (value - 1 + filteredTenants.length) % filteredTenants.length
      )
      return
    }
    if (event.key === "Enter") {
      event.preventDefault()
      const target = filteredTenants[highlight] ?? filteredTenants[0]
      selectTenant(target.uid)
    }
  }

  const site = config?.site
  const year = new Date().getFullYear()
  const copyright =
    site?.copyrightText.trim() || siteName || t("site.brandFallback")
  const filingIcpNumber = site?.filingIcpNumber.trim() ?? ""
  const filingIcpUrl = site?.filingIcpUrl.trim() || "https://beian.miit.gov.cn/"
  const filingPoliceNumber = site?.filingPoliceNumber.trim() ?? ""
  const filingPoliceUrl =
    site?.filingPoliceUrl.trim() || "https://beian.mps.gov.cn/"
  const filingExtraText = site?.filingExtraText.trim() ?? ""
  const contactEmail = site?.contactEmail.trim() ?? ""
  const repositoryUrl = site?.repositoryUrl.trim() ?? ""

  const accountMenu = (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className="h-9 max-w-48 gap-2 px-1.5"
          aria-label={`${t("site.account")}: ${displayName}`}
        >
          <span
            aria-hidden="true"
            className="relative grid size-7 shrink-0 place-items-center overflow-hidden rounded-full bg-primary/10 text-xs font-semibold text-primary-ink ring-1 ring-border/70"
          >
            {initials(displayName)}
            {user?.avatar ? (
              <img
                src={user.avatar}
                alt=""
                className="absolute inset-0 size-full object-cover"
              />
            ) : null}
          </span>
          <span className="hidden min-w-0 truncate text-sm sm:inline">
            {displayName}
          </span>
          <ChevronDown
            aria-hidden="true"
            className="hidden size-3.5 shrink-0 text-muted-foreground sm:block"
          />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align="end"
        sideOffset={8}
        className="min-w-56 max-w-[calc(100vw-1rem)]"
      >
        <DropdownMenuLabel className="px-2.5 py-2">
          <span className="block truncate text-sm font-semibold">
            {displayName}
          </span>
          {user?.email ? (
            <span className="block truncate text-xs font-normal text-muted-foreground">
              {user.email}
            </span>
          ) : null}
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className="min-h-10 gap-2.5 px-2.5"
          disabled={entering}
          onClick={() => void enterAdmin()}
        >
          {entering ? (
            <Loader2 className="animate-spin" aria-hidden="true" />
          ) : (
            <LayoutDashboard aria-hidden="true" />
          )}
          {t("site.enterAdmin")}
        </DropdownMenuItem>
        <DropdownMenuItem
          className="min-h-10 gap-2.5 px-2.5"
          onClick={() => navigate("/profile")}
        >
          <UserRound aria-hidden="true" />
          {t("site.profile")}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          variant="destructive"
          className="min-h-10 gap-2.5 px-2.5"
          onClick={logout}
        >
          <LogOut aria-hidden="true" />
          {t("site.logout")}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )

  return (
    <SiteAccountContext.Provider
      value={{
        loggedIn,
        user,
        entering,
        enterAdmin: () => void enterAdmin(),
        loginHref,
        registerHref,
      }}
    >
      <main className="flex min-h-svh flex-col bg-background">
        <header className="sticky top-0 z-30 border-b border-border/60 bg-background/85 backdrop-blur-sm">
          <div className="mx-auto flex w-full max-w-6xl min-w-0 items-center gap-3 px-4 py-3 sm:px-6">
            <Link
              to="/"
              className="flex min-w-0 items-center gap-2 rounded-md outline-none focus-visible:ring-3 focus-visible:ring-ring"
            >
              <img
                src={logoUrl}
                alt=""
                className="size-8 shrink-0 rounded-lg"
              />
              <span className="truncate font-heading text-base font-semibold tracking-tight">
                {siteName}
              </span>
            </Link>
            <div className="ml-auto flex min-w-0 items-center gap-1.5">
              <AppearanceMenu />
              <LocaleMenu />
              {loggedIn ? (
                accountMenu
              ) : (
                <>
                  <Button variant="ghost" size="sm" asChild>
                    <Link to={loginHref}>{t("site.login")}</Link>
                  </Button>
                  <Button size="sm" asChild>
                    <Link to={registerHref}>{t("site.register")}</Link>
                  </Button>
                </>
              )}
            </div>
          </div>
        </header>

        <div className="flex min-w-0 flex-1 flex-col">{children}</div>

        <footer className="border-t border-border/60 bg-muted/20">
          <div className="mx-auto flex w-full max-w-6xl flex-col gap-3 px-4 py-8 sm:px-6">
            <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-2 text-sm">
              <p className="min-w-0 break-words text-muted-foreground">
                {t("site.footer.copyright", { year, name: copyright })}
              </p>
              <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
                {contactEmail ? (
                  <a
                    href={`mailto:${contactEmail}`}
                    className="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
                  >
                    {t("site.footer.contact")}
                  </a>
                ) : null}
                {repositoryUrl ? (
                  <a
                    href={repositoryUrl}
                    target="_blank"
                    rel="noreferrer"
                    className="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline"
                  >
                    {t("site.footer.repository")}
                  </a>
                ) : null}
              </div>
            </div>
            {filingIcpNumber ? (
              <p className="text-xs text-muted-foreground">
                <a
                  href={filingIcpUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="underline-offset-4 hover:text-foreground hover:underline"
                >
                  {filingIcpNumber}
                </a>
              </p>
            ) : null}
            {filingPoliceNumber ? (
              <p className="text-xs text-muted-foreground">
                <a
                  href={filingPoliceUrl}
                  target="_blank"
                  rel="noreferrer"
                  className="underline-offset-4 hover:text-foreground hover:underline"
                >
                  {filingPoliceNumber}
                </a>
              </p>
            ) : null}
            {filingExtraText ? (
              <p className="text-xs break-words text-muted-foreground">
                {filingExtraText}
              </p>
            ) : null}
          </div>
        </footer>

        <Dialog
          open={tenantDialog !== null}
          onOpenChange={(open) => {
            if (!open) {
              setTenantDialog(null)
            }
          }}
        >
          <DialogContent className="sm:max-w-lg">
            <DialogHeader>
              <DialogTitle>
                {tenantDialog === "pick"
                  ? t("site.tenant.pickTitle")
                  : t("site.tenant.noneTitle")}
              </DialogTitle>
              <DialogDescription>
                {tenantDialog === "pick"
                  ? t("site.tenant.pickDescription")
                  : t("site.tenant.noneBody")}
              </DialogDescription>
            </DialogHeader>
            {tenantDialog === "pick" ? (
              <>
                <div className="shrink-0">
                  <Input
                    type="search"
                    value={query}
                    aria-label={t("site.tenant.search")}
                    aria-controls={`${prefix}-tenant-list`}
                    aria-activedescendant={
                      filteredTenants[highlight]
                        ? `${prefix}-tenant-${filteredTenants[highlight].uid}`
                        : undefined
                    }
                    placeholder={t("site.tenant.searchPlaceholder")}
                    autoComplete="off"
                    onChange={(event) => {
                      setQuery(event.target.value)
                      setHighlight(0)
                    }}
                    onKeyDown={onSearchKeyDown}
                  />
                </div>
                <DialogBody className="grid content-start gap-1">
                  {filteredTenants.length === 0 ? (
                    <p className="px-2.5 py-6 text-center text-sm text-muted-foreground">
                      {t("site.tenant.noResults")}
                    </p>
                  ) : (
                    <div
                      id={`${prefix}-tenant-list`}
                      role="listbox"
                      aria-label={t("site.tenant.listLabel")}
                      className="grid gap-0.5"
                    >
                      {filteredTenants.map((tenant, index) => (
                        <TenantOption
                          key={tenant.uid}
                          item={tenant}
                          id={`${prefix}-tenant-${tenant.uid}`}
                          active={index === highlight}
                          onHover={() => setHighlight(index)}
                          onSelect={() => selectTenant(tenant.uid)}
                        />
                      ))}
                    </div>
                  )}
                </DialogBody>
              </>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setTenantDialog(null)}
              >
                {t("site.tenant.cancel")}
              </Button>
              {tenantDialog === "none" ? (
                <Button asChild>
                  <Link to="/namespaces" onClick={() => setTenantDialog(null)}>
                    {t("site.tenant.create")}
                  </Link>
                </Button>
              ) : null}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </main>
    </SiteAccountContext.Provider>
  )
}
