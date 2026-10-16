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
