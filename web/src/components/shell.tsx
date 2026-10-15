import { useTranslation } from "react-i18next"
import { NavLink, Navigate, Outlet, useNavigate } from "react-router-dom"

import { SessionProvider, useSession } from "@/auth/session"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useTheme } from "@/components/theme-provider"

const links = [
  { to: "/profile", label: "nav.profile", permission: "" },
  { to: "/namespaces", label: "nav.namespaces", permission: "namespace:read" },
  { to: "/users", label: "nav.users", permission: "user:read" },
  { to: "/members", label: "nav.members", permission: "member:read" },
  { to: "/roles", label: "nav.roles", permission: "role:read" },
]

function ShellFrame() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme } = useTheme()
  const session = useSession()
  const navigate = useNavigate()
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

  const visible = links.filter((link) => {
    if (!link.permission) {
      return true
    }
    if (link.to === "/namespaces") {
      return true
    }
    return session.namespaceUid !== "" && session.can(link.permission)
  })

  return (
    <div className="flex min-h-screen">
      <aside className="w-56 border-r p-4">
        <p className="mb-4 text-sm font-semibold">Alioth</p>
        <nav className="flex flex-col gap-1">
          {visible.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              className={({ isActive }) => `rounded-md px-3 py-2 text-sm ${isActive ? "bg-muted" : "hover:bg-muted"}`}
            >
              {t(link.label)}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-end gap-2 border-b px-4 py-3">
          {session.namespaces.length > 0 ? (
            <Select value={session.namespaceUid} onValueChange={session.setNamespace}>
              <SelectTrigger className="w-56" aria-label={t("namespace.switch")}>
                <SelectValue placeholder={t("namespace.switch")} />
              </SelectTrigger>
              <SelectContent>
                {session.namespaces.map((item) => (
                  <SelectItem key={item.uid} value={item.uid}>{item.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : null}
          <Button variant="outline" size="sm" onClick={() => setLanguage(i18n.language === "zh" ? "en" : "zh")}>
            {i18n.language === "zh" ? t("lang.en") : t("lang.zh")}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setTheme(theme === "light" ? "dark" : theme === "dark" ? "system" : "light")}
          >
            {t(`theme.${theme}`)}
          </Button>
          <Button variant="ghost" size="sm" onClick={logout}>{t("action.logout")}</Button>
        </header>
        <main className="p-6">
          {session.loading ? <p>{t("state.loading")}</p> : <Outlet />}
        </main>
      </div>
    </div>
  )
}

export function Shell() {
  return (
    <SessionProvider>
      <ShellFrame />
    </SessionProvider>
  )
}
