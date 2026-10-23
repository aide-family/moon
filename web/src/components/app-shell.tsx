import {
  Activity,
  Bell,
  BellRing,
  Building2,
  ChevronRight,
  Database,
  FileText,
  Gauge,
  History,
  LayoutList,
  Mail,
  MessageSquare,
  Server,
  Settings,
  Shield,
  Terminal,
  User,
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
  { to: "/users", label: "nav.users", permission: "system:user:read", icon: Users },
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
                      (link.to === "/ssh-commands" && location.pathname.startsWith("/ssh-commands/"))
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

function ShellFrame() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const location = useLocation()
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
    (link) => session.namespaceUid !== "" && session.can(link.permission),
  )
  const notifies = notifyLinks.filter(
    (link) => session.namespaceUid !== "" && session.can(link.permission),
  )
  const edges = edgeLinks.filter(
    (link) => session.namespaceUid !== "" && session.can(link.permission),
  )
  const systemActive = visible.some((link) => location.pathname === link.to)
  const alertActive = alerts.some((link) => location.pathname === link.to)
  const notifyActive = notifies.some((link) => location.pathname === link.to)
  const edgeActive = edges.some(
    (link) =>
      location.pathname === link.to ||
      (link.to === "/ssh-commands" && location.pathname.startsWith("/ssh-commands/")),
  )
  const token = localStorage.getItem("token")
  if (!token) {
    return <Navigate to="/login" replace />
  }
  // 后端接口失败时直接给出标准 502 页面，而不是在内容区残留“加载失败 + 重试”。
  if (session.failed) {
    return <ServerErrorPage detail={session.loadError} onRetry={session.reload} />
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
            <div className="ml-auto flex h-8 min-w-0 items-center gap-2 overflow-x-auto overflow-y-hidden">
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
                    {session.unreadMessages > 99 ? "99+" : session.unreadMessages}
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
                    aria-label={t("shell.account")}
                  >
                    <User />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {session.user?.email ? (
                    <DropdownMenuLabel>{session.user.email}</DropdownMenuLabel>
                  ) : null}
                  {session.user?.email ? <DropdownMenuSeparator /> : null}
                  <DropdownMenuItem onClick={() => navigate("/profile")}>
                    {t("nav.profile")}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => navigate("/invitations")}>
                    {t("nav.invitations")}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => navigate("/messages")}>
                    {t("nav.messages")}
                    {session.unreadMessages > 0 ? (
                      <span className="ml-auto rounded-full bg-destructive px-1.5 text-[10px] leading-4 text-white">
                        {session.unreadMessages > 99 ? "99+" : session.unreadMessages}
                      </span>
                    ) : null}
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={logout}>
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
