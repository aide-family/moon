import { get, put } from "@/api/http"

import type {
  EmailSettings,
  LoginMethods,
  OAuthProvider,
  OAuthProviderSettings,
  SaveEmailSettingsParams,
  SaveLoginMethodsParams,
  SaveOAuthProviderParams,
  SaveSiteSettingsParams,
  SiteSettings,
} from "./types"

/** GET /v1/system/site */
export function getSiteSettings(): Promise<{ site: SiteSettings }> {
  return get<{ site: SiteSettings }>("/system/site")
}

/** PUT /v1/system/site */
export function updateSiteSettings(
  body: SaveSiteSettingsParams
): Promise<{ site: SiteSettings }> {
  return put<{ site: SiteSettings }>("/system/site", body)
}

/** GET /v1/system/auth/login-methods */
export function getLoginMethods(): Promise<{ loginMethods: LoginMethods }> {
  return get<{ loginMethods: LoginMethods }>("/system/auth/login-methods")
}

/** PUT /v1/system/auth/login-methods */
export function updateLoginMethods(
  body: SaveLoginMethodsParams
): Promise<{ loginMethods: LoginMethods }> {
  return put<{ loginMethods: LoginMethods }>("/system/auth/login-methods", body)
}

/** GET /v1/system/auth/email-settings */
export function getEmailSettings(): Promise<{ email: EmailSettings }> {
  return get<{ email: EmailSettings }>("/system/auth/email-settings")
}

/** PUT /v1/system/auth/email-settings */
export function updateEmailSettings(
  body: SaveEmailSettingsParams
): Promise<{ email: EmailSettings }> {
  return put<{ email: EmailSettings }>("/system/auth/email-settings", body)
}

/** GET /v1/system/auth/oauth-providers */
export function getOAuthProviders(): Promise<{
  items: OAuthProviderSettings[]
}> {
  return get<{ items: OAuthProviderSettings[] }>("/system/auth/oauth-providers")
}

/** PUT /v1/system/auth/oauth-providers/{provider} */
export function updateOAuthProvider(
  provider: OAuthProvider,
  body: SaveOAuthProviderParams
): Promise<{ item: OAuthProviderSettings }> {
  return put<{ item: OAuthProviderSettings }>(
    `/system/auth/oauth-providers/${provider}`,
    body
  )
}

export type {
  EmailSettings,
  LoginMethods,
  OAuthProvider,
  OAuthProviderSettings,
  SaveEmailSettingsParams,
  SaveLoginMethodsParams,
  SaveOAuthProviderParams,
  SaveSiteSettingsParams,
  SiteConfig,
  SiteSettings,
} from "./types"
