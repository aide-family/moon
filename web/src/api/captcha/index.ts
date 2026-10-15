import { get } from "@/api/http"

import type { Captcha } from "./types"

/** GET /v1/captcha */
export function getCaptcha(): Promise<Captcha> {
  return get<Captcha>("/captcha")
}
