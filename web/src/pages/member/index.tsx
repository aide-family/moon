import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { deleteMember, inviteMember, listMembers, updateMemberRoles } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import { selectRoles } from "@/api/role"
import { useSession } from "@/auth/session"
import { useListQuery } from "@/lib/list-query"
import { Button } from "@/components/ui/button"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"

type RoleOption = { uid: string; name: string }

export function MemberPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [items, setItems] = useState<MemberItem[]>([])
  const [total, setTotal] = useState(0)
  const [roles, setRoles] = useState<RoleOption[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState("")
  const [roleUid, setRoleUid] = useState("")
  const [assigning, setAssigning] = useState<MemberItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const [result, options] = await Promise.all([
          listMembers({ page: query.page, pageSize: query.pageSize, name: query.name }),
          selectRoles(100),
        ])
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
        const available = (options.items ?? []).filter((item) => item.name !== "owner")
        setRoles(available)
        setRoleUid(available[0]?.uid ?? "")
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

  async function onInvite(event: React.FormEvent) {
    event.preventDefault()
    setPending(true)
    setError("")
    try {
      await inviteMember({ email, roleUid })
      setOpen(false)
      setEmail("")
      setAttempt((value) => value + 1)
    } catch (err) {
      setError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  async function onAssign(event: React.FormEvent) {
    event.preventDefault()
    if (!assigning) {
      return
    }
    setPending(true)
    setError("")
    try {
      await updateMemberRoles(assigning.uid, selected)
      setAssigning(null)
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
        <h1 className="text-xl font-semibold">{t("member.title")}</h1>
        {session.can("member:write") ? <Button onClick={() => { setOpen(true); setError("") }}>{t("action.invite")}</Button> : null}
      </div>
      <form className="flex gap-2" onSubmit={(event) => { event.preventDefault(); query.update({ name: draft }) }}>
        <Input value={draft} placeholder={t("user.name")} onChange={(event) => setDraft(event.target.value)} />
        <Button type="submit" variant="outline">{t("action.search")}</Button>
      </form>
      {loading ? <p>{t("state.loading")}</p> : null}
      {failed ? <p>{t("state.failed")}</p> : null}
      {!loading && !failed && items.length === 0 ? <p>{t("state.empty")}</p> : null}
      {items.length > 0 ? (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("member.email")}</TableHead>
              <TableHead>{t("member.role")}</TableHead>
              <TableHead>{t("member.status")}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.uid}>
                <TableCell>{item.email}</TableCell>
                <TableCell>{(item.roleNames ?? []).join(", ")}</TableCell>
                <TableCell>{item.status}</TableCell>
                <TableCell className="space-x-2 text-right">
                  {session.can("member:write") && !(item.roleNames ?? []).includes("owner") ? (
                    <>
                      <Button variant="outline" size="sm" onClick={() => { setAssigning(item); setSelected(item.roleUids ?? []); setError("") }}>
                        {t("action.assign")}
                      </Button>
                      <Button variant="destructive" size="sm" onClick={() => void deleteMember(item.uid).then(() => setAttempt((value) => value + 1))}>
                        {t("action.delete")}
                      </Button>
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
          <DialogHeader><DialogTitle>{t("action.invite")}</DialogTitle></DialogHeader>
          <form className="space-y-4" onSubmit={onInvite}>
            <div className="space-y-2">
              <Label htmlFor="invite-email">{t("member.email")}</Label>
              <Input id="invite-email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="invite-role">{t("member.role")}</Label>
              <select id="invite-role" className="w-full rounded-md border bg-background px-3 py-2" value={roleUid} onChange={(event) => setRoleUid(event.target.value)}>
                {roles.map((role) => <option key={role.uid} value={role.uid}>{role.name}</option>)}
              </select>
            </div>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter><Button type="submit" disabled={pending || !roleUid}>{t("action.invite")}</Button></DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={assigning !== null} onOpenChange={(value) => { if (!value) setAssigning(null) }}>
        <DialogContent>
          <DialogHeader><DialogTitle>{t("action.assign")}</DialogTitle></DialogHeader>
          <form className="space-y-4" onSubmit={onAssign}>
            {roles.map((role) => (
              <label key={role.uid} className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={selected.includes(role.uid)}
                  onChange={(event) => {
                    setSelected((current) => event.target.checked ? [...current, role.uid] : current.filter((uid) => uid !== role.uid))
                  }}
                />
                {role.name}
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
