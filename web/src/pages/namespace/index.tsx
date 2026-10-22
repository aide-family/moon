import { useState } from "react"
import { useTranslation } from "react-i18next"
import { formMessage, readFormError } from "@/lib/form-error"

import { listMembers } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import {
  createNamespace,
  deleteNamespace,
  transferNamespaceOwner,
  updateNamespace,
  updateNamespaceStatus,
} from "@/api/namespace"
import type { NamespaceItem } from "@/api/namespace/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { ListFrame } from "@/components/list-frame"
import { Badge } from "@/components/ui/badge"
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
import { enumLabel } from "@/lib/enum-label"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireSelected,
  requireTrimmed,
} from "@/lib/form"

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
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NamespaceItem | null>(null)
  const [confirm, setConfirm] = useState<
    { type: "delete" | "disable"; uid: string } | { type: "transfer" } | null
  >(null)
  const [confirmError, setConfirmError] = useState("")

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
      const result = await createNamespace({ name: name.trim(), remark })
      localStorage.setItem("namespace", result.namespace.uid)
      setOpen(false)
      setName("")
      setRemark("")
      session.reload()
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
      localStorage.setItem("namespace", editing)
      await updateNamespace({ uid: editing, name: name.trim(), remark })
      setEditing(null)
      session.reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onStatus(uid: string) {
    localStorage.setItem("namespace", uid)
    await updateNamespaceStatus(uid, "ACTIVE")
    session.reload()
  }

  async function runConfirm() {
    if (!confirm) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      if (confirm.type === "delete") {
        localStorage.setItem("namespace", confirm.uid)
        await deleteNamespace(confirm.uid)
        localStorage.removeItem("namespace")
      } else if (confirm.type === "disable") {
        localStorage.setItem("namespace", confirm.uid)
        await updateNamespaceStatus(confirm.uid, "DISABLED")
      } else if (transferUid && target) {
        localStorage.setItem("namespace", transferUid)
        await transferNamespaceOwner(transferUid, target)
        setTransferUid(null)
        setTarget("")
      }
      setConfirm(null)
      session.reload()
    } catch (err) {
      setConfirmError(formMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function openTransfer(uid: string) {
    localStorage.setItem("namespace", uid)
    setTransferUid(uid)
    setError("")
    const result = await listMembers({ page: 1, pageSize: 100 })
    const items = (result.items ?? []).filter(
      (item) => item.userUid !== session.user?.uid && item.status === "JOINED"
    )
    setMembers(items)
    setTarget(items[0]?.userUid ?? "")
  }

  function requestTransfer(event: React.FormEvent) {
    event.preventDefault()
    if (!transferUid) {
      return
    }
    const errors = collectFieldErrors([
      ["target", requireSelected(target, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setConfirmError("")
    setConfirm({ type: "transfer" })
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

  function closeTransfer(value: boolean) {
    if (!value) {
      setTransferUid(null)
      setTarget("")
      setError("")
      setFieldErrors({})
    }
  }

  function namespaceActions(item: NamespaceItem): RowAction[] {
    if (item.uid !== session.namespaceUid) {
      return []
    }
    const actions: RowAction[] = []
    if (session.can("namespace:write")) {
      actions.push({
        key: "edit",
        label: t("action.edit"),
        onSelect: () => {
          setEditing(item.uid)
          setName(item.name)
          setRemark(item.remark)
          setError("")
          setFieldErrors({})
        },
      })
      if (item.status === "ACTIVE") {
        actions.push({
          key: "disable",
          label: t("action.disable"),
          destructive: true,
          onSelect: () => {
            setConfirmError("")
            setConfirm({ type: "disable", uid: item.uid })
          },
        })
      } else {
        actions.push({
          key: "enable",
          label: t("action.enable"),
          onSelect: () => void onStatus(item.uid),
        })
      }
    }
    if (session.can("namespace:transfer")) {
      actions.push({
        key: "transfer",
        label: t("action.transfer"),
        onSelect: () => void openTransfer(item.uid),
      })
    }
    if (session.can("namespace:delete")) {
      actions.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirm({ type: "delete", uid: item.uid })
        },
      })
    }
    return actions
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("namespace.title")}
        description={t("namespace.description")}
      >
        <Button
          onClick={() => {
            setName("")
            setRemark("")
            setError("")
            setFieldErrors({})
            setOpen(true)
          }}
        >
          {t("namespace.create")}
        </Button>
      </PageHeader>
      <ListFrame
        total={session.namespaces.length}
        page={1}
        pageSize={Math.max(session.namespaces.length, 1)}
        showPager={!session.loading && !session.failed && session.namespaces.length > 0}
        onPrev={() => undefined}
        onNext={() => undefined}
      >
      {session.loading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!session.loading && session.failed ? (
        <div className="flex items-center gap-2">
          <p className="text-sm text-destructive">{session.loadError || t("state.failed")}</p>
          <Button variant="outline" onClick={session.reload}>
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
      {!session.loading &&
      !session.failed &&
      session.namespaces.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("namespace.guide")}</p>
      ) : null}
      {!session.loading && !session.failed && session.namespaces.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>{t("namespace.name")}</TableHead>
                  <TableHead>{t("namespace.status")}</TableHead>
                  <TableHead className="text-right">{t("action.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {session.namespaces.map((item) => (
                  <TableRow key={item.uid}>
                    <TableCell>
                      <span className="flex items-center gap-2">
                        {item.name}
                        {item.uid === session.namespaceUid ? (
                          <Badge>{t("namespace.current")}</Badge>
                        ) : null}
                      </span>
                    </TableCell>
                    <TableCell>{enumLabel(t, item.status)}</TableCell>
                    <TableCell>
                      <RowActions
                        detailLabel={t("action.detail")}
                        moreLabel={t("action.more")}
                        onDetail={() => setDetail(item)}
                        actions={namespaceActions(item)}
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
            <DialogTitle>{t("namespace.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={onCreate}>
            <FormField
              label={t("namespace.name")}
              htmlFor="ns-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="ns-name"
                value={name}
                placeholder={t("namespace.namePlaceholder")}
                onChange={(event) => setName(event.target.value)}
                required
              />
            </FormField>
            <FormField label={t("namespace.remark")} htmlFor="ns-remark">
              <Input
                id="ns-remark"
                value={remark}
                placeholder={t("namespace.remarkPlaceholder")}
                onChange={(event) => setRemark(event.target.value)}
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
              label={t("namespace.name")}
              htmlFor="edit-name"
              required
              error={fieldErrors.name}
            >
              <Input
                id="edit-name"
                value={name}
                placeholder={t("namespace.namePlaceholder")}
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
      <Dialog open={transferUid !== null} onOpenChange={closeTransfer}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.transfer")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="grid gap-4" onSubmit={requestTransfer}>
            {members.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t("state.empty")}
              </p>
            ) : (
              <FormField
                label={t("member.email")}
                hint={t("namespace.transferHint")}
                required
                error={fieldErrors.target}
              >
                <Select value={target || undefined} onValueChange={setTarget}>
                  <SelectTrigger
                    className="w-full"
                    aria-label={t("member.email")}
                  >
                    <SelectValue
                      placeholder={t("namespace.memberPlaceholder")}
                    />
                  </SelectTrigger>
                  <SelectContent>
                    {members.map((item) => (
                      <SelectItem key={item.uid} value={item.userUid}>
                        {item.email}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormField>
            )}
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
          { label: t("namespace.name"), value: detail?.name ?? "" },
          { label: t("namespace.status"), value: enumLabel(t, detail?.status ?? "") },
          { label: t("namespace.remark"), value: detail?.remark ?? "" },
          { label: t("detail.uid"), value: detail?.uid ?? "" },
        ]}
      />
      <ConfirmAction
        open={confirm !== null}
        title={
          confirm?.type === "delete"
            ? t("confirm.deleteTitle")
            : confirm?.type === "disable"
              ? t("confirm.disableTitle")
              : t("confirm.transferTitle")
        }
        description={
          confirm?.type === "delete"
            ? t("confirm.deleteDescription")
            : confirm?.type === "disable"
              ? t("confirm.disableDescription")
              : t("confirm.transferDescription")
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
