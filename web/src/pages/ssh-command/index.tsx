import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { failureText, formMessage } from "@/lib/form-error"

import { listAgents } from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
import {
  approveSSHCommand,
  createSSHCommand,
  deleteSSHCommand,
  executeSSHCommand,
  listSSHCommandAudits,
  listSSHCommands,
  rejectSSHCommand,
  updateSSHCommand,
} from "@/api/ssh-command"
import type { SSHCommandAuditItem, SSHCommandItem } from "@/api/ssh-command/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Textarea } from "@/components/ui/textarea"
import { firstError, requireTrimmed } from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { useListQuery } from "@/lib/list-query"

export function SSHCommandPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<SSHCommandItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<SSHCommandItem | null>(null)
  const [name, setName] = useState("")
  const [command, setCommand] = useState("")
  const [workdir, setWorkdir] = useState("")
  const [env, setEnv] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<SSHCommandItem | null>(null)
  const [execOpen, setExecOpen] = useState(false)
  const [executing, setExecuting] = useState<SSHCommandItem | null>(null)
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [agentUids, setAgentUids] = useState<string[]>([])
  const [audits, setAudits] = useState<SSHCommandAuditItem[]>([])
  const [auditOpen, setAuditOpen] = useState(false)
  const [lastAudits, setLastAudits] = useState<SSHCommandAuditItem[]>([])
  const [confirm, setConfirm] = useState<SSHCommandItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canWrite = session.can("ssh:write")
  const canExec = session.can("ssh:exec")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listSSHCommands({ page: query.page, pageSize: query.pageSize, name: query.name })
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureText(err, ""))
          setItems([])
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
  }, [query.page, query.pageSize, query.name, attempt, session.namespaceUid])

  function openCreate() {
    setEditing(null)
    setName("")
    setCommand("")
    setWorkdir("")
    setEnv("")
    setRemark("")
    setError("")
    setOpen(true)
  }

  function openEdit(item: SSHCommandItem) {
    setEditing(item)
    setName(item.name)
    setCommand(item.command)
    setWorkdir(item.workdir)
    setEnv(item.env)
    setRemark(item.remark)
    setError("")
    setOpen(true)
  }

  async function openExec(item: SSHCommandItem) {
    setExecuting(item)
    setAgentUids([])
    setError("")
    try {
      const result = await listAgents({ page: 1, pageSize: 100 })
      setAgents((result.items ?? []).filter((agent) => agent.online && agent.status === "ACTIVE"))
    } catch {
      setAgents([])
    }
    setExecOpen(true)
  }

  async function openAudits(item: SSHCommandItem) {
    setDetail(item)
    try {
      const result = await listSSHCommandAudits({ page: 1, pageSize: 20, commandUid: item.uid })
      setAudits(result.items ?? [])
    } catch {
      setAudits([])
    }
    setAuditOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const invalid = firstError(requireTrimmed(name, t("form.required")), requireTrimmed(command, t("form.required")))
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      const body = { name: name.trim(), command: command.trim(), workdir, env, remark }
      if (editing) {
        await updateSSHCommand({ ...body, uid: editing.uid })
      } else {
        await createSSHCommand(body)
      }
      setOpen(false)
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function onExecute(event: React.FormEvent) {
    event.preventDefault()
    if (!executing) {
      return
    }
    if (agentUids.length === 0) {
      setError(t("form.required"))
      return
    }
    setPending(true)
    setError("")
    try {
      const result = await executeSSHCommand(executing.uid, agentUids)
      setExecOpen(false)
      setLastAudits(result.audits ?? [])
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await deleteSSHCommand(confirm.uid)
      setConfirm(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: SSHCommandItem): RowAction[] {
    const list: RowAction[] = [{ key: "audits", label: t("ssh.audits"), onSelect: () => void openAudits(item) }]
    if (canWrite) {
      list.push({ key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) })
      if (item.reviewStatus === "APPROVED") {
        list.push({ key: "reject", label: t("ssh.reject"), destructive: true, onSelect: () => void rejectSSHCommand(item.uid).then(() => setAttempt((value) => value + 1)) })
      } else {
        list.push({ key: "approve", label: t("ssh.approve"), onSelect: () => void approveSSHCommand(item.uid).then(() => setAttempt((value) => value + 1)) })
      }
      list.push({ key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm(item) })
    }
    if (canExec && item.reviewStatus === "APPROVED") {
      list.push({ key: "exec", label: t("ssh.execute"), onSelect: () => void openExec(item) })
    }
    return list
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("ssh.title")} description={t("ssh.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("ssh.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("ssh.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("ssh.name")}</TableHead>
                <TableHead>{t("ssh.command")}</TableHead>
                <TableHead>{t("ssh.reviewStatus")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell className="max-w-md truncate font-mono text-xs">{item.command}</TableCell>
                  <TableCell>{enumLabel(t, item.reviewStatus)}</TableCell>
                  <TableCell>
                    <RowActions detailLabel={t("action.detail")} moreLabel={t("action.more")} onDetail={() => setDetail(item)} actions={actions(item)} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField label={t("ssh.name")} htmlFor="ssh-name">
              <Input id="ssh-name" value={name} placeholder={t("ssh.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField label={t("ssh.workdir")} htmlFor="ssh-workdir">
              <Input id="ssh-workdir" value={workdir} placeholder={t("ssh.workdirPlaceholder")} onChange={(event) => setWorkdir(event.target.value)} />
            </FormField>
            <FormField label={t("ssh.command")} htmlFor="ssh-command" className="sm:col-span-full">
              <Textarea id="ssh-command" value={command} placeholder={t("ssh.commandPlaceholder")} onChange={(event) => setCommand(event.target.value)} required />
            </FormField>
            <FormField label={t("ssh.env")} htmlFor="ssh-env" className="sm:col-span-full">
              <Textarea id="ssh-env" value={env} onChange={(event) => setEnv(event.target.value)} placeholder={t("ssh.envPlaceholder")} />
            </FormField>
            <FormField label={t("ssh.remark")} htmlFor="ssh-remark" className="sm:col-span-full">
              <Textarea id="ssh-remark" value={remark} placeholder={t("ssh.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
            </FormField>
            {error ? <p className="sm:col-span-full text-sm text-destructive">{error}</p> : null}
            <DialogFooter className="sm:col-span-full">
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={execOpen} onOpenChange={setExecOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("ssh.execute")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onExecute}>
            <p className="text-sm text-muted-foreground">{t("ssh.pickAgents")}</p>
            <div className="grid max-h-64 gap-2 overflow-auto">
              {agents.map((agent) => {
                const checked = agentUids.includes(agent.uid)
                return (
                  <label key={agent.uid} className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={checked}
                      onCheckedChange={(value) => {
                        setAgentUids((current) =>
                          value === true ? [...current, agent.uid] : current.filter((uid) => uid !== agent.uid),
                        )
                      }}
                    />
                    <span>{agent.name} ({agent.hostname || agent.localIp || agent.uid})</span>
                  </label>
                )
              })}
            </div>
            {agents.length === 0 ? <p className="text-sm text-muted-foreground">{t("ssh.noOnlineAgent")}</p> : null}
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setExecOpen(false)}>
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending || agentUids.length === 0}>
                {t("ssh.execute")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={lastAudits.length > 0} onOpenChange={(value) => !value && setLastAudits([])}>
        <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("ssh.execResult")}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-3 text-sm">
            {lastAudits.map((audit) => (
              <div key={audit.uid || audit.agentUid} className="grid gap-1 rounded-md border p-2">
                <p>{audit.agentUid}</p>
                <p>{t("ssh.exitCode")}: {audit.exitCode}</p>
                <p>{t("ssh.status")}: {enumLabel(t, audit.status)}</p>
                <pre className="max-h-40 overflow-auto rounded-md border p-2 text-xs">{audit.output || audit.detail || "-"}</pre>
              </div>
            ))}
          </div>
        </DialogContent>
      </Dialog>
      <Dialog open={auditOpen} onOpenChange={setAuditOpen}>
        <DialogContent className="max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{detail?.name ?? t("ssh.audits")}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-3 text-sm">
            {audits.length === 0 ? <p className="text-muted-foreground">{t("state.empty")}</p> : null}
            {audits.map((audit) => (
              <div key={audit.uid} className="rounded-md border p-2">
                <p>{enumLabel(t, audit.kind)} / {enumLabel(t, audit.status)}</p>
                <p>{audit.createdAt}</p>
                {audit.kind === "EXECUTE" ? <p>{t("ssh.exitCode")}: {audit.exitCode}</p> : null}
                {audit.output ? <pre className="mt-1 max-h-32 overflow-auto text-xs">{audit.output}</pre> : null}
              </div>
            ))}
          </div>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detail !== null && !auditOpen}
        title={detail?.name ?? ""}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={
          detail
            ? [
                { label: t("detail.uid"), value: detail.uid },
                { label: t("ssh.command"), value: detail.command },
                { label: t("ssh.workdir"), value: detail.workdir || "-" },
                { label: t("ssh.env"), value: detail.env || "-" },
                { label: t("ssh.reviewStatus"), value: enumLabel(t, detail.reviewStatus) },
                { label: t("ssh.remark"), value: detail.remark || "-" },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={t("action.delete")}
        description={t("ssh.confirmDelete")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void runConfirm()}
        onOpenChange={(value) => {
          if (!value) {
            setConfirm(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
