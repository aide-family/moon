import { BRAND_IDS, type BrandId } from "@/lib/brand-palette.generated"

export type ThemeMode = "light" | "dark" | "system"
export type ResolvedMode = "light" | "dark"

/**
 * 沿用既有键名 `theme`，老用户的明暗偏好不会丢；`brand` 是新键。
 * index.html 里那段首帧引导脚本必须读同一组键名（它没法 import 本文件）。
 */
export const STORAGE_KEYS = { mode: "theme", brand: "brand" } as const
export const DEFAULT_MODE: ThemeMode = "system"
export const DEFAULT_BRAND: BrandId = "sunset"

export function isThemeMode(value: unknown): value is ThemeMode {
  return value === "light" || value === "dark" || value === "system"
}

export function isBrandId(value: unknown): value is BrandId {
  return typeof value === "string" && (BRAND_IDS as readonly string[]).includes(value)
}

/** 系统当前偏好；只在浏览器环境调用（本项目纯 CSR）。 */
export function systemMode(): ResolvedMode {
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
}

/** 读取已保存的外观；localStorage 不可用（隐私模式）时回落到默认值。 */
export function readStoredAppearance(): { mode: ThemeMode; brand: BrandId } {
  let mode: string | null = null
  let brand: string | null = null

  try {
    mode = localStorage.getItem(STORAGE_KEYS.mode)
    brand = localStorage.getItem(STORAGE_KEYS.brand)
  } catch {
    // 隐私模式下 localStorage 会抛错，忽略即可。
  }

  return {
    mode: isThemeMode(mode) ? mode : DEFAULT_MODE,
    brand: isBrandId(brand) ? brand : DEFAULT_BRAND,
  }
}
