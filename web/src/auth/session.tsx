import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react"

import { ApiError } from "@/api/http"
import { listSelfNamespaces, getSelfInfo, getSelfPermissions } from "@/api/self"
import { getUnreadMessageCount } from "@/api/message"
import i18n from "@/i18n"
import { failureMessage } from "@/lib/form-error"
import type { SelfPermissions } from "@/api/self/types"
import type { NamespaceItem } from "@/api/namespace/types"
import type { UserItem } from "@/api/user/types"

type SwitchKind = "switch" | "readonly"

type SwitchRequest = {
  from: string
  to: string
  kind: SwitchKind
  // 切换是 effect 驱动的：调用方拿着这个 Promise 决定失败后是否导航。
  resolve: (ok: boolean) => void
}

// 系统管理员只读视角：不是目标租户的成员，但平台角色允许跨租户只读。
function isSystemAdmin(info: {
  systemRole?: string
  systemPermissions?: string[]
}): boolean {
  const role = info.systemRole ?? "USER"
  return (
    (role === "SUPER_ADMIN" || role === "SYSTEM_ADMIN") &&
    (info.systemPermissions ?? []).includes("system:namespace:read")
  )
}

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
  /** 当前是否处于系统管理员只读视角；写按钮随 permissions 自动隐藏。 */
  readOnly: boolean
  /** 只读视角正在查看的租户；普通成员视角为 null。 */
  viewingNamespace: NamespaceItem | null
  loading: boolean
  switching: boolean
  switchingTo: string
  switchError: string
  failed: boolean
  loadError: string
  loadStatus: number
  reload: () => void
  setNamespace: (uid: string) => void
  /** 系统管理员以只读身份切入任意租户；resolve(false) 表示没有只读权限或租户不存在。 */
  enterNamespace: (uid: string) => Promise<boolean>
  /** 退出只读视角，回落到第一个成员租户（没有则空），并清掉持久化的租户。 */
  exitReadOnly: () => void
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
  const [readOnly, setReadOnly] = useState(false)
  const [viewingNamespace, setViewingNamespace] =
    useState<NamespaceItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [switching, setSwitching] = useState(false)
  const [switchingTo, setSwitchingTo] = useState("")
  const [switchError, setSwitchError] = useState("")
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [loadStatus, setLoadStatus] = useState(0)
  const namespaceUidRef = useRef("")
  const switchRef = useRef<SwitchRequest | null>(null)

  const reload = useCallback(() => setAttempt((value) => value + 1), [])

  useEffect(() => {
    let active = true
    const transition = switchRef.current
    async function load() {
      setLoading(true)
      setFailed(false)
      setLoadError("")
      setLoadStatus(0)
      try {
        const [info, spaces, unread] = await Promise.all([
          getSelfInfo({ silent: true }),
          listSelfNamespaces({ silent: true }),
          getUnreadMessageCount()
            .then((result) => result.count ?? 0)
            .catch(() => 0),
        ])
        if (!active) {
          return
        }
        const items = spaces.items ?? []
        const stored = localStorage.getItem("namespace") ?? ""
        const requested = transition?.to ?? stored
        // 系统管理员可以只读进入任意租户：目标不在成员列表时先探测权限，
        // 只有服务端确认 readOnly 且返回了租户信息才接受，否则回落到成员租户。
        const canProbeReadOnly = isSystemAdmin(info)
        let next =
          requested && items.some((item) => item.uid === requested)
            ? requested
            : ""
        let probed: SelfPermissions | null = null
        if (!next && requested && canProbeReadOnly) {
          try {
            probed = await getSelfPermissions(requested, { silent: true })
          } catch {
            probed = null
          }
          if (!active) {
            return
          }
          if (probed?.readOnly && probed.namespace?.uid === requested) {
            next = requested
          }
        }
        // 切换目标可能已被删除或成员身份被撤销，这时不要静默落到别的租户。
        if (transition && next !== transition.to) {
          throw new Error(
            i18n.t(
              transition.kind === "readonly"
                ? "readonly.unavailable"
                : "namespace.switcher.missing"
            )
          )
        }
        if (!next) {
          next = items[0]?.uid ?? ""
        }
        // 权限按目标租户读取，但不提前改 localStorage / namespaceUid：
        // 拉起中的页面轮询仍然使用旧租户，避免新旧权限交叉。
        let granted: string[] = []
        let nextReadOnly = false
        let nextViewing: NamespaceItem | null = null
        if (next) {
          const result =
            probed && next === requested
              ? probed
              : await getSelfPermissions(next, { silent: true })
          if (!active) {
            return
          }
          granted = result.permissions ?? []
          // 只有只读视角才把租户信息放进会话；成员视角的租户信息以成员列表为准。
          if (result.readOnly && result.namespace) {
            nextReadOnly = true
            nextViewing = result.namespace
          }
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
        setReadOnly(nextReadOnly)
        setViewingNamespace(nextViewing)
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
          transition.resolve(true)
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
          transition.resolve(false)
          return
        }
        setFailed(true)
        // 会话加载失败时保留状态码，外壳据此区分 502（网关）与 503（维护）。
        setLoadStatus(err instanceof ApiError ? err.status : 502)
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
      // 本次加载被下一次 attempt 取代时，别让切换调用方一直等不到结果；
      // 已成功完成的切换会把 switchRef 清空，这里只兜底中断的情况。
      if (switchRef.current === transition) {
        transition?.resolve(false)
      }
    }
  }, [attempt])

  // 切换只在 effect 里真正执行；resolve(true) 表示新租户已经加载成功。
  const switchNamespace = useCallback(
    (uid: string, kind: SwitchKind): Promise<boolean> => {
      if (!uid || uid === namespaceUidRef.current) {
        return Promise.resolve(true)
      }
      if (switchRef.current) {
        return Promise.resolve(false)
      }
      return new Promise<boolean>((resolve) => {
        switchRef.current = {
          from: namespaceUidRef.current,
          to: uid,
          kind,
          resolve,
        }
        setSwitchError("")
        setSwitching(true)
        setSwitchingTo(uid)
        setAttempt((value) => value + 1)
      })
    },
    []
  )

  const setNamespace = useCallback(
    (uid: string) => {
      void switchNamespace(uid, "switch")
    },
    [switchNamespace]
  )

  const enterNamespace = useCallback(
    (uid: string) => switchNamespace(uid, "readonly"),
    [switchNamespace]
  )

  // 退出只读：清掉持久化的租户，重新加载会回落到第一个成员租户（没有则空）。
  const exitReadOnly = useCallback(() => {
    if (switchRef.current) {
      return
    }
    localStorage.removeItem("namespace")
    namespaceUidRef.current = ""
    setNamespaceUid("")
    setPermissions([])
    setReadOnly(false)
    setViewingNamespace(null)
    setSwitchError("")
    // 先进入加载态，避免下一次 effect 执行前用空的权限集渲染一帧。
    setLoading(true)
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
      isSystemAdmin:
        systemRole === "SUPER_ADMIN" || systemRole === "SYSTEM_ADMIN",
      isSuperAdmin: systemRole === "SUPER_ADMIN",
      unreadMessages,
      namespaceUid,
      readOnly,
      viewingNamespace,
      loading,
      switching,
      switchingTo,
      switchError,
      failed,
      loadError,
      loadStatus,
      reload,
      setNamespace,
      enterNamespace,
      exitReadOnly,
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
      readOnly,
      viewingNamespace,
      loading,
      switching,
      switchingTo,
      switchError,
      failed,
      loadError,
      loadStatus,
      reload,
      setNamespace,
      enterNamespace,
      exitReadOnly,
      clearSwitchError,
      refreshUnreadMessages,
    ]
  )

  return (
    <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
  )
}

export function useSession() {
  const value = useContext(SessionContext)
  if (!value) {
    throw new Error("useSession must be used in SessionProvider")
  }
  return value
}
