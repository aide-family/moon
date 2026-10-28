import * as React from "react"

import {
  DEFAULT_BRAND,
  DEFAULT_MODE,
  STORAGE_KEYS,
  isBrandId,
  isThemeMode,
  readStoredAppearance,
  systemMode,
  type ResolvedMode,
  type ThemeMode,
} from "@/lib/appearance"
import { type BrandId } from "@/lib/brand-palette.generated"

/** 兼容既有调用点的旧名字。 */
type Theme = ThemeMode
type ResolvedTheme = ResolvedMode

type ThemeProviderProps = {
  children: React.ReactNode
  /** 旧 API 保留：未存过偏好时的默认模式。 */
  defaultTheme?: Theme
  /** 旧 API 保留：主题存储键，默认 `theme`。 */
  storageKey?: string
  disableTransitionOnChange?: boolean
}

type ThemeProviderState = {
  theme: Theme
  resolvedTheme: ResolvedTheme
  setTheme: (theme: Theme) => void
  brand: BrandId
  setBrand: (brand: BrandId) => void
}

const COLOR_SCHEME_QUERY = "(prefers-color-scheme: dark)"
const THEME_COLOR_META = 'meta[name="theme-color"]'

const ThemeProviderContext = React.createContext<ThemeProviderState | undefined>(undefined)

/**
 * 把外观状态写到 <html>：类名、data-brand、原生 color-scheme、浏览器主题色。
 * 只改 `<html>` 上的属性，7 套品牌 × 2 套模式的差异全部由 CSS 变量承担，
 * 组件不需要知道当前是哪套主题。
 */
function syncDocument(brand: BrandId, resolved: ResolvedMode) {
  const root = document.documentElement

  root.classList.toggle("dark", resolved === "dark")
  root.classList.toggle("light", resolved === "light")
  root.dataset.brand = brand
  root.style.colorScheme = resolved
  // 引导脚本为了消白闪写过内联底色，挂载后交还给 token。
  root.style.removeProperty("background-color")

  // 等这一帧的颜色算完，再把浏览器 UI 用的 theme-color 对上页面底色。
  window.requestAnimationFrame(() => {
    const background = window.getComputedStyle(root).getPropertyValue("--background").trim()
    const meta = document.querySelector<HTMLMetaElement>(THEME_COLOR_META)
    if (meta && background) {
      meta.content = background
    }
  })
}

/** 切换这一帧禁用过渡，否则整屏颜色会各自用 150ms 追着变。 */
function withTransitionSuppressed(apply: () => void) {
  const root = document.documentElement
  root.classList.add("theme-switching")
  apply()
  // 读一次布局，确保新值已经落位，再恢复过渡。
  void root.offsetHeight
  window.requestAnimationFrame(() => {
    window.requestAnimationFrame(() => {
      root.classList.remove("theme-switching")
    })
  })
}

export function ThemeProvider({
  children,
  defaultTheme,
  storageKey = STORAGE_KEYS.mode,
  disableTransitionOnChange = true,
}: ThemeProviderProps) {
  const stored = React.useMemo(() => readStoredAppearance(), [])
  const [theme, setThemeState] = React.useState<ThemeMode>(stored.mode)
  const [brand, setBrandState] = React.useState<BrandId>(stored.brand)
  const [systemDark, setSystemDark] = React.useState(() => systemMode() === "dark")

  const resolvedTheme: ResolvedTheme = theme === "system" ? (systemDark ? "dark" : "light") : theme

  const apply = React.useCallback(
    (nextBrand: BrandId, nextResolved: ResolvedTheme) => {
      if (disableTransitionOnChange) {
        withTransitionSuppressed(() => syncDocument(nextBrand, nextResolved))
        return
      }

      syncDocument(nextBrand, nextResolved)
    },
    [disableTransitionOnChange]
  )

  // useLayoutEffect：在浏览器绘制前落类名，React 挂载不会再闪一次。
  React.useLayoutEffect(() => {
    apply(brand, resolvedTheme)
  }, [apply, brand, resolvedTheme])

  // 跟随系统时，系统外观一变就重算。
  React.useEffect(() => {
    const mediaQuery = window.matchMedia(COLOR_SCHEME_QUERY)
    const handleChange = () => setSystemDark(mediaQuery.matches)
    handleChange()
    mediaQuery.addEventListener("change", handleChange)
    return () => mediaQuery.removeEventListener("change", handleChange)
  }, [])

  const setTheme = React.useCallback(
    (next: ThemeMode) => {
      try {
        localStorage.setItem(storageKey, next)
      } catch {
        // 隐私模式：内存里生效即可。
      }
      setThemeState(next)
    },
    [storageKey]
  )

  const setBrand = React.useCallback((next: BrandId) => {
    try {
      localStorage.setItem(STORAGE_KEYS.brand, next)
    } catch {
      // 同上。
    }
    setBrandState(next)
  }, [])

  // 跨标签页同步：两个键都要听，否则 A 页换品牌 B 页不跟。
  React.useEffect(() => {
    const handleStorageChange = (event: StorageEvent) => {
      if (event.storageArea !== localStorage) {
        return
      }

      if (event.key === storageKey) {
        setThemeState(isThemeMode(event.newValue) ? event.newValue : (defaultTheme ?? DEFAULT_MODE))
        return
      }

      if (event.key === STORAGE_KEYS.brand) {
        setBrandState(isBrandId(event.newValue) ? event.newValue : DEFAULT_BRAND)
      }
    }

    window.addEventListener("storage", handleStorageChange)
    return () => window.removeEventListener("storage", handleStorageChange)
  }, [defaultTheme, storageKey])

  const value = React.useMemo(
    () => ({ theme, resolvedTheme, setTheme, brand, setBrand }),
    [brand, resolvedTheme, setBrand, setTheme, theme]
  )

  return <ThemeProviderContext.Provider value={value}>{children}</ThemeProviderContext.Provider>
}

export const useTheme = () => {
  const context = React.useContext(ThemeProviderContext)

  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider")
  }

  return context
}
