import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { listAgents } from "@/api/agent"
import type { AgentItem } from "@/api/agent/types"
import { deleteSFTP, listSFTP, mkdirSFTP, uploadSFTP } from "@/api/sftp"
import type { SFTPEntry } from "@/api/sftp/types"
import { useSession } from "@/auth/session"
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

export function SFTPPage() {
  const { t } = useTranslation()
  const session = useSession()
  const [agents, setAgents] = useState<AgentItem[]>([])
  const [agentUid, setAgentUid] = useState("")
  const [path, setPath] = useState("/")
  const [entries, setEntries] = useState<SFTPEntry[]>([])
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)
  const [mkdirName, setMkdirName] = useState("")
  const canRead = session.can("sftp:read")
  const canWrite = session.can("sftp:write")

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

  async function load(nextPath = path) {
    if (!agentUid) return
    setPending(true)
    setError("")
    try {
      const result = await listSFTP({ agentUid, path: nextPath })
      setEntries(result.entries ?? [])
      setPath(nextPath)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
      setEntries([])
    } finally {
      setPending(false)
    }
  }

  async function onMkdir() {
    if (!agentUid || !mkdirName) return
    setPending(true)
    try {
      const target = `${path.replace(/\/$/, "")}/${mkdirName}`
      await mkdirSFTP({ agentUid, path: target })
      setMkdirName("")
      await load(path)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
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
      const content = btoa(binary)
      await uploadSFTP({ agentUid, path: `${path.replace(/\/$/, "")}/${file.name}`, content })
      await load(path)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onDelete(entry: SFTPEntry) {
    if (!agentUid || !canWrite) return
    setPending(true)
    try {
      await deleteSFTP({ agentUid, path: entry.path })
      await load(path)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  if (!canRead) {
    return <p className="text-sm text-muted-foreground">{t("state.failed")}</p>
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("sftp.title")} description={t("sftp.description")} />
      <div className="flex flex-wrap gap-2">
        <Select value={agentUid} onValueChange={setAgentUid}>
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
        <Input className="max-w-md" value={path} onChange={(e) => setPath(e.target.value)} />
        <Button disabled={!agentUid || pending} onClick={() => void load()}>
          {t("action.refresh")}
        </Button>
      </div>
      {canWrite ? (
        <div className="flex flex-wrap gap-2">
          <FormField label={t("sftp.mkdir")} htmlFor="mkdir">
            <Input id="mkdir" value={mkdirName} onChange={(e) => setMkdirName(e.target.value)} />
          </FormField>
          <Button disabled={pending} onClick={() => void onMkdir()}>
            {t("sftp.mkdir")}
          </Button>
          <Input type="file" onChange={(e) => void onUpload(e.target.files?.[0] ?? null)} />
        </div>
      ) : null}
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("sftp.name")}</TableHead>
            <TableHead>{t("sftp.size")}</TableHead>
            <TableHead>{t("sftp.mode")}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {entries.map((entry) => (
            <TableRow key={entry.path}>
              <TableCell>
                {entry.isDir ? (
                  <button className="text-primary underline" type="button" onClick={() => void load(entry.path)}>
                    {entry.name}/
                  </button>
                ) : (
                  entry.name
                )}
              </TableCell>
              <TableCell>{entry.size}</TableCell>
              <TableCell>{entry.mode}</TableCell>
              <TableCell>
                {canWrite ? (
                  <Button variant="outline" size="sm" onClick={() => void onDelete(entry)}>
                    {t("action.delete")}
                  </Button>
                ) : null}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </section>
  )
}
