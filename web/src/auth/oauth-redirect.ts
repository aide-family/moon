import type { OAuth2Provider } from "@/api/auth/types"

// 第三方登录跳转前的现场：state、来源、回调地址和登录后回跳路径。
// 只写进 sessionStorage（key 含 state），回调页凭 URL 里的 state 取回并校验。
export type OAuthRedirect = {
  state: string
  provider: OAuth2Provider
  redirectUri: string
  next: string
}

const storagePrefix = "alioth.oauth."

// 第三方展示名。飞书后端支持但管理台不展示，保留名称以便回调页兜底。
const providerNames: Record<OAuth2Provider, string> = {
  GITHUB: "GitHub",
  GITEE: "Gitee",
  FEISHU: "Feishu",
  DINGTALK: "DingTalk",
}

/** 第三方展示名，用于提示文案。 */
export function oauthProviderName(provider: OAuth2Provider): string {
  return providerNames[provider]
}

function storageKey(state: string): string {
  return `${storagePrefix}${state}`
}

function readStorage(key: string): string | null {
  try {
    return sessionStorage.getItem(key)
  } catch {
    return null
  }
}

// 各平台回调里授权码的查询参数名。钉钉返回的是 authCode，其余平台都是 code。
const codeParams: Partial<Record<OAuth2Provider, string>> = {
  DINGTALK: "authCode",
}

/** 第三方回调里授权码的查询参数名。 */
export function oauthCodeParam(provider: OAuth2Provider): string {
  return codeParams[provider] ?? "code"
}

function isOAuth2Provider(value: unknown): value is OAuth2Provider {
  return (
    value === "GITHUB" ||
    value === "GITEE" ||
    value === "FEISHU" ||
    value === "DINGTALK"
  )
}

/** 前端固定的授权回调地址，必须与后端配置的 redirect_uri 一致。 */
export function oauthRedirectUri(): string {
  return `${window.location.origin}/auth/callback`
}

/** 一次性 OAuth state；非安全上下文没有 crypto.randomUUID 时退回随机串。 */
export function newOAuthState(): string {
  try {
    return crypto.randomUUID()
  } catch {
    return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 12)}`
  }
}

export function storeOAuthRedirect(redirect: OAuthRedirect): void {
  try {
    sessionStorage.setItem(storageKey(redirect.state), JSON.stringify(redirect))
  } catch {
    // 隐私模式下 sessionStorage 可能不可写：跳转仍会发生，回调页会提示 state 失效。
  }
}

/** 读取并校验 state 对应的现场；缺失或来源不合法时返回 null。 */
export function readOAuthRedirect(state: string | null): OAuthRedirect | null {
  if (!state) {
    return null
  }
  const raw = readStorage(storageKey(state))
  if (!raw) {
    return null
  }
  try {
    const parsed = JSON.parse(raw) as Partial<OAuthRedirect>
    if (parsed.state !== state || !isOAuth2Provider(parsed.provider)) {
      return null
    }
    return {
      state,
      provider: parsed.provider,
      redirectUri:
        typeof parsed.redirectUri === "string" ? parsed.redirectUri : "",
      next: typeof parsed.next === "string" ? parsed.next : "",
    }
  } catch {
    return null
  }
}

export function clearOAuthRedirect(state: string): void {
  try {
    sessionStorage.removeItem(storageKey(state))
  } catch {
    // 清理失败不影响登录结果。
  }
}
