import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { changeRemark, getSelfInfo } from "@/api/self"
import type { UserItem } from "@/api/user/types"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"

export function ProfilePage() {
  const { t } = useTranslation()
  const [user, setUser] = useState<UserItem | null>(null)
  const [remark, setRemark] = useState("")
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState(false)
  const [pending, setPending] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const info = await getSelfInfo()
        if (!active) {
          return
        }
        setUser(info.user)
        setRemark(info.user.remark)
      } catch {
        if (active) {
          setFailed(true)
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [attempt])

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    setSaved(false)
    try {
      const info = await changeRemark(remark)
      setUser(info.user)
      setSaved(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="flex flex-col gap-6">
      <PageHeader
        title={t("profile.title")}
        description={t("profile.description")}
      />
      {loading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!loading && (failed || !user) ? (
        <div className="flex items-center gap-2">
          <p className="text-sm text-destructive">{t("state.failed")}</p>
          <Button
            variant="outline"
            onClick={() => setAttempt((value) => value + 1)}
          >
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
      {!loading && user ? (
        <Card className="max-w-xl">
          <CardHeader>
            <CardDescription>{user.email}</CardDescription>
          </CardHeader>
          <CardContent>
            <form className="grid gap-4" onSubmit={onSubmit}>
              <FormField label={t("profile.remark")} htmlFor="remark">
                <Input
                  id="remark"
                  value={remark}
                  placeholder={t("profile.remarkPlaceholder")}
                  onChange={(event) => setRemark(event.target.value)}
                />
              </FormField>
              {error ? (
                <p className="text-sm text-destructive">{error}</p>
              ) : null}
              {saved ? (
                <p className="text-sm text-muted-foreground">
                  {t("profile.saved")}
                </p>
              ) : null}
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </form>
          </CardContent>
        </Card>
      ) : null}
    </section>
  )
}
