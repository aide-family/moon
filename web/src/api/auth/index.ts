import { post } from "@/api/http"

import type { EmailLoginParams, LoginResult, OAuth2LoginParams, PasswordLoginParams, RegisterParams } from "./types"

/** POST /v1/auth/register */
export function register(body: RegisterParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/register", body)
}

/** POST /v1/auth/password/login */
export function passwordLogin(body: PasswordLoginParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/password/login", body)
}

/** POST /v1/auth/email/login/code */
export function sendEmailLoginCode(email: string): Promise<Record<string, never>> {
  return post<Record<string, never>>("/auth/email/login/code", { email })
}

/** POST /v1/auth/email/login */
export function emailLogin(body: EmailLoginParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/email/login", body)
}

/** POST /v1/auth/oauth2/login */
export function oauth2Login(body: OAuth2LoginParams): Promise<LoginResult> {
  return post<LoginResult>("/auth/oauth2/login", body)
}
