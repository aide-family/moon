import {
  Bell,
  Building2,
  ChevronRight,
  Database,
  Gauge,
  Languages,
  Moon,
  Settings,
  Shield,
  Sun,
  SunMoon,
  User,
  UserRound,
  Users,
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
import { useTheme } from "@/components/theme-provider"
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
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
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
    to: "/alert-levels",
    label: "nav.alertLevels",
    permission: "alert_rule:read",
    icon: Gauge,
  },
]

function NavGroup({
  label,
  icon: Icon,
  open,
  active,
  links,
  onOpenChange,
}: {
  label: string
  icon: LucideIcon
  open: boolean
  active: boolean
  links: { to: string; label: string; icon: LucideIcon }[]
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  const location = useLocation()
  if (links.length === 0) {
    return null
  }
  return (
    <Collapsible
      asChild
      open={open}
      onOpenChange={onOpenChange}
      className="group/collapsible"
    >
      <SidebarMenuItem>
        <CollapsibleTrigger asChild>
          <SidebarMenuButton tooltip={t(label)} isActive={active}>
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
  const { t, i18n } = useTranslation()
  const { theme, setTheme } = useTheme()
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
  const systemActive = visible.some((link) => location.pathname === link.to)
  const alertActive = alerts.some((link) => location.pathname === link.to)
  const [systemOpen, setSystemOpen] = useState(systemActive)
  const [alertOpen, setAlertOpen] = useState(alertActive)
  useEffect(() => {
    if (systemActive) {
      setSystemOpen(true)
    }
  }, [systemActive])
  useEffect(() => {
    if (alertActive) {
      setAlertOpen(true)
    }
  }, [alertActive])
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

  function setLanguage(language: string) {
    localStorage.setItem("lang", language)
    void i18n.changeLanguage(language)
  }

  function onTheme(value: string) {
    if (value === "light" || value === "dark" || value === "system") {
      setTheme(value)
    }
  }

  const language = i18n.language.startsWith("zh") ? "zh" : "en"
  const ThemeIcon = theme === "dark" ? Moon : theme === "system" ? SunMoon : Sun

  return (
    <TooltipProvider>
      <SidebarProvider className="h-svh">
        <Sidebar collapsible="icon">
          <SidebarHeader>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton size="lg" asChild tooltip="Alioth">
                  <div>
                    <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-sidebar-primary text-sm font-semibold text-sidebar-primary-foreground">
                      A
                    </span>
                    <span className="font-semibold">Alioth</span>
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
                    label="nav.alert"
                    icon={Bell}
                    open={alertOpen}
                    active={alertActive}
                    links={alerts}
                    onOpenChange={setAlertOpen}
                  />
                  <NavGroup
                    label="nav.system"
                    icon={Settings}
                    open={systemOpen}
                    active={systemActive}
                    links={visible}
                    onOpenChange={setSystemOpen}
                  />
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarRail />
        </Sidebar>
        <SidebarInset className="min-h-0">
          <header className="flex h-14 items-center gap-2 border-b px-4">
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
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={t("lang.label")}
                  >
                    <Languages />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuRadioGroup
                    value={language}
                    onValueChange={setLanguage}
                  >
                    <DropdownMenuRadioItem value="zh">
                      {t("lang.zh")}
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="en">
                      {t("lang.en")}
                    </DropdownMenuRadioItem>
                  </DropdownMenuRadioGroup>
                </DropdownMenuContent>
              </DropdownMenu>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    aria-label={t("theme.label")}
                  >
                    <ThemeIcon />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuRadioGroup value={theme} onValueChange={onTheme}>
                    <DropdownMenuRadioItem value="light">
                      {t("theme.light")}
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="dark">
                      {t("theme.dark")}
                    </DropdownMenuRadioItem>
                    <DropdownMenuRadioItem value="system">
                      {t("theme.system")}
                    </DropdownMenuRadioItem>
                  </DropdownMenuRadioGroup>
                </DropdownMenuContent>
              </DropdownMenu>
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
          <div className="flex min-h-0 flex-1 flex-col gap-6 overflow-hidden p-4 md:p-6">
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
