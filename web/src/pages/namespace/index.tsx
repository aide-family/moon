import { useState } from "react"
import { useTranslation } from "react-i18next"

import { listMembers } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import { createNamespace, deleteNamespace, transferNamespaceOwner, updateNamespace, updateNamespaceStatus } from "@/api/namespace"
import { useSession } from "@/auth/session"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

export function NamespacePage() {
  const { t } = useTranslation()
  const session = useSession()
  const [open, setOpen] = useState(false)
  const [name, setName] = useState("")
  const [remark, setRemark] = useState("")
  const [editing, setEditing] = useState<string | null>(null)
  const [transferUid, setTransferUid] = useState<string | null>(null)
  const [members, setMembers] = useState<MemberItem[]>([])
  const [target, setTarget] = useState("")
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  async function onCreate(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      const result = await createNamespace({ name, remark })
      localStorage.setItem("namespace", result.namespace.uid)
      setOpen(false)
      setName("")
      setRemark("")
      session.reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onRename(event: React.FormEvent) {
    event.preventDefault()
    if (!editing) {
      return
    }
    setPending(true)
    setError("")
    try {
      localStorage.setItem("namespace", editing)
      await updateNamespace({ uid: editing, name, remark })
      setEditing(null)
      session.reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onStatus(uid: string, status: string) {
    localStorage.setItem("namespace", uid)
    await updateNamespaceStatus(uid, status === "ACTIVE" ? "DISABLED" : "ACTIVE")
    session.reload()
  }

  async function onDelete(uid: string) {
    localStorage.setItem("namespace", uid)
    await deleteNamespace(uid)
    localStorage.removeItem("namespace")
    session.reload()
  }

  async function openTransfer(uid: string) {
    localStorage.setItem("namespace", uid)
    setTransferUid(uid)
    setError("")
    const result = await listMembers({ page: 1, pageSize: 100 })
    const items = (result.items ?? []).filter((item) => item.userUid !== session.user?.uid && item.status === "JOINED")
    setMembers(items)
    setTarget(items[0]?.userUid ?? "")
  }

  async function onTransfer(event: React.FormEvent) {
    event.preventDefault()
    if (!transferUid || !target) {
      return
    }
    setPending(true)
    setError("")
    try {
      localStorage.setItem("namespace", transferUid)
      await transferNamespaceOwner(transferUid, target)
      setTransferUid(null)
      session.reload()
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t("namespace.title")}</h1>
        <Button onClick={() => { setOpen(true); setError("") }}>{t("namespace.create")}</Button>
      </div>
      {session.namespaces.length === 0 ? <p className="text-sm text-muted-foreground">{t("namespace.guide")}</p> : null}
      {session.loading ? <p>{t("state.loading")}</p> : null}
      {session.failed ? <p>{t("state.failed")}</p> : null}
      {!session.loading && session.namespaces.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("namespace.name")}</TableHead>
              <TableHead>{t("namespace.status")}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {session.namespaces.map((item) => (
              <TableRow key={item.uid}>
                <TableCell>
                  {item.name} {item.uid === session.namespaceUid ? <Badge>{t("namespace.current")}</Badge> : null}
                </TableCell>
                <TableCell>{item.status}</TableCell>
                <TableCell className="space-x-2 text-right">
                  {session.can("namespace:write") && item.uid === session.namespaceUid ? (
                    <>
                      <Button variant="outline" size="sm" onClick={() => { setEditing(item.uid); setName(item.name); setRemark(item.remark); setError("") }}>
                        {t("action.edit")}
                      </Button>
                      <Button variant="outline" size="sm" onClick={() => void onStatus(item.uid, item.status)}>
                        {item.status === "ACTIVE" ? t("action.disable") : t("action.enable")}
                      </Button>
                    </>
                  ) : null}
                  {session.can("namespace:transfer") && item.uid === session.namespaceUid ? (
                    <Button variant="outline" size="sm" onClick={() => void openTransfer(item.uid)}>
                      {t("action.transfer")}
                    </Button>
                  ) : null}
                  {session.can("namespace:delete") && item.uid === session.namespaceUid ? (
                    <Button variant="destructive" size="sm" onClick={() => void onDelete(item.uid)}>
                      {t("action.delete")}
                    </Button>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : null}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("namespace.create")}</DialogTitle>
          </DialogHeader>
          <form className="space-y-4" onSubmit={onCreate}>
            <div className="space-y-2">
              <Label htmlFor="ns-name">{t("namespace.name")}</Label>
              <Input id="ns-name" value={name} onChange={(event) => setName(event.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="ns-remark">{t("namespace.remark")}</Label>
              <Input id="ns-remark" value={remark} onChange={(event) => setRemark(event.target.value)} />
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>{t("action.create")}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={editing !== null} onOpenChange={(value) => { if (!value) setEditing(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.edit")}</DialogTitle>
          </DialogHeader>
          <form className="space-y-4" onSubmit={onRename}>
            <div className="space-y-2">
              <Label htmlFor="edit-name">{t("namespace.name")}</Label>
              <Input id="edit-name" value={name} onChange={(event) => setName(event.target.value)} required />
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>{t("action.save")}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={transferUid !== null} onOpenChange={(value) => { if (!value) setTransferUid(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.transfer")}</DialogTitle>
          </DialogHeader>
          <form className="space-y-4" onSubmit={onTransfer}>
            {members.length === 0 ? <p>{t("state.empty")}</p> : (
              <select className="w-full rounded-md border bg-background px-3 py-2" value={target} onChange={(event) => setTarget(event.target.value)}>
                {members.map((item) => (
                  <option key={item.uid} value={item.userUid}>{item.email}</option>
                ))}
              </select>
            )}
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending || !target}>{t("action.save")}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </section>
  )
}
