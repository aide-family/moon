/**
 * 只接受站内绝对路径作为登录/注册后的回跳目标，避免开放重定向。
 *
 * 没有显式 next 时回落到 "/" 着陆页：着陆页按当前用户回放上次访问的页面；
 * 首次登录或上次页面已不存在时再进入概览。显式 next（含 OAuth 现场）仍然优先。
 */
export function readNextPath(value: string | null): string {
  if (!value || !value.startsWith("/") || value.startsWith("//")) {
    return "/"
  }
  // 反斜杠和控制字符会被 URL 解析成跨站地址（例如 "/\evil.com" 解析为 https://evil.com/）。
  // 路由层虽然会拦截并抛错，但用户此时已登录、却会卡在登录页看到英文报错，这里直接拒绝。
  for (const char of value) {
    const code = char.codePointAt(0) ?? 0
    if (char === "\\" || code < 0x20 || code === 0x7f) {
      return "/"
    }
  }
  try {
    const url = new URL(value, window.location.origin)
    if (url.origin !== window.location.origin) {
      return "/"
    }
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return "/"
  }
}
