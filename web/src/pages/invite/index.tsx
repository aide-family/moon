import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useNavigate, useParams } from "react-router-dom"

import { acceptInvitation, declineInvitation, getInvitation } from "@/api/invitation"
import type { MemberInviteItem } from "@/api/invitation/types"
import { Button } from "@/components/ui/button"
import { enumLabel } from "@/lib/enum-label"
import { failureText } from "@/lib/form-error"
import { AuthShell } from "@/pages/auth-shell"

export function InvitePage() {
  const { t } = useTranslation()
  const { token = "" } = useParams()
  const navigate = useNavigate()
  const [invitation, setInvitation] = useState<MemberInviteItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const loggedIn = Boolean(localStorage.getItem("token"))
  const next = `/invite/${token}`

  useEffect(() => {
    let active = true
    void getInvitation(token)
      .then((result) => {
        if (active) {
          setInvitation(result.invitation)
        }
      })
      .catch((err) => {
        if (active) {
          setError(failureText(err, t("state.failed")))
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [token, t])

  useEffect(() => {
    if (!loggedIn && invitation && invitation.status === "PENDING" && !invitation.registered) {
      navigate(
        `/register?next=${encodeURIComponent(next)}&email=${encodeURIComponent(invitation.email)}`,
        { replace: true },
      )
    }
  }, [invitation, loggedIn, navigate, next])

  async function onAccept() {
    setPending(true)
    setError("")
    try {
      const result = await acceptInvitation({ token })
      localStorage.setItem("namespace", result.member.namespaceUid)
      navigate("/members")
    } catch (err) {
      setError(failureText(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function onDecline() {
    setPending(true)
    setError("")
    try {
      await declineInvitation({ token })
      setInvitation((current) => (current ? { ...current, status: "DECLINED" } : current))
    } catch (err) {
      setError(failureText(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  return (
    <AuthShell title={t("invitation.title")} subtitle={t("invitation.description")}>
      {loading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!loading && error && !invitation ? <p className="text-sm text-destructive">{error}</p> : null}
      {!loading && invitation ? (
        <div className="grid gap-3 text-sm">
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">{t("invitation.namespace")}</span>
            <span className="font-medium">{invitation.namespaceName}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">{t("member.role")}</span>
            <span>{invitation.roleName}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">{t("invitation.inviter")}</span>
            <span>{invitation.inviterName}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">{t("member.expiresAt")}</span>
            <span>{new Date(invitation.expiresAt * 1000).toLocaleString()}</span>
          </div>
          <div className="flex justify-between gap-4">
            <span className="text-muted-foreground">{t("member.status")}</span>
            <span>{enumLabel(t, invitation.status)}</span>
          </div>
          {invitation.status !== "PENDING" ? (
            <p className="text-muted-foreground">{t("invitation.notPending")}</p>
          ) : loggedIn ? (
            <div className="flex gap-2">
              <Button className="flex-1" disabled={pending} onClick={() => void onAccept()}>
                {t("invitation.accept")}
              </Button>
              <Button
                className="flex-1"
                variant="outline"
                disabled={pending}
                onClick={() => void onDecline()}
              >
                {t("invitation.decline")}
              </Button>
            </div>
          ) : (
            <div className="grid gap-2">
              <Button asChild>
                <Link to={`/login?next=${encodeURIComponent(next)}`}>
                  {t("invitation.loginToAccept")}
                </Link>
              </Button>
              <Button asChild variant="outline">
                <Link
                  to={`/register?next=${encodeURIComponent(next)}&email=${encodeURIComponent(invitation.email)}`}
                >
                  {t("invitation.registerToAccept")}
                </Link>
              </Button>
            </div>
          )}
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
        </div>
      ) : null}
    </AuthShell>
  )
}
