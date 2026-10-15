import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { createRole, deleteRole, listPermissions, listRoles, updateRole, updateRolePermissions, updateRoleStatus } from "@/api/role"
import type { PermissionItem, RoleItem } from "@/api/role/types"
import { useSession } from "@/auth/session"
import { useListQuery } from "@/lib/list-query"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

export function RolePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<RoleItem[]>([])
  const [total, setTotal] = useState(0)
  const [catalog, setCatalog] = useState<PermissionItem[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [name, setName] = useState("")
  const [editing, setEditing] = useState<RoleItem | null>(null)
  const [granting, setGranting] = useState<RoleItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const [result, permissions] = await Promise.all([
          listRoles({ page: query.page, pageSize: query.pageSize, name: query.name }),
          listPermissions(),
        ])
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
        setCatalog((permissions.items ?? []).filter((item) => item.name !== "namespace:delete" && item.name !== "namespace:transfer"))
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

  async function onCreate(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      await createRole({ name })
      setOpen(false)
      setName("")
      setAttempt((value) => value + 1)
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
      await updateRole({ uid: editing.uid, name })
      setEditing(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onGrant(event: React.FormEvent) {
    event.preventDefault()
    if (!granting) {
      return
    }
    setPending(true)
    setError("")
    try {
      await updateRolePermissions(granting.uid, selected)
      setGranting(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">{t("role.title")}</h1>
        {session.can("role:write") ? <Button onClick={() => { setOpen(true); setError("") }}>{t("action.create")}</Button> : null}
      </div>
      <form className="flex gap-2" onSubmit={(event) => { event.preventDefault(); query.update({ name: draft }) }}>
        <Input value={draft} placeholder={t("role.name")} onChange={(event) => setDraft(event.target.value)} />
        <Button type="submit" variant="outline">{t("action.search")}</Button>
      </form>
      {loading ? <p>{t("state.loading")}</p> : null}
      {failed ? <p>{t("state.failed")}</p> : null}
      {!loading && !failed && items.length === 0 ? <p>{t("state.empty")}</p> : null}
      {items.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("role.name")}</TableHead>
              <TableHead>{t("namespace.status")}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.uid}>
                <TableCell>{item.name}{item.isSystem ? ` (${t("role.system")})` : ""}</TableCell>
                <TableCell>{item.status}</TableCell>
                <TableCell className="space-x-2 text-right">
                  {session.can("role:write") && !item.isSystem ? (
                    <>
                      <Button variant="outline" size="sm" onClick={() => { setEditing(item); setName(item.name); setError("") }}>{t("action.edit")}</Button>
                      <Button variant="outline" size="sm" onClick={() => { setGranting(item); setSelected(item.permissions ?? []); setError("") }}>{t("action.permissions")}</Button>
                      <Button variant="outline" size="sm" onClick={() => void updateRoleStatus(item.uid, item.status === "ACTIVE" ? "DISABLED" : "ACTIVE").then(() => setAttempt((value) => value + 1))}>
                        {item.status === "ACTIVE" ? t("action.disable") : t("action.enable")}
                      </Button>
                      <Button variant="destructive" size="sm" onClick={() => void deleteRole(item.uid).then(() => setAttempt((value) => value + 1))}>{t("action.delete")}</Button>
                    </>
                  ) : null}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      ) : null}
      <div className="flex items-center gap-2">
        <Button variant="outline" disabled={query.page <= 1} onClick={() => query.update({ page: query.page - 1 })}>{t("action.prev")}</Button>
        <span className="text-sm">{query.page} / {Math.max(1, Math.ceil(total / query.pageSize))}</span>
        <Button variant="outline" disabled={query.page * query.pageSize >= total} onClick={() => query.update({ page: query.page + 1 })}>{t("action.next")}</Button>
      </div>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader><DialogTitle>{t("action.create")}</DialogTitle></DialogHeader>
          <form className="space-y-4" onSubmit={onCreate}>
            <div className="space-y-2">
              <Label htmlFor="role-name">{t("role.name")}</Label>
              <Input id="role-name" value={name} onChange={(event) => setName(event.target.value)} required />
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter><Button type="submit" disabled={pending}>{t("action.create")}</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={editing !== null} onOpenChange={(value) => { if (!value) setEditing(null) }}>
        <DialogContent>
          <DialogHeader><DialogTitle>{t("action.edit")}</DialogTitle></DialogHeader>
          <form className="space-y-4" onSubmit={onRename}>
            <Input value={name} onChange={(event) => setName(event.target.value)} required />
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter><Button type="submit" disabled={pending}>{t("action.save")}</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={granting !== null} onOpenChange={(value) => { if (!value) setGranting(null) }}>
        <DialogContent>
          <DialogHeader><DialogTitle>{t("action.permissions")}</DialogTitle></DialogHeader>
          <form className="max-h-80 space-y-2 overflow-auto" onSubmit={onGrant}>
            {catalog.map((item) => (
              <label key={item.uid} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={selected.includes(item.name)}
                  onChange={(event) => {
                    setSelected((current) => event.target.checked ? [...current, item.name] : current.filter((name) => name !== item.name))
                  }}
                />
                {item.name}
              </label>
            ))}
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter><Button type="submit" disabled={pending || selected.length === 0}>{t("action.save")}</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </section>
  )
}
