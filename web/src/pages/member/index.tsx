import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"

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
import {
  collectFieldErrors,
  hasFieldErrors,
  requireSelected,
  requireTrimmed,
} from "@/lib/form"
import { enumLabel } from "@/lib/enum-label"
import { usePagedList } from "@/hooks/use-paged-list"
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
  const [roles, setRoles] = useState<RoleOption[]>([])
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState("")
  const [roleUid, setRoleUid] = useState("")
  const [assigning, setAssigning] = useState<MemberItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<MemberItem | null>(null)
  const [removing, setRemoving] = useState<MemberItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<MemberItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listMembers({ page, pageSize, name: query.name }),
  })

  useEffect(() => {
    let active = true
    void selectRoles(100)
      .then((options) => {
        if (!active) return
        const available = (options.items ?? []).filter((item) => item.name !== "owner")
        setRoles(available)
        setRoleUid((current) => current || available[0]?.uid || "")
      })
      .catch(() => {
        if (active) setRoles([])
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  async function onInvite(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["email", requireTrimmed(email, t("form.required"))],
      ["role", requireSelected(roleUid, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await inviteMember({ email: email.trim(), roleUid })
      setOpen(false)
      setEmail("")
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onAssign(event: React.FormEvent) {
    event.preventDefault()
    if (!assigning) {
      return
    }
    const errors = collectFieldErrors([
      ["roles", selected.length === 0 ? t("form.required") : null],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await updateMemberRoles(assigning.uid, selected)
      setAssigning(null)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  function closeInvite(value: boolean) {
    setOpen(value)
    if (!value) {
      setEmail("")
      setError("")
      setFieldErrors({})
    }
  }

  function closeAssign(value: boolean) {
    if (!value) {
      setAssigning(null)
      setSelected([])
      setError("")
      setFieldErrors({})
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
          setFieldErrors({})
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
      reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
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
              setFieldErrors({})
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
              <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
              <Button
                variant="outline"
                onClick={() => reload()}
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
                    <TableCell>{enumLabel(t, item.status)}</TableCell>
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
          <form noValidate className="grid gap-4" onSubmit={onInvite}>
            <FormField
              label={t("member.email")}
              htmlFor="invite-email"
              required
              error={fieldErrors.email}
            >
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
              required
              error={fieldErrors.role}
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
              <Button type="submit" disabled={pending}>
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
          <form noValidate className="grid gap-4" onSubmit={onAssign}>
            <FormField
              label={t("member.role")}
              required
              error={fieldErrors.roles}
            >
              <div className="grid gap-2">
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
              </div>
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>
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
          { label: t("member.status"), value: enumLabel(t, detail?.status ?? "") },
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
