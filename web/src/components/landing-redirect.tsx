import { Navigate } from "react-router-dom"

import { useSession } from "@/auth/session"
import { readLastPath } from "@/lib/last-page"

// "/" 的着陆页：登录后先按会话选租户，再回放该用户上次访问的页面。
// - 没有租户：去租户页（选择或创建），与 ShellFrame 的租户守卫同一判据，不会来回跳；
// - 有租户且记得有效页面：回到该页面；首次登录或页面已不存在则进概览；
// - 一律 replace："/" 不在历史里留下会被反复回放的中间记录。
export function LandingRedirect() {
  const session = useSession()

  if (session.namespaceUid === "") {
    return <Navigate to="/namespaces" replace />
  }

  const lastPath = session.user ? readLastPath(session.user.uid) : null
  return <Navigate to={lastPath ?? "/overview"} replace />
}
