import {
  Activity,
  Bell,
  BellRing,
  Boxes,
  Building2,
  CalendarClock,
  ChevronRight,
  Database,
  FileText,
  Gauge,
  History,
  Landmark,
  LayoutDashboard,
  LayoutList,
  LogOut,
  Mail,
  MessageSquare,
  Package,
  ReceiptText,
  Server,
  Settings,
  Shield,
  ShieldAlert,
  Target,
  Terminal,
  UserRound,
  Users,
  UsersRound,
  type LucideIcon,
} from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import {
  NavLink,
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom"

import { SessionProvider, useSession } from "@/auth/session"
import { AppearanceMenu } from "@/components/appearance-menu"
import { GitHubIcon } from "@/components/github-icon"
import { LocaleMenu } from "@/components/locale-menu"
import { StatusBadge } from "@/components/status-badge"
import { TenantSwitcher } from "@/components/tenant-switcher"
import { Button } from "@/components/ui/button"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
  SidebarRail,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { TooltipProvider } from "@/components/ui/tooltip"
import { MENU_CAPABILITIES } from "@/lib/feature-catalog"
import {
  isKnownAppPath,
  normalizeAppPath,
  rememberLastPath,
} from "@/lib/last-page"
import { ServerErrorPage } from "@/pages/server-error"

const repositoryURL = "https://github.com/aide-family/moon"

// 这些页面不依赖当前租户，没有租户也可以直接打开；其余登录内页面会被送回租户页。
const tenantFreePaths = new Set([
  "/namespaces",
  "/orders",
  "/tenants",
  "/packages",
  "/users",
  "/workers",
  "/invitations",
  "/messages",
  "/profile",
  "/forbidden",
])

const links: {
  to: string
  label: string
  permission: string
  /** 菜单能力开关：配置后还要求当前租户已开通该能力（system 菜单不配置）。 */
  capability?: string
  icon: LucideIcon
}[] = [
  {
    to: "/namespaces",
    label: "nav.namespaces",
    permission: "namespace:read",
    icon: Building2,
  },
  {
    to: "/orders",
    label: "nav.orders",
    // 订单只属于当前登录用户，刚注册、还没有任何租户权限的用户也要能查看。
    permission: "",
    icon: ReceiptText,
  },
  {
    to: "/tenants",
    label: "nav.tenants",
    permission: "system:namespace:read",
    icon: Landmark,
  },
  {
    to: "/packages",
    label: "nav.packages",
    permission: "system:package:read",
    icon: Package,
  },
  {
    to: "/users",
    label: "nav.users",
    permission: "system:user:read",
    icon: Users,
  },
  {
    to: "/members",
    label: "nav.members",
    permission: "member:read",
    capability: MENU_CAPABILITIES["/members"],
    icon: UserRound,
  },
  {
    to: "/roles",
    label: "nav.roles",
    permission: "role:read",
    capability: MENU_CAPABILITIES["/roles"],
    icon: Shield,
  },
  {
    to: "/workers",
    label: "nav.workers",
    permission: "system:user:read",
    icon: Activity,
  },
]

const alertLinks: {
  to: string
  label: string
  permission: string
  capability?: string
  icon: LucideIcon
}[] = [
  {
    to: "/alerts",
    label: "nav.realtime",
    permission: "alert_realtime:read",
    capability: MENU_CAPABILITIES["/alerts"],
    icon: Activity,
  },
  {
    to: "/alert-history",
    label: "nav.history",
    permission: "alert_history:read",
    capability: MENU_CAPABILITIES["/alert-history"],
    icon: History,
  },
  {
    to: "/datasources",
    label: "nav.datasources",
    permission: "datasource:read",
    capability: MENU_CAPABILITIES["/datasources"],
    icon: Database,
  },
  {
    to: "/probe-targets",
    label: "nav.probeTargets",
    permission: "probe:read",
    capability: MENU_CAPABILITIES["/probe-targets"],
    icon: Target,
  },
  {
    to: "/alert-rules",
    label: "nav.alertRules",
    permission: "alert_rule:read",
    capability: MENU_CAPABILITIES["/alert-rules"],
    icon: Bell,
  },
  {
    to: "/alert-dictionaries",
    label: "nav.alertDictionaries",
    permission: "alert_dictionary:read",
    capability: MENU_CAPABILITIES["/alert-dictionaries"],
    icon: Gauge,
  },
  {
    to: "/alert-pages",
    label: "nav.alertPages",
    permission: "alert_page:read",
    capability: MENU_CAPABILITIES["/alert-pages"],
    icon: LayoutList,
  },
]

const notifyLinks: {
  to: string
  label: string
  permission: string
  capability?: string
  icon: LucideIcon
}[] = [
  {
    to: "/notification-channels",
    label: "nav.channels",
    permission: "notification_channel:read",
    capability: MENU_CAPABILITIES["/notification-channels"],
    icon: Mail,
  },
  {
    to: "/notification-templates",
    label: "nav.templates",
    permission: "notification_template:read",
    capability: MENU_CAPABILITIES["/notification-templates"],
    icon: FileText,
  },
  {
    to: "/notification-groups",
    label: "nav.groups",
    permission: "notification_group:read",
    capability: MENU_CAPABILITIES["/notification-groups"],
    icon: UsersRound,
  },
  {
    to: "/notification-subscriptions",
    label: "nav.subscriptions",
    permission: "notification_subscription:read",
    capability: MENU_CAPABILITIES["/notification-subscriptions"],
    icon: BellRing,
  },
  {
    to: "/notification-oncalls",
    label: "nav.onCalls",
    permission: "notification_oncall:read",
    capability: MENU_CAPABILITIES["/notification-oncalls"],
    icon: CalendarClock,
  },
  {
    to: "/message-logs",
    label: "nav.messageLogs",
    permission: "message_log:read",
    capability: MENU_CAPABILITIES["/message-logs"],
    icon: MessageSquare,
  },
]

const kubeLinks: {
  to: string
  label: string
  permission: string
  capability?: string
  icon: LucideIcon
}[] = [
  {
    to: "/clusters",
    label: "nav.clusters",
    permission: "cluster:read",
    capability: MENU_CAPABILITIES["/clusters"],
    icon: Boxes,
  },
]

const edgeLinks: {
  to: string
  label: string
  permission: string
  capability?: string
  icon: LucideIcon
}[] = [
  {
    to: "/agents",
    label: "nav.agents",
    permission: "agent:read",
    capability: MENU_CAPABILITIES["/agents"],
    icon: Server,
  },
  {
    to: "/ssh-commands",
    label: "nav.sshCommands",
    permission: "ssh:read",
    capability: MENU_CAPABILITIES["/ssh-commands"],
    icon: Terminal,
  },
]

function NavGroup({
  label,
  icon: Icon,
  active,
  links,
}: {
  label: string
  icon: LucideIcon
  active: boolean
  links: { to: string; label: string; icon: LucideIcon }[]
}) {
  const { t } = useTranslation()
  const location = useLocation()
  const [open, setOpen] = useState(active)
  if (links.length === 0) {
    return null
  }
  return (
    <Collapsible
      asChild
      open={open}
      onOpenChange={setOpen}
      className="group/collapsible"
    >
      <SidebarMenuItem>
        <CollapsibleTrigger asChild>
          <SidebarMenuButton
            tooltip={t(label)}
            isActive={active}
            className="border-l-2 border-transparent data-active:border-primary"
          >
            <Icon />
            <span>{t(label)}</span>
            <ChevronRight className="ml-auto transition-transform group-data-[collapsible=icon]:hidden group-data-[state=open]/collapsible:rotate-90" />
          </SidebarMenuButton>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <SidebarMenuSub>
            {links.map((link) => {
              const LinkIcon = link.icon
              return (
                <SidebarMenuSubItem key={link.to}>
                  <SidebarMenuSubButton
                    asChild
                    isActive={
                      location.pathname === link.to ||
                      location.pathname.startsWith(`${link.to}/`)
                    }
                    className="border-l-2 border-transparent data-active:border-primary"
                  >
                    <NavLink to={link.to}>
                      <LinkIcon />
                      <span>{t(link.label)}</span>
                    </NavLink>
                  </SidebarMenuSubButton>
                </SidebarMenuSubItem>
              )
            })}
          </SidebarMenuSub>
        </CollapsibleContent>
      </SidebarMenuItem>
    </Collapsible>
  )
}

function AccountAvatar({
  avatar,
  name,
  className,
}: {
  avatar?: string
  name: string
  className: string
}) {
  const [imageFailed, setImageFailed] = useState(false)
  const initial = name.trim().charAt(0).toLocaleUpperCase()

  return (
    <span
      aria-hidden="true"
      className={`relative grid shrink-0 place-items-center overflow-hidden rounded-full bg-primary/10 font-semibold text-primary-ink ring-1 ring-border/70 ${className}`}
    >
      {initial || "A"}
      {avatar && !imageFailed ? (
        <img
          src={avatar}
          alt=""
          className="absolute inset-0 size-full rounded-full object-cover"
          onError={() => setImageFailed(true)}
        />
      ) : null}
    </span>
  )
}

function ShellFrame() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const location = useLocation()
  const userUid = session.user?.uid ?? ""
  // 每次实际落在某个页面就记一下，供下次登录的 "/" 着陆页回放。
  // "/" 是着陆页本身、/forbidden 是权限结果页，二者由 last-page 过滤，不覆盖有效记录。
  useEffect(() => {
    rememberLastPath(userUid, location.pathname)
  }, [userUid, location.pathname])
  const accountName = session.user?.nickname || session.user?.name || ""
  const accountLabel = accountName || session.user?.email || t("shell.account")
  const avatarName = accountName || session.user?.email || "Alioth"
  // 租户菜单三重判定：菜单权限（menu:<能力>）、租户已开通能力、资源权限。
  // 旧后端没有 menu:* 权限或 features 字段时，canMenu / hasFeature 都不拦截；
  // 系统菜单不受能力限制，只看系统权限。
  const tenantVisible = (link: { permission: string; capability?: string }) => {
    if (session.namespaceUid === "" || !session.can(link.permission)) {
      return false
    }
    if (!link.capability) {
      return true
    }
    // 入口可能是能力码，也可能是模块码（数据源 / 探测 / 通知渠道按类型拆分后），
    // 由 session.canFeatureEntry 统一判断「任一类型 + 菜单授权」。
    return session.canFeatureEntry(link.capability)
  }
  const visible = links.filter((link) => {
    // 租户页和订单页不依赖任何租户权限：还没有加入租户的用户也要能看到入口。
    if (link.to === "/namespaces" || link.to === "/orders") {
      return true
    }
    if (link.permission.startsWith("system:")) {
      return session.canSystem(link.permission)
    }
    return tenantVisible(link)
  })
  const alerts = alertLinks.filter(tenantVisible)
  const notifies = notifyLinks.filter(tenantVisible)
  const kubes = kubeLinks.filter(tenantVisible)
  const edges = edgeLinks.filter(tenantVisible)
  const systemActive = visible.some(
    (link) =>
      location.pathname === link.to ||
      location.pathname.startsWith(`${link.to}/`)
  )
  const alertActive = alerts.some((link) => location.pathname === link.to)
  const notifyActive = notifies.some((link) => location.pathname === link.to)
  const kubeActive = kubes.some(
    (link) =>
      location.pathname === link.to ||
      location.pathname.startsWith(`${link.to}/`)
  )
  const edgeActive = edges.some(
    (link) =>
      location.pathname === link.to ||
      location.pathname.startsWith(`${link.to}/`)
  )
  // 概览是唯一不带领域前缀的一级入口：它是值班着陆页，跟随实时告警能力。
  const overviewActive = location.pathname === "/overview"
  const canOverview = tenantVisible({
    permission: "alert_realtime:read",
    capability: MENU_CAPABILITIES["/overview"],
  })
  const token = localStorage.getItem("token")
  if (!token) {
    return <Navigate to="/login" replace />
  }
  // 后端接口失败时直接给出标准错误页（502/503），而不是在内容区残留“加载失败 + 重试”。
  if (session.failed) {
    return (
      <ServerErrorPage
        status={session.loadStatus === 503 ? 503 : 502}
        detail={session.loadError}
        onRetry={session.reload}
      />
    )
  }
  // 会话加载完仍没有租户时，租户相关页面统一回到租户页（选择/创建）。
  // 租户无关页面（用户、边缘节点、消息、个人中心等）和 "/" 着陆页不受此守卫影响，
  // 因此不会与 LandingRedirect 互相跳转。路径先归一化，尾斜杠地址不会误判。
  const appPath = normalizeAppPath(location.pathname)
  // 租户订单是 /tenants 的子路由，和租户管理页一样不依赖当前租户。
  const tenantFree =
    tenantFreePaths.has(appPath) || appPath.startsWith("/tenants/")
  if (
    !session.loading &&
    !session.switching &&
    session.namespaceUid === "" &&
    isKnownAppPath(appPath) &&
    !tenantFree
  ) {
    return <Navigate to="/namespaces" replace />
  }

  function logout() {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    // 不清理 namespace：下次登录继续进入离开时的租户（不存在时会回落到第一个租户）。
    // 也不清理 lastPath：下次登录回到离开时的页面。
    navigate("/login")
  }

  // 退出只读回落到第一个成员租户；租户管理页不属于租户相关页面，不会被守卫拦截。
  function exitReadOnlyView() {
    session.exitReadOnly()
    navigate("/tenants")
  }

  return (
    <TooltipProvider>
      <SidebarProvider className="h-svh overflow-hidden">
        <Sidebar
          collapsible="icon"
          variant="sidebar"
          className="md:border-r md:border-border/60"
        >
          <SidebarHeader>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton size="lg" asChild tooltip="Alioth">
                  <div>
                    <img
                      src="/alioth.svg"
                      alt=""
                      className="size-8 shrink-0 rounded-[9px]"
                    />
                    <span className="font-heading font-semibold">Alioth</span>
                  </div>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  {canOverview ? (
                    <SidebarMenuItem>
                      <SidebarMenuButton
                        asChild
                        isActive={overviewActive}
                        tooltip={t("nav.overview")}
                        className="border-l-2 border-transparent data-active:border-primary"
                      >
                        <NavLink to="/overview">
                          <LayoutDashboard />
                          <span>{t("nav.overview")}</span>
                        </NavLink>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ) : null}
                  <NavGroup
                    key={`alert-${alertActive}`}
                    label="nav.alert"
                    icon={Bell}
                    active={alertActive}
                    links={alerts}
                  />
                  <NavGroup
                    key={`notify-${notifyActive}`}
                    label="nav.notify"
                    icon={Mail}
                    active={notifyActive}
                    links={notifies}
                  />
                  <NavGroup
                    key={`kube-${kubeActive}`}
                    label="nav.kubernetes"
                    icon={Boxes}
                    active={kubeActive}
                    links={kubes}
                  />
                  <NavGroup
                    key={`edge-${edgeActive}`}
                    label="nav.edge"
                    icon={Server}
                    active={edgeActive}
                    links={edges}
                  />
                  <NavGroup
                    key={`system-${systemActive}`}
                    label="nav.system"
                    icon={Settings}
                    active={systemActive}
                    links={visible}
                  />
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarRail />
        </Sidebar>
        <SidebarInset className="relative min-h-0 min-w-0 bg-transparent">
          {session.switching ? (
            <div
              className="pointer-events-none absolute inset-x-0 top-0 z-30 h-0.5 overflow-hidden bg-primary/15"
              role="progressbar"
              aria-label={t("namespace.switcher.switchingGeneric")}
            >
              <div className="h-full w-full animate-pulse bg-primary" />
            </div>
          ) : null}
          <header className="flex h-12 items-center gap-2 border-b border-border/60 bg-background/80 px-4 backdrop-blur-sm">
            <SidebarTrigger />
            <div className="ml-auto flex h-10 min-w-0 items-center gap-2 overflow-x-auto overflow-y-hidden">
              {/* 只读视角下当前租户不在成员列表里，切换器会显示「无租户」，
                  上下文由内容区顶部的只读横幅承担，这里直接隐藏。 */}
              {session.readOnly ? null : <TenantSwitcher />}
              <Button
                variant="ghost"
                size="icon"
                className="relative shrink-0"
                aria-label={t("nav.messages")}
                onClick={() => navigate("/messages")}
              >
                <Bell />
                {session.unreadMessages > 0 ? (
                  <span className="absolute -top-0.5 -right-0.5 flex min-w-4 items-center justify-center rounded-full bg-danger-strong px-1 text-[10px] leading-4 text-danger-strong-foreground">
                    {session.unreadMessages > 99
                      ? "99+"
                      : session.unreadMessages}
                  </span>
                ) : null}
              </Button>
              <AppearanceMenu />
              <LocaleMenu />
              <Button variant="ghost" size="icon" className="shrink-0" asChild>
                <a
                  href={repositoryURL}
                  target="_blank"
                  rel="noreferrer"
                  aria-label={t("shell.github")}
                >
                  <GitHubIcon />
                </a>
              </Button>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="size-10 shrink-0 rounded-full p-1 transition-colors hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset data-[state=open]:bg-accent"
                    aria-label={
                      accountName || session.user?.email
                        ? `${t("shell.account")}: ${accountLabel}`
                        : t("shell.account")
                    }
                  >
                    <AccountAvatar
                      avatar={session.user?.avatar}
                      name={avatarName}
                      className="size-8 text-xs"
                    />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent
                  align="end"
                  sideOffset={8}
                  className="max-w-[calc(100vw-1rem)] min-w-72 p-1.5"
                >
                  <DropdownMenuLabel className="px-2.5 py-2.5">
                    <div className="flex min-w-0 items-center gap-3">
                      <AccountAvatar
                        avatar={session.user?.avatar}
                        name={avatarName}
                        className="size-10 text-sm"
                      />
                      <div className="min-w-0 flex-1">
                        <p className="truncate text-sm font-semibold text-foreground">
                          {session.user?.nickname ||
                            session.user?.name ||
                            t("shell.account")}
                        </p>
                        <p className="mt-0.5 truncate text-xs font-normal text-muted-foreground">
                          {session.user?.email || "—"}
                        </p>
                        <StatusBadge
                          value={
                            session.systemRole ||
                            session.user?.systemRole ||
                            "USER"
                          }
                          className="mt-1.5 px-1.5 py-0"
                        />
                      </div>
                    </div>
                  </DropdownMenuLabel>
                  <DropdownMenuSeparator className="my-1.5" />
                  <DropdownMenuItem
                    onClick={() => navigate("/profile")}
                    className="min-h-11 gap-2.5 px-2.5"
                  >
                    <UserRound className="text-muted-foreground" />
                    {t("nav.profile")}
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate("/invitations")}
                    className="min-h-11 gap-2.5 px-2.5"
                  >
                    <Mail className="text-muted-foreground" />
                    {t("nav.invitations")}
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    onClick={() => navigate("/messages")}
                    className="min-h-11 gap-2.5 px-2.5"
                  >
                    <MessageSquare className="text-muted-foreground" />
                    {t("nav.messages")}
                    {session.unreadMessages > 0 ? (
                      <span className="ml-auto min-w-5 rounded-full bg-danger-strong px-1.5 text-center text-[10px] leading-4 font-medium text-danger-strong-foreground tabular-nums">
                        {session.unreadMessages > 99
                          ? "99+"
                          : session.unreadMessages}
                      </span>
                    ) : null}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator className="my-1.5" />
                  <DropdownMenuItem
                    onClick={logout}
                    variant="destructive"
                    className="min-h-11 gap-2.5 px-2.5"
                  >
                    <LogOut />
                    {t("action.logout")}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </header>
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-6 overflow-hidden bg-transparent p-4 md:p-6">
            {/* 系统管理员只读横幅：写操作已按 permissions 隐藏，服务端 403 兜底。 */}
            {session.readOnly ? (
              <div
                role="status"
                className="flex shrink-0 items-start gap-3 rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-sm"
              >
                <ShieldAlert
                  aria-hidden="true"
                  className="mt-0.5 size-4 shrink-0 text-warning"
                />
                <div className="min-w-0 flex-1">
                  <p className="font-medium break-words text-warning">
                    {t("readonly.banner", {
                      name:
                        session.viewingNamespace?.name || t("namespace.none"),
                    })}
                  </p>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {t("readonly.hint")}
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="shrink-0"
                  onClick={exitReadOnlyView}
                >
                  {t("readonly.exit")}
                </Button>
              </div>
            ) : null}
            {session.loading && !session.switching ? (
              <p className="text-sm text-muted-foreground">
                {t("state.loading")}
              </p>
            ) : null}
            {!session.loading || session.switching ? <Outlet /> : null}
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}

export function AppShell() {
  return (
    <SessionProvider>
      <ShellFrame />
    </SessionProvider>
  )
}
