import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate } from "react-router-dom"

import { register } from "@/api/auth"
import { getCaptcha } from "@/api/captcha"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

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
      const result = await register({ email, password, name, captchaId, captchaCode })
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
    <main className="mx-auto flex min-h-screen max-w-md items-center p-6">
      <Card className="w-full">
        <CardHeader>
          <CardTitle>{t("register.title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={onSubmit}>
            <div className="space-y-2">
              <Label htmlFor="name">{t("register.name")}</Label>
              <Input id="name" value={name} onChange={(event) => setName(event.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">{t("login.emailLabel")}</Label>
              <Input id="email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">{t("login.passwordLabel")}</Label>
              <Input id="password" type="password" minLength={8} value={password} onChange={(event) => setPassword(event.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="captcha">{t("register.captcha")}</Label>
              <button type="button" onClick={() => void refreshCaptcha()} className="block">
                {captchaImage ? <img src={captchaImage} alt={t("register.captcha")} className="h-12 rounded border" /> : t("state.loading")}
              </button>
              <Input id="captcha" value={captchaCode} onChange={(event) => setCaptchaCode(event.target.value)} required />
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <Button type="submit" disabled={pending}>
              {t("action.register")}
            </Button>
          </form>
          <p className="mt-4 text-sm">
            <Link to="/login">{t("register.hasAccount")}</Link>
          </p>
        </CardContent>
      </Card>
    </main>
  )
}
