// 上次访问页面的记忆：按用户隔离，登录后由 "/" 着陆页回放。
//
// 存储格式：localStorage["lastPath:<userUid>"] = "/alerts"
// - 值是 App.tsx 里注册的登录内路径（不含查询串与 hash），查询串常带一次性筛选，
//   回放时不应把上次的筛选条件带进下一次登录。
// - 路径已被删除时读出来的值视为无效，着陆页回落到概览，无需额外清理。
// - 只读写 localStorage，不依赖 React，便于单独测试。

const storagePrefix = "lastPath:"

// 与 App.tsx 里受登录保护的路由一一对应；增删页面时必须同步这里，
// 否则新页面不会被记住，已删除页面的旧记录还会被当成有效路径。
const knownPaths = new Set([
  "/overview",
  "/profile",
  "/forbidden",
  "/namespaces",
  "/orders",
  "/tenants",
  "/packages",
  "/users",
  "/workers",
  "/invitations",
  "/messages",
  "/members",
  "/roles",
  "/datasources",
  "/probe-targets",
  "/alert-rules",
  "/alert-dictionaries",
  "/alert-pages",
  "/alerts",
  "/alert-history",
  "/notification-channels",
  "/notification-templates",
  "/notification-templates/presets",
  "/notification-groups",
  "/notification-subscriptions",
  "/notification-oncalls",
  "/message-logs",
  "/clusters",
  "/agents",
  "/ssh-commands",
])

// 带 uid 的详情页无法穷举，用模式匹配。
const knownPatterns = [
  /^\/clusters\/[^/]+$/,
  /^\/tenants\/[^/]+\/orders$/,
  /^\/agents\/[^/]+\/shell$/,
  /^\/ssh-commands\/[^/]+\/results$/,
]

// 根路径是着陆页本身，/forbidden 是权限结果页：两者都不代表“用户想回到的页面”。
const ignoredPaths = new Set(["/", "/forbidden"])

// 浏览器地址栏可能带尾斜杠（例如 /alerts/），按注册路径归一化后再判断。
// 导出给外壳守卫复用：租户页判定必须和路由表使用同一套归一化，否则 /profile/
// 这类手写地址会被误判成租户相关页面。
export function normalizeAppPath(pathname: string): string {
  if (pathname.length > 1 && pathname.endsWith("/")) {
    return pathname.replace(/\/+$/, "") || "/"
  }
  return pathname
}

function storageKey(userUid: string): string {
  return `${storagePrefix}${userUid}`
}

/** pathname 是否是 App.tsx 里注册的登录内页面（不含登录/注册/OAuth 等公开页）。 */
export function isKnownAppPath(pathname: string): boolean {
  const normalized = normalizeAppPath(pathname)
  return (
    knownPaths.has(normalized) ||
    knownPatterns.some((pattern) => pattern.test(normalized))
  )
}

/** 记住用户当前所在的页面；根路径、/forbidden 与未知路径不写入。 */
export function rememberLastPath(userUid: string, pathname: string): void {
  if (!userUid) {
    return
  }
  const normalized = normalizeAppPath(pathname)
  if (!isKnownAppPath(normalized) || ignoredPaths.has(normalized)) {
    return
  }
  try {
    localStorage.setItem(storageKey(userUid), normalized)
  } catch {
    // 隐私模式下 localStorage 可能不可写：下次登录回落到概览即可。
  }
}

/** 读取用户上次访问的页面；记录已失效（页面被删除或被排除）时返回 null。 */
export function readLastPath(userUid: string): string | null {
  if (!userUid) {
    return null
  }
  let stored: string | null
  try {
    stored = localStorage.getItem(storageKey(userUid))
  } catch {
    return null
  }
  if (!stored) {
    return null
  }
  const normalized = normalizeAppPath(stored)
  if (!isKnownAppPath(normalized) || ignoredPaths.has(normalized)) {
    return null
  }
  return normalized
}
