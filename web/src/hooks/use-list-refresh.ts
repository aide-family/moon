/**
 * 列表页的统一刷新事件。列表容器（ListFrame）的刷新按钮默认派发它，
 * usePagedList 收到后按当前页码和筛选条件重新请求，
 * 页面不需要为每个列表再传一个刷新函数。
 */
export const LIST_REFRESH_EVENT = "alioth:list-refresh"
