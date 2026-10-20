import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react"

import { listSelfNamespaces, getSelfInfo, getSelfPermissions } from "@/api/self"
import { getUnreadMessageCount } from "@/api/message"
import { failureText } from "@/lib/form-error"
import type { NamespaceItem } from "@/api/namespace/types"
import type { UserItem } from "@/api/user/types"

type SessionValue = {
  user: UserItem | null
  namespaces: NamespaceItem[]
  permissions: string[]
  systemPermissions: string[]
  systemRole: string
  isSystemAdmin: boolean
  isSuperAdmin: boolean
  unreadMessages: number
  namespaceUid: string
  loading: boolean
  failed: boolean
  loadError: string
  reload: () => void
  setNamespace: (uid: string) => void
  can: (code: string) => boolean
  canSystem: (code: string) => boolean
  refreshUnreadMessages: () => void
}

const SessionContext = createContext<SessionValue | null>(null)

export function SessionProvider({ children }: { children: ReactNode }) {
  const [attempt, setAttempt] = useState(0)
  const [user, setUser] = useState<UserItem | null>(null)
  const [namespaces, setNamespaces] = useState<NamespaceItem[]>([])
  const [permissions, setPermissions] = useState<string[]>([])
  const [systemPermissions, setSystemPermissions] = useState<string[]>([])
  const [systemRole, setSystemRole] = useState("USER")
  const [unreadMessages, setUnreadMessages] = useState(0)
  const [namespaceUid, setNamespaceUid] = useState("")
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")

  const reload = useCallback(() => setAttempt((value) => value + 1), [])

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setLoadError("")
      try {
        const [info, spaces, unread] = await Promise.all([
          getSelfInfo(),
          listSelfNamespaces(),
          getUnreadMessageCount().then((result) => result.count ?? 0).catch(() => 0),
        ])
        if (!active) {
          return
        }
        const items = spaces.items ?? []
        setUser(info.user)
        setNamespaces(items)
        setSystemPermissions(info.systemPermissions ?? [])
        setSystemRole(info.systemRole ?? "USER")
        setUnreadMessages(unread)
        const stored = localStorage.getItem("namespace") ?? ""
        const next = items.some((item) => item.uid === stored) ? stored : (items[0]?.uid ?? "")
        if (next) {
          localStorage.setItem("namespace", next)
        } else {
          localStorage.removeItem("namespace")
        }
        setNamespaceUid(next)
        if (!next) {
          setPermissions([])
          return
        }
        const granted = await getSelfPermissions()
        if (!active) {
          return
        }
        setPermissions(granted.permissions ?? [])
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureText(err, ""))
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [attempt])

  const setNamespace = useCallback((uid: string) => {
    localStorage.setItem("namespace", uid)
    setNamespaceUid(uid)
    setAttempt((value) => value + 1)
  }, [])

  const refreshUnreadMessages = useCallback(() => {
    void getUnreadMessageCount()
      .then((result) => setUnreadMessages(result.count ?? 0))
      .catch(() => undefined)
  }, [])

  useEffect(() => {
    const timer = window.setInterval(refreshUnreadMessages, 30000)
    return () => window.clearInterval(timer)
  }, [refreshUnreadMessages])

  const value = useMemo<SessionValue>(
    () => ({
      user,
      namespaces,
      permissions,
      systemPermissions,
      systemRole,
      isSystemAdmin: systemRole === "SUPER_ADMIN" || systemRole === "SYSTEM_ADMIN",
      isSuperAdmin: systemRole === "SUPER_ADMIN",
      unreadMessages,
      namespaceUid,
      loading,
      failed,
      loadError,
      reload,
      setNamespace,
      can: (code: string) => permissions.includes(code),
      canSystem: (code: string) => systemPermissions.includes(code),
      refreshUnreadMessages,
    }),
    [
      user,
      namespaces,
      permissions,
      systemPermissions,
      systemRole,
      unreadMessages,
      namespaceUid,
      loading,
      failed,
      loadError,
      reload,
      setNamespace,
      refreshUnreadMessages,
    ],
  )

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
}

export function useSession() {
  const value = useContext(SessionContext)
  if (!value) {
    throw new Error("useSession must be used in SessionProvider")
  }
  return value
}
