import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Building2 } from "lucide-react"
import { formErrorMessage, readFormError } from "@/lib/form-error"

import { listMembers } from "@/api/member"
import type { MemberItem } from "@/api/member/types"
import {
  deleteNamespace,
  listNamespaces,
  transferNamespaceOwner,
  updateNamespace,
} from "@/api/namespace"
import type { NamespaceItem } from "@/api/namespace/types"
import {
  createOrder,
  createRenewalOrder,
  createUpgradeOrder,
  getNamespacePackage,
  listPublicPackages,
  payOrder,
} from "@/api/order"
import type {
  NamespacePackageInfo,
  OrderItem,
  PublicPackageItem,
} from "@/api/order/types"
import { cn } from "@/lib/class-names"
import { formatPriceCents } from "@/lib/price"
import { toast } from "sonner"
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
  // 创建租户分两步：先填资料并选套餐生成待支付订单，再虚拟支付完成创建。
  const [createStep, setCreateStep] = useState<"form" | "pay">("form")
  const [packages, setPackages] = useState<PublicPackageItem[]>([])
  const [packagesState, setPackagesState] = useState<
    "idle" | "loading" | "ready" | "failed"
  >("idle")
  const [packageUid, setPackageUid] = useState("")
  const [order, setOrder] = useState<OrderItem | null>(null)
  // 当前租户套餐：升级/续期入口展示名称、价格、人数上限和到期时间。
  const [currentPackage, setCurrentPackage] = useState<NamespacePackageInfo | null>(null)
  const [packageReload, setPackageReload] = useState(0)
  const [upgradeOpen, setUpgradeOpen] = useState(false)
  const [renewOpen, setRenewOpen] = useState(false)
  // 打开续费弹窗时固定「现在」，渲染阶段不再调用 Date.now（保持组件纯函数）。
  const [renewBaseMs, setRenewBaseMs] = useState(0)
  const [upgradePackages, setUpgradePackages] = useState<PublicPackageItem[]>([])
  const [upgradePackagesState, setUpgradePackagesState] = useState<
    "idle" | "loading" | "ready" | "failed"
  >("idle")
  const [upgradeUid, setUpgradeUid] = useState("")
  const [packagePending, setPackagePending] = useState(false)
  const [packageError, setPackageError] = useState("")
  const [editing, setEditing] = useState<string | null>(null)
  const [transferUid, setTransferUid] = useState<string | null>(null)
  const [members, setMembers] = useState<MemberItem[]>([])
  const [target, setTarget] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [detail, setDetail] = useState<NamespaceItem | null>(null)
  const [confirm, setConfirm] = useState<
    { type: "delete"; uid: string } | { type: "transfer" } | null
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

  // 切换租户后重新读取当前套餐；升级/续期支付成功后也会刷新。
  useEffect(() => {
    const uid = session.namespaceUid
    if (!uid) {
      return
    }
    let active = true
    void getNamespacePackage(uid, { silent: true })
      .then((result) => {
        if (active) {
          setCurrentPackage(result)
        }
      })
      .catch(() => {
        if (active) {
          setCurrentPackage(null)
        }
      })
    return () => {
      active = false
    }
  }, [session.namespaceUid, packageReload])

  async function openUpgrade() {
    setPackageError("")
    setUpgradeUid("")
    setUpgradeOpen(true)
    setUpgradePackagesState("loading")
    try {
      const result = await listPublicPackages()
      const items = (result.items ?? []).filter(
        (item) => item.uid !== currentPackage?.packageUid
      )
      setUpgradePackages(items)
      setUpgradePackagesState("ready")
      setUpgradeUid(items[0]?.uid ?? "")
    } catch {
      setUpgradePackages([])
      setUpgradePackagesState("failed")
    }
  }

  function openRenew() {
    setPackageError("")
    setRenewBaseMs(Date.now())
    setRenewOpen(true)
  }

  // 升级/续期都是一次「下单 + 虚拟支付」：支付失败时订单保留在订单页，可以稍后支付。
  async function runPackageOrder(
    create: () => Promise<{ order: OrderItem }>,
    successMessage: string
  ) {
    setPackagePending(true)
    setPackageError("")
    try {
      const { order: placed } = await create()
      await payOrder(placed.uid)
      toast.success(successMessage)
      setUpgradeOpen(false)
      setRenewOpen(false)
      setPackageReload((value) => value + 1)
      refresh()
    } catch (err) {
      setPackageError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPackagePending(false)
    }
  }

  // 套餐到期时间：0 / 缺省表示永久。续期从当前到期时间（未过期）或现在起算。
  function packageExpiryText(): string {
    if (!currentPackage?.expiresAt) {
      return t("namespace.permanent")
    }
    return new Date(currentPackage.expiresAt * 1000).toLocaleDateString()
  }

  function renewalTargetText(): string {
    if (!currentPackage) {
      return "—"
    }
    const base = new Date(
      Math.max(renewBaseMs, (currentPackage.expiresAt ?? 0) * 1000)
    )
    base.setDate(base.getDate() + (currentPackage.durationDays ?? 0))
    return base.toLocaleDateString()
  }

  // 打开创建弹窗时拉取公开套餐；失败时留在表单里让用户重试，不阻塞填写名称。
  async function loadPackages() {
    setPackagesState("loading")
    try {
      const result = await listPublicPackages()
      const items = result.items ?? []
      setPackages(items)
      setPackagesState("ready")
      setPackageUid((current) =>
        current && items.some((item) => item.uid === current)
          ? current
          : (items[0]?.uid ?? "")
      )
    } catch {
      setPackages([])
      setPackagesState("failed")
    }
  }

  function openCreate() {
    setName("")
    setRemark("")
    setError("")
    setFieldErrors({})
    setCreateStep("form")
    setOrder(null)
    setPackageUid("")
    setOpen(true)
    void loadPackages()
  }

  // 第一步：校验收资料与套餐，生成待支付订单；租户要等支付成功才创建。
  async function onCreate(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      ["name", requireTrimmed(name, t("form.required"))],
      ["package", requireSelected(packageUid, t("form.required"))],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      const result = await createOrder({
        name: name.trim(),
        remark: remark.trim(),
        packageUid,
      })
      setOrder(result.order)
      setCreateStep("pay")
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  // 第二步：虚拟支付。后端在同一个事务里创建租户、绑定套餐并把订单置为已支付。
  async function onPay() {
    if (!order) {
      return
    }
    setPending(true)
    setError("")
    try {
      const result = await payOrder(order.uid)
      localStorage.setItem("namespace", result.namespace.uid)
      toast.success(t("namespace.created"))
      setOpen(false)
      refresh()
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
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
      // 关闭时丢弃本地订单状态；已经生成的待支付订单仍会出现在订单页，可以继续支付。
      setName("")
      setRemark("")
      setError("")
      setFieldErrors({})
      setCreateStep("form")
      setOrder(null)
      setPackageUid("")
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
        <Button className="shrink-0 self-start" onClick={openCreate}>
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
          </div>
        </section>
      </div>

      <section
        aria-label={t("namespace.packageSection")}
        className="glass-card flex flex-wrap items-center justify-between gap-3 rounded-xl p-4"
      >
        <div className="min-w-0">
          <p className="text-xs font-medium tracking-[0.12em] text-muted-foreground uppercase">
            {t("namespace.packageSection")}
          </p>
          <p className="mt-1 truncate text-sm font-semibold">
            {currentPackage?.packageName || "—"}
          </p>
          <p className="mt-0.5 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
            <span>
              {t("namespace.expiresAt")}：{packageExpiryText()}
            </span>
            <span>
              {t("namespace.maxMembers")}：
              {currentPackage?.maxMembers
                ? t("namespace.membersValue", {
                    count: currentPackage.maxMembers,
                  })
                : t("namespace.membersUnlimited")}
            </span>
            <span className="tabular-nums">
              {currentPackage
                ? (currentPackage.priceCents ?? 0) > 0
                  ? t("namespace.priceValue", {
                      amount: formatPriceCents(
                        currentPackage.priceCents,
                        currentPackage.currency
                      ),
                    })
                  : t("namespace.free")
                : ""}
            </span>
          </p>
        </div>
        {session.can("namespace:write") &&
        currentNamespace?.status === "ACTIVE" ? (
          <div className="flex shrink-0 flex-wrap gap-2">
            <Button variant="outline" size="sm" onClick={() => void openUpgrade()}>
              {t("namespace.upgrade")}
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={(currentPackage?.durationDays ?? 0) <= 0}
              title={
                (currentPackage?.durationDays ?? 0) <= 0
                  ? t("namespace.renewPermanentHint")
                  : undefined
              }
              onClick={openRenew}
            >
              {t("namespace.renew")}
            </Button>
          </div>
        ) : null}
      </section>

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
      <Dialog
        open={upgradeOpen}
        onOpenChange={(value) => {
          if (!value && !packagePending) {
            setUpgradeOpen(false)
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{t("namespace.upgradeTitle")}</DialogTitle>
          </DialogHeader>
          <div className="flex min-h-0 flex-1 flex-col gap-4">
            <DialogBody className="grid gap-3">
              <p className="text-sm text-muted-foreground">
                {t("namespace.upgradeHint")}
              </p>
              {upgradePackagesState === "loading" ? (
                <p className="text-sm text-muted-foreground">
                  {t("state.loading")}
                </p>
              ) : null}
              {upgradePackagesState === "failed" ? (
                <div className="flex items-center justify-between gap-2 rounded-lg border border-destructive/25 bg-destructive/5 p-3">
                  <p className="text-sm text-destructive">
                    {t("namespace.packageLoadFailed")}
                  </p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void openUpgrade()}
                  >
                    {t("action.retry")}
                  </Button>
                </div>
              ) : null}
              {upgradePackagesState === "ready" &&
              upgradePackages.length === 0 ? (
                <p className="rounded-lg border border-dashed border-border p-3 text-sm text-muted-foreground">
                  {t("namespace.noUpgradePackages")}
                </p>
              ) : null}
              {upgradePackagesState === "ready" && upgradePackages.length > 0 ? (
                <div
                  aria-label={t("namespace.upgradeTitle")}
                  className="grid gap-2"
                  role="radiogroup"
                >
                  {upgradePackages.map((item) => {
                    const checked = upgradeUid === item.uid
                    return (
                      <button
                        aria-checked={checked}
                        className={cn(
                          "flex min-w-0 flex-col gap-1.5 rounded-lg border p-3 text-left transition-colors",
                          checked
                            ? "border-primary bg-primary/5"
                            : "border-border hover:bg-muted/40"
                        )}
                        key={item.uid}
                        role="radio"
                        type="button"
                        onClick={() => setUpgradeUid(item.uid)}
                      >
                        <span className="flex min-w-0 flex-wrap items-center gap-2">
                          <span className="text-sm font-semibold">
                            {item.name}
                          </span>
                          <span className="ml-auto font-mono text-sm font-semibold tabular-nums">
                            {(item.priceCents ?? 0) > 0
                              ? t("namespace.priceValue", {
                                  amount: formatPriceCents(
                                    item.priceCents,
                                    item.currency
                                  ),
                                })
                              : t("namespace.free")}
                          </span>
                        </span>
                        <span className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
                          <span>
                            {(item.durationDays ?? 0) > 0
                              ? t("namespace.durationValue", {
                                  count: item.durationDays,
                                })
                              : t("namespace.permanent")}
                          </span>
                          <span>
                            {item.maxMembers
                              ? t("namespace.membersValue", {
                                  count: item.maxMembers,
                                })
                              : t("namespace.membersUnlimited")}
                          </span>
                          <span>
                            {t("namespace.featureCount", {
                              count: item.features?.length ?? 0,
                            })}
                          </span>
                        </span>
                      </button>
                    )
                  })}
                </div>
              ) : null}
            </DialogBody>
            {packageError ? (
              <p className="text-sm text-destructive">{packageError}</p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                disabled={packagePending}
                onClick={() => setUpgradeOpen(false)}
              >
                {t("action.cancel")}
              </Button>
              <Button
                type="button"
                disabled={
                  packagePending ||
                  upgradePackagesState !== "ready" ||
                  !upgradeUid
                }
                onClick={() =>
                  void runPackageOrder(
                    () => createUpgradeOrder(upgradeUid),
                    t("namespace.upgraded")
                  )
                }
              >
                {packagePending ? t("namespace.paying") : t("namespace.payUpgrade")}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog
        open={renewOpen}
        onOpenChange={(value) => {
          if (!value && !packagePending) {
            setRenewOpen(false)
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t("namespace.renewTitle")}</DialogTitle>
          </DialogHeader>
          <div className="flex min-h-0 flex-1 flex-col gap-4">
            <DialogBody className="grid gap-3">
              <p className="text-sm text-muted-foreground">
                {t("namespace.renewHint")}
              </p>
              <dl className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4 sm:grid-cols-2">
                <div className="min-w-0">
                  <dt className="text-xs text-muted-foreground">
                    {t("namespace.package")}
                  </dt>
                  <dd className="mt-0.5 text-sm font-medium break-words">
                    {currentPackage?.packageName || "—"}
                  </dd>
                </div>
                <div className="min-w-0">
                  <dt className="text-xs text-muted-foreground">
                    {t("namespace.amount")}
                  </dt>
                  <dd className="mt-0.5 font-mono text-lg font-semibold tabular-nums">
                    {(currentPackage?.priceCents ?? 0) > 0
                      ? t("namespace.priceValue", {
                          amount: formatPriceCents(
                            currentPackage?.priceCents,
                            currentPackage?.currency
                          ),
                        })
                      : t("namespace.free")}
                  </dd>
                </div>
                <div className="min-w-0">
                  <dt className="text-xs text-muted-foreground">
                    {t("namespace.expiresAt")}
                  </dt>
                  <dd className="mt-0.5 text-sm">{packageExpiryText()}</dd>
                </div>
                <div className="min-w-0">
                  <dt className="text-xs text-muted-foreground">
                    {t("namespace.renewTo")}
                  </dt>
                  <dd className="mt-0.5 text-sm">{renewalTargetText()}</dd>
                </div>
              </dl>
            </DialogBody>
            {packageError ? (
              <p className="text-sm text-destructive">{packageError}</p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                disabled={packagePending}
                onClick={() => setRenewOpen(false)}
              >
                {t("action.cancel")}
              </Button>
              <Button
                type="button"
                disabled={packagePending}
                onClick={() =>
                  void runPackageOrder(
                    () => createRenewalOrder(),
                    t("namespace.renewed")
                  )
                }
              >
                {packagePending ? t("namespace.paying") : t("namespace.payRenew")}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={open} onOpenChange={closeCreate}>
        <DialogContent className={createStep === "form" ? "sm:max-w-2xl" : undefined}>
          <DialogHeader>
            <DialogTitle>
              {createStep === "form"
                ? t("namespace.create")
                : t("namespace.payTitle")}
            </DialogTitle>
          </DialogHeader>
          {createStep === "form" ? (
            <form
              noValidate
              className="flex min-h-0 flex-1 flex-col gap-4"
              onSubmit={onCreate}
            >
              <DialogBody className="grid gap-4">
                <div className="grid gap-4 sm:grid-cols-2">
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
                </div>
                <div className="grid min-w-0 gap-2">
                  <div className="flex flex-wrap items-baseline gap-2">
                    <span className="text-sm font-medium">
                      {t("namespace.package")}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {t("namespace.packageHint")}
                    </span>
                  </div>
                  {packagesState === "loading" ? (
                    <p className="text-sm text-muted-foreground">
                      {t("state.loading")}
                    </p>
                  ) : null}
                  {packagesState === "failed" ? (
                    <div
                      className="flex flex-col items-start gap-2 rounded-lg border border-destructive/25 bg-destructive/5 p-3 sm:flex-row sm:items-center sm:justify-between"
                      role="alert"
                    >
                      <p className="text-sm text-destructive">
                        {t("namespace.packageLoadFailed")}
                      </p>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => void loadPackages()}
                      >
                        {t("action.retry")}
                      </Button>
                    </div>
                  ) : null}
                  {packagesState === "ready" && packages.length === 0 ? (
                    <p className="rounded-lg border border-dashed border-border p-3 text-sm text-muted-foreground">
                      {t("namespace.noPublicPackages")}
                    </p>
                  ) : null}
                  {packagesState === "ready" && packages.length > 0 ? (
                    <div
                      aria-label={t("namespace.package")}
                      className="grid gap-2"
                      role="radiogroup"
                    >
                      {packages.map((item) => {
                        const checked = packageUid === item.uid
                        return (
                          <button
                            aria-checked={checked}
                            className={cn(
                              "flex min-w-0 flex-col gap-1.5 rounded-lg border p-3 text-left transition-colors",
                              checked
                                ? "border-primary bg-primary/5"
                                : "border-border hover:bg-muted/40"
                            )}
                            key={item.uid}
                            role="radio"
                            type="button"
                            onClick={() => setPackageUid(item.uid)}
                          >
                            <span className="flex min-w-0 flex-wrap items-center gap-2">
                              <span className="text-sm font-semibold">
                                {item.name}
                              </span>
                              {item.isBase ? (
                                <Badge
                                  className="h-5 px-1.5 text-[10px]"
                                  variant="secondary"
                                >
                                  {t("namespace.packageBase")}
                                </Badge>
                              ) : null}
                              <span className="ml-auto font-mono text-sm font-semibold tabular-nums">
                                {(item.priceCents ?? 0) > 0
                                  ? t("namespace.priceValue", {
                                      amount: formatPriceCents(
                                        item.priceCents,
                                        item.currency
                                      ),
                                    })
                                  : t("namespace.free")}
                              </span>
                            </span>
                            <span className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
                              <span>
                                {(item.durationDays ?? 0) > 0
                                  ? t("namespace.durationValue", {
                                      count: item.durationDays,
                                    })
                                  : t("namespace.permanent")}
                              </span>
                              <span>
                                {item.maxMembers
                                  ? t("namespace.membersValue", {
                                      count: item.maxMembers,
                                    })
                                  : t("namespace.membersUnlimited")}
                              </span>
                              <span>
                                {t("namespace.featureCount", {
                                  count: item.features?.length ?? 0,
                                })}
                              </span>
                            </span>
                            {item.remark?.trim() ? (
                              <span className="line-clamp-2 text-xs text-muted-foreground">
                                {item.remark}
                              </span>
                            ) : null}
                          </button>
                        )
                      })}
                    </div>
                  ) : null}
                </div>
              </DialogBody>
              {error ? (
                <p className="text-sm text-destructive">{error}</p>
              ) : null}
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => closeCreate(false)}
                >
                  {t("action.cancel")}
                </Button>
                <Button
                  type="submit"
                  disabled={
                    pending ||
                    packagesState !== "ready" ||
                    packages.length === 0
                  }
                >
                  {pending ? t("action.saving") : t("namespace.createOrder")}
                </Button>
              </DialogFooter>
            </form>
          ) : (
            <div className="flex min-h-0 flex-1 flex-col gap-4">
              <DialogBody className="grid gap-4">
                <p className="text-sm text-muted-foreground">
                  {t("namespace.payHint")}
                </p>
                <dl className="grid gap-3 rounded-lg border border-border bg-muted/20 p-4 sm:grid-cols-2">
                  <div className="min-w-0">
                    <dt className="text-xs text-muted-foreground">
                      {t("namespace.name")}
                    </dt>
                    <dd className="mt-0.5 text-sm font-medium break-words">
                      {order?.namespaceName ?? name}
                    </dd>
                  </div>
                  <div className="min-w-0">
                    <dt className="text-xs text-muted-foreground">
                      {t("namespace.package")}
                    </dt>
                    <dd className="mt-0.5 text-sm font-medium break-words">
                      {order?.packageName ?? "—"}
                    </dd>
                  </div>
                  <div className="min-w-0">
                    <dt className="text-xs text-muted-foreground">
                      {t("namespace.amount")}
                    </dt>
                    <dd className="mt-0.5 font-mono text-lg font-semibold tabular-nums">
                      {(order?.priceCents ?? 0) > 0
                        ? t("namespace.priceValue", {
                            amount: formatPriceCents(
                              order?.priceCents,
                              order?.currency
                            ),
                          })
                        : t("namespace.free")}
                    </dd>
                  </div>
                  <div className="min-w-0">
                    <dt className="text-xs text-muted-foreground">
                      {t("namespace.duration")}
                    </dt>
                    <dd className="mt-0.5 text-sm">
                      {(order?.durationDays ?? 0) > 0
                        ? t("namespace.durationValue", {
                            count: order?.durationDays,
                          })
                        : t("namespace.permanent")}
                    </dd>
                  </div>
                  <div className="min-w-0">
                    <dt className="text-xs text-muted-foreground">
                      {t("namespace.maxMembers")}
                    </dt>
                    <dd className="mt-0.5 text-sm">
                      {order?.maxMembers
                        ? t("namespace.membersValue", {
                            count: order.maxMembers,
                          })
                        : t("namespace.membersUnlimited")}
                    </dd>
                  </div>
                </dl>
              </DialogBody>
              {error ? (
                <p className="text-sm text-destructive">{error}</p>
              ) : null}
              <DialogFooter>
                <Button
                  type="button"
                  variant="outline"
                  disabled={pending}
                  onClick={() => {
                    setError("")
                    setCreateStep("form")
                  }}
                >
                  {t("action.goBack")}
                </Button>
                <Button
                  type="button"
                  disabled={pending || !order}
                  onClick={() => void onPay()}
                >
                  {pending ? t("namespace.paying") : t("namespace.pay")}
                </Button>
              </DialogFooter>
            </div>
          )}
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
          {
            label: t("namespace.package"),
            value: detail?.packageName || "—",
          },
          {
            label: t("namespace.members"),
            value: detail
              ? `${detail.memberCount ?? 0} / ${
                  detail.maxMembers
                    ? detail.maxMembers
                    : t("namespace.membersUnlimited")
                }`
              : "—",
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
            : t("confirm.transferTitle")
        }
        description={
          confirm?.type === "delete"
            ? t("confirm.deleteDescription")
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
