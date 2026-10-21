import { useEffect, useMemo, useState } from "react"
import { ArrowLeft, Check, CircleAlert, CircleCheck, CircleX, Server, Terminal } from "lucide-react"
import { useTranslation } from "react-i18next"
import { Link, useParams, useSearchParams } from "react-router-dom"

import { listAgents } from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
import { getSSHCommand, listSSHCommandAudits } from "@/api/ssh-command"
import type { SSHCommandAuditItem, SSHCommandItem } from "@/api/ssh-command/types"
import { Button } from "@/components/ui/button"
import { enumLabel } from "@/lib/enum-label"

function formatTimestamp(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

export function SSHCommandResultsPage() {
  const { t } = useTranslation()
  const { uid = "" } = useParams()
  const [searchParams] = useSearchParams()
  const auditKey = searchParams.getAll("audit").join("|")
  const isNewRun = searchParams.get("run") === "1"
  const [command, setCommand] = useState<SSHCommandItem | null>(null)
  const [audits, setAudits] = useState<SSHCommandAuditItem[]>([])
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [selectedAuditUid, setSelectedAuditUid] = useState("")
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [reloadKey, setReloadKey] = useState(0)

  useEffect(() => {
    let active = true
    const requestedAuditUids = auditKey ? auditKey.split("|") : []

    Promise.all([
      getSSHCommand(uid),
      listSSHCommandAudits({ page: 1, pageSize: 100, commandUid: uid }),
      listAgents({ page: 1, pageSize: 100 }).catch(() => ({ items: [] as AgentItem[] })),
    ])
      .then(([commandResult, auditResult, agentResult]) => {
        if (!active) return
        const executionAudits = (auditResult.items ?? []).filter((audit) => audit.kind === "EXECUTE")
        let currentRun: SSHCommandAuditItem[]
        if (requestedAuditUids.length > 0 || isNewRun) {
          const requested = new Set(requestedAuditUids)
          currentRun = executionAudits.filter((audit) => requested.has(audit.uid))
        } else {
          const latestByAgent = new Map<string, SSHCommandAuditItem>()
          for (const audit of executionAudits) {
            if (!latestByAgent.has(audit.agentUid)) latestByAgent.set(audit.agentUid, audit)
          }
          currentRun = Array.from(latestByAgent.values())
        }
        setCommand(commandResult.command)
        setAudits(currentRun)
        setAgents(agentResult.items ?? [])
        setSelectedAuditUid((current) => currentRun.some((audit) => audit.uid === current) ? current : currentRun[0]?.uid ?? "")
      })
      .catch(() => {
        if (!active) return
        setFailed(true)
      })
      .finally(() => {
        if (active) setLoading(false)
      })

    return () => {
      active = false
    }
  }, [uid, auditKey, isNewRun, reloadKey])

  const agentsByUid = useMemo(() => new Map(agents.map((agent) => [agent.uid, agent])), [agents])
  const selectedAudit = audits.find((audit) => audit.uid === selectedAuditUid) ?? audits[0] ?? null
  const failedCount = audits.filter((audit) => audit.status === "FAILED").length
  const machineName = (audit: SSHCommandAuditItem) => {
    const agent = agentsByUid.get(audit.agentUid)
    return agent?.name || agent?.hostname || agent?.localIp || audit.agentUid
  }

  return (
    <section className="-m-4 flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden md:-m-6">
      <header className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b bg-background px-4 py-3 md:px-6">
        <div className="flex min-w-0 items-center gap-3">
          <Button asChild variant="ghost" size="icon" aria-label={t("ssh.backToCommands")}>
            <Link to="/ssh-commands"><ArrowLeft /></Link>
          </Button>
          <div className="min-w-0">
            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{t("ssh.execResult")}</p>
            <h1 className="truncate font-heading text-lg font-semibold tracking-tight">{command?.name || t("ssh.title")}</h1>
          </div>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <span className="tabular-nums">{t("ssh.machineResults", { count: audits.length })}</span>
          {failedCount > 0 ? <span className="rounded-full bg-destructive/10 px-2 py-0.5 font-medium text-destructive">{t("ssh.failedCount", { count: failedCount })}</span> : null}
        </div>
      </header>

      <div className="flex min-h-0 min-w-0 flex-1 flex-col md:flex-row">
        <aside className="flex max-h-56 shrink-0 flex-col border-b bg-muted/20 md:max-h-none md:w-72 md:border-r md:border-b-0">
          <div className="flex items-center justify-between px-4 py-3 md:px-5">
            <div>
              <h2 className="text-sm font-semibold">{t("ssh.agents")}</h2>
              <p className="text-xs text-muted-foreground">{t("ssh.selectMachineHint")}</p>
            </div>
            <span className="rounded-md bg-muted px-2 py-1 text-xs font-medium tabular-nums">{audits.length}</span>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto px-2 pb-3">
            {audits.map((audit) => {
              const selected = selectedAudit?.uid === audit.uid
              const isFailed = audit.status === "FAILED"
              const AgentStatusIcon = isFailed ? CircleX : CircleCheck
              return (
                <button
                  key={audit.uid}
                  type="button"
                  aria-pressed={selected}
                  onClick={() => setSelectedAuditUid(audit.uid)}
                  className={`flex min-h-14 w-full items-center gap-3 rounded-md px-3 py-2 text-left transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring ${selected ? "bg-muted" : ""}`}
                >
                  <AgentStatusIcon aria-hidden="true" className={`size-4 shrink-0 ${isFailed ? "text-destructive" : "text-emerald-600 dark:text-emerald-400"}`} />
                  <span className="min-w-0 flex-1">
                    <span className={`block truncate text-sm font-medium ${isFailed ? "text-destructive" : "text-foreground"}`}>{machineName(audit)}</span>
                    <span className="block truncate text-xs text-muted-foreground">{agentsByUid.get(audit.agentUid)?.hostname || agentsByUid.get(audit.agentUid)?.localIp || audit.agentUid}</span>
                  </span>
                  <span className={`shrink-0 text-xs ${isFailed ? "text-destructive" : "text-muted-foreground"}`}>{enumLabel(t, audit.status)}</span>
                </button>
              )
            })}
            {!loading && !failed && audits.length === 0 ? <p className="px-3 py-5 text-sm text-muted-foreground">{t("ssh.noResults")}</p> : null}
          </div>
        </aside>

        <main className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          {loading ? <p className="p-6 text-sm text-muted-foreground">{t("state.loading")}</p> : null}
          {!loading && failed ? (
            <div className="m-auto grid justify-items-center gap-3 p-6 text-center">
              <CircleAlert className="size-8 text-destructive" aria-hidden="true" />
              <p className="text-sm text-muted-foreground">{t("state.failed")}</p>
              <Button
                variant="outline"
                onClick={() => {
                  setLoading(true)
                  setFailed(false)
                  setReloadKey((value) => value + 1)
                }}
              >
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && selectedAudit ? (
            <>
              <div className="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3 md:px-6">
                <div className="flex min-w-0 items-center gap-3">
                  <Server className="size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                  <div className="min-w-0">
                    <h2 className={`truncate text-base font-semibold ${selectedAudit.status === "FAILED" ? "text-destructive" : ""}`}>{machineName(selectedAudit)}</h2>
                    <p className="truncate text-xs text-muted-foreground">{formatTimestamp(selectedAudit.createdAt)}</p>
                  </div>
                </div>
                {selectedAudit.operatorUid ? (
                  <div className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                    {selectedAudit.operatorAvatar ? (
                      <img src={selectedAudit.operatorAvatar} alt="" className="size-6 rounded-full object-cover" />
                    ) : (
                      <span className="flex size-6 items-center justify-center rounded-full bg-muted font-medium text-foreground">
                        {(selectedAudit.operatorNickname || selectedAudit.operatorName || selectedAudit.operatorUid).slice(0, 1).toUpperCase()}
                      </span>
                    )}
                    <span className="truncate">
                      {t("ssh.operator")}: <span className="font-medium text-foreground">{selectedAudit.operatorNickname || selectedAudit.operatorName || selectedAudit.operatorUid}</span>
                      {selectedAudit.operatorNickname && selectedAudit.operatorName && selectedAudit.operatorNickname !== selectedAudit.operatorName ? (
                        <span className="ml-1">({selectedAudit.operatorName})</span>
                      ) : null}
                    </span>
                  </div>
                ) : null}
                <div className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium ${selectedAudit.status === "FAILED" ? "bg-destructive/10 text-destructive" : "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400"}`}>
                  {selectedAudit.status === "FAILED" ? <CircleX className="size-3.5" aria-hidden="true" /> : <Check className="size-3.5" aria-hidden="true" />}
                  {enumLabel(t, selectedAudit.status)}
                </div>
              </div>
              <div className="min-h-0 flex-1 space-y-5 overflow-y-auto p-4 md:p-6">
                <section className="grid gap-2">
                  <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("ssh.command")}</h3>
                  <div className="flex items-start gap-3 rounded-md border bg-muted/30 p-3">
                    <Terminal className="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden="true" />
                    <pre className="min-w-0 whitespace-pre-wrap break-words font-mono text-sm leading-6">{selectedAudit.command || command?.command || "—"}</pre>
                  </div>
                </section>

                <section className="grid gap-2">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("ssh.output")}</h3>
                    <span className="font-mono text-xs text-muted-foreground">{t("ssh.exitCode")}: <span className={selectedAudit.exitCode === 0 ? "text-foreground" : "text-destructive"}>{selectedAudit.exitCode}</span></span>
                  </div>
                  <pre className="min-h-36 overflow-auto rounded-md border bg-muted/30 p-4 font-mono text-xs leading-6 whitespace-pre-wrap break-words">{selectedAudit.output || selectedAudit.detail || t("ssh.noOutput")}</pre>
                  {selectedAudit.truncated ? <p className="text-xs text-muted-foreground">{t("ssh.outputTruncated")}</p> : null}
                </section>
                {selectedAudit.detail && selectedAudit.output ? (
                  <section className="grid gap-2">
                    <h3 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{t("ssh.errorDetail")}</h3>
                    <p className="rounded-md border border-destructive/20 bg-destructive/5 p-3 text-sm text-destructive">{selectedAudit.detail}</p>
                  </section>
                ) : null}
              </div>
            </>
          ) : null}
          {!loading && !failed && !selectedAudit ? <p className="m-auto p-6 text-sm text-muted-foreground">{t("ssh.noResults")}</p> : null}
        </main>
      </div>
    </section>
  )
}

export default SSHCommandResultsPage
