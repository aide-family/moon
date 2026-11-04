import { useTranslation } from "react-i18next"

/** 公开官网的请求语言：跟随 LocaleMenu 的当前语言，只取后端支持的 zh/en。 */
export function useSiteLocale(): "zh" | "en" {
  const { i18n } = useTranslation()
  return i18n.language.startsWith("zh") ? "zh" : "en"
}
