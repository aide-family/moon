// Grafana 面板地址处理。
//
// 大盘只保存面板的 Grafana 分享/嵌入地址，浏览器直接加载 iframe，控制面不代理。
// 校验规则与服务端 app/alioth/internal/biz/board 保持一致，用于表单即时提示；
// 真正生效的仍是服务端校验，前端提示不能替代后端拒绝。
//
// 嵌入地址会追加/覆盖 kiosk、theme、from、to：kiosk 已有值时保留，
// theme 跟随 Alioth 的明暗模式，from/to 跟随查看页选择的时间范围；
// 其余 query（var-* / refresh 等）保留。

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

/** 注入到 Grafana URL 的展示参数。 */
export interface GrafanaEmbedOptions {
  /** Grafana 明暗主题，跟随 Alioth 的 resolvedTheme。 */
  theme?: "light" | "dark"
  /** Grafana 时间范围起点，例如 now-3h。 */
  from?: string
  /** Grafana 时间范围终点，例如 now。 */
  to?: string
}

/** 取出参数的键名；非法编码时按原文比较，不抛错。 */
function queryKey(pair: string): string {
  const rawKey = pair.split("=", 1)[0] ?? ""
  try {
    return decodeURIComponent(rawKey)
  } catch {
    return rawKey
  }
}

/**
 * 生成 iframe 使用的嵌入地址：保留原始 query，追加/覆盖 theme、from、to，
 * 并确保 kiosk 存在。
 *
 * 注意：Grafana 只把裸 `kiosk`（或 `kiosk=tv`）当作嵌入模式；`kiosk=` 空值
 * 会被当成普通页面，露出导航和搜索栏。因此这里不能交给 URLSearchParams
 * 序列化（它会把空值写成 `kiosk=`），而是手工拼 query，并把历史空值规范回裸
 * `kiosk`。地址本身非法时原样返回，交给 validateGrafanaURL 拦截。
 */
export function buildGrafanaEmbedURL(
  raw: string,
  options: GrafanaEmbedOptions = {}
): string {
  const trimmed = raw.trim()
  if (trimmed === "") {
    return trimmed
  }
  const hashIndex = trimmed.indexOf("#")
  const fragment = hashIndex >= 0 ? trimmed.slice(hashIndex) : ""
  const withoutFragment = hashIndex >= 0 ? trimmed.slice(0, hashIndex) : trimmed
  const queryIndex = withoutFragment.indexOf("?")
  const base = queryIndex >= 0 ? withoutFragment.slice(0, queryIndex) : withoutFragment
  try {
    new URL(base)
  } catch {
    return trimmed
  }
  const query = queryIndex >= 0 ? withoutFragment.slice(queryIndex + 1) : ""
  const pairs = query === "" ? [] : query.split("&").filter(Boolean)
  const replaced = new Set(["theme", "from", "to"])
  const kept: string[] = []
  let kiosk: string | null = null
  for (const pair of pairs) {
    const key = queryKey(pair)
    if (key === "kiosk") {
      // `kiosk=` 是历史生成物；规范回裸 kiosk。`kiosk=tv` 保留原值。
      kiosk = pair === "kiosk=" ? "kiosk" : pair
      continue
    }
    if (replaced.has(key)) {
      continue
    }
    kept.push(pair)
  }
  if (kiosk === null) {
    kiosk = "kiosk"
  }
  const extra: string[] = [kiosk]
  if (options.theme) {
    extra.push(`theme=${encodeURIComponent(options.theme)}`)
  }
  if (options.from) {
    extra.push(`from=${encodeURIComponent(options.from)}`)
  }
  if (options.to) {
    extra.push(`to=${encodeURIComponent(options.to)}`)
  }
  return `${base}?${[...kept, ...extra].join("&")}${fragment}`
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
