import { useCallback, useEffect, useSyncExternalStore } from "react"

import { fetchSiteConfig } from "@/api/site/public"
import type { SiteConfig } from "@/api/site/types"

interface SiteConfigState {
  config: SiteConfig | null
  loading: boolean
  failed: boolean
}

// 站点配置在页面会话内几乎不变，模块级缓存让登录页、注册页与官网共用一次请求。
// 未加载时先保持 loading，避免登录页先渲染默认入口、拿到配置后再收起造成的闪动。
let state: SiteConfigState = { config: null, loading: true, failed: false }
let inFlight: Promise<void> | null = null
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

function load(force: boolean): Promise<void> {
  if (!force && state.config) {
    if (state.loading || state.failed) {
      setState({ loading: false, failed: false })
    }
    return Promise.resolve()
  }
  // 并发挂载共享同一笔请求；StrictMode 的重复 effect 也只发一次网络调用。
  if (inFlight) {
    return inFlight
  }
  setState({ loading: true, failed: false })
  inFlight = fetchSiteConfig()
    .then((config) => {
      setState({ config, loading: false, failed: false })
    })
    .catch(() => {
      // 拿不到配置时由调用方回落默认值，不阻塞登录与注册。
      setState({ loading: false, failed: true })
    })
    .finally(() => {
      inFlight = null
    })
  return inFlight
}

/**
 * 公开站点配置（站点信息 + 登录方式开关）。
 *
 * 模块级缓存：同一页面里的登录页页脚、登录入口与注册页页脚只请求一次；
 * 失败后由调用方决定是否展示默认值，`reload` 会强制重新拉取。
 */
export function useSiteConfig() {
  const snapshot = useSyncExternalStore(subscribe, getSnapshot, getSnapshot)

  useEffect(() => {
    void load(false)
  }, [])

  const reload = useCallback(() => {
    void load(true)
  }, [])

  return {
    config: snapshot.config,
    loading: snapshot.loading,
    failed: snapshot.failed,
    reload,
  }
}
