import crypto from "node:crypto"
import path from "node:path"

import { test as base } from "@playwright/test"

/**
 * E2E 夹具：直接用 JWT 注入登录态，不走验证码登录。
 * 默认值对应本机开发环境；CI 或其它环境用环境变量覆盖。
 */
export const E2E_SECRET = process.env.E2E_JWT_SECRET ?? "dev-only-change-me"
export const E2E_UID =
  process.env.E2E_USER_UID ?? "0ab93de2a42571f26e563d1b10dc8e9d"
export const E2E_NAMESPACE =
  process.env.E2E_NAMESPACE ?? "1b8307a4abf29e8f783bb9c0dd651695"
export const E2E_API = process.env.E2E_API_URL ?? "http://127.0.0.1:8100"
export const E2E_WEB = process.env.E2E_WEB_URL ?? "http://127.0.0.1:8080"
export const SHOT_DIR = path.resolve(import.meta.dirname, "../../tmp/acceptance")

export function mintToken(uid = E2E_UID, ttlSeconds = 3600): string {
  const encode = (value: Buffer | string) =>
    Buffer.from(value).toString("base64url")
  const header = encode(JSON.stringify({ alg: "HS256", typ: "JWT" }))
  const now = Math.floor(Date.now() / 1000)
  const payload = encode(
    JSON.stringify({ uid, kind: "access", iat: now, exp: now + ttlSeconds })
  )
  const signature = crypto
    .createHmac("sha256", E2E_SECRET)
    .update(`${header}.${payload}`)
    .digest("base64url")
  return `${header}.${payload}.${signature}`
}

/** 会话状态：token / 租户 / 语言，主题单独给（浅色与深色各一份）。 */
export function storageState(theme: "light" | "dark" = "light") {
  return {
    cookies: [],
    origins: [
      {
        origin: E2E_WEB,
        localStorage: [
          { name: "token", value: mintToken() },
          { name: "namespace", value: E2E_NAMESPACE },
          { name: "theme", value: theme },
          { name: "lang", value: "zh" },
        ],
      },
    ],
  }
}

interface SitePageSeed {
  uid: string
  slug: string
}

export const test = base.extend<{
  api: ReturnType<typeof createApi>
  seedPage: SitePageSeed
}>({
  // eslint-disable-next-line no-empty-pattern
  api: async ({}, use) => {
    const api = createApi()
    await use(api)
    await api.dispose()
  },
  seedPage: async ({ api }, use) => {
    const slug = `e2e-doc-${Date.now()}`
    const bodyMd = [
      "# E2E 文档标题",
      "",
      "一段正文。",
      "",
      "## 配置示例",
      "",
      "```bash",
      "aliothctl version",
      "```",
      "",
      "## 下一步",
      "",
      "看数据模型。",
    ].join("\n")
    const created = await api.post("/v1/system/site/pages", {
      slug,
      locale: "zh",
      title: "E2E 文档标题",
      summary: "E2E 摘要",
      bodyMd,
    })
    const uid = created.page.item.uid as string
    await api.post(`/v1/system/site/pages/${uid}/publish`, {})
    await use({ uid, slug })
    await api.delete(`/v1/system/site/pages/${uid}`)
  },
})

export function createApi() {
  const headers: Record<string, string> = {
    Accept: "application/json",
    "Content-Type": "application/json",
    Authorization: `Bearer ${mintToken()}`,
    "X-Namespace": E2E_NAMESPACE,
  }
  async function call(method: string, url: string, data?: unknown) {
    const response = await fetch(`${E2E_API}${url}`, {
      method,
      headers,
      body: data === undefined ? undefined : JSON.stringify(data),
    })
    const text = await response.text()
    let json: any = null
    try {
      json = text ? JSON.parse(text) : null
    } catch {
      // 非 JSON 响应
    }
    return { status: response.status, json, text }
  }
  return {
    get: (url: string) => call("GET", url),
    post: (url: string, data?: unknown) => call("POST", url, data),
    put: (url: string, data?: unknown) => call("PUT", url, data),
    delete: (url: string) => call("DELETE", url),
    dispose: async () => {},
  }
}
