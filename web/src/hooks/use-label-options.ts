import { useEffect, useState } from "react"

import { listLabelOptions, type LabelOption } from "@/api/label"
import { useSession } from "@/auth/session"

/**
 * 当前租户的标签集合。失败时返回空数组，调用方仍允许手工输入集合外的键值。
 * 按租户缓存由后端 Redis 负责，组件只在挂载/租户切换时请求一次。
 */
export function useLabelOptions(): LabelOption[] {
  const session = useSession()
  const [options, setOptions] = useState<LabelOption[]>([])
  useEffect(() => {
    let active = true
    listLabelOptions()
      .then((result) => {
        if (active) {
          setOptions(result.items ?? [])
        }
      })
      .catch(() => {
        if (active) {
          setOptions([])
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])
  return options
}
