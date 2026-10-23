import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react"

import { ApiError } from "@/api/http"
import { listSelfNamespaces, getSelfInfo, getSelfPermissions } from "@/api/self"
import { getUnreadMessageCount } from "@/api/message"
import i18n from "@/i18n"
import { failureMessage } from "@/lib/form-error"
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
  switching: boolean
  switchingTo: string
  switchError: string
  failed: boolean
  loadError: string
  reload: () => void
  setNamespace: (uid: string) => void
  clearSwitchError: () => void
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
  const [switching, setSwitching] = useState(false)
  const [switchingTo, setSwitchingTo] = useState("")
  const [switchError, setSwitchError] = useState("")
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const namespaceUidRef = useRef("")
  const switchRef = useRef<{ from: string; to: string } | null>(null)

  const reload = useCallback(() => setAttempt((value) => value + 1), [])

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setLoadError("")
      const transition = switchRef.current
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
        const stored = localStorage.getItem("namespace") ?? ""
        const requested = transition?.to ?? stored
        const next = items.some((item) => item.uid === requested)
          ? requested
          : (items[0]?.uid ?? "")
        // 切换目标可能已被删除或成员身份被撤销，这时不要静默落到别的租户。
        if (transition && next !== transition.to) {
          throw new Error(i18n.t("namespace.switcher.missing"))
        }
        // 权限按目标租户读取，但不提前改 localStorage / namespaceUid：
        // 拉起中的页面轮询仍然使用旧租户，避免新旧权限交叉。
        let granted: string[] = []
        if (next) {
          const result = await getSelfPermissions(next)
          if (!active) {
            return
          }
          granted = result.permissions ?? []
        }
        if (!active) {
          return
        }
        setUser(info.user)
        setNamespaces(items)
        setSystemPermissions(info.systemPermissions ?? [])
        setSystemRole(info.systemRole ?? "USER")
        setUnreadMessages(unread)
        setPermissions(granted)
        namespaceUidRef.current = next
        setNamespaceUid(next)
        if (next) {
          localStorage.setItem("namespace", next)
        } else {
          localStorage.removeItem("namespace")
        }
        if (transition) {
          switchRef.current = null
          setSwitching(false)
          setSwitchingTo("")
        }
      } catch (err) {
        if (!active) {
          return
        }
        if (transition) {
          switchRef.current = null
          setSwitching(false)
          setSwitchingTo("")
          setSwitchError(
            err instanceof ApiError ? err.message : i18n.t("state.failed")
          )
          return
        }
        setFailed(true)
        setLoadError(failureMessage(err, ""))
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
    if (!uid || uid === namespaceUidRef.current || switchRef.current) {
      return
    }
    switchRef.current = { from: namespaceUidRef.current, to: uid }
    setSwitchError("")
    setSwitching(true)
    setSwitchingTo(uid)
    setAttempt((value) => value + 1)
  }, [])

  const clearSwitchError = useCallback(() => setSwitchError(""), [])

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
      switching,
      switchingTo,
      switchError,
      failed,
      loadError,
      reload,
      setNamespace,
      clearSwitchError,
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
      switching,
      switchingTo,
      switchError,
      failed,
      loadError,
      reload,
      setNamespace,
      clearSwitchError,
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
