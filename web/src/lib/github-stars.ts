// GitHub 仓库 star 数量：公共 API 未认证时限流 60 次/小时，因此一律先读
// sessionStorage 缓存，未过期就不再请求；请求失败也只是没有数字，不影响登录页。
const CACHE_PREFIX = "alioth.github-stars."
// 十分钟内重复进登录页（含注册页跳回、刷新）不再打 API；star 数不是实时指标。
const CACHE_TTL_MS = 10 * 60 * 1000
const REQUEST_TIMEOUT_MS = 8_000
const API_BASE = "https://api.github.com/repos/"

type StarCacheEntry = {
  count: number
  fetchedAt: number
}

// 同一 slug 的并发请求共用一次 fetch：开发模式 StrictMode 会挂载两次，
// 真实浏览器里多标签页也可能同时进入登录页，没必要各打一次。
const inFlight = new Map<string, Promise<number | null>>()

/** 从仓库地址取出 owner/repo；API 路径与缓存 key 都用它，也避免两个常量各写一份。 */
export function repositorySlug(repositoryURL: string): string | null {
  try {
    const slug = new URL(repositoryURL).pathname.replace(/^\/+|\/+$/g, "")
    return slug === "" ? null : slug
  } catch {
    return null
  }
}

function readCache(slug: string): StarCacheEntry | null {
  try {
    const raw = sessionStorage.getItem(CACHE_PREFIX + slug)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as Partial<StarCacheEntry>
    if (
      typeof parsed.count !== "number" ||
      !Number.isFinite(parsed.count) ||
      parsed.count < 0 ||
      typeof parsed.fetchedAt !== "number" ||
      !Number.isFinite(parsed.fetchedAt)
    ) {
      return null
    }
    return { count: parsed.count, fetchedAt: parsed.fetchedAt }
  } catch {
    // 隐私模式下 sessionStorage 不可读写：当作没有缓存，退回每次请求。
    return null
  }
}

function writeCache(slug: string, count: number): void {
  const entry: StarCacheEntry = { count, fetchedAt: Date.now() }
  try {
    sessionStorage.setItem(CACHE_PREFIX + slug, JSON.stringify(entry))
  } catch {
    // 写不进去只影响下次是否还要请求，不影响这次展示。
  }
}

/** 仍在有效期内的缓存值，供首屏同步渲染，避免回到登录页时数字再“长”出来。 */
export function cachedRepositoryStars(slug: string): number | null {
  const cached = readCache(slug)
  if (!cached || Date.now() - cached.fetchedAt >= CACHE_TTL_MS) {
    return null
  }
  return cached.count
}

async function requestStars(slug: string): Promise<number | null> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)
  try {
    const response = await fetch(API_BASE + slug, {
      headers: { Accept: "application/vnd.github+json" },
      signal: controller.signal,
      // 缓存由上面的 TTL 负责，不再叠一层浏览器缓存，否则限流时也看不出刷新。
      cache: "no-store",
    })
    if (!response.ok) {
      return null
    }
    const payload = (await response.json()) as { stargazers_count?: unknown }
    const count = payload.stargazers_count
    return typeof count === "number" && Number.isFinite(count) && count >= 0
      ? count
      : null
  } catch {
    // 超时、断网、403 限流都走这里：调用方按“没有数字”渲染，不弹错。
    return null
  } finally {
    window.clearTimeout(timer)
  }
}

/**
 * 读取仓库 star 数：有效缓存直接用，过期才请求刷新；刷新失败退回过期缓存，
 * 两者都没有就返回 null。永不抛错。
 */
export function readRepositoryStars(slug: string): Promise<number | null> {
  const cached = readCache(slug)
  if (cached && Date.now() - cached.fetchedAt < CACHE_TTL_MS) {
    return Promise.resolve(cached.count)
  }
  const pending = inFlight.get(slug)
  if (pending) {
    return pending
  }
  const request = requestStars(slug)
    .then((count) => {
      if (count !== null) {
        writeCache(slug, count)
        return count
      }
      // 刷新失败时旧值仍然可读，比突然只剩一个 Star 图标稳。
      return cached?.count ?? null
    })
    .finally(() => {
      inFlight.delete(slug)
    })
  inFlight.set(slug, request)
  return request
}

/**
 * star 数量摘要：不足一千显示原数，过千用 k 缩写、最多一位小数（1234 → 1.2k），
 * 十万以上省掉小数（123456 → 123k）。中英文习惯一致，输出保持 ASCII，
 * 这样两种语言下按钮宽度相同，切换语言也不会跳动。
 */
export function formatStarCount(count: number): string {
  if (!Number.isFinite(count) || count < 0) {
    return ""
  }
  if (count < 1000) {
    return String(Math.floor(count))
  }
  if (count < 100_000) {
    return `${(count / 1000).toFixed(1).replace(/\.0$/, "")}k`
  }
  if (count < 1_000_000) {
    return `${Math.floor(count / 1000)}k`
  }
  return `${(count / 1_000_000).toFixed(1).replace(/\.0$/, "")}m`
}
