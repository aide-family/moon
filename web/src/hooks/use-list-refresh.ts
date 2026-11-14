import { useEffect, useRef } from "react"

/**
 * 列表页的统一刷新事件。列表容器（ListFrame）的刷新按钮默认派发它，
 * usePagedList 收到后按当前页码和筛选条件重新请求，
 * 页面不需要为每个列表再传一个刷新函数。
 */
export const LIST_REFRESH_EVENT = "alioth:list-refresh"

/** 让任意列表加载逻辑接入统一的刷新按钮，回调始终读最新一次渲染的闭包。 */
export function useListRefresh(callback: () => void) {
  const callbackRef = useRef(callback)
  useEffect(() => {
    callbackRef.current = callback
  }, [callback])
  useEffect(() => {
    const listener = () => callbackRef.current()
    window.addEventListener(LIST_REFRESH_EVENT, listener)
    return () => window.removeEventListener(LIST_REFRESH_EVENT, listener)
  }, [])
}
