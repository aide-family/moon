import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"

import {
  createAgent,
  deleteAgent,
  listAgents,
  revokeAgent,
  rotateAgentToken,
  updateAgent,
} from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
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
import { requireTrimmed, collectFieldErrors, hasFieldErrors } from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/lib/list-query"

export function AgentPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [open, setOpen] = useState(false)
  const [editing, setEditing] = useState<AgentItem | null>(null)
  const [name, setName] = useState("")
  const [sshHost, setSshHost] = useState("")
  const [sshPort, setSshPort] = useState("22")
  const [sshUser, setSshUser] = useState("")
  const [sshPassword, setSshPassword] = useState("")
  const [sshPrivateKey, setSshPrivateKey] = useState("")
  const [remark, setRemark] = useState("")
  const [token, setToken] = useState("")
  const [tokenOpen, setTokenOpen] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AgentItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "revoke" | "delete"; item: AgentItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const canWrite = session.can("agent:write")
  const canExec = session.can("ssh:exec")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<AgentItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listAgents({ page, pageSize, name: query.name }),
  })

  function openCreate() {
    setEditing(null)
    setName("")
    setSshHost("")
    setSshPort("22")
    setSshUser("")
    setSshPassword("")
    setSshPrivateKey("")
    setRemark("")
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  function openEdit(item: AgentItem) {
    setEditing(item)
    setName(item.name)
    setSshHost(item.sshHost)
    setSshPort(String(item.sshPort || 22))
    setSshUser(item.sshUser)
    setSshPassword("")
    setSshPrivateKey("")
    setRemark(item.remark)
    setError("")
    setFieldErrors({})
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const body = {
        name: name.trim(),
        sshHost,
        sshPort: Number(sshPort) || 22,
        sshUser,
        sshPassword,
        sshPrivateKey,
        remark,
      }
      if (editing) {
        await updateAgent({ ...body, uid: editing.uid })
        setOpen(false)
      } else {
        const result = await createAgent(body)
        setOpen(false)
        setToken(result.token)
        setTokenOpen(true)
      }
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onRotate(item: AgentItem) {
    setPending(true)
    try {
      const result = await rotateAgentToken(item.uid)
      setToken(result.token)
      setTokenOpen(true)
      reload()
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
      if (confirm.action === "delete") {
        await deleteAgent(confirm.item.uid)
      } else {
        await revokeAgent(confirm.item.uid)
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AgentItem): RowAction[] {
    const list: RowAction[] = []
    if (canExec && item.online) {
      list.push({ key: "login", label: t("agent.login"), onSelect: () => navigate(`/agents/${item.uid}/shell`) })
    }
    if (!canWrite) {
      return list
    }
    list.push(
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      { key: "rotate", label: t("agent.rotate"), onSelect: () => void onRotate(item) },
      { key: "revoke", label: t("agent.revoke"), destructive: true, onSelect: () => setConfirm({ action: "revoke", item }) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    )
    return list
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("agent.title")} description={t("agent.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <form
        className="flex flex-wrap gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input className="max-w-sm" value={draft} placeholder={t("agent.namePlaceholder")} onChange={(event) => setDraft(event.target.value)} />
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
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("agent.empty")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("agent.name")}</TableHead>
                <TableHead>{t("agent.hostname")}</TableHead>
                <TableHead>{t("agent.localIp")}</TableHead>
                <TableHead>{t("agent.os")}</TableHead>
                <TableHead>{t("agent.cpu")}</TableHead>
                <TableHead>{t("agent.memory")}</TableHead>
                <TableHead>{t("agent.disk")}</TableHead>
                <TableHead>{t("agent.online")}</TableHead>
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead className="text-right">{t("action.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{item.hostname || "-"}</TableCell>
                  <TableCell>{item.localIp || "-"}</TableCell>
                  <TableCell>{item.osVersion || item.os || "-"}</TableCell>
                  <TableCell className="max-w-48 truncate" title={item.cpu}>{item.cpu || "-"}</TableCell>
                  <TableCell>{item.memory || "-"}</TableCell>
                  <TableCell>{item.disk || "-"}</TableCell>
                  <TableCell>{item.online ? t("agent.onlineYes") : t("agent.onlineNo")}</TableCell>
                  <TableCell>{enumLabel(t, item.status)}</TableCell>
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
          <form noValidate className="grid gap-4 sm:grid-cols-3" onSubmit={onSave}>
            <FormField
              label={t("agent.name")}
              htmlFor="agent-name"
              className="sm:col-span-full"
              required
              error={fieldErrors.name}
            >
              <Input id="agent-name" value={name} placeholder={t("agent.namePlaceholder")} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField label={t("agent.sshHost")} htmlFor="agent-host">
              <Input id="agent-host" value={sshHost} placeholder={t("agent.sshHostPlaceholder")} onChange={(event) => setSshHost(event.target.value)} />
            </FormField>
            <FormField label={t("agent.sshPort")} htmlFor="agent-port">
              <Input id="agent-port" value={sshPort} placeholder={t("agent.sshPortPlaceholder")} onChange={(event) => setSshPort(event.target.value)} />
            </FormField>
            <FormField label={t("agent.sshUser")} htmlFor="agent-user">
              <Input id="agent-user" value={sshUser} placeholder={t("agent.sshUserPlaceholder")} onChange={(event) => setSshUser(event.target.value)} />
            </FormField>
            <FormField label={t("agent.sshPassword")} htmlFor="agent-password" className="sm:col-span-full">
              <Input id="agent-password" type="password" value={sshPassword} onChange={(event) => setSshPassword(event.target.value)} placeholder={t("agent.keepSecret")} />
            </FormField>
            <FormField label={t("agent.sshPrivateKey")} htmlFor="agent-key" className="sm:col-span-full">
              <Textarea id="agent-key" value={sshPrivateKey} onChange={(event) => setSshPrivateKey(event.target.value)} placeholder={editing ? t("agent.keepSecret") : t("agent.sshPrivateKeyPlaceholder")} />
            </FormField>
            <FormField label={t("agent.remark")} htmlFor="agent-remark" className="sm:col-span-full">
              <Textarea id="agent-remark" value={remark} placeholder={t("agent.remarkPlaceholder")} onChange={(event) => setRemark(event.target.value)} />
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
      <Dialog open={tokenOpen} onOpenChange={setTokenOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("agent.tokenTitle")}</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{t("agent.tokenHint")}</p>
          <Input readOnly value={token} />
          <DialogFooter>
            <Button onClick={() => setTokenOpen(false)}>{t("action.confirm")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <DetailSheet
        wide
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
                { label: t("agent.hostname"), value: detail.hostname || "-" },
                { label: t("agent.machineId"), value: detail.machineId || "-" },
                { label: t("agent.localIp"), value: detail.localIp || "-" },
                { label: t("agent.os"), value: detail.os || "-" },
                { label: t("agent.arch"), value: detail.arch || "-" },
                { label: t("agent.osVersion"), value: detail.osVersion || "-" },
                { label: t("agent.cpu"), value: detail.cpu || "-" },
                { label: t("agent.disk"), value: detail.disk || "-" },
                { label: t("agent.memory"), value: detail.memory || "-" },
                { label: t("agent.bootTime"), value: detail.bootTime || "-" },
                { label: t("agent.timeServer"), value: detail.timeServer || "-" },
                { label: t("agent.version"), value: detail.version || "-" },
                { label: t("agent.online"), value: detail.online ? t("agent.onlineYes") : t("agent.onlineNo") },
                { label: t("agent.sshHost"), value: detail.sshHost || "-" },
                { label: t("agent.sshPort"), value: String(detail.sshPort || 22) },
                { label: t("agent.sshUser"), value: detail.sshUser || "-" },
                { label: t("agent.remark"), value: detail.remark || "-" },
              ]
            : []
        }
      />
      <ConfirmAction
        open={confirm !== null}
        title={confirm?.action === "delete" ? t("action.delete") : t("agent.revoke")}
        description={t("agent.confirmBody")}
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
