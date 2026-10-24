import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import { banUser, listUsers, permitUser, updateUserSystemRole } from "@/api/user"
import type { UserItem } from "@/api/user/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { collectFieldErrors, hasFieldErrors, requireSelected } from "@/lib/form"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { ListFrame } from "@/components/list-frame"
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
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"

export function UserPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [detail, setDetail] = useState<UserItem | null>(null)
  const [banning, setBanning] = useState<UserItem | null>(null)
  const [granting, setGranting] = useState<UserItem | null>(null)
  const [role, setRole] = useState("")
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [confirmError, setConfirmError] = useState("")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<UserItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: query.name,
    fetch: (page, pageSize) => listUsers({ page, pageSize, name: query.name }),
  })

  if (!session.loading && !session.canSystem("system:user:read")) {
    return <Navigate to="/forbidden" replace />
  }

  async function onPermit(item: UserItem) {
    await permitUser(item.uid)
    reload()
  }

  async function onBan() {
    if (!banning) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await banUser(banning.uid)
      setBanning(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function onGrant(event: React.FormEvent) {
    event.preventDefault()
    if (!granting) {
      return
    }
    const errors = collectFieldErrors([
      ["role", requireSelected(role, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await updateUserSystemRole(granting.uid, role)
      setGranting(null)
      setRole("")
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  function userActions(item: UserItem): RowAction[] {
    const actions: RowAction[] = []
    if (
      session.canSystem("system:user:write") &&
      item.uid !== session.user?.uid &&
      item.systemRole !== "SUPER_ADMIN"
    ) {
      if (item.status === "ACTIVE") {
        actions.push({
          key: "ban",
          label: t("action.ban"),
          destructive: true,
          onSelect: () => {
            setConfirmError("")
            setBanning(item)
          },
        })
      } else {
        actions.push({
          key: "permit",
          label: t("action.permit"),
          onSelect: () => void onPermit(item),
        })
      }
    }
    if (
      session.isSuperAdmin &&
      item.uid !== session.user?.uid &&
      item.systemRole !== "SUPER_ADMIN"
    ) {
      actions.push({
        key: "grant",
        label: t("user.grantRole"),
        onSelect: () => {
          setGranting(item)
          setRole(item.systemRole === "SYSTEM_ADMIN" ? "SYSTEM_ADMIN" : "USER")
          setError("")
          setFieldErrors({})
        },
      })
    }
    return actions
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("user.title")} description={t("user.description")} />
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
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("user.empty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("user.name")}</TableHead>
                <TableHead>{t("login.emailLabel")}</TableHead>
                <TableHead>{t("user.systemRole")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>{item.name}</TableCell>
                  <TableCell>{item.email}</TableCell>
                  <TableCell><StatusBadge value={item.systemRole} /></TableCell>
                  <TableCell><StatusBadge value={item.status} /></TableCell>
                  <TableCell>
                    <RowActions nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => setDetail(item)}
                      actions={userActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <DetailSheet
        open={detail !== null}
        title={detail?.name || detail?.email || t("action.detail")}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={[
          { label: t("user.name"), value: detail?.name ?? "" },
          { label: t("user.nickname"), value: detail?.nickname ?? "" },
          { label: t("login.emailLabel"), value: detail?.email ?? "" },
          { label: t("user.systemRole"), value: <StatusBadge value={detail?.systemRole ?? ""} /> },
          { label: t("namespace.status"), value: <StatusBadge value={detail?.status ?? ""} /> },
          { label: t("profile.remark"), value: detail?.remark ?? "" },
          { label: t("detail.uid"), value: detail?.uid ?? "" },
        ]}
      />
      <ConfirmAction
        open={banning !== null}
        title={t("confirm.banTitle")}
        description={t("confirm.banDescription")}
        confirmLabel={t("action.confirm")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void onBan()}
        onOpenChange={(value) => {
          if (!value) {
            setBanning(null)
            setConfirmError("")
          }
        }}
      />
      <Dialog
        open={granting !== null}
        onOpenChange={(value) => {
          if (!value) {
            setGranting(null)
            setError("")
            setFieldErrors({})
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("user.grantRole")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onGrant}>
            <FormField
              label={t("user.systemRole")}
              htmlFor="system-role"
              hint={t("user.grantRoleHint")}
              required
              error={fieldErrors.role}
            >
              <Select value={role} onValueChange={setRole}>
                <SelectTrigger id="system-role">
                  <SelectValue placeholder={t("user.systemRolePlaceholder")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="SYSTEM_ADMIN">{t("user.systemAdmin")}</SelectItem>
                  <SelectItem value="USER">{t("user.normalUser")}</SelectItem>
                </SelectContent>
              </Select>
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => setGranting(null)}
              >
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </section>
  )
}
