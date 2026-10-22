import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { changePhone, changeRemark, getSelfInfo } from "@/api/self"
import type { UserItem } from "@/api/user/types"
import { useSession } from "@/auth/session"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import { StatusBadge } from "@/components/status-badge"
import { Skeleton } from "@/components/ui/skeleton"
import { failureMessage, readFormError } from "@/lib/form-error"

// 备注上限以后端 ChangeRemark 的 buf.validate 为准。
const remarkLimit = 200

export function ProfilePage() {
  const { t } = useTranslation()
  const session = useSession()
  const [user, setUser] = useState<UserItem | null>(null)
  const [remark, setRemark] = useState("")
  const [phone, setPhone] = useState("")
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState("")
  const [saved, setSaved] = useState(false)
  const [pending, setPending] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setLoadError("")
      try {
        // session 里已有同一次 self/info 的身份信息，这里再取一次只为了最新的备注。
        const info = await getSelfInfo()
        if (!active) {
          return
        }
        setUser(info.user)
        setRemark(info.user.remark)
        setPhone(info.user.phone ?? "")
        setSaved(false)
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureMessage(err, ""))
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

  const profile = user ?? session.user
  const phoneDirty = !!user && phone !== (user.phone ?? "")
  const dirty = !!user && (remark !== user.remark || phoneDirty)
  const displayName =
    profile?.nickname || profile?.name || profile?.email || "—"

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault()
    if (pending || !dirty) {
      return
    }
    setPending(true)
    setFieldErrors({})
    setFormError("")
    setSaved(false)
    try {
      let info = await getSelfInfo()
      if (phoneDirty) {
        info = await changePhone(phone.trim())
      }
      if (remark !== info.user.remark) {
        info = await changeRemark(remark.trim())
      }
      setUser(info.user)
      setRemark(info.user.remark)
      setPhone(info.user.phone ?? "")
      setSaved(true)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setFormError(failure.form)
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("profile.title")}
        description={t("profile.description")}
      />
      <div className="min-h-0 flex-1 overflow-y-auto">
        <Card className="glass-card max-w-2xl">
          <CardHeader>
            <CardTitle>{displayName}</CardTitle>
            <CardDescription className="break-all">
              {profile?.email || "—"}
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-6">
            <dl className="grid gap-4 sm:grid-cols-2">
              <div className="grid gap-1">
                <dt className="text-sm text-muted-foreground">
                  {t("user.systemRole")}
                </dt>
                <dd><StatusBadge value={profile?.systemRole ?? ""} /></dd>
              </div>
              <div className="grid gap-1">
                <dt className="text-sm text-muted-foreground">
                  {t("profile.status")}
                </dt>
                <dd><StatusBadge value={profile?.status ?? ""} /></dd>
              </div>
            </dl>
            <Separator />
            {loading ? (
              <div className="grid gap-2">
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-8 w-full" />
              </div>
            ) : failed ? (
              <div className="grid gap-3">
                <p className="text-sm text-destructive">
                  {loadError || t("state.failed")}
                </p>
                <Button
                  variant="outline"
                  className="justify-self-start"
                  onClick={() => setAttempt((value) => value + 1)}
                >
                  {t("action.retry")}
                </Button>
              </div>
            ) : (
              <form
                id="profile-form"
                className="grid gap-4"
                noValidate
                onSubmit={onSubmit}
              >
                <FormField
                  label={t("profile.phone")}
                  htmlFor="profile-phone"
                  hint={t("profile.phoneHint")}
                  error={fieldErrors.phone}
                >
                  <Input
                    id="profile-phone"
                    value={phone}
                    placeholder={t("profile.phonePlaceholder")}
                    aria-invalid={fieldErrors.phone ? true : undefined}
                    onChange={(event) => {
                      setPhone(event.target.value)
                      setSaved(false)
                      setFormError("")
                      setFieldErrors({})
                    }}
                  />
                </FormField>
                <FormField
                  label={t("profile.remark")}
                  htmlFor="profile-remark"
                  hint={t("profile.remarkHint")}
                  error={fieldErrors.remark}
                >
                  <Input
                    id="profile-remark"
                    value={remark}
                    maxLength={remarkLimit}
                    placeholder={t("profile.remarkPlaceholder")}
                    aria-invalid={fieldErrors.remark ? true : undefined}
                    onChange={(event) => {
                      setRemark(event.target.value)
                      setSaved(false)
                      setFormError("")
                      setFieldErrors({})
                    }}
                  />
                </FormField>
                {formError ? (
                  <p role="alert" className="text-sm text-destructive">
                    {formError}
                  </p>
                ) : null}
              </form>
            )}
          </CardContent>
          {!loading && !failed ? (
            <CardFooter className="justify-end gap-2">
              <p
                role="status"
                aria-live="polite"
                className="mr-auto text-sm text-muted-foreground"
              >
                {saved ? t("profile.saved") : null}
              </p>
              <Button
                type="submit"
                form="profile-form"
                aria-busy={pending || undefined}
                disabled={pending || !dirty}
              >
                {pending ? t("action.saving") : t("action.save")}
              </Button>
            </CardFooter>
          ) : null}
        </Card>
      </div>
    </section>
  )
}
