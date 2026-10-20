export type RangePreset = "1h" | "3h" | "6h" | "12h" | "1d" | "7d" | "custom"

export const RANGE_PRESET_SECONDS: Record<Exclude<RangePreset, "custom">, number> = {
  "1h": 3600,
  "3h": 10800,
  "6h": 21600,
  "12h": 43200,
  "1d": 86400,
  "7d": 604800,
}

/** 把预设或自定义时间解析成 Unix 秒。自定义输入不合法时返回 NaN，由调用方提示。 */
export function resolveRange(
  preset: RangePreset,
  customStart: string,
  customEnd: string,
): { start: number; end: number } {
  if (preset === "custom") {
    return { start: Number(customStart), end: Number(customEnd) }
  }
  const end = Math.floor(Date.now() / 1000)
  return { start: end - RANGE_PRESET_SECONDS[preset], end }
}
