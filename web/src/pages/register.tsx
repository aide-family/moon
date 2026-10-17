import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router-dom"

import { register } from "@/api/auth"
import { getCaptcha } from "@/api/captcha"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { readFormError } from "@/lib/form-error"
import { AuthShell } from "@/pages/auth-shell"

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
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    async function load() {
      try {
        const captcha = await getCaptcha(controller.signal)
        if (!active) {
          return
        }
        setCaptchaId(captcha.id)
        setCaptchaImage(captcha.image)
      } catch (err) {
        if (!active || controller.signal.aborted) {
          return
        }
        const failure = readFormError(err, t("state.failed"))
        setFieldErrors(failure.fields)
        setError(failure.form)
      }
    }
    void load()
    return () => {
      active = false
      controller.abort()
    }
    // Initial load only; refreshCaptcha handles user-triggered reloads.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- avoid refetch when i18n `t` identity changes
  }, [])

  async function refreshCaptcha() {
    try {
      const captcha = await getCaptcha()
      setCaptchaId(captcha.id)
      setCaptchaImage(captcha.image)
      setCaptchaCode("")
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setError(failure.form || t("state.failed"))
    }
  }

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    setFieldErrors({})
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
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
      await refreshCaptcha()
    } finally {
      setPending(false)
    }
  }

  return (
    <AuthShell title={t("register.welcome")} subtitle={t("register.welcomeSub")}>
          <form className="grid gap-4" onSubmit={onSubmit}>
            <FormField label={t("register.name")} htmlFor="name" error={fieldErrors.name}>
              <Input
                id="name"
                value={name}
                placeholder={t("register.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
              />
            </FormField>
            <FormField label={t("login.emailLabel")} htmlFor="email" error={fieldErrors.email}>
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
              error={fieldErrors.password}
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
              error={fieldErrors.captchaCode}
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
                    className="h-24 w-auto rounded border"
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
    </AuthShell>
  )
}
