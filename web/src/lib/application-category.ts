import type { TFunction } from "i18next"

/** 内置分类；界面只把它们作为输入提示，分类本身是自由字符串（proto ≤32）。 */
export const APPLICATION_CATEGORY_PRESETS = [
  "DATABASE",
  "MIDDLEWARE",
  "OBSERVABILITY",
  "TOOLING",
  "OTHER",
] as const

export type ApplicationCategoryPreset = (typeof APPLICATION_CATEGORY_PRESETS)[number]

/** 留空按 OTHER 处理；长度上限与 proto 一致。 */
export const APPLICATION_CATEGORY_MAX = 32

export function normalizeApplicationCategory(value: string): string {
  const trimmed = value.trim()
  return trimmed === "" ? "OTHER" : trimmed
}

export function isPresetApplicationCategory(value: string): value is ApplicationCategoryPreset {
  return (APPLICATION_CATEGORY_PRESETS as readonly string[]).includes(value)
}

/**
 * 分类展示文案：内置分类翻译成中文/英文，自定义分类原样显示。
 * 用显式判断而不是拼接 i18n key，避免自定义值渲染出缺失的 key。
 */
export function applicationCategoryLabel(t: TFunction, value: string): string {
  const category = normalizeApplicationCategory(value)
  return isPresetApplicationCategory(category)
    ? t(`appStore.categories.${category}`)
    : category
}
