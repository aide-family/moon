import { useCallback, useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getCaptcha } from "@/api/captcha"
import type { Captcha, CaptchaKind } from "@/api/captcha/types"
import { formErrorMessage } from "@/lib/form-error"

export interface CaptchaController {
  /** 图形验证码标识，发码和提交登录/注册时随请求上送。 */
  captchaId: string
  /** 图形验证码图片的 data URL；为空时界面显示同尺寸占位。 */
  image: string
  /** 题型；前端据此切换输入提示，空值表示尚未加载。 */
  kind: CaptchaKind
  /** 首次加载或刷新中。 */
  loading: boolean
  /** 加载失败原因，直接显示在字段下方，不会永远停在“加载中”。 */
  error: string
  /** 重新获取一张图形验证码。 */
  refresh: () => Promise<void>
}

/** 图形验证码的加载与刷新。题型随机，页面只负责输入框里的答案。 */
export function useCaptcha(): CaptchaController {
  const { t } = useTranslation()
  const [captchaId, setCaptchaId] = useState("")
  const [image, setImage] = useState("")
  const [kind, setKind] = useState<CaptchaKind>("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  const commit = useCallback(
    async (request: Promise<Captcha>, signal?: AbortSignal) => {
      try {
        const captcha = await request
        if (signal?.aborted) {
          return
        }
        setCaptchaId(captcha.id)
        setImage(captcha.image)
        setKind(captcha.kind ?? "")
        setError("")
      } catch (err) {
        if (signal?.aborted) {
          return
        }
        setCaptchaId("")
        setImage("")
        setKind("")
        setError(formErrorMessage(err, t("form.captchaUnavailable")))
      } finally {
        if (!signal?.aborted) {
          setLoading(false)
        }
      }
    },
    [t]
  )

  useEffect(() => {
    const controller = new AbortController()
    async function load() {
      await commit(getCaptcha(controller.signal), controller.signal)
    }
    void load()
    return () => controller.abort()
  }, [commit])

  const refresh = useCallback(async () => {
    setLoading(true)
    await commit(getCaptcha())
  }, [commit])

  return { captchaId, image, kind, loading, error, refresh }
}
