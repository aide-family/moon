import { get, put } from "@/api/http"

import type {
  LoginMethods,
  SaveLoginMethodsParams,
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

export type {
  LoginMethods,
  SaveLoginMethodsParams,
  SaveSiteSettingsParams,
  SiteConfig,
  SiteSettings,
} from "./types"
