import {
  Activity,
  Bell,
  BellRing,
  Building2,
  CalendarClock,
  ChevronRight,
  Database,
  FileText,
  Gauge,
  History,
  LayoutDashboard,
  LayoutList,
  LogOut,
  Mail,
  MessageSquare,
  Server,
  Settings,
  Shield,
  Terminal,
  UserRound,
  Users,
  UsersRound,
  type LucideIcon,
} from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import {
  NavLink,
  Navigate,
  Outlet,
  useLocation,
  useNavigate,
} from "react-router-dom"

import { SessionProvider, useSession } from "@/auth/session"
import { GitHubIcon } from "@/components/github-icon"
import { LocaleThemeMenu } from "@/components/locale-theme-menu"
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
import { ServerErrorPage } from "@/pages/server-error"

const repositoryURL = "https://github.com/aide-family/Alioth"

const links: {
  to: string
  label: string
  permission: string
  icon: LucideIcon
}[] = [
  {
    to: "/namespaces",
    label: "nav.namespaces",
    permission: "namespace:read",
    icon: Building2,
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
    icon: UserRound,
  },
  { to: "/roles", label: "nav.roles", permission: "role:read", icon: Shield },
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
  icon: LucideIcon
}[] = [
  {
    to: "/alerts",
    label: "nav.realtime",
    permission: "alert:read",
    icon: Activity,
  },
  {
    to: "/alert-history",
    label: "nav.history",
    permission: "alert:read",
    icon: History,
  },
  {
    to: "/datasources",
    label: "nav.datasources",
    permission: "datasource:read",
    icon: Database,
  },
  {
    to: "/alert-rules",
    label: "nav.alertRules",
    permission: "alert_rule:read",
    icon: Bell,
  },
  {
    to: "/alert-dictionaries",
    label: "nav.alertDictionaries",
    permission: "alert_rule:read",
    icon: Gauge,
  },
  {
    to: "/alert-pages",
    label: "nav.alertPages",
    permission: "alert:read",
    icon: LayoutList,
  },
]

const notifyLinks: {
  to: string
  label: string
  permission: string
  icon: LucideIcon
}[] = [
  {
    to: "/notification-channels",
    label: "nav.channels",
    permission: "notification:read",
    icon: Mail,
  },
  {
    to: "/notification-templates",
    label: "nav.templates",
    permission: "notification:read",
    icon: FileText,
  },
  {
    to: "/notification-groups",
    label: "nav.groups",
    permission: "notification:read",
    icon: UsersRound,
  },
  {
    to: "/notification-subscriptions",
    label: "nav.subscriptions",
    permission: "notification:read",
    icon: BellRing,
  },
  {
    to: "/notification-oncalls",
    label: "nav.onCalls",
    permission: "notification:read",
    icon: CalendarClock,
  },
  {
    to: "/message-logs",
    label: "nav.messageLogs",
    permission: "notification:read",
    icon: MessageSquare,
  },
]

const edgeLinks: {
  to: string
  label: string
  permission: string
  icon: LucideIcon
}[] = [
  {
    to: "/agents",
    label: "nav.agents",
    permission: "agent:read",
    icon: Server,
  },
  {
    to: "/ssh-commands",
    label: "nav.sshCommands",
    permission: "ssh:read",
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
                      (link.to === "/ssh-commands" &&
                        location.pathname.startsWith("/ssh-commands/"))
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
      className={`relative grid shrink-0 place-items-center overflow-hidden rounded-full bg-primary/10 font-semibold text-primary ring-1 ring-border/70 ${className}`}
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
  const accountName = session.user?.nickname || session.user?.name || ""
  const accountLabel = accountName || session.user?.email || t("shell.account")
  const avatarName = accountName || session.user?.email || "Alioth"
  const visible = links.filter((link) => {
    if (link.to === "/namespaces") {
      return true
    }
    if (link.permission.startsWith("system:")) {
      return session.canSystem(link.permission)
    }
    return session.namespaceUid !== "" && session.can(link.permission)
  })
  const alerts = alertLinks.filter(
    (link) => session.namespaceUid !== "" && session.can(link.permission)
  )
  const notifies = notifyLinks.filter(
    (link) => session.namespaceUid !== "" && session.can(link.permission)
  )
  const edges = edgeLinks.filter(
    (link) => session.namespaceUid !== "" && session.can(link.permission)
  )
  const systemActive = visible.some((link) => location.pathname === link.to)
  const alertActive = alerts.some((link) => location.pathname === link.to)
  const notifyActive = notifies.some((link) => location.pathname === link.to)
  const edgeActive = edges.some(
    (link) =>
      location.pathname === link.to ||
      (link.to === "/ssh-commands" &&
        location.pathname.startsWith("/ssh-commands/"))
  )
  // 概览是唯一不带领域前缀的一级入口：它是值班着陆页，不属于某个资源域。
  const overviewActive = location.pathname === "/overview"
  const canOverview = session.namespaceUid !== "" && session.can("alert:read")
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

  function logout() {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    localStorage.removeItem("namespace")
    navigate("/login")
  }

  return (
    <TooltipProvider>
      <SidebarProvider className="h-svh overflow-hidden">
        <Sidebar collapsible="icon" variant="inset">
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
              <TenantSwitcher />
              <Button
                variant="ghost"
                size="icon"
                className="relative shrink-0"
                aria-label={t("nav.messages")}
                onClick={() => navigate("/messages")}
              >
                <Bell />
                {session.unreadMessages > 0 ? (
                  <span className="absolute -top-0.5 -right-0.5 flex min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] leading-4 text-white">
                    {session.unreadMessages > 99
                      ? "99+"
                      : session.unreadMessages}
                  </span>
                ) : null}
              </Button>
              <LocaleThemeMenu />
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
                      <span className="ml-auto min-w-5 rounded-full bg-destructive px-1.5 text-center text-[10px] leading-4 font-medium text-white tabular-nums">
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
