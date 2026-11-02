// Grafana 面板地址处理。
//
// 大盘只保存面板的 Grafana 分享/嵌入地址，浏览器直接加载 iframe，控制面不代理。
// 校验规则与服务端 app/alioth/internal/biz/board 保持一致，用于表单即时提示；
// 真正生效的仍是服务端校验，前端提示不能替代后端拒绝。
//
// 嵌入地址在原始 query 之后追加 kiosk（隐藏 Grafana 侧边栏），已有 kiosk 时
// 保留原值，其余 query（from / to / var-* / refresh 等）原样保留。

/** 地址最大字节数，与后端 maxURLBytes 一致（按 UTF-8 字节计）。 */
export const GRAFANA_URL_MAX_BYTES = 2048

/** 面板路径必须包含的三类 Grafana 分享路径之一。 */
const GRAFANA_PATHS = ["/d/", "/d-solo/", "/public-dashboards/"]

/** query 参数名里出现这些片段即视为把凭据写进了地址，不区分大小写。 */
const CREDENTIAL_KEYWORDS = ["token", "secret", "password", "api_key", "apikey"]

/**
 * URL 校验结果：null 表示合法；否则是可用于 i18n 的错误码。
 * 页面按 `board.form.urlIssue.<code>` 取文案。
 */
export type GrafanaURLIssue =
  | "required"
  | "tooLong"
  | "invalid"
  | "scheme"
  | "userinfo"
  | "fragment"
  | "path"
  | "credentials"

const KIOSK_PATTERN = /[?&]kiosk(?:=[^&#]*)?(?:[&#]|$)/

function hasKiosk(raw: string): boolean {
  if (KIOSK_PATTERN.test(raw)) {
    return true
  }
  try {
    return new URL(raw).searchParams.has("kiosk")
  } catch {
    return false
  }
}

/**
 * 生成 iframe 使用的嵌入地址：保留原 query，没有 kiosk 参数时追加 `kiosk`，
 * 已有 kiosk 时保留原值（例如 `kiosk=tv` 不被覆盖）。
 * 地址本身就是非法 URL 时原样返回，交给 validateGrafanaURL 拦截。
 */
export function buildGrafanaEmbedURL(raw: string): string {
  const trimmed = raw.trim()
  if (trimmed === "") {
    return trimmed
  }
  if (hasKiosk(trimmed)) {
    return trimmed
  }
  // 手工拼接而不是 URL/URLSearchParams，避免把原 query 的编码形式改掉。
  const hashIndex = trimmed.indexOf("#")
  const base = hashIndex >= 0 ? trimmed.slice(0, hashIndex) : trimmed
  const fragment = hashIndex >= 0 ? trimmed.slice(hashIndex) : ""
  if (base === "" || base.endsWith("?") || base.endsWith("&")) {
    return `${base}kiosk${fragment}`
  }
  const separator = base.includes("?") ? "&" : "?"
  return `${base}${separator}kiosk${fragment}`
}

/**
 * 校验 Grafana 分享/嵌入地址。规则与后端 validateGrafanaURL 一致：
 * 绝对 http(s) 地址、不带 userinfo / fragment、路径含 Grafana 面板路径、
 * query 不带凭据类参数名。返回 null 表示通过，否则返回可 i18n 的错误码。
 */
export function validateGrafanaURL(raw: string): GrafanaURLIssue | null {
  const trimmed = raw.trim()
  if (trimmed === "") {
    return "required"
  }
  if (new TextEncoder().encode(trimmed).length > GRAFANA_URL_MAX_BYTES) {
    return "tooLong"
  }
  let parsed: URL
  try {
    parsed = new URL(trimmed)
  } catch {
    return "invalid"
  }
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return "scheme"
  }
  if (parsed.host === "") {
    return "invalid"
  }
  // new URL 会把空 userinfo（http://@host/...）归一化掉，按原始 authority 再查一次。
  const authority = /^[a-zA-Z][\w+.-]*:\/\/([^/?#]*)/.exec(trimmed)?.[1] ?? ""
  if (
    parsed.username !== "" ||
    parsed.password !== "" ||
    authority.includes("@")
  ) {
    return "userinfo"
  }
  // 后端拒绝任何 fragment；# 也可能出现在 query 值里，以解析结果为准。
  if (parsed.hash !== "") {
    return "fragment"
  }
  if (!GRAFANA_PATHS.some((segment) => parsed.pathname.includes(segment))) {
    return "path"
  }
  for (const name of parsed.searchParams.keys()) {
    const lower = name.toLowerCase()
    if (CREDENTIAL_KEYWORDS.some((keyword) => lower.includes(keyword))) {
      return "credentials"
    }
  }
  return null
}
