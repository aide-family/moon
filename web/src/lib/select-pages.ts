/**
 * 下拉数据的全量拉取：接口本身按游标（last_uid）或页码分页，页大小上限 100。
 * 组件只取第一页会把第 101 项之后的数据悄悄丢掉，这里按页取完再交给下拉框。
 */
export async function loadAllSelectItems<T>(
  fetchPage: (lastUid: string) => Promise<{ items: T[]; lastUid: string }>,
  pageLimit: number,
  maxPages = 50
): Promise<T[]> {
  const items: T[] = []
  const seen = new Set<string>()
  let lastUid = ""
  for (let page = 0; page < maxPages; page++) {
    const result = await fetchPage(lastUid)
    const batch = result.items ?? []
    for (const item of batch) {
      if (seen.has(lastUidOf(item))) {
        continue
      }
      seen.add(lastUidOf(item))
      items.push(item)
    }
    // 不足一页、没有下一页游标或游标没有前进时结束，避免服务端异常导致的死循环。
    if (
      batch.length < pageLimit ||
      !result.lastUid ||
      result.lastUid === lastUid
    ) {
      break
    }
    lastUid = result.lastUid
  }
  return items
}

/**
 * 页码分页（page/pageSize + total）的全量拉取。
 * 与游标版本一样，最多翻 maxPages 页，防止异常数据把页面拖死。
 */
export async function loadAllPagedItems<T>(
  fetchPage: (
    page: number,
    pageSize: number
  ) => Promise<{ items: T[]; total: number }>,
  pageSize = 100,
  maxPages = 50
): Promise<T[]> {
  const items: T[] = []
  for (let page = 1; page <= maxPages; page++) {
    const result = await fetchPage(page, pageSize)
    const batch = result.items ?? []
    items.push(...batch)
    if (batch.length === 0 || items.length >= (result.total ?? 0)) {
      break
    }
  }
  return items
}

function lastUidOf(item: unknown): string {
  if (item && typeof item === "object" && "uid" in item) {
    const uid = (item as { uid?: unknown }).uid
    return typeof uid === "string" ? uid : ""
  }
  return ""
}
