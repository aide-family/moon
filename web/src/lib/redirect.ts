/** 只接受站内绝对路径作为登录/注册后的回跳目标，避免开放重定向。 */
export function readNextPath(value: string | null): string {
  if (!value || !value.startsWith("/") || value.startsWith("//")) {
    return "/namespaces"
  }
  return value
}
