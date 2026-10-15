import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router-dom"

import { emailLogin, passwordLogin, sendEmailLoginCode } from "@/api/auth"
import type { LoginResult } from "@/api/auth/types"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

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
  const [pending, setPending] = useState(false)

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const result = mode === "password" ? await passwordLogin({ email, password }) : await emailLogin({ email, code })
      storeSession(result)
      navigate("/namespaces")
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onSendCode() {
    setPending(true)
    setError("")
    try {
      await sendEmailLoginCode(email)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-md items-center p-6">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>{t("login.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mb-4 flex gap-2">
            <Button type="button" variant={mode === "password" ? "default" : "outline"} onClick={() => setMode("password")}>
              {t("login.password")}
            </Button>
            <Button type="button" variant={mode === "email" ? "default" : "outline"} onClick={() => setMode("email")}>
              {t("login.email")}
            </Button>
          </div>
          <form className="space-y-4" onSubmit={onSubmit}>
            <div className="space-y-2">
              <Label htmlFor="email">{t("login.emailLabel")}</Label>
              <Input id="email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
            </div>
            {mode === "password" ? (
              <div className="space-y-2">
                <Label htmlFor="password">{t("login.passwordLabel")}</Label>
                <Input id="password" type="password" value={password} onChange={(event) => setPassword(event.target.value)} required />
              </div>
            ) : (
              <div className="space-y-2">
                <Label htmlFor="code">{t("login.codeLabel")}</Label>
                <div className="flex gap-2">
                  <Input id="code" value={code} onChange={(event) => setCode(event.target.value)} required />
                  <Button type="button" variant="outline" disabled={pending || !email} onClick={() => void onSendCode()}>
                    {t("action.sendCode")}
                  </Button>
                </div>
              </div>
            )}
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <Button type="submit" disabled={pending}>
              {t("action.login")}
            </Button>
          </form>
          <p className="mt-4 text-sm text-muted-foreground">{t("login.oauthSkipped")}</p>
          <p className="mt-2 text-sm">
            <Link to="/register">{t("login.noAccount")}</Link>
          </p>
        </CardContent>
      </Card>
    </main>
  )
}
