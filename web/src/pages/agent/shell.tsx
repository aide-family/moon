import { useEffect, useRef, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { FitAddon } from "@xterm/addon-fit"
import { Terminal } from "@xterm/xterm"
import "@xterm/xterm/css/xterm.css"

import { issueShellTicket } from "@/api/agent"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { formMessage } from "@/lib/form-error"

export function AgentShellPage() {
  const { t } = useTranslation()
  const { uid = "" } = useParams()
  const host = useRef<HTMLDivElement>(null)
  const [error, setError] = useState("")

  useEffect(() => {
    const el = host.current
    if (!el || !uid) {
      return
    }
    let active = true
    let socket: WebSocket | null = null
    const term = new Terminal({
      cursorBlink: true,
      fontFamily: "JetBrains Mono, ui-monospace, monospace",
      fontSize: 14,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(el)
    fit.fit()

    function sendResize() {
      fit.fit()
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows }))
      }
    }

    term.onData((data) => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(new TextEncoder().encode(data))
      }
    })

    void issueShellTicket(uid)
      .then(({ ticket }) => {
        if (!active) {
          return
        }
        const proto = window.location.protocol === "https:" ? "wss" : "ws"
        socket = new WebSocket(`${proto}//${window.location.host}/v1/agent/${uid}/shell?ticket=${encodeURIComponent(ticket)}`)
        socket.binaryType = "arraybuffer"
        socket.onopen = () => sendResize()
        socket.onmessage = (event) => {
          if (typeof event.data === "string") {
            try {
              const payload = JSON.parse(event.data) as { type?: string; reason?: string }
              if (payload.type === "exit") {
                setError(payload.reason || t("agent.shellClosed"))
                return
              }
            } catch {
              term.write(event.data)
            }
            return
          }
          term.write(new Uint8Array(event.data as ArrayBuffer))
        }
        socket.onerror = () => {
          if (active) {
            setError(t("agent.shellFailed"))
          }
        }
      })
      .catch((err: unknown) => {
        if (active) {
          setError(formMessage(err, t("agent.shellFailed")))
        }
      })

    window.addEventListener("resize", sendResize)
    return () => {
      active = false
      window.removeEventListener("resize", sendResize)
      socket?.close()
      term.dispose()
    }
  }, [t, uid])

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("agent.shellTitle")} description={t("agent.shellDescription")}>
        <Button variant="outline" asChild>
          <Link to="/agents">{t("agent.back")}</Link>
        </Button>
      </PageHeader>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <div ref={host} className="min-h-96 flex-1 overflow-hidden rounded-xl bg-black p-2" />
    </section>
  )
}
