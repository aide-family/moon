import { Bell, CheckCheck } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { useNavigate } from "react-router-dom"
import { toast } from "sonner"

import {
  listSelfMessages,
  markAllMessagesRead,
  markMessageRead,
} from "@/api/message"
import type { SystemMessageItem } from "@/api/message/types"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"
import { failureText } from "@/lib/form-error"

export function MessagePage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const [unreadOnly, setUnreadOnly] = useState(false)
  const [pendingUid, setPendingUid] = useState("")
  const [markingAll, setMarkingAll] = useState(false)
  const { items, total, loading, failed, error, reload } = usePagedList<SystemMessageItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: unreadOnly ? "unread" : "all",
    fetch: (page, pageSize) => listSelfMessages({ page, pageSize, unreadOnly }),
  })

  function title(item: SystemMessageItem): string {
    switch (item.kind) {
      case "INVITE_RECEIVED":
        return t("message.inviteReceivedTitle")
      case "INVITE_ACCEPTED":
        return t("message.inviteAcceptedTitle")
      case "INVITE_DECLINED":
        return t("message.inviteDeclinedTitle")
      case "MEMBER_REMOVED":
        return t("message.memberRemovedTitle")
      default:
        return t("message.title")
    }
  }

  function body(item: SystemMessageItem): string {
    switch (item.kind) {
      case "INVITE_RECEIVED":
        return t("message.inviteReceivedBody", {
          actor: item.actorName,
          namespace: item.namespaceName,
          role: item.roleName,
        })
      case "INVITE_ACCEPTED":
        return t("message.inviteAcceptedBody", {
          actor: item.actorName,
          namespace: item.namespaceName,
        })
      case "INVITE_DECLINED":
        return t("message.inviteDeclinedBody", {
          actor: item.actorName,
          namespace: item.namespaceName,
        })
      case "MEMBER_REMOVED":
        return t("message.memberRemovedBody", {
          actor: item.actorName,
          namespace: item.namespaceName,
        })
      default:
        return ""
    }
  }

  async function onMarkRead(item: SystemMessageItem) {
    if (item.read) {
      return
    }
    setPendingUid(item.uid)
    try {
      await markMessageRead(item.uid)
      session.refreshUnreadMessages()
      reload()
    } catch (err) {
      toast.error(failureText(err, t("state.failed")))
    } finally {
      setPendingUid("")
    }
  }

  async function onMarkAll() {
    setMarkingAll(true)
    try {
      await markAllMessagesRead()
      session.refreshUnreadMessages()
      reload()
    } catch (err) {
      toast.error(failureText(err, t("state.failed")))
    } finally {
      setMarkingAll(false)
    }
  }

  function openMessage(item: SystemMessageItem) {
    void onMarkRead(item)
    if (item.kind === "INVITE_RECEIVED") {
      navigate("/invitations")
      return
    }
    if (item.kind === "MEMBER_REMOVED") {
      session.reload()
      navigate("/namespaces")
      return
    }
    if (item.namespaceUid) {
      session.setNamespace(item.namespaceUid)
    }
    navigate("/members")
  }

  return (
    <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-6">
      <PageHeader title={t("message.title")} description={t("message.description")}>
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          <div className="flex items-center gap-2">
            <Checkbox
              id="message-unread-only"
              checked={unreadOnly}
              onCheckedChange={(value) => {
                setUnreadOnly(value === true)
                if (query.page !== 1) {
                  query.update({ page: 1 })
                }
              }}
            />
            <Label htmlFor="message-unread-only">{t("message.unreadOnly")}</Label>
          </div>
          <Button
            variant="outline"
            disabled={markingAll || session.unreadMessages === 0}
            onClick={() => void onMarkAll()}
          >
            <CheckCheck />
            {t("message.markAllRead")}
          </Button>
        </div>
      </PageHeader>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? (
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{error || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("message.empty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <div className="grid gap-3 overflow-y-auto">
            {items.map((item) => (
              <article
                key={item.uid}
                className={
                  item.read
                    ? "rounded-lg border border-border/70 p-4"
                    : "rounded-lg border border-primary/30 bg-primary/5 p-4"
                }
              >
                <div className="flex items-start gap-3">
                  <span className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-muted">
                    <Bell className="size-4" />
                  </span>
                  <div className="min-w-0 flex-1">
                    <div className="flex shrink-0 flex-wrap items-center gap-2">
                      <h3 className="text-sm font-medium">{title(item)}</h3>
                      {item.read ? null : (
                        <span className="size-2 rounded-full bg-primary" aria-label={t("message.unread")} />
                      )}
                    </div>
                    <p className="mt-1 text-sm text-muted-foreground">{body(item)}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {new Date(item.createdAt * 1000).toLocaleString()}
                    </p>
                  </div>
                  <div className="flex shrink-0 flex-wrap justify-end gap-2">
                    {item.read ? null : (
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={pendingUid === item.uid}
                        onClick={() => void onMarkRead(item)}
                      >
                        {t("message.markRead")}
                      </Button>
                    )}
                    <Button size="sm" onClick={() => openMessage(item)}>
                      {item.kind === "INVITE_RECEIVED"
                        ? t("message.viewInvitation")
                        : item.kind === "MEMBER_REMOVED"
                          ? t("message.viewNamespaces")
                          : t("message.viewMembers")}
                    </Button>
                  </div>
                </div>
              </article>
            ))}
          </div>
        ) : null}
      </ListFrame>
    </section>
  )
}
