import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Building2 } from "lucide-react"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import { listMembers } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import {
  createNamespace,
  deleteNamespace,
  listNamespaces,
  transferNamespaceOwner,
  updateNamespace,
  updateNamespaceStatus,
} from "@/api/namespace"
import type { NamespaceItem } from "@/api/namespace/types"
import { TABLE_GRID_MAIN_IDENTITY_STATUS_ACTIONS } from "@/lib/table-layout"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { ListFrame } from "@/components/list-frame"
import { Badge } from "@/components/ui/badge"
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  collectFieldErrors,
  hasFieldErrors,
  requireSelected,
  requireTrimmed,
} from "@/lib/form"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"

export function NamespacePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  // 表头和每一行都是独立的 grid 容器，必须共用同一份列模板才能对齐。
  const rowGrid = TABLE_GRID_MAIN_IDENTITY_STATUS_ACTIONS
  const [draft, setDraft] = useState(query.name)
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

  // 列表走服务端分页：page / name 来自 URL，和成员、角色等列表页同一套约定。
  // 还没选定租户时没有 X-Namespace 可带，接口必然 401，所以直接给空列表，
  // 让「创建第一个租户」这条路仍然可用。
  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<NamespaceItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: query.name,
    fetch: (page, pageSize) =>
      session.namespaceUid
        ? listNamespaces({ page, pageSize, name: query.name })
        : Promise.resolve({ items: [], total: 0 }),
  })

  // 页码可能来自历史 URL，或被删除、筛选后的结果集收缩：越界时回到最后一页，
  // 否则列表空着但分页控件已被隐藏，用户只能手改 URL。
  useEffect(() => {
    if (loading || failed || query.page <= 1) {
      return
    }
    const lastPage = Math.max(1, Math.ceil(total / query.pageSize))
    if (
      lastPage !== query.page &&
      (items.length === 0 || query.page > lastPage)
    ) {
      query.update({ page: lastPage })
    }
    // query.update 每次渲染都是新函数，这里用页码与总数做真正的依赖。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, failed, items.length, total, query.page, query.pageSize])

  // 租户写操作既影响分页列表，也影响会话里的租户集（切换器、统计和权限），两处都要刷新。
  function refresh() {
    reload()
    session.reload()
  }

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
      refresh()
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
      refresh()
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
    refresh()
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
      refresh()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
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

  const currentNamespace = session.namespaces.find(
    (item) => item.uid === session.namespaceUid
  )
  // 统计口径跟着会话里的完整租户集走，不受列表筛选与分页影响。
  const activeCount = session.namespaces.filter(
    (item) => item.status === "ACTIVE"
  ).length
  const disabledCount = session.namespaces.filter(
    (item) => item.status === "DISABLED"
  ).length

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-5">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            {t("namespace.title")}
          </h1>
          <p className="text-sm text-muted-foreground">
            {t("namespace.description")}
          </p>
        </div>
        <Button
          className="shrink-0 self-start"
          onClick={() => {
            setName("")
            setRemark("")
            setError("")
            setFieldErrors({})
            setOpen(true)
          }}
        >
          <Building2 aria-hidden="true" className="mr-2 size-4" />
          {t("namespace.create")}
        </Button>
      </header>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(17rem,0.8fr)]">
        <section
          aria-label={t("namespace.switcher.currentSection")}
          className="relative isolate overflow-hidden rounded-xl border border-primary/20 bg-primary/5 p-5 text-foreground shadow-sm"
        >
          <div
            aria-hidden="true"
            className="pointer-events-none absolute -top-14 -right-10 size-44 rounded-full border border-primary/10"
          >
            <span className="absolute inset-5 rounded-full border border-primary/10" />
            <span className="absolute inset-10 rounded-full border border-primary/10" />
          </div>
          <div className="relative flex min-w-0 items-center gap-4">
            <span className="grid size-12 shrink-0 place-items-center rounded-xl bg-primary text-primary-foreground ring-1 ring-primary/20">
              <Building2 aria-hidden="true" className="size-5" />
            </span>
            <div className="min-w-0">
              <p className="text-xs font-medium tracking-wide text-muted-foreground">
                {t("namespace.switcher.currentSection")}
              </p>
              <h2 className="mt-1 truncate font-heading text-xl font-semibold tracking-tight">
                {currentNamespace?.name || t("namespace.none")}
              </h2>
              {currentNamespace ? (
                <p className="mt-1 truncate text-sm text-muted-foreground">
                  {currentNamespace.ownerName ||
                    currentNamespace.ownerEmail ||
                    currentNamespace.ownerUid}
                </p>
              ) : null}
            </div>
          </div>
          <span
            aria-hidden="true"
            className="absolute top-5 right-5 size-2 rounded-full bg-primary ring-4 ring-primary/15"
          />
        </section>

        <section
          aria-label={t("namespace.status")}
          className="glass-card flex min-w-0 flex-col justify-center gap-3 rounded-xl p-4 sm:flex-row sm:items-center sm:justify-around sm:gap-5 xl:flex-col xl:items-stretch xl:gap-3"
        >
          <div className="min-w-0 sm:text-center xl:text-left">
            <p className="text-xs font-medium tracking-[0.12em] text-muted-foreground uppercase">
              {t("namespace.switcher.listLabel")}
            </p>
            <p className="mt-1 font-mono text-2xl font-semibold tracking-tight tabular-nums">
              {session.loading ? "—" : session.namespaces.length}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2 sm:justify-center xl:justify-start">
            <span className="inline-flex items-center gap-2 rounded-lg bg-muted/60 px-3 py-2">
              <StatusBadge value="ACTIVE" />
              <span className="font-mono text-sm font-semibold tabular-nums">
                {session.loading ? "—" : activeCount}
              </span>
            </span>
            <span className="inline-flex items-center gap-2 rounded-lg bg-muted/60 px-3 py-2">
              <StatusBadge value="DISABLED" />
              <span className="font-mono text-sm font-semibold tabular-nums">
                {session.loading ? "—" : disabledCount}
              </span>
            </span>
          </div>
        </section>
      </div>

      <form
        className="flex gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input
          className="max-w-sm"
          aria-label={t("namespace.switcher.search")}
          placeholder={t("namespace.switcher.searchPlaceholder")}
          value={draft}
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
          <div
            aria-busy="true"
            aria-live="polite"
            className="grid gap-3"
            role="status"
          >
            <span className="sr-only">{t("state.loading")}</span>
            {[0, 1, 2].map((item) => (
              <div
                aria-hidden="true"
                className="h-[4.5rem] animate-pulse rounded-lg bg-muted/70 motion-reduce:animate-none"
                key={item}
              />
            ))}
          </div>
        ) : null}
        {!loading && failed ? (
          <div
            className="flex flex-col items-start gap-3 rounded-lg border border-destructive/25 bg-destructive/5 p-5 sm:flex-row sm:items-center sm:justify-between"
            role="alert"
          >
            <p className="text-sm text-destructive">
              {loadError || t("state.failed")}
            </p>
            <Button variant="outline" onClick={() => reload()}>
              {t("action.retry")}
            </Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 && query.name ? (
          <div className="flex min-h-44 flex-col items-center justify-center px-5 py-8 text-center">
            <p className="text-sm font-medium">
              {t("namespace.switcher.noResults")}
            </p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("namespace.switcher.noResultsHint", { query: query.name })}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 && !query.name ? (
          <div className="flex min-h-44 flex-col items-center justify-center rounded-lg border border-dashed border-border px-5 py-8 text-center">
            <span className="mb-3 grid size-11 place-items-center rounded-xl bg-primary/10 text-primary-ink">
              <Building2 aria-hidden="true" className="size-5" />
            </span>
            <p className="max-w-md text-sm text-muted-foreground">
              {t("namespace.guide")}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden rounded-lg border border-border/70">
            <div
              className={`hidden shrink-0 items-center gap-4 bg-muted/35 px-5 py-3 text-[11px] font-medium tracking-[0.1em] text-muted-foreground uppercase md:grid ${rowGrid}`}
            >
              <span>{t("namespace.name")}</span>
              <span>{t("namespace.leader")}</span>
              <span>{t("namespace.status")}</span>
              <span className="text-right">{t("action.actions")}</span>
            </div>
            {/* 表体独立滚动，表头留在滚动区外：长列表不会再被 ListFrame 的 overflow-hidden 裁掉。 */}
            <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain">
              <ul
                aria-label={t("namespace.switcher.listLabel")}
                className="divide-y divide-border/70"
              >
                {items.map((item) => {
                  const isCurrent = item.uid === session.namespaceUid
                  return (
                    <li
                      className={`grid min-w-0 gap-3 p-4 md:items-center md:gap-4 md:px-5 ${rowGrid} ${isCurrent ? "bg-primary/[0.035]" : "bg-card/30"}`}
                      key={item.uid}
                    >
                      <div className="flex min-w-0 items-center gap-3">
                        <span
                          aria-hidden="true"
                          className={`relative grid size-10 shrink-0 place-items-center overflow-hidden rounded-lg ring-1 ${isCurrent ? "bg-primary/10 text-primary-ink ring-primary/20" : "bg-muted/70 text-muted-foreground ring-border/70"}`}
                        >
                          <Building2 className="size-4" />
                          {item.logo ? (
                            <img
                              alt=""
                              className="absolute inset-0 size-full object-cover"
                              loading="lazy"
                              onError={(event) => {
                                event.currentTarget.hidden = true
                              }}
                              src={item.logo}
                            />
                          ) : null}
                        </span>
                        <div className="min-w-0">
                          <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                            <h3 className="text-sm font-semibold break-words">
                              {item.name}
                            </h3>
                            {isCurrent ? (
                              <Badge
                                variant="secondary"
                                className="h-5 px-1.5 text-[10px]"
                              >
                                {t("namespace.current")}
                              </Badge>
                            ) : null}
                          </div>
                          {item.remark ? (
                            <p className="mt-1 truncate text-xs text-muted-foreground">
                              {item.remark}
                            </p>
                          ) : null}
                        </div>
                      </div>

                      <div className="min-w-0 pl-[3.25rem] md:pl-0">
                        <p className="mb-1 text-[11px] font-medium tracking-wide text-muted-foreground uppercase md:hidden">
                          {t("namespace.leader")}
                        </p>
                        <span className="flex min-w-0 flex-col">
                          <span className="text-sm break-all">
                            {item.ownerName || item.ownerEmail || item.ownerUid}
                          </span>
                          {item.ownerName && item.ownerEmail ? (
                            <span className="text-xs break-all text-muted-foreground">
                              {item.ownerEmail}
                            </span>
                          ) : null}
                        </span>
                      </div>

                      <div className="flex items-center gap-2 pl-[3.25rem] md:pl-0">
                        <span className="text-[11px] font-medium tracking-wide text-muted-foreground uppercase md:hidden">
                          {t("namespace.status")}
                        </span>
                        <StatusBadge value={item.status} />
                      </div>

                      <div className="flex min-w-0 justify-start md:justify-end">
                        <RowActions
                          detailLabel={t("action.detail")}
                          moreLabel={t("action.more")}
                          onDetail={() => setDetail(item)}
                          actions={namespaceActions(item)}
                          nowrap
                        />
                      </div>
                    </li>
                  )
                })}
              </ul>
            </div>
          </div>
        ) : null}
      </ListFrame>
      <Dialog open={open} onOpenChange={closeCreate}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("namespace.create")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onCreate}>
            <DialogBody className="grid gap-4">
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
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onRename}>
            <DialogBody className="grid gap-4">
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
      <Dialog open={transferUid !== null} onOpenChange={closeTransfer}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("action.transfer")}</DialogTitle>
          </DialogHeader>
          <form noValidate className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={requestTransfer}>
            <DialogBody className="grid gap-4">
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
          {
            label: t("namespace.leader"),
            value: (
              <span className="flex flex-col">
                <span>
                  {detail?.ownerName || detail?.ownerEmail || detail?.ownerUid}
                </span>
                {detail?.ownerName && detail?.ownerEmail ? (
                  <span className="text-xs text-muted-foreground">
                    {detail.ownerEmail}
                  </span>
                ) : null}
              </span>
            ),
          },
          {
            label: t("namespace.status"),
            value: <StatusBadge value={detail?.status ?? ""} />,
          },
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
