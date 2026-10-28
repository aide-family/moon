import { Languages } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

/**
 * 语言切换。主题/品牌已经是独立的 AppearanceMenu，
 * 这里只留语言，避免同一件事有两个入口。
 */
export function LocaleMenu() {
  const { t, i18n } = useTranslation()
  const language = i18n.language.startsWith("zh") ? "zh" : "en"

  function setLanguage(next: string) {
    localStorage.setItem("lang", next)
    void i18n.changeLanguage(next)
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label={t("lang.label")}>
          <Languages />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuRadioGroup value={language} onValueChange={setLanguage}>
          <DropdownMenuRadioItem value="zh">{t("lang.zh")}</DropdownMenuRadioItem>
          <DropdownMenuRadioItem value="en">{t("lang.en")}</DropdownMenuRadioItem>
        </DropdownMenuRadioGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
