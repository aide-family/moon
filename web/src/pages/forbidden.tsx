import { useTranslation } from "react-i18next"
import { Link } from "react-router-dom"

import error403Illustration from "@/assets/undraw/error-403.svg"
import { ErrorPage } from "@/components/error-page"
import { Button } from "@/components/ui/button"

export function ForbiddenPage() {
  const { t } = useTranslation()

  return (
    <ErrorPage
      status={403}
      label="Forbidden"
      title={t("error.forbiddenTitle")}
      description={t("error.forbiddenDescription")}
      image={error403Illustration}
    >
      <Button asChild variant="outline">
        <Link to="/profile">{t("nav.profile")}</Link>
      </Button>
    </ErrorPage>
  )
}
