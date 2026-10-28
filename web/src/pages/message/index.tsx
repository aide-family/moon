import {
  ArrowRight,
  Bell,
  CheckCheck,
  ChevronLeft,
  ChevronRight,
  Inbox,
  UserRoundCheck,
  UserRoundMinus,
  UserRoundPlus,
  UserRoundX,
} from "lucide-react"
import type { LucideIcon } from "lucide-react"
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
import { Button } from "@/components/ui/button"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { failureMessage } from "@/lib/form-error"

type MessageGroup = {
  key: string
  label: string
  items: SystemMessageItem[]
}

function groupByDay(
  items: SystemMessageItem[],
  locale: string,
  todayLabel: string,
  yesterdayLabel: string
): MessageGroup[] {
  const now = new Date()
  const today = new Date(
    now.getFullYear(),
    now.getMonth(),
    now.getDate()
  ).getTime()
  const yesterday = today - 24 * 60 * 60 * 1000
  const groups: MessageGroup[] = []

  for (const item of items) {
    const date = new Date(item.createdAt * 1000)
    const day = new Date(
      date.getFullYear(),
      date.getMonth(),
      date.getDate()
    ).getTime()
    const key = Number.isNaN(day) ? "unknown" : String(day)
    let label = date.toLocaleDateString(locale, {
      year: "numeric",
      month: "long",
      day: "numeric",
    })
    if (day === today) {
      label = todayLabel
    } else if (day === yesterday) {
      label = yesterdayLabel
    }

    let group = groups[groups.length - 1]
    if (!group || group.key !== key) {
      group = { key, label, items: [] }
      groups.push(group)
    }
    group.items.push(item)
  }

  return groups
}

function messageIcon(kind: string): LucideIcon {
  switch (kind) {
    case "INVITE_RECEIVED":
      return UserRoundPlus
    case "INVITE_ACCEPTED":
      return UserRoundCheck
    case "INVITE_DECLINED":
      return UserRoundX
    case "MEMBER_REMOVED":
      return UserRoundMinus
    default:
      return Bell
  }
}

function messageTone(kind: string): string {
  switch (kind) {
    case "INVITE_RECEIVED":
      return "bg-primary/10 text-primary-ink"
    case "INVITE_ACCEPTED":
      return "bg-success/10 text-success"
    case "INVITE_DECLINED":
    case "MEMBER_REMOVED":
      return "bg-muted text-muted-foreground"
    default:
      return "bg-muted text-muted-foreground"
  }
}

export function MessagePage() {
  const { t, i18n } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const [unreadOnly, setUnreadOnly] = useState(false)
  const [pendingUid, setPendingUid] = useState("")
  const [markingAll, setMarkingAll] = useState(false)
  const { items, total, loading, failed, error, reload } =
    usePagedList<SystemMessageItem>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: unreadOnly ? "unread" : "all",
      fetch: (page, pageSize) =>
        listSelfMessages({ page, pageSize, unreadOnly }),
    })
  const groups = groupByDay(
    items,
    i18n.language,
    t("message.today"),
    t("message.yesterday")
  )
  const pageCount = Math.max(1, Math.ceil(total / Math.max(query.pageSize, 1)))

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
      toast.error(failureMessage(err, t("state.failed")))
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
      toast.error(failureMessage(err, t("state.failed")))
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

  function selectFilter(nextUnreadOnly: boolean) {
    setUnreadOnly(nextUnreadOnly)
    if (query.page !== 1) {
      query.update({ page: 1 })
    }
  }

  return (
    <section className="flex min-h-0 min-w-0 flex-1 flex-col gap-5">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div className="min-w-0">
          <p className="mb-2 flex items-center gap-2 text-[11px] font-semibold tracking-[0.16em] text-muted-foreground uppercase">
            <span
              className="size-1.5 rounded-full bg-primary"
              aria-hidden="true"
            />
            {t("message.inboxLabel")}
          </p>
          <h1 className="font-heading text-2xl font-semibold tracking-tight text-balance">
            {t("message.title")}
          </h1>
          <p className="mt-1 max-w-2xl text-sm text-muted-foreground">
            {t("message.description")}
          </p>
        </div>
        <Button
          variant="outline"
          disabled={markingAll || session.unreadMessages === 0}
          onClick={() => void onMarkAll()}
          className="h-10 gap-2"
        >
          <CheckCheck aria-hidden="true" />
          {markingAll ? t("state.loading") : t("message.markAllRead")}
        </Button>
      </header>

      <div className="grid min-h-0 min-w-0 flex-1 gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
        <aside className="flex min-w-0 flex-col gap-3">
          <section className="relative overflow-hidden rounded-xl border border-primary/15 bg-primary/[0.035] p-5">
            <div className="relative z-10">
              <p className="text-xs font-medium text-muted-foreground">
                {t("message.unreadCount")}
              </p>
              <p className="mt-2 font-heading text-4xl leading-none font-semibold tracking-tight tabular-nums">
                {session.unreadMessages}
              </p>
              <p className="mt-3 text-xs leading-5 text-muted-foreground">
                {t("message.inboxHint")}
              </p>
            </div>
            <Inbox
              className="absolute -right-3 -bottom-4 size-24 rotate-[-10deg] text-primary-ink/[0.08]"
              aria-hidden="true"
            />
          </section>

          <nav
            aria-label={t("message.inboxLabel")}
            className="rounded-xl border border-border/60 bg-card/50 p-2"
          >
            <p className="px-3 pt-1 pb-2 text-[10px] font-semibold tracking-[0.14em] text-muted-foreground uppercase">
              {t("message.views")}
            </p>
            <button
              type="button"
              aria-pressed={!unreadOnly}
              onClick={() => selectFilter(false)}
              className={`flex min-h-11 w-full items-center gap-3 rounded-lg px-3 text-left text-sm transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none ${
                !unreadOnly
                  ? "bg-primary/10 font-medium text-foreground"
                  : "text-muted-foreground hover:bg-muted/70 hover:text-foreground"
              }`}
            >
              <Inbox
                className={`size-4 shrink-0 ${!unreadOnly ? "text-primary-ink" : ""}`}
                aria-hidden="true"
              />
              <span className="min-w-0 flex-1">{t("message.allMessages")}</span>
              <span className="font-mono text-xs tabular-nums">
                {unreadOnly ? "" : total}
              </span>
            </button>
            <button
              type="button"
              aria-pressed={unreadOnly}
              onClick={() => selectFilter(true)}
              className={`flex min-h-11 w-full items-center gap-3 rounded-lg px-3 text-left text-sm transition-colors focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none ${
                unreadOnly
                  ? "bg-primary/10 font-medium text-foreground"
                  : "text-muted-foreground hover:bg-muted/70 hover:text-foreground"
              }`}
            >
              <span
                className={`relative flex size-4 shrink-0 items-center justify-center ${unreadOnly ? "text-primary-ink" : ""}`}
                aria-hidden="true"
              >
                <span className="size-2 rounded-full bg-current" />
              </span>
              <span className="min-w-0 flex-1">{t("message.unreadOnly")}</span>
              {session.unreadMessages > 0 ? (
                <span className="rounded-full bg-primary/10 px-2 py-0.5 font-mono text-[10px] leading-4 text-foreground tabular-nums">
                  {session.unreadMessages > 99 ? "99+" : session.unreadMessages}
                </span>
              ) : null}
            </button>
          </nav>

          <div className="hidden rounded-xl border border-border/50 px-4 py-3 text-xs leading-5 text-muted-foreground lg:block">
            <span className="mb-1 block font-medium text-foreground">
              {t("message.activityTitle")}
            </span>
            {t("message.activityHint")}
          </div>
        </aside>

        <main className="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-xl border border-border/60 bg-card/65">
          <div className="flex min-h-14 flex-wrap items-center justify-between gap-3 border-b border-border/60 px-4 py-3 sm:px-5">
            <div className="flex min-w-0 items-baseline gap-2">
              <h2 className="truncate text-sm font-semibold">
                {t("message.activityTitle")}
              </h2>
              <span className="font-mono text-xs text-muted-foreground tabular-nums">
                {total}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <span className="hidden text-xs text-muted-foreground sm:inline">
                {t("message.latestFirst")}
              </span>
              <div
                role="group"
                className="flex items-center gap-1"
                aria-label={t("message.pagination", {
                  page: query.page,
                  pages: pageCount,
                })}
              >
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={t("action.prev")}
                  disabled={query.page <= 1 || loading}
                  onClick={() =>
                    query.update({ page: Math.max(1, query.page - 1) })
                  }
                  className="size-9"
                >
                  <ChevronLeft aria-hidden="true" />
                </Button>
                <span className="min-w-12 text-center font-mono text-xs text-muted-foreground tabular-nums">
                  {query.page} / {pageCount}
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={t("action.next")}
                  disabled={query.page * query.pageSize >= total || loading}
                  onClick={() => query.update({ page: query.page + 1 })}
                  className="size-9"
                >
                  <ChevronRight aria-hidden="true" />
                </Button>
              </div>
            </div>
          </div>

          <div
            className="min-h-0 flex-1 overflow-y-auto px-4 py-5 sm:px-6"
            aria-busy={loading}
          >
            {loading ? (
              <div className="grid gap-5" role="status">
                <span className="sr-only">{t("state.loading")}</span>
                {[0, 1, 2].map((item) => (
                  <div
                    key={item}
                    className="flex gap-4 motion-safe:animate-pulse"
                  >
                    <span className="size-9 shrink-0 rounded-full bg-muted" />
                    <span className="grid flex-1 gap-2 py-1">
                      <span className="h-3 w-1/3 rounded bg-muted" />
                      <span className="h-3 w-3/4 rounded bg-muted/70" />
                    </span>
                  </div>
                ))}
              </div>
            ) : null}

            {!loading && failed ? (
              <div className="flex min-h-48 flex-col items-center justify-center gap-3 text-center">
                <p className="max-w-md text-sm text-destructive">
                  {error || t("state.failed")}
                </p>
                <Button
                  variant="outline"
                  onClick={() => reload()}
                  className="h-10"
                >
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}

            {!loading && !failed && items.length === 0 ? (
              <div className="flex min-h-64 flex-col items-center justify-center px-6 text-center">
                <span className="mb-4 flex size-12 items-center justify-center rounded-2xl bg-muted/80 text-muted-foreground">
                  <Inbox className="size-5" aria-hidden="true" />
                </span>
                <h3 className="text-sm font-semibold">
                  {unreadOnly ? t("message.allCaughtUp") : t("message.empty")}
                </h3>
                <p className="mt-1 max-w-sm text-sm leading-6 text-muted-foreground">
                  {unreadOnly
                    ? t("message.allCaughtUpHint")
                    : t("message.emptyHint")}
                </p>
              </div>
            ) : null}

            {!loading && !failed && items.length > 0 ? (
              <div className="mx-auto max-w-3xl">
                {groups.map((group) => (
                  <section
                    key={group.key}
                    aria-label={group.label}
                    className="mb-7 last:mb-0"
                  >
                    <h3 className="mb-4 flex items-center gap-3 text-[11px] font-semibold tracking-[0.12em] text-muted-foreground uppercase">
                      <span>{group.label}</span>
                      <span
                        className="h-px flex-1 bg-border/60"
                        aria-hidden="true"
                      />
                    </h3>
                    <div className="relative ml-4 border-l border-border/70 pl-6 sm:ml-5 sm:pl-8">
                      <div className="grid gap-3">
                        {group.items.map((item) => {
                          const Icon = messageIcon(item.kind)
                          const timestamp = new Date(item.createdAt * 1000)
                          return (
                            <article
                              key={item.uid}
                              className={`relative rounded-xl p-4 transition-colors sm:p-5 ${
                                item.read
                                  ? "border border-border/50 bg-background/45 hover:bg-background/75"
                                  : "border border-primary/20 bg-primary/[0.045] hover:bg-primary/[0.07]"
                              }`}
                            >
                              <span
                                className={`absolute top-5 -left-[2.15rem] flex size-8 items-center justify-center rounded-full border-2 border-background sm:-left-[2.65rem] ${messageTone(item.kind)}`}
                                aria-hidden="true"
                              >
                                <Icon className="size-3.5" />
                              </span>
                              <div className="flex flex-wrap items-start justify-between gap-x-4 gap-y-3">
                                <div className="flex min-w-0 flex-1 items-start gap-3">
                                  <div className="min-w-0 flex-1">
                                    <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                                      <h4 className="text-sm leading-5 font-semibold">
                                        {title(item)}
                                      </h4>
                                      {!item.read ? (
                                        <span className="inline-flex items-center gap-1.5 rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary-ink">
                                          <span
                                            className="size-1.5 rounded-full bg-primary"
                                            aria-hidden="true"
                                          />
                                          {t("message.unread")}
                                        </span>
                                      ) : null}
                                    </div>
                                    <p className="mt-1.5 text-sm leading-6 text-muted-foreground">
                                      {body(item)}
                                    </p>
                                    <time
                                      className="mt-2 block font-mono text-[11px] text-muted-foreground tabular-nums"
                                      dateTime={timestamp.toISOString()}
                                    >
                                      {timestamp.toLocaleTimeString(
                                        i18n.language,
                                        { hour: "2-digit", minute: "2-digit" }
                                      )}
                                    </time>
                                  </div>
                                </div>
                                <div className="ml-11 flex shrink-0 flex-wrap items-center gap-2 sm:ml-0">
                                  {!item.read ? (
                                    <Button
                                      size="default"
                                      variant="outline"
                                      disabled={pendingUid === item.uid}
                                      onClick={() => void onMarkRead(item)}
                                      className="h-10"
                                    >
                                      {t("message.markRead")}
                                    </Button>
                                  ) : null}
                                  <Button
                                    size="default"
                                    variant="outline"
                                    onClick={() => openMessage(item)}
                                    className="h-10 gap-2 border-primary/25"
                                  >
                                    {item.kind === "INVITE_RECEIVED"
                                      ? t("message.viewInvitation")
                                      : item.kind === "MEMBER_REMOVED"
                                        ? t("message.viewNamespaces")
                                        : t("message.viewMembers")}
                                    <ArrowRight
                                      className="size-3.5"
                                      aria-hidden="true"
                                    />
                                  </Button>
                                </div>
                              </div>
                            </article>
                          )
                        })}
                      </div>
                    </div>
                  </section>
                ))}
              </div>
            ) : null}
          </div>
        </main>
      </div>
    </section>
  )
}
