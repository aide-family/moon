import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"

import {
  acceptInvitation,
  declineInvitation,
  listSelfInvitations,
} from "@/api/invitation"
import type { MemberInviteItem } from "@/api/invitation/types"
import { useSession } from "@/auth/session"
import { PageHeader } from "@/components/page-header"
import { enumLabel } from "@/lib/enum-label"
import { failureText } from "@/lib/form-error"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

export function InvitationPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const [items, setItems] = useState<MemberInviteItem[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [pendingUid, setPendingUid] = useState("")
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let active = true
    // eslint-disable-next-line react-hooks/set-state-in-effect -- reset the loading flag when the query changes
    setLoading(true)
    void listSelfInvitations()
      .then((result) => {
        if (active) {
          setFailed(false)
          setItems(result.items ?? [])
        }
      })
      .catch(() => {
        if (active) {
          setFailed(true)
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
  }, [tick])

  async function onAccept(item: MemberInviteItem) {
    setPendingUid(item.uid)
    try {
      const result = await acceptInvitation({ uid: item.uid })
      toast.success(t("invitation.accepted"))
      session.setNamespace(result.member.namespaceUid)
      navigate("/members")
    } catch (err) {
      toast.error(failureText(err, t("state.failed")))
    } finally {
      setPendingUid("")
    }
  }

  async function onDecline(item: MemberInviteItem) {
    setPendingUid(item.uid)
    try {
      await declineInvitation({ uid: item.uid })
      setTick((value) => value + 1)
    } catch (err) {
      toast.error(failureText(err, t("state.failed")))
    } finally {
      setPendingUid("")
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("invitation.title")}
        description={t("invitation.description")}
      />
      {loading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!loading && failed ? (
        <div className="flex items-center gap-2">
          <p className="text-sm text-destructive">{t("state.failed")}</p>
          <Button variant="outline" onClick={() => setTick((value) => value + 1)}>
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
      {!loading && !failed && items.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("invitation.empty")}</p>
      ) : null}
      {!loading && !failed && items.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("invitation.namespace")}</TableHead>
              <TableHead>{t("member.role")}</TableHead>
              <TableHead>{t("invitation.inviter")}</TableHead>
              <TableHead>{t("member.expiresAt")}</TableHead>
              <TableHead>{t("member.status")}</TableHead>
              <TableHead className="text-right">{t("action.actions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.uid}>
                <TableCell>{item.namespaceName}</TableCell>
                <TableCell>{item.roleName}</TableCell>
                <TableCell>{item.inviterName}</TableCell>
                <TableCell>
                  {new Date(item.expiresAt * 1000).toLocaleString()}
                </TableCell>
                <TableCell>{enumLabel(t, item.status)}</TableCell>
                <TableCell>
                  {item.status === "PENDING" ? (
                    <div className="flex justify-end gap-2">
                      <Button
                        size="sm"
                        disabled={pendingUid === item.uid}
                        onClick={() => void onAccept(item)}
                      >
                        {t("invitation.accept")}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={pendingUid === item.uid}
                        onClick={() => void onDecline(item)}
                      >
                        {t("invitation.decline")}
                      </Button>
                    </div>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : null}
    </section>
  )
}
