import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { changeRemark, getSelfInfo } from "@/api/self"
import type { UserItem } from "@/api/user/types"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

export function ProfilePage() {
  const { t } = useTranslation()
  const [user, setUser] = useState<UserItem | null>(null)
  const [remark, setRemark] = useState("")
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [error, setError] = useState("")
  const [saved, setSaved] = useState(false)
  const [pending, setPending] = useState(false)

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
  }, [])

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

  if (loading) {
    return <p>{t("state.loading")}</p>
  }
  if (failed || !user) {
    return <p>{t("state.failed")}</p>
  }

  return (
    <Card className="max-w-xl">
      <CardHeader>
        <CardTitle>{t("profile.title")}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="mb-4 text-sm text-muted-foreground">{user.email}</p>
        <form className="space-y-4" onSubmit={onSubmit}>
          <div className="space-y-2">
            <Label htmlFor="remark">{t("profile.remark")}</Label>
            <Input id="remark" value={remark} onChange={(event) => setRemark(event.target.value)} />
          </div>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          {saved ? <p className="text-sm">{t("profile.saved")}</p> : null}
          <Button type="submit" disabled={pending}>
            {t("action.save")}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
