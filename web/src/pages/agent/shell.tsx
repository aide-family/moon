import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, useParams } from "react-router-dom"

import { issueShellTicket } from "@/api/agent"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  XtermTerminal,
  type XtermSessionState,
} from "@/components/xterm-terminal"

export function AgentShellPage() {
  const { t } = useTranslation()
  const { uid = "" } = useParams()
  const [attempt, setAttempt] = useState(0)
  const [state, setState] = useState<XtermSessionState>({ state: "connecting" })

  const errorMessage =
    state.state === "failed"
      ? state.reason === "timeout"
        ? t("agent.shellTimeout")
        : state.message || t("agent.shellFailed")
      : state.state === "closed"
        ? state.reason || t("agent.shellClosed")
        : ""

  function reconnect() {
    setState({ state: "connecting" })
    setAttempt((value) => value + 1)
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("agent.shellTitle")}
        description={t("agent.shellDescription")}
      >
        <Button variant="outline" asChild>
          <Link to="/agents">{t("agent.back")}</Link>
        </Button>
      </PageHeader>
      {errorMessage ? (
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-sm text-destructive">{errorMessage}</p>
          <Button variant="outline" onClick={reconnect}>
            {t("agent.shellReconnect")}
          </Button>
        </div>
      ) : null}
      <XtermTerminal
        sessionKey={`${uid}|${attempt}`}
        issueTicket={async () => (await issueShellTicket(uid)).ticket}
        socketUrl={(ticket) => {
          const proto = window.location.protocol === "https:" ? "wss" : "ws"
          return `${proto}//${window.location.host}/v1/agent/${encodeURIComponent(uid)}/shell?ticket=${encodeURIComponent(ticket)}`
        }}
        onStateChange={setState}
        className="min-h-96 flex-1"
      />
    </section>
  )
}
