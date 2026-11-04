import { useCallback, useEffect, useSyncExternalStore } from "react"

import { fetchSiteConfig } from "@/api/site/public"
import type { SiteConfig } from "@/api/site/types"

import { useSiteLocale } from "./use-site-locale"

interface SiteConfigState {
  config: SiteConfig | null
  /** config 对应的请求语言；空串表示还没有成功请求过。 */
  locale: string
  loading: boolean
  failed: boolean
}

// 站点配置在页面会话内几乎不变，模块级缓存让登录页、注册页、官网与官网首页共用一次
// 请求。首页文案与能力卡片按语言下发，所以缓存同时记录 config 属于哪种语言。
// 未加载时先保持 loading，避免登录页先渲染默认入口、拿到配置后再收起造成的闪动。
let state: SiteConfigState = {
  config: null,
  locale: "",
  loading: true,
  failed: false,
}
let inFlight: { locale: string; promise: Promise<void> } | null = null
const listeners = new Set<() => void>()

function setState(next: Partial<SiteConfigState>) {
  state = { ...state, ...next }
  for (const listener of listeners) {
    listener()
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

function getSnapshot(): SiteConfigState {
  return state
}

function load(locale: string, force: boolean): Promise<void> {
  if (!force && state.config && state.locale === locale) {
    if (state.loading || state.failed) {
      setState({ loading: false, failed: false })
    }
    return Promise.resolve()
  }
  // 并发挂载共享同一笔请求；StrictMode 的重复 effect 也只发一次网络调用。
  if (inFlight) {
    if (inFlight.locale === locale) {
      return inFlight.promise
    }
    // 换语言时先等上一笔结束，避免迟到的旧语言响应覆盖新语言内容。
    return inFlight.promise.then(() => load(locale, force))
  }
  setState({ loading: true, failed: false })
  const promise = fetchSiteConfig(locale)
    .then((config) => {
      setState({ config, locale, loading: false, failed: false })
    })
    .catch(() => {
      // 拿不到配置时由调用方决定展示可重试错误；旧语言的 config 不再当作当前语言。
      setState({ loading: false, failed: true })
    })
    .finally(() => {
      inFlight = null
    })
  inFlight = { locale, promise }
  return promise
}

/**
 * 公开站点配置（站点信息 + 登录方式开关 + 当前语言的首页文案与能力卡片）。
 *
 * 模块级缓存：同一页面里的登录页页脚、登录入口、官网页头与官网首页只请求一次；
 * 语言切换后按新语言重新请求，`reload` 会强制重新拉取当前语言。
 */
export function useSiteConfig() {
  const locale = useSiteLocale()
  const snapshot = useSyncExternalStore(subscribe, getSnapshot, getSnapshot)

  useEffect(() => {
    void load(locale, false)
  }, [locale])

  const reload = useCallback(() => {
    void load(locale, true)
  }, [locale])

  return {
    config: snapshot.config,
    /** config 实际对应的语言；和当前语言不一致时调用方应视为仍在加载。 */
    locale: snapshot.locale,
    loading: snapshot.loading,
    failed: snapshot.failed,
    reload,
  }
}
