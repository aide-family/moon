import { Languages, Moon, Sun, SunMoon } from "lucide-react"
import { useTranslation } from "react-i18next"

import { useTheme } from "@/components/theme-provider"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

export function LocaleThemeMenu() {
  const { t, i18n } = useTranslation()
  const { theme, setTheme } = useTheme()
  const language = i18n.language.startsWith("zh") ? "zh" : "en"
  const ThemeIcon = theme === "dark" ? Moon : theme === "system" ? SunMoon : Sun

  function setLanguage(next: string) {
    localStorage.setItem("lang", next)
    void i18n.changeLanguage(next)
  }

  function onTheme(value: string) {
    if (value === "light" || value === "dark" || value === "system") {
      setTheme(value)
    }
  }

  return (
    <div className="flex items-center gap-2">
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={t("lang.label")}>
            <Languages />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuRadioGroup value={language} onValueChange={setLanguage}>
            <DropdownMenuRadioItem value="zh">
              {t("lang.zh")}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="en">
              {t("lang.en")}
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label={t("theme.label")}>
            <ThemeIcon />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuRadioGroup value={theme} onValueChange={onTheme}>
            <DropdownMenuRadioItem value="light">
              <Sun />
              {t("theme.light")}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="dark">
              <Moon />
              {t("theme.dark")}
            </DropdownMenuRadioItem>
            <DropdownMenuRadioItem value="system">
              <SunMoon />
              {t("theme.system")}
            </DropdownMenuRadioItem>
          </DropdownMenuRadioGroup>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}
