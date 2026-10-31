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
  updateRoleRemark,
  updateRoleStatus,
} from "@/api/role"
import type {
  PermissionItem,
  PermissionMenu,
  RoleItem,
} from "@/api/role/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { PermissionEditor } from "@/components/permission-editor"
import { PermissionSummary } from "@/components/permission-summary"
import { RoleTypeTag } from "@/components/role-type-tag"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TABLE_COLUMN_WIDTH, TABLE_MIN_WIDTH } from "@/lib/table-layout"
import { collectFieldErrors, hasFieldErrors, requireTrimmed } from "@/lib/form"
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
import { Textarea } from "@/components/ui/textarea"
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

// 备注按字符数（而不是 UTF-16 长度）校验，emoji 不会被算成两个。
const REMARK_MAX = 5000

function remarkLength(value: string): number {
  return Array.from(value).length
}

export function RolePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [catalog, setCatalog] = useState<PermissionItem[]>([])
  // 菜单目录由后端返回，授权页按「分组 → 菜单 → 子权限」渲染。
  const [menus, setMenus] = useState<PermissionMenu[]>([])
  const [catalogStatus, setCatalogStatus] = useState<"loading" | "ready" | "failed">("loading")
  const [catalogTick, setCatalogTick] = useState(0)
  const [open, setOpen] = useState(false)
  const [name, setName] = useState("")
  const [remark, setRemark] = useState("")
  const [editing, setEditing] = useState<RoleItem | null>(null)
  const [granting, setGranting] = useState<RoleItem | null>(null)
  const [error, setError] = useState("")
  // 行内操作（启用/停用）的错误独立于表单错误，避免弹窗和页面同时报同一句话。
  const [actionError, setActionError] = useState("")
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
        setCatalog(permissions.items ?? [])
        setMenus(permissions.menus ?? [])
        setCatalogStatus("ready")
      })
      .catch(() => {
        if (!active) return
        setCatalog([])
        setMenus([])
        setCatalogStatus("failed")
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid, catalogTick])

  function validateRemark(value: string): string | null {
    return remarkLength(value.trim()) > REMARK_MAX ? t("role.remarkTooLong") : null
  }

  async function onCreate(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["remark", validateRemark(remark)],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await createRole({ name: name.trim(), remark: remark.trim() })
      setOpen(false)
      setName("")
      setRemark("")
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onSaveEdit(event: React.FormEvent) {
    event.preventDefault()
    if (!editing) {
      return
    }
    const errors = collectFieldErrors([
      // 系统角色只写备注，名称由后端锁死，前端也不再校验。
      ["name", editing.isSystem ? null : requireTrimmed(name, t("form.required"))],
      ["remark", validateRemark(remark)],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      // 备注先写：改名失败时备注不会白填；两个请求都是全量覆盖，重试幂等。
      await updateRoleRemark({ uid: editing.uid, remark: remark.trim() })
      if (!editing.isSystem && name.trim() !== editing.name) {
        await updateRole({ uid: editing.uid, name: name.trim() })
      }
      setEditing(null)
      setName("")
      setRemark("")
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
      setRemark("")
      setError("")
      setFieldErrors({})
    }
  }

  function closeEdit(value: boolean) {
    if (!value) {
      setEditing(null)
      setName("")
      setRemark("")
      setError("")
      setFieldErrors({})
    }
  }

  function openEdit(item: RoleItem) {
    setEditing(item)
    setName(item.name)
    setRemark(item.remark ?? "")
    setError("")
    setFieldErrors({})
  }

  function roleActions(item: RoleItem): RowAction[] {
    if (!session.can("role:read")) {
      return []
    }
    // 系统角色：权限只读查看，但备注是运维说明，允许编辑（名称由后端锁死）。
    if (item.isSystem) {
      const actions: RowAction[] = [
        {
          key: "permissions",
          label: t("action.permissions"),
          onSelect: () => setDetail(item),
        },
      ]
      if (session.can("role:write")) {
        actions.push({
          key: "edit",
          label: t("action.edit"),
          onSelect: () => openEdit(item),
        })
      }
      return actions
    }
    if (!session.can("role:write")) {
      return []
    }
    return [
      {
        key: "edit",
        label: t("action.edit"),
        onSelect: () => openEdit(item),
      },
      {
        key: "permissions",
        label: t("action.permissions"),
        onSelect: () => {
          setGranting(item)
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
                setActionError,
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

  const canWrite = session.can("role:write")
  const remarkCount = remarkLength(remark)

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("role.title")} description={t("role.description")}>
        {canWrite ? (
          <Button
            onClick={() => {
              setName("")
              setRemark("")
              setError("")
              setFieldErrors({})
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
      {actionError ? <p className="text-sm text-destructive">{actionError}</p> : null}
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
          <p className="text-sm text-muted-foreground">{t("role.empty")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("role.name")}</TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.type}>{t("role.type")}</TableHead>
                <TableHead className="w-[18rem]">{t("role.remark")}</TableHead>
                <TableStatusHead>{t("namespace.status")}</TableStatusHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="max-w-0 whitespace-normal break-words font-medium">
                    {item.name}
                  </TableCell>
                  <TableCell>
                    <RoleTypeTag isSystem={item.isSystem} />
                  </TableCell>
                  <TableCell className="max-w-0 whitespace-normal">
                    {item.remark?.trim() ? (
                      // 列表只露两行，超出省略；不挂 Tooltip，全文去详情里看。
                      <span className="line-clamp-2 break-words text-muted-foreground">
                        {item.remark}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
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
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onCreate}>
            <DialogBody className="grid gap-4">
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
              <FormField
                label={t("role.remark")}
                htmlFor="role-remark"
                hint={t("role.remarkHint")}
                error={fieldErrors.remark}
              >
                <Textarea
                  id="role-remark"
                  value={remark}
                  rows={3}
                  maxLength={REMARK_MAX}
                  placeholder={t("role.remarkPlaceholder")}
                  className="resize-none"
                  onChange={(event) => setRemark(event.target.value)}
                />
                <p className="text-right text-xs text-muted-foreground">
                  {t("role.remarkCount", { count: remarkCount })}
                </p>
              </FormField>
            </DialogBody>
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
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onSaveEdit}>
            <DialogBody className="grid gap-4">
              <FormField
                label={t("role.name")}
                htmlFor="role-edit-name"
                required={!editing?.isSystem}
                hint={editing?.isSystem ? t("role.systemNameLocked") : undefined}
                error={fieldErrors.name}
              >
                <Input
                  id="role-edit-name"
                  value={name}
                  placeholder={t("role.namePlaceholder")}
                  disabled={editing?.isSystem}
                  onChange={(event) => setName(event.target.value)}
                />
              </FormField>
              <FormField
                label={t("role.remark")}
                htmlFor="role-edit-remark"
                hint={t("role.remarkHint")}
                error={fieldErrors.remark}
              >
                <Textarea
                  id="role-edit-remark"
                  value={remark}
                  rows={3}
                  maxLength={REMARK_MAX}
                  placeholder={t("role.remarkPlaceholder")}
                  className="resize-none"
                  onChange={(event) => setRemark(event.target.value)}
                />
                <p className="text-right text-xs text-muted-foreground">
                  {t("role.remarkCount", { count: remarkCount })}
                </p>
              </FormField>
            </DialogBody>
            {error ? <p className="text-sm text-destructive">{error}</p> : null}
            <DialogFooter>
              <Button type="submit" disabled={pending}>
                {t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <PermissionEditor
        key={granting?.uid ?? "none"}
        open={granting !== null}
        role={granting}
        catalog={catalog}
        menus={menus}
        catalogLoading={catalogStatus === "loading"}
        catalogFailed={catalogStatus === "failed"}
        onRetryCatalog={() => {
          setCatalogStatus("loading")
          setCatalogTick((value) => value + 1)
        }}
        onOpenChange={(value) => {
          if (!value) setGranting(null)
        }}
        onSave={async (permissions) => {
          if (!granting) return
          await updateRolePermissions(granting.uid, permissions)
          setGranting(null)
          reload()
        }}
      />

      <DetailSheet
        open={detail !== null}
        wide
        title={detail?.name ?? t("action.detail")}
        description={detail?.isSystem ? t("role.grant.readonly") : undefined}
        onOpenChange={(value) => {
          if (!value) setDetail(null)
        }}
        fields={[
          { label: t("role.name"), value: detail?.name ?? "" },
          {
            label: t("role.type"),
            value: detail ? <RoleTypeTag isSystem={detail.isSystem} /> : "",
          },
          { label: t("namespace.status"), value: <StatusBadge value={detail?.status ?? ""} /> },
          {
            label: t("role.remark"),
            value: (
              // 详情看全文：长备注在自己的滚动区里，不把权限分组挤出屏幕。
              <span className="block max-h-64 overflow-y-auto whitespace-pre-wrap break-words">
                {detail?.remark?.trim() || t("role.remarkEmpty")}
              </span>
            ),
          },
          {
            label: t("action.permissions"),
            span: 2,
            value: (
              <PermissionSummary
                permissions={detail?.permissions ?? []}
                catalog={catalog}
                menus={menus}
              />
            ),
          },
          { label: t("detail.uid"), value: detail?.uid ?? "" },
        ]}
      />

      <ConfirmAction
        open={confirm !== null}
        title={
          confirm?.type === "delete" ? t("confirm.deleteTitle") : t("confirm.disableTitle")
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
