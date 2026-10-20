import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { File, Folder } from "lucide-react"
import { formMessage } from "@/lib/form-error"

import { listAgents } from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
import { deleteSFTP, downloadSFTP, listSFTP, listSFTPAudits, mkdirSFTP, uploadSFTP } from "@/api/sftp"
import type { SFTPAuditItem, SFTPEntry } from "@/api/sftp/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

function crumbs(path: string) {
  const parts = path.split("/").filter(Boolean)
  const items = [{ label: "/", path: "/" }]
  let current = ""
  for (const part of parts) {
    current += `/${part}`
    items.push({ label: part, path: current })
  }
  return items
}

function saveFile(name: string, content: string) {
  const binary = atob(content)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i)
  }
  const url = URL.createObjectURL(new Blob([bytes]))
  const link = document.createElement("a")
  link.href = url
  link.download = name
  link.click()
  URL.revokeObjectURL(url)
}

export function SFTPPage() {
  const { t } = useTranslation()
  const session = useSession()
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [agentUid, setAgentUid] = useState("")
  const [path, setPath] = useState("/")
  const [entries, setEntries] = useState<SFTPEntry[]>([])
  const [selected, setSelected] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [mkdirName, setMkdirName] = useState("")
  const [confirmEntry, setConfirmEntry] = useState<SFTPEntry | null>(null)
  const [audits, setAudits] = useState<SFTPAuditItem[]>([])
  const [auditError, setAuditError] = useState("")
  const [auditAttempt, setAuditAttempt] = useState(0)
  const canRead = session.can("sftp:read")
  const canWrite = session.can("sftp:write")
  const ordered = useMemo(
    () =>
      [...entries].sort((a, b) => {
        if (a.isDir !== b.isDir) {
          return a.isDir ? -1 : 1
        }
        return a.name.localeCompare(b.name)
      }),
    [entries],
  )

  useEffect(() => {
    let active = true
    void listAgents({ page: 1, pageSize: 100 }).then((result) => {
      if (!active) return
      setAgents((result.items ?? []).filter((item) => item.online))
    })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  useEffect(() => {
    let active = true
    if (!agentUid) {
      return
    }
    void listSFTPAudits({ page: 1, pageSize: 20, agentUid })
      .then((result) => {
        if (active) setAudits(result.items ?? [])
      })
      .catch((err) => {
        if (active) setAuditError(formMessage(err, t("state.failed")))
      })
    return () => {
      active = false
    }
  }, [agentUid, auditAttempt, session.namespaceUid, t])

  async function load(nextPath = path) {
    if (!agentUid) return
    setPending(true)
    setError("")
    setSelected("")
    try {
      const result = await listSFTP({ agentUid, path: nextPath })
      setEntries(result.entries ?? [])
      setPath(nextPath)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
      setEntries([])
    } finally {
      setPending(false)
    }
  }

  async function onMkdir() {
    if (!agentUid || !mkdirName.trim()) return
    setPending(true)
    setError("")
    try {
      const target = `${path.replace(/\/$/, "")}/${mkdirName.trim()}`
      await mkdirSFTP({ agentUid, path: target })
      setMkdirName("")
      await load(path)
      setAuditAttempt((value) => value + 1)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function onUpload(file: File | null) {
    if (!file || !agentUid) return
    setPending(true)
    try {
      const buf = await file.arrayBuffer()
      const bytes = new Uint8Array(buf)
      let binary = ""
      bytes.forEach((b) => {
        binary += String.fromCharCode(b)
      })
      await uploadSFTP({ agentUid, path: `${path.replace(/\/$/, "")}/${file.name}`, content: btoa(binary) })
      await load(path)
      setAuditAttempt((value) => value + 1)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function onDownload(entry: SFTPEntry) {
    if (!agentUid || entry.isDir) return
    setPending(true)
    try {
      const result = await downloadSFTP({ agentUid, path: entry.path })
      saveFile(entry.name, result.content)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function onDelete() {
    if (!agentUid || !canWrite || !confirmEntry) return
    setPending(true)
    try {
      await deleteSFTP({ agentUid, path: confirmEntry.path })
      setConfirmEntry(null)
      await load(path)
      setAuditAttempt((value) => value + 1)
    } catch (err) {
      setError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  if (!canRead) {
    return <p className="text-sm text-muted-foreground">{t("state.failed")}</p>
  }

  const current = ordered.find((entry) => entry.path === selected)

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("sftp.title")} description={t("sftp.description")} />
      <div className="flex flex-wrap items-center gap-2">
        <Select
          value={agentUid}
          onValueChange={(value) => {
            setAgentUid(value)
            setAudits([])
            setAuditError("")
          }}
        >
          <SelectTrigger className="w-56">
            <SelectValue placeholder={t("sftp.pickAgent")} />
          </SelectTrigger>
          <SelectContent>
            {agents.map((agent) => (
              <SelectItem key={agent.uid} value={agent.uid}>
                {agent.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Button disabled={!agentUid || pending} onClick={() => void load()}>
          {t("action.refresh")}
        </Button>
        {canWrite ? (
          <label className="inline-flex h-8 cursor-pointer items-center rounded-lg border px-3 text-sm">
            {t("sftp.upload")}
            <Input className="hidden" type="file" onChange={(event) => void onUpload(event.target.files?.[0] ?? null)} />
          </label>
        ) : null}
        <Button variant="outline" disabled={!current || current.isDir || pending} onClick={() => current && void onDownload(current)}>
          {t("sftp.download")}
        </Button>
        {canWrite ? (
          <Button variant="outline" disabled={!current || pending} onClick={() => current && setConfirmEntry(current)}>
            {t("action.delete")}
          </Button>
        ) : null}
      </div>
      {canWrite ? (
        <div className="flex flex-wrap items-end gap-2">
          <FormField label={t("sftp.mkdir")} htmlFor="mkdir" required>
            <Input id="mkdir" value={mkdirName} placeholder={t("sftp.mkdirPlaceholder")} onChange={(event) => setMkdirName(event.target.value)} />
          </FormField>
          <Button disabled={pending || !mkdirName.trim()} onClick={() => void onMkdir()}>
            {t("sftp.mkdir")}
          </Button>
        </div>
      ) : null}
      <div className="flex flex-wrap items-center gap-1 text-sm">
        {crumbs(path).map((crumb) => (
          <button key={crumb.path} className="text-primary" type="button" onClick={() => void load(crumb.path)}>
            {crumb.label === "/" ? crumb.label : ` / ${crumb.label}`}
          </button>
        ))}
      </div>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <div className="glass-card min-h-0 flex-1 overflow-auto rounded-xl p-4">
        {pending ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!pending && ordered.length === 0 ? <p className="text-sm text-muted-foreground">{t("sftp.empty")}</p> : null}
        <div className="grid grid-cols-[repeat(auto-fill,minmax(7rem,1fr))] gap-3">
          {ordered.map((entry) => (
            <button
              key={entry.path}
              type="button"
              className={`flex flex-col items-center gap-2 rounded-lg p-3 text-center text-sm ${selected === entry.path ? "bg-accent" : "hover:bg-accent/60"}`}
              onClick={() => setSelected(entry.path)}
              onDoubleClick={() => {
                if (entry.isDir) {
                  void load(entry.path)
                } else {
                  void onDownload(entry)
                }
              }}
            >
              {entry.isDir ? <Folder className="size-10 text-primary" /> : <File className="size-10 text-muted-foreground" />}
              <span className="w-full truncate" title={entry.name}>{entry.name}</span>
            </button>
          ))}
        </div>
      </div>
      <div className="grid gap-2">
        <h3 className="text-sm font-medium">{t("sftp.audits")}</h3>
        {auditError ? <p className="text-sm text-destructive">{auditError}</p> : null}
        {audits.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("sftp.empty")}</p>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("sftp.time")}</TableHead>
                <TableHead>{t("sftp.action")}</TableHead>
                <TableHead>{t("sftp.path")}</TableHead>
                <TableHead>{t("sftp.result")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {audits.map((audit) => (
                <TableRow key={audit.uid}>
                  <TableCell className="font-mono text-xs">{audit.createdAt}</TableCell>
                  <TableCell>{audit.action}</TableCell>
                  <TableCell className="font-mono text-xs">{audit.path}</TableCell>
                  <TableCell>{audit.result}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </div>
      <ConfirmAction
        open={confirmEntry !== null}
        title={t("confirm.deleteTitle")}
        description={t("confirm.deleteDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        onConfirm={() => void onDelete()}
        onOpenChange={(open) => {
          if (!open) setConfirmEntry(null)
        }}
      />
    </section>
  )
}
