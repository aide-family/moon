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
  return value
}
