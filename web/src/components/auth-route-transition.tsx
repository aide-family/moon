import { ViewTransition } from "react"
import type { ReactNode } from "react"

/**
 * 认证页与 app 壳之间的整页换场。
 *
 * React 的 `<ViewTransition>` 只在 `startTransition` 触发的更新上生效：登录成功时
 * 认证页边界 unmount（`auth-exit` 淡出）、app 壳边界 mount（`auth-enter` 淡入），
 * 两张快照同时交接。`default="none"` 关掉 update 与 share，所以 app 内部换页、
 * 表格重排都不会被卷进这次过渡；退出登录反向走同一对类名。
 *
 * 放在路由 element 上而不是页面组件内部：它要包住整页，且不能在 DOM 节点里层。
 */
export function AuthRouteTransition({ children }: { children: ReactNode }) {
  return (
    <ViewTransition enter="auth-enter" exit="auth-exit" default="none">
      {children}
    </ViewTransition>
  )
}
