import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router-dom"

import { register } from "@/api/auth"
import { getCaptcha } from "@/api/captcha"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"

export function RegisterPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [name, setName] = useState("")
  const [captchaId, setCaptchaId] = useState("")
  const [captchaImage, setCaptchaImage] = useState("")
  const [captchaCode, setCaptchaCode] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  useEffect(() => {
    let active = true
    async function load() {
      try {
        const captcha = await getCaptcha()
        if (!active) {
          return
        }
        setCaptchaId(captcha.id)
        setCaptchaImage(captcha.image)
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : t("state.failed"))
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [t])

  async function refreshCaptcha() {
    setError("")
    const captcha = await getCaptcha()
    setCaptchaId(captcha.id)
    setCaptchaImage(captcha.image)
    setCaptchaCode("")
  }

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const result = await register({
        email,
        password,
        name,
        captchaId,
        captchaCode,
      })
      localStorage.setItem("token", result.session.accessToken)
      localStorage.setItem("refreshToken", result.session.refreshToken)
      navigate("/namespaces")
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
      try {
        await refreshCaptcha()
      } catch {
        setError(t("state.failed"))
      }
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="flex min-h-svh items-center justify-center bg-muted/40 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>{t("register.title")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4">
          <form className="grid gap-4" onSubmit={onSubmit}>
            <FormField label={t("register.name")} htmlFor="name">
              <Input
                id="name"
                value={name}
                placeholder={t("register.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
              />
            </FormField>
            <FormField label={t("login.emailLabel")} htmlFor="email">
              <Input
                id="email"
                type="email"
                value={email}
                placeholder={t("login.emailPlaceholder")}
                onChange={(event) => setEmail(event.target.value)}
                required
              />
            </FormField>
            <FormField
              label={t("login.passwordLabel")}
              htmlFor="password"
              hint={t("login.passwordHint")}
            >
              <Input
                id="password"
                type="password"
                minLength={8}
                value={password}
                placeholder={t("login.passwordPlaceholder")}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
            </FormField>
            <FormField
              label={t("register.captcha")}
              htmlFor="captcha"
              hint={t("register.captchaHint")}
            >
              <button
                type="button"
                onClick={() => void refreshCaptcha()}
                className="block"
              >
                {captchaImage ? (
                  <img
                    src={captchaImage}
                    alt={t("register.captcha")}
                    className="h-12 rounded border"
                  />
                ) : (
                  t("state.loading")
                )}
              </button>
              <Input
                id="captcha"
                value={captchaCode}
                placeholder={t("register.captchaPlaceholder")}
                onChange={(event) => setCaptchaCode(event.target.value)}
                required
              />
            </FormField>
            <Button className="w-full" type="submit" disabled={pending}>
              {t("action.register")}
            </Button>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
          </form>
          <p className="text-sm">
            <Link
              className="text-primary underline-offset-4 hover:underline"
              to="/login"
            >
              {t("register.hasAccount")}
            </Link>
          </p>
        </CardContent>
      </Card>
    </main>
  )
}
