import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  deleteMember,
  inviteMember,
  listMembers,
  updateMemberRoles,
} from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import { selectRoles } from "@/api/role"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { firstError, requireSelected, requireTrimmed } from "@/lib/form"
import { useListQuery } from "@/lib/list-query"
import { ListFrame } from "@/components/list-frame"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
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
  const [detail, setDetail] = useState<MemberItem | null>(null)
  const [removing, setRemoving] = useState<MemberItem | null>(null)
  const [confirmError, setConfirmError] = useState("")

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const [result, options] = await Promise.all([
          listMembers({
            page: query.page,
            pageSize: query.pageSize,
            name: query.name,
          }),
          selectRoles(100),
        ])
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
        const available = (options.items ?? []).filter(
          (item) => item.name !== "owner"
        )
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
    const invalid = firstError(requireTrimmed(email, t("form.required")), requireSelected(roleUid, t("form.required")))
    if (invalid) {
      setError(invalid)
      return
    }
    setPending(true)
    setError("")
    try {
      await inviteMember({ email: email.trim(), roleUid })
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

  function closeInvite(value: boolean) {
    setOpen(value)
    if (!value) {
      setEmail("")
      setError("")
    }
  }

  function closeAssign(value: boolean) {
    if (!value) {
      setAssigning(null)
      setSelected([])
      setError("")
    }
  }

  function memberActions(item: MemberItem): RowAction[] {
    if (
      !session.can("member:write") ||
      (item.roleNames ?? []).includes("owner")
    ) {
      return []
    }
    return [
      {
        key: "assign",
        label: t("action.assign"),
        onSelect: () => {
          setAssigning(item)
          setSelected(item.roleUids ?? [])
          setError("")
        },
      },
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setRemoving(item)
        },
      },
    ]
  }

  async function onDelete() {
    if (!removing) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await deleteMember(removing.uid)
      setRemoving(null)
      setAttempt((value) => value + 1)
    } catch (err) {
      setConfirmError(err instanceof Error ? err.message : t("state.failed"))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("member.title")}
        description={t("member.description")}
      >
        {session.can("member:write") ? (
          <Button
            onClick={() => {
              setEmail("")
              setError("")
              setOpen(true)
            }}
          >
            {t("action.invite")}
          </Button>
        ) : null}
      </PageHeader>
      <form
        className="flex gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input
          className="max-w-sm"
          value={draft}
          placeholder={t("user.name")}
          onChange={(event) => setDraft(event.target.value)}
        />
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
          {loading ? (
            <p className="text-sm text-muted-foreground">
              {t("state.loading")}
            </p>
          ) : null}
          {!loading && failed ? (
            <div className="flex items-center gap-2">
              <p className="text-sm text-destructive">{t("state.failed")}</p>
              <Button
                variant="outline"
                onClick={() => setAttempt((value) => value + 1)}
              >
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {!loading && !failed && items.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("member.empty")}</p>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
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
                    <TableCell>
                      <RowActions
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => setDetail(item)}
                        actions={memberActions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={closeInvite}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.invite")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onInvite}>
            <FormField label={t("member.email")} htmlFor="invite-email">
              <Input
                id="invite-email"
                type="email"
                value={email}
                placeholder={t("member.emailPlaceholder")}
                onChange={(event) => setEmail(event.target.value)}
                required
              />
            </FormField>
            <FormField
              label={t("member.role")}
              htmlFor="invite-role"
              hint={t("member.roleHint")}
            >
              <Select value={roleUid || undefined} onValueChange={setRoleUid}>
                <SelectTrigger id="invite-role" className="w-full">
                  <SelectValue placeholder={t("member.rolePlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  {roles.map((role) => (
                    <SelectItem key={role.uid} value={role.uid}>
                      {role.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending || !roleUid}>
                {t("action.invite")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={assigning !== null} onOpenChange={closeAssign}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.assign")}</DialogTitle>
          </DialogHeader>
          <form className="grid gap-4" onSubmit={onAssign}>
            {roles.map((role) => (
              <div key={role.uid} className="flex items-center gap-2">
                <Checkbox
                  id={`assign-${role.uid}`}
                  checked={selected.includes(role.uid)}
                  onCheckedChange={(checked) => {
                    setSelected((current) =>
                      checked === true
                        ? [...current, role.uid]
                        : current.filter((uid) => uid !== role.uid)
                    )
                  }}
                />
                <Label htmlFor={`assign-${role.uid}`}>{role.name}</Label>
              </div>
            ))}
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending || selected.length === 0}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <DetailSheet
        open={detail !== null}
        title={detail?.email ?? t("action.detail")}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={[
          { label: t("member.email"), value: detail?.email ?? "" },
          { label: t("user.name"), value: detail?.name ?? "" },
          {
            label: t("member.role"),
            value: (detail?.roleNames ?? []).join(", "),
          },
          { label: t("member.status"), value: detail?.status ?? "" },
          { label: t("detail.uid"), value: detail?.uid ?? "" },
        ]}
      />
      <ConfirmAction
        open={removing !== null}
        title={t("confirm.deleteTitle")}
        description={t("confirm.deleteDescription")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void onDelete()}
        onOpenChange={(value) => {
          if (!value) {
            setRemoving(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
