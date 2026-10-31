import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import { banUser, deleteUser, listUsers, permitUser, updateUserSystemRole } from "@/api/user"
import type { UserItem } from "@/api/user/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { cn } from "@/lib/class-names"
import { collectFieldErrors, hasFieldErrors, requireSelected } from "@/lib/form"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
import { ListFrame } from "@/components/list-frame"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"

// 平台角色选项：卡片单选，带一句能力说明。系统只读是新增角色（可查看平台
// 全部信息、不可修改），用醒目的提示区分，避免误授予写权限。
const SYSTEM_ROLE_OPTIONS = [
  {
    value: "SYSTEM_ADMIN",
    labelKey: "user.systemAdmin",
    descriptionKey: "user.systemAdminDescription",
  },
  {
    value: "SYSTEM_READONLY",
    labelKey: "user.systemReadonly",
    descriptionKey: "user.systemReadonlyDescription",
    highlight: true,
  },
  {
    value: "USER",
    labelKey: "user.normalUser",
    descriptionKey: "user.normalUserDescription",
  },
] as const

export function UserPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [detail, setDetail] = useState<UserItem | null>(null)
  const [banning, setBanning] = useState<UserItem | null>(null)
  const [deleting, setDeleting] = useState<UserItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
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

  // 删除用于清理错误账号。确认弹窗明确不可恢复；后端还会拒绝租户负责人和有订单的用户，
  // 失败原因通过 messageKeys 映射成中文提示。
  async function onDelete() {
    if (!deleting) {
      return
    }
    setDeletePending(true)
    setDeleteError("")
    try {
      await deleteUser(deleting.uid)
      setDeleting(null)
      reload()
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
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
      actions.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setDeleteError("")
          setDeleting(item)
        },
      })
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
          setRole(
            item.systemRole === "SYSTEM_ADMIN" ||
              item.systemRole === "SYSTEM_READONLY"
              ? item.systemRole
              : "USER"
          )
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
      <ConfirmAction
        open={deleting !== null}
        title={t("confirm.deleteUserTitle")}
        description={t("confirm.deleteUserDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={deletePending}
        error={deleteError}
        onConfirm={() => void onDelete()}
        onOpenChange={(value) => {
          if (!value) {
            setDeleting(null)
            setDeleteError("")
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
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onGrant}>
            <DialogBody className="grid gap-4">
              <FormField
                label={t("user.systemRole")}
                required
                error={fieldErrors.role}
              >
                <fieldset className="grid gap-2" role="radiogroup" aria-label={t("user.systemRole")}>
                  <legend className="sr-only">{t("user.systemRole")}</legend>
                  {SYSTEM_ROLE_OPTIONS.map((option) => {
                    const selected = role === option.value
                    const highlight = "highlight" in option && option.highlight
                    return (
                      <label
                        key={option.value}
                        className={cn(
                          "flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors",
                          selected
                            ? "border-primary bg-primary/5"
                            : "border-border hover:bg-muted/50",
                          highlight &&
                            !selected &&
                            "border-violet-border bg-violet-soft/60"
                        )}
                      >
                        <input
                          type="radio"
                          name="system-role"
                          value={option.value}
                          checked={selected}
                          onChange={() => setRole(option.value)}
                          className="mt-1 size-4 shrink-0 accent-primary focus-visible:ring-2 focus-visible:ring-ring"
                        />
                        <span className="grid min-w-0 gap-0.5">
                          <span className="flex flex-wrap items-center gap-2">
                            <span className="text-sm font-medium">
                              {t(option.labelKey)}
                            </span>
                            <Badge
                              variant="outline"
                              className="px-1.5 py-0 font-mono text-[10px] font-normal text-muted-foreground"
                            >
                              {option.value}
                            </Badge>
                            {highlight ? (
                              <Badge
                                variant="outline"
                                className="border-violet-border bg-violet-soft px-1.5 py-0 font-normal text-violet"
                              >
                                {t("user.systemReadonlyBadge")}
                              </Badge>
                            ) : null}
                          </span>
                          <span className="text-xs break-words text-muted-foreground">
                            {t(option.descriptionKey)}
                          </span>
                        </span>
                      </label>
                    )
                  })}
                  <p className="rounded-lg border border-border bg-muted/40 px-3 py-2 text-xs break-words text-muted-foreground">
                    {t("user.grantRoleHint")}
                  </p>
                </fieldset>
              </FormField>
            </DialogBody>
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
