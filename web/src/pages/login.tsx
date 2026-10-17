import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router-dom"

import { emailLogin, passwordLogin, sendEmailLoginCode } from "@/api/auth"
import type { LoginResult } from "@/api/auth/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { readFormError } from "@/lib/form-error"
import { AuthShell } from "@/pages/auth-shell"

function storeSession(result: LoginResult) {
  localStorage.setItem("token", result.session.accessToken)
  localStorage.setItem("refreshToken", result.session.refreshToken)
}

export function LoginPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [mode, setMode] = useState<"password" | "email">("password")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const result =
        mode === "password"
          ? await passwordLogin({ email, password })
          : await emailLogin({ email, code })
      storeSession(result)
      navigate("/namespaces")
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onSendCode() {
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await sendEmailLoginCode(email)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <AuthShell title={t("login.welcome")} subtitle={t("login.welcomeSub")}>
          <div className="flex gap-2">
            <Button
              type="button"
              className="flex-1"
              variant={mode === "password" ? "default" : "outline"}
              onClick={() => setMode("password")}
            >
              {t("login.password")}
            </Button>
            <Button
              type="button"
              className="flex-1"
              variant={mode === "email" ? "default" : "outline"}
              onClick={() => setMode("email")}
            >
              {t("login.email")}
            </Button>
          </div>
          <form className="grid gap-4" onSubmit={onSubmit}>
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
            {mode === "password" ? (
              <FormField
                label={t("login.passwordLabel")}
                htmlFor="password"
                hint={t("login.passwordHint")}
              >
                <Input
                  id="password"
                  type="password"
                  value={password}
                  placeholder={t("login.passwordPlaceholder")}
                  onChange={(event) => setPassword(event.target.value)}
                  required
                />
              </FormField>
            ) : (
              <FormField
                label={t("login.codeLabel")}
                htmlFor="code"
                hint={t("login.codeHint")}
              >
                <div className="flex gap-2">
                  <Input
                    id="code"
                    value={code}
                    placeholder={t("login.codePlaceholder")}
                    onChange={(event) => setCode(event.target.value)}
                    required
                  />
                  <Button
                    type="button"
                    variant="outline"
                    disabled={pending || !email}
                    onClick={() => void onSendCode()}
                  >
                    {t("action.sendCode")}
                  </Button>
                </div>
              </FormField>
            )}
            <Button className="w-full" type="submit" disabled={pending}>
              {t("action.login")}
            </Button>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
          </form>
          <p className="text-sm text-muted-foreground">
            {t("login.oauthSkipped")}
          </p>
          <p className="text-sm">
            <Link
              className="text-primary underline-offset-4 hover:underline"
              to="/register"
            >
              {t("login.noAccount")}
            </Link>
          </p>
    </AuthShell>
  )
}
