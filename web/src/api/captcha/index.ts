import { get } from "@/api/http"

import type { Captcha } from "./types"

/** GET /v1/captcha */
export function getCaptcha(signal?: AbortSignal): Promise<Captcha> {
  return get<Captcha>("/captcha", undefined, signal ? { signal } : undefined)
}
