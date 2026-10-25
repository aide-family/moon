import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router-dom"

import error404Illustration from "@/assets/undraw/error-404.svg"
import { ErrorPage } from "@/components/error-page"
import { Button } from "@/components/ui/button"

export function NotFoundPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  return (
    <ErrorPage
      status={404}
      label="Not Found"
      title={t("error.notFoundTitle")}
      description={t("error.notFoundDescription")}
      image={error404Illustration}
    >
      <Button asChild>
        <Link to="/overview">{t("nav.overview")}</Link>
      </Button>
      <Button variant="outline" onClick={() => navigate(-1)}>
        {t("action.goBack")}
      </Button>
    </ErrorPage>
  )
}
