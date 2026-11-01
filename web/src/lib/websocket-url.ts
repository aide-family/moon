/**
 * 拼绝对 WebSocket 地址。必须带 `://`：缺冒号时浏览器会把 `ws//host/...`
 * 当成相对路径，按当前页面 URL 解析到前端 SPA 路由上，握手会一直 pending。
 */
export function absoluteWebSocketURL(path: string): string {
  const proto = window.location.protocol === "https:" ? "wss" : "ws"
  const suffix = path.startsWith("/") ? path : `/${path}`
  return `${proto}://${window.location.host}${suffix}`
}
