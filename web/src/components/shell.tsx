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
import { LocaleThemeMenu } from "@/components/locale-theme-menu"
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
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
  { to: "/users", label: "nav.users", permission: "user:read", icon: Users },
  {
    to: "/members",
    label: "nav.members",
    permission: "member:read",
    icon: UserRound,
  },
  { to: "/roles", label: "nav.roles", permission: "role:read", icon: Shield },
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
  {
    to: "/sftp",
    label: "nav.sftp",
    permission: "sftp:read",
    icon: FileText,
  },
  {
    to: "/k8s",
    label: "nav.k8s",
    permission: "k8s:read",
    icon: Gauge,
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
                    isActive={location.pathname === link.to}
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
  const edgeActive = edges.some((link) => location.pathname === link.to)
  const token = localStorage.getItem("token")
  if (!token) {
    return <Navigate to="/login" replace />
  }

  function logout() {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    localStorage.removeItem("namespace")
    navigate("/login")
  }

  return (
    <TooltipProvider>
      <SidebarProvider className="h-svh">
        <Sidebar collapsible="icon" variant="inset">
          <SidebarHeader>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton size="lg" asChild tooltip="Alioth">
                  <div>
                    <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-sidebar-primary text-sm font-semibold text-sidebar-primary-foreground">
                      A
                    </span>
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
        <SidebarInset className="min-h-0 bg-transparent">
          <header className="flex h-12 items-center gap-2 border-b border-border/60 bg-background/80 px-4 backdrop-blur-sm">
            <SidebarTrigger />
            <div className="ml-auto flex min-w-0 items-center gap-2 overflow-x-auto">
              {session.namespaces.length > 0 ? (
                <Select
                  value={session.namespaceUid}
                  onValueChange={session.setNamespace}
                >
                  <SelectTrigger
                    className="w-28 shrink-0 md:w-56"
                    aria-label={t("namespace.switch")}
                  >
                    <SelectValue placeholder={t("namespace.switch")} />
                  </SelectTrigger>
                  <SelectContent>
                    {session.namespaces.map((item) => (
                      <SelectItem key={item.uid} value={item.uid}>
                        {item.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : null}
              <LocaleThemeMenu />
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
                  <DropdownMenuItem onClick={logout}>
                    {t("action.logout")}
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </header>
          <div className="flex min-h-0 flex-1 flex-col gap-6 overflow-hidden bg-transparent p-4 md:p-6">
            {session.loading ? (
              <p className="text-sm text-muted-foreground">
                {t("state.loading")}
              </p>
            ) : null}
            {!session.loading && session.failed ? (
              <div className="flex items-center gap-2">
                <p className="text-sm text-destructive">{t("state.failed")}</p>
                <Button variant="outline" onClick={session.reload}>
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {!session.loading && !session.failed ? <Outlet /> : null}
          </div>
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  )
}

export function Shell() {
  return (
    <SessionProvider>
      <ShellFrame />
    </SessionProvider>
  )
}
