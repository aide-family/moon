import { useCallback, useEffect, useState } from "react"

/** 与服务端 emailCodeInterval 一致：同一邮箱 60 秒内只能发一次验证码。 */
export const SEND_CODE_COOLDOWN_SECONDS = 60

/** 服务端节流错误 metadata 里的剩余秒数字段。 */
export const RETRY_AFTER_SECONDS_KEY = "retryAfterSeconds"

const STORAGE_PREFIX = "alioth:send-code-cooldown:"

function storageKey(scope: string): string {
  return `${STORAGE_PREFIX}${scope}`
}

function remainingSeconds(key: string): number {
  try {
    const raw = localStorage.getItem(key)
    if (!raw) {
      return 0
    }
    const deadline = Number(raw)
    if (!Number.isFinite(deadline)) {
      localStorage.removeItem(key)
      return 0
    }
    return Math.max(0, Math.ceil((deadline - Date.now()) / 1000))
  } catch {
    // 隐私模式下 localStorage 不可用，退化为本次会话内的倒计时。
    return 0
  }
}

export interface SendCodeCooldown {
  /** 距离可以再次发送还有多少秒；0 表示可以发送。 */
  remaining: number
  /** 开始倒计时。服务端返回剩余秒数时传实际值，否则用默认 60 秒。 */
  start: (seconds?: number) => void
}

/**
 * 发码倒计时。截止时间写入 localStorage，刷新页面后按剩余秒数继续。
 * 键只带用途（login / register）：刷新后邮箱输入框是空的，按邮箱记录会让
 * 倒计时“消失”；同一浏览器 60 秒内换邮箱也保守地继续倒计时。
 */
export function useSendCodeCooldown(scope: string): SendCodeCooldown {
  const key = storageKey(scope)
  const [remaining, setRemaining] = useState(() => remainingSeconds(key))

  // 每秒按截止时间重算，标签页被浏览器节流后回到前台也能立即对齐。
  useEffect(() => {
    const tick = () => setRemaining(remainingSeconds(key))
    tick()
    const timer = window.setInterval(tick, 1000)
    return () => window.clearInterval(timer)
  }, [key])

  const start = useCallback(
    (seconds = SEND_CODE_COOLDOWN_SECONDS) => {
      const value = Math.max(1, Math.ceil(seconds))
      try {
        localStorage.setItem(key, String(Date.now() + value * 1000))
      } catch {
        // 写不进去时仍保留本次倒计时，只是刷新后不保留。
      }
      setRemaining(value)
    },
    [key]
  )

  return { remaining, start }
}
