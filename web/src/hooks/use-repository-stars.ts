import { useEffect, useState } from "react"

import { cachedRepositoryStars, readRepositoryStars } from "@/lib/github-stars"

/**
 * 仓库 star 数，取不到时为 null。
 *
 * 只暴露数量、不暴露 loading / error：入口在没有数字时直接显示 Star 文案，
 * 不做骨架占位、不弹错，所以首屏不会因为这次请求多一次跳动。
 * 初始值同步读缓存，回到登录页时数字直接就在。
 */
export function useRepositoryStars(slug: string | null): number | null {
  const [count, setCount] = useState<number | null>(() =>
    slug ? cachedRepositoryStars(slug) : null
  )

  useEffect(() => {
    if (!slug) {
      return
    }
    let active = true
    void readRepositoryStars(slug).then((next) => {
      // 卸载或 slug 变化后丢弃结果，避免把旧仓库的数字写到新仓库上。
      if (active && next !== null) {
        setCount(next)
      }
    })
    return () => {
      active = false
    }
  }, [slug])

  return count
}
