export interface LabelPair {
  key: string
  value: string
}

export function parseLabelPairs(raw: string): LabelPair[] {
  try {
    const data = JSON.parse(raw || "{}") as Record<string, string>
    const pairs = Object.entries(data).map(([key, value]) => ({ key, value: String(value) }))
    return pairs.length > 0 ? pairs : [{ key: "", value: "" }]
  } catch {
    return [{ key: "", value: "" }]
  }
}

export function serializeLabelPairs(pairs: LabelPair[]): string {
  const out: Record<string, string> = {}
  for (const pair of pairs) {
    const key = pair.key.trim()
    if (!key) {
      continue
    }
    out[key] = pair.value.trim()
  }
  return JSON.stringify(out)
}

/**
 * 按后端语义判断标签值：标签值是完整匹配的正则。
 * 浏览器和 Go/PostgreSQL 的正则方言不完全一致，非法表达式退回字面比较，
 * 只用于表单预览时保持和后端相同的退路。
 */
export function labelValueMatches(pattern: string, value: string): boolean {
  try {
    return new RegExp(`^(?:${pattern})$`).test(value)
  } catch {
    return value === pattern
  }
}

/** 判断一组标签匹配器是否全部命中目标标签；键不存在时不命中。 */
export function labelMatchersMatch(matchers: LabelPair[], labels: Record<string, string>): boolean {
  return matchers
    .filter((pair) => pair.key.trim() !== "")
    .every((pair) => {
      const key = pair.key.trim()
      if (!Object.prototype.hasOwnProperty.call(labels, key)) {
        return false
      }
      return labelValueMatches(pair.value.trim(), labels[key])
    })
}
