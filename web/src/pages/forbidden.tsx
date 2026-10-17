import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import { Button } from "@/components/ui/button"

export function ForbiddenPage() {
  const { t } = useTranslation()

  return (
    <section className="flex min-h-0 flex-1 flex-col items-start justify-center gap-4">
      <h1 className="font-heading text-2xl font-semibold tracking-tight">{t("error.forbiddenTitle")}</h1>
      <p className="text-sm text-muted-foreground">{t("error.forbiddenDescription")}</p>
      <Button asChild variant="outline">
        <Link to="/profile">{t("nav.profile")}</Link>
      </Button>
    </section>
  )
}
