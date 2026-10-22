import { useState } from "react"
import { useTranslation } from "react-i18next"
import { createSearchParams, Link, useNavigate } from "react-router-dom"
import { formMessage, readFormError } from "@/lib/form-error"

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
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireTrimmed,
  type FieldErrors,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"

export function SSHCommandPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<SSHCommandItem | null>(null)
  const [name, setName] = useState("")
  const [command, setCommand] = useState("")
  const [workdir, setWorkdir] = useState("")
  const [env, setEnv] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<SSHCommandItem | null>(null)
  const [execOpen, setExecOpen] = useState(false)
  const [executing, setExecuting] = useState<SSHCommandItem | null>(null)
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [agentUids, setAgentUids] = useState<string[]>([])
  const [audits, setAudits] = useState<SSHCommandAuditItem[]>([])
  const [auditOpen, setAuditOpen] = useState(false)
  const [auditTarget, setAuditTarget] = useState<SSHCommandItem | null>(null)
  const [confirm, setConfirm] = useState<SSHCommandItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("ssh:write")
  const canExec = session.can("ssh:exec")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<SSHCommandItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listSSHCommands({ page, pageSize, name: query.name }),
  })

  function openCreate() {
    setEditing(null)
    setName("")
    setCommand("")
    setWorkdir("")
    setEnv("")
    setRemark("")
    setError("")
    setFieldErrors({})
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
    setFieldErrors({})
    setOpen(true)
  }

  async function openExec(item: SSHCommandItem) {
    setExecuting(item)
    setAgentUids([])
    setError("")
    setFieldErrors({})
    try {
      const result = await listAgents({ page: 1, pageSize: 100 })
      setAgents((result.items ?? []).filter((agent) => agent.online && agent.status === "ACTIVE"))
    } catch {
      setAgents([])
    }
    setExecOpen(true)
  }

  async function openAudits(item: SSHCommandItem) {
    setAuditTarget(item)
    try {
      const result = await listSSHCommandAudits({ page: 1, pageSize: 100, commandUid: item.uid })
      setAudits(result.items ?? [])
    } catch {
      setAudits([])
    }
    setAuditOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["command", requireTrimmed(command, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const body = { name: name.trim(), command: command.trim(), workdir, env, remark }
      if (editing) {
        await updateSSHCommand({ ...body, uid: editing.uid })
      } else {
        await createSSHCommand(body)
      }
      setOpen(false)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onExecute(event: React.FormEvent) {
    event.preventDefault()
    if (!executing) {
      return
    }
    const errors = collectFieldErrors([
      ["agents", agentUids.length === 0 ? t("form.required") : null],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const result = await executeSSHCommand(executing.uid, agentUids)
      setExecOpen(false)
      reload()
      const params = createSearchParams({ run: "1" })
      for (const audit of result.audits ?? []) {
        if (audit.uid) params.append("audit", audit.uid)
      }
      navigate({
        pathname: `/ssh-commands/${executing.uid}/results`,
        search: params.toString(),
      })
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
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
      reload()
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
        list.push({ key: "reject", label: t("ssh.reject"), destructive: true, onSelect: () => void rejectSSHCommand(item.uid).then(() => reload()) })
      } else {
        list.push({ key: "approve", label: t("ssh.approve"), onSelect: () => void approveSSHCommand(item.uid).then(() => reload()) })
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
            <Button variant="outline" onClick={() => reload()}>
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
                <TableHead className="text-right">{t("action.actions")}</TableHead>
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
          <form noValidate className="grid gap-4 sm:grid-cols-2" onSubmit={onSave}>
            <FormField
              label={t("ssh.name")}
              htmlFor="ssh-name"
              required
              error={fieldErrors.name}
            >
              <Input id="ssh-name" value={name} placeholder={t("ssh.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField label={t("ssh.workdir")} htmlFor="ssh-workdir">
              <Input id="ssh-workdir" value={workdir} placeholder={t("ssh.workdirPlaceholder")} onChange={(event) => setWorkdir(event.target.value)} />
            </FormField>
            <FormField
              label={t("ssh.command")}
              htmlFor="ssh-command"
              className="sm:col-span-full"
              required
              error={fieldErrors.command}
            >
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
          <form noValidate className="grid gap-4" onSubmit={onExecute}>
            <FormField
              label={t("ssh.agents")}
              hint={t("ssh.pickAgents")}
              required
              error={fieldErrors.agents}
            >
              <div
                className={`grid max-h-64 gap-2 overflow-auto rounded-md border p-3 ${
                  fieldErrors.agents ? "border-destructive" : ""
                }`}
              >
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
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setExecOpen(false)}>
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                {t("ssh.execute")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Sheet
        open={auditOpen}
        onOpenChange={(value) => {
          setAuditOpen(value)
          if (!value) setAuditTarget(null)
        }}
      >
        <SheetContent className="gap-0 p-0" style={{ width: "60vw", maxWidth: "none" }}>
          <SheetHeader className="border-b px-6 py-5 pr-14">
            <SheetTitle>{auditTarget?.name ?? t("ssh.audits")}</SheetTitle>
            <SheetDescription>{t("ssh.auditTimelineDescription")}</SheetDescription>
          </SheetHeader>
          <div className="min-h-0 flex-1 overflow-y-auto px-6 py-5">
            {audits.length === 0 ? (
              <p className="py-8 text-center text-sm text-muted-foreground">{t("state.empty")}</p>
            ) : (
              <ol className="grid">
                {audits.map((audit, index) => {
                  const isFailed = audit.status === "FAILED"
                  const operatorName = audit.operatorNickname || audit.operatorName || audit.operatorUid
                  return (
                    <li key={audit.uid} className="grid grid-cols-[1.25rem_minmax(0,1fr)] gap-x-3 pb-7 last:pb-0">
                      <span className="relative flex justify-center">
                        <span className={`relative z-10 mt-1 size-2.5 rounded-full ring-4 ring-background ${isFailed ? "bg-destructive" : audit.kind === "EXECUTE" ? "bg-emerald-600 dark:bg-emerald-400" : "bg-muted-foreground"}`} />
                        {index < audits.length - 1 ? <span className="absolute top-4 bottom-[-1.75rem] w-px bg-border" /> : null}
                      </span>
                      <article className="min-w-0 rounded-lg border bg-card p-4">
                        <div className="flex flex-wrap items-center justify-between gap-2">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{enumLabel(t, audit.kind)}</span>
                            <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${isFailed ? "bg-destructive/10 text-destructive" : "bg-muted text-muted-foreground"}`}>
                              {enumLabel(t, audit.status)}
                            </span>
                          </div>
                          <time className="text-xs text-muted-foreground">{new Date(audit.createdAt).toLocaleString()}</time>
                        </div>
                        {audit.operatorUid ? (
                          <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
                            {audit.operatorAvatar ? (
                              <img src={audit.operatorAvatar} alt="" className="size-6 rounded-full object-cover" />
                            ) : (
                              <span className="flex size-6 items-center justify-center rounded-full bg-muted font-medium text-foreground">
                                {operatorName.slice(0, 1).toUpperCase()}
                              </span>
                            )}
                            <span>
                              {t("ssh.operator")}: <span className="font-medium text-foreground">{operatorName}</span>
                              {audit.operatorNickname && audit.operatorName && audit.operatorNickname !== audit.operatorName ? (
                                <span className="ml-1">({audit.operatorName})</span>
                              ) : null}
                            </span>
                          </div>
                        ) : null}
                        {audit.kind === "EXECUTE" ? (
                          <Link
                            to={`/ssh-commands/${audit.commandUid || auditTarget?.uid}/results?${createSearchParams({ audit: audit.uid })}`}
                            className="mt-3 inline-flex items-center gap-2 text-sm font-medium text-primary hover:underline"
                          >
                            {t("ssh.viewExecutionRecord")}
                            <span aria-hidden="true">→</span>
                          </Link>
                        ) : (
                          <>
                            {audit.command ? <pre className="mt-3 whitespace-pre-wrap break-words rounded-md bg-muted/50 p-3 font-mono text-xs leading-5">$ {audit.command}</pre> : null}
                            {audit.detail ? <p className="mt-3 text-sm text-muted-foreground">{audit.detail}</p> : null}
                          </>
                        )}
                      </article>
                    </li>
                  )
                })}
              </ol>
            )}
          </div>
        </SheetContent>
      </Sheet>
      <DetailSheet
        open={detail !== null}
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
