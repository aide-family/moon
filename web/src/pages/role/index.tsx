import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import {
  createRole,
  deleteRole,
  listPermissions,
  listRoles,
  updateRole,
  updateRolePermissions,
  updateRoleStatus,
} from "@/api/role"
import type { PermissionItem, RoleItem } from "@/api/role/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { collectFieldErrors, hasFieldErrors, requireTrimmed } from "@/lib/form"
import { usePagedList } from "@/hooks/use-paged-list"
import { useListQuery } from "@/hooks/use-list-query"
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { runAction } from "@/lib/action"

export function RolePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [catalog, setCatalog] = useState<PermissionItem[]>([])
  const [open, setOpen] = useState(false)
  const [name, setName] = useState("")
  const [editing, setEditing] = useState<RoleItem | null>(null)
  const [granting, setGranting] = useState<RoleItem | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<RoleItem | null>(null)
  const [confirm, setConfirm] = useState<{
    type: "delete" | "disable"
    item: RoleItem
  } | null>(null)
  const [confirmError, setConfirmError] = useState("")
  const { items, total, loading, failed, error: loadError, reload } = usePagedList<RoleItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${query.name}`,
    fetch: (page, pageSize) => listRoles({ page, pageSize, name: query.name }),
  })

  useEffect(() => {
    let active = true
    void listPermissions()
      .then((permissions) => {
        if (!active) return
        setCatalog(
          (permissions.items ?? []).filter(
            (item) => item.name !== "namespace:delete" && item.name !== "namespace:transfer",
          ),
        )
      })
      .catch(() => {
        if (active) setCatalog([])
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid])

  async function onCreate(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await createRole({ name: name.trim() })
      setOpen(false)
      setName("")
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onRename(event: React.FormEvent) {
    event.preventDefault()
    if (!editing) {
      return
    }
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await updateRole({ uid: editing.uid, name: name.trim() })
      setEditing(null)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
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
      ["permissions", selected.length === 0 ? t("form.required") : null],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await updateRolePermissions(granting.uid, selected)
      setGranting(null)
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }


  function closeCreate(value: boolean) {
    setOpen(value)
    if (!value) {
      setName("")
      setError("")
      setFieldErrors({})
    }
  }

  function closeEdit(value: boolean) {
    if (!value) {
      setEditing(null)
      setName("")
      setError("")
      setFieldErrors({})
    }
  }

  function closeGrant(value: boolean) {
    if (!value) {
      setGranting(null)
      setSelected([])
      setError("")
      setFieldErrors({})
    }
  }

  function roleActions(item: RoleItem): RowAction[] {
    if (!session.can("role:write") || item.isSystem) {
      return []
    }
    return [
      {
        key: "edit",
        label: t("action.edit"),
        onSelect: () => {
          setEditing(item)
          setName(item.name)
          setError("")
        },
      },
      {
        key: "permissions",
        label: t("action.permissions"),
        onSelect: () => {
          setGranting(item)
          setSelected(item.permissions ?? [])
          setError("")
          setFieldErrors({})
        },
      },
      item.status === "ACTIVE"
        ? {
            key: "disable",
            label: t("action.disable"),
            destructive: true,
            onSelect: () => {
              setConfirmError("")
              setConfirm({ type: "disable", item })
            },
          }
        : {
            key: "enable",
            label: t("action.enable"),
            onSelect: () =>
              runAction(
                updateRoleStatus(item.uid, "ACTIVE").then(() => reload()),
                setError,
              ),
          },
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirm({ type: "delete", item })
        },
      },
    ]
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm.type === "delete") {
        await deleteRole(confirm.item.uid)
      } else {
        await updateRoleStatus(confirm.item.uid, "DISABLED")
      }
      setConfirm(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("role.title")} description={t("role.description")}>
        {session.can("role:write") ? (
          <Button
            onClick={() => {
              setName("")
              setError("")
              setOpen(true)
            }}
          >
            {t("action.create")}
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
          placeholder={t("role.name")}
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
            <p className="text-sm text-muted-foreground">{t("role.empty")}</p>
          ) : null}
          {!loading && !failed && items.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("role.name")}</TableHead>
                  <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                  <TableActionsHead>{t("action.actions")}</TableActionsHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {items.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>
                      {item.name}
                      {item.isSystem ? ` (${t("role.system")})` : ""}
                    </TableCell>
                    <TableCell><StatusBadge value={item.status} /></TableCell>
                    <TableCell>
                      <RowActions nowrap
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => setDetail(item)}
                        actions={roleActions(item)}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={closeCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onCreate}>
            <FormField
              label={t("role.name")}
              htmlFor="role-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="role-name"
                value={name}
                placeholder={t("role.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
                required
              />
            </FormField>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>
                {t("action.create")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
      <Dialog open={editing !== null} onOpenChange={closeEdit}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.edit")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onRename}>
            <FormField
              label={t("role.name")}
              htmlFor="role-edit-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="role-edit-name"
                value={name}
                placeholder={t("role.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
                required
              />
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
      <Dialog open={granting !== null} onOpenChange={closeGrant}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.permissions")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onGrant}>
            <div className="grid min-h-0 flex-1 content-start gap-4 overflow-y-auto">
            <FormField
              label={t("action.permissions")}
              hint={t("role.permissionHint")}
              required
              error={fieldErrors.permissions}
            >
              <div className="grid gap-2">
                {catalog.map((item) => (
                  <div key={item.uid} className="flex items-center gap-2">
                    <Checkbox
                      id={`perm-${item.uid}`}
                      checked={selected.includes(item.name)}
                      onCheckedChange={(checked) => {
                        setSelected((current) =>
                          checked === true
                            ? [...current, item.name]
                            : current.filter((name) => name !== item.name)
                        )
                      }}
                    />
                    <Label htmlFor={`perm-${item.uid}`}>{item.name}</Label>
                  </div>
                ))}
              </div>
            </FormField>
            </div>
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
        title={detail?.name ?? t("action.detail")}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={[
          { label: t("role.name"), value: detail?.name ?? "" },
          { label: t("namespace.status"), value: <StatusBadge value={detail?.status ?? ""} /> },
          {
            label: t("role.system"),
            value: detail?.isSystem ? t("role.system") : "",
          },
          {
            label: t("action.permissions"),
            value: (detail?.permissions ?? []).join(", "),
          },
          { label: t("detail.uid"), value: detail?.uid ?? "" },
        ]}
      />
      <ConfirmAction
        open={confirm !== null}
        title={
          confirm?.type === "delete"
            ? t("confirm.deleteTitle")
            : t("confirm.disableTitle")
        }
        description={
          confirm?.type === "delete"
            ? t("confirm.deleteDescription")
            : t("confirm.disableDescription")
        }
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
