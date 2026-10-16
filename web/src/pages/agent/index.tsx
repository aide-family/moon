import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

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
import { useListQuery } from "@/lib/list-query"

export function AgentPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<AgentItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
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
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<AgentItem | null>(null)
  const [confirm, setConfirm] = useState<{ action: "revoke" | "delete"; item: AgentItem } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const shown = useRef(confirm)
  if (confirm) {
    shown.current = confirm
  }
  const canWrite = session.can("agent:write")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listAgents({ page: query.page, pageSize: query.pageSize, name: query.name })
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
      } catch {
        if (active) {
          setFailed(true)
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
    setSshHost("")
    setSshPort("22")
    setSshUser("")
    setSshPassword("")
    setSshPrivateKey("")
    setRemark("")
    setError("")
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
    setOpen(true)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const body = {
        name,
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
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
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
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
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
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  function actions(item: AgentItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    return [
      { key: "edit", label: t("action.edit"), onSelect: () => openEdit(item) },
      { key: "rotate", label: t("agent.rotate"), onSelect: () => void onRotate(item) },
      { key: "revoke", label: t("agent.revoke"), destructive: true, onSelect: () => setConfirm({ action: "revoke", item }) },
      { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setConfirm({ action: "delete", item }) },
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("agent.title")} description={t("agent.description")}>
        {canWrite ? <Button onClick={openCreate}>{t("action.create")}</Button> : null}
      </PageHeader>
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
            <p className="text-sm text-destructive">{t("state.failed")}</p>
            <Button variant="outline" onClick={() => setAttempt((value) => value + 1)}>
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
                <TableHead>{t("agent.online")}</TableHead>
                <TableHead>{t("namespace.status")}</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{item.hostname || "-"}</TableCell>
                  <TableCell>{item.localIp || "-"}</TableCell>
                  <TableCell>{item.online ? t("agent.onlineYes") : t("agent.onlineNo")}</TableCell>
                  <TableCell>{item.status}</TableCell>
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
        <DialogContent className="max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editing ? t("action.edit") : t("action.create")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onSave}>
            <FormField label={t("agent.name")} htmlFor="agent-name">
              <Input id="agent-name" value={name} onChange={(event) => setName(event.target.value)} required />
            </FormField>
            <FormField label={t("agent.sshHost")} htmlFor="agent-host">
              <Input id="agent-host" value={sshHost} onChange={(event) => setSshHost(event.target.value)} />
            </FormField>
            <FormField label={t("agent.sshPort")} htmlFor="agent-port">
              <Input id="agent-port" value={sshPort} onChange={(event) => setSshPort(event.target.value)} />
            </FormField>
            <FormField label={t("agent.sshUser")} htmlFor="agent-user">
              <Input id="agent-user" value={sshUser} onChange={(event) => setSshUser(event.target.value)} />
            </FormField>
            <FormField label={t("agent.sshPassword")} htmlFor="agent-password">
              <Input id="agent-password" type="password" value={sshPassword} onChange={(event) => setSshPassword(event.target.value)} placeholder={editing ? t("agent.keepSecret") : ""} />
            </FormField>
            <FormField label={t("agent.sshPrivateKey")} htmlFor="agent-key">
              <Textarea id="agent-key" value={sshPrivateKey} onChange={(event) => setSshPrivateKey(event.target.value)} placeholder={editing ? t("agent.keepSecret") : ""} />
            </FormField>
            <FormField label={t("agent.remark")} htmlFor="agent-remark">
              <Textarea id="agent-remark" value={remark} onChange={(event) => setRemark(event.target.value)} />
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
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
        title={shown.current?.action === "delete" ? t("action.delete") : t("agent.revoke")}
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
