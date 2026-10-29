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
  /** 当前租户已开通的能力 code；旧后端未返回 features 时为空数组且按「不过滤」处理。 */
  features: string[]
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
  /**
   * 当前租户是否开通了某个能力（GET /v1/self/permissions 的 features）。
   * 旧后端未返回 features 时返回 true（不拦截菜单），避免升级期间租户菜单全部隐藏。
   */
  hasFeature: (code: string) => boolean
  /**
   * 菜单权限（permissions 里的 menu:<能力>）。旧后端没有 features 字段时视为
   * 尚未启用菜单权限并放行，菜单是否可见仍由 can(资源权限) 兜底；
   * 新后端严格按 menu:<能力> 判断。
   */
  canMenu: (capability: string) => boolean
  refreshUnreadMessages: () => void
}

const SessionContext = createContext<SessionValue | null>(null)

export function SessionProvider({ children }: { children: ReactNode }) {
  const [attempt, setAttempt] = useState(0)
  const [user, setUser] = useState<UserItem | null>(null)
  const [namespaces, setNamespaces] = useState<NamespaceItem[]>([])
  const [permissions, setPermissions] = useState<string[]>([])
  // features 为当前租户已开通功能；featuresKnown 区分「后端明确返回空数组」与
  // 「旧后端没有 features 字段」，后者不做功能过滤，避免升级期间菜单消失。
  const [features, setFeatures] = useState<string[]>([])
  const [featuresKnown, setFeaturesKnown] = useState(false)
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
        let grantedFeatures: string[] = []
        let featureFilterKnown = false
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
          // features 是后端新增字段：只有它出现时才启用菜单能力过滤。
          if (result.features !== undefined) {
            grantedFeatures = result.features ?? []
            featureFilterKnown = true
          }
          // 只有只读视角才把租户信息放进会话；成员视角的租户信息以成员列表为准。
          if (result.readOnly && result.namespace) {
            nextReadOnly = true
            nextViewing = result.namespace
          }
          // 租户被停用后后端只返回空权限：这时不要停在空壳租户上，
          // 回落到仍在启用的成员租户，避免导航/接口全部 403 却没有任何出口。
          const member = items.find((item) => item.uid === next)
          const unusable = member ? member.status === "DISABLED" : true
          if (granted.length === 0 && unusable) {
            const fallback = items.find(
              (item) => item.uid !== next && item.status !== "DISABLED"
            )
            if (fallback) {
              next = fallback.uid
              const recovered = await getSelfPermissions(next, { silent: true })
              if (!active) {
                return
              }
              granted = recovered.permissions ?? []
              grantedFeatures = recovered.features ?? []
              featureFilterKnown = recovered.features !== undefined
              nextReadOnly = false
              nextViewing = null
            }
          }
        }
        if (!active) {
          return
        }
        // 自愈可能改了目标租户；显式切换时不能悄悄落到别的租户，按切换失败处理。
        if (transition && next !== transition.to) {
          const targetExists = items.some(
            (item) => item.uid === transition.to
          )
          throw new Error(
            i18n.t(
              transition.kind === "readonly"
                ? "readonly.unavailable"
                : targetExists
                  ? "error.namespaceDisabled"
                  : "namespace.switcher.missing"
            )
          )
        }
        setUser(info.user)
        setNamespaces(items)
        setSystemPermissions(info.systemPermissions ?? [])
        setSystemRole(info.systemRole ?? "USER")
        setUnreadMessages(unread)
        setPermissions(granted)
        setFeatures(grantedFeatures)
        setFeaturesKnown(featureFilterKnown)
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
    setFeatures([])
    setFeaturesKnown(false)
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
      features,
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
      hasFeature: (code: string) =>
        !featuresKnown || features.includes(code),
      // 旧后端（没有 features 字段）按未启用菜单权限放行；新后端严格判断。
      canMenu: (capability: string) =>
        !featuresKnown || permissions.includes(`menu:${capability}`),
      refreshUnreadMessages,
    }),
    [
      user,
      namespaces,
      permissions,
      features,
      featuresKnown,
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
