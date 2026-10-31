import { Package } from "lucide-react"
import { useEffect, useState } from "react"
import type { TFunction } from "i18next"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import {
  createPackage,
  deletePackage,
  listFeatures,
  listPackageNamespaces,
  listPackages,
  syncPackage,
  updatePackage,
} from "@/api/package"
import type {
  FeatureItem,
  PackageItem,
  PackageNamespaceItem,
} from "@/api/package/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FeaturePicker } from "@/components/feature-picker"
import { FeatureTags } from "@/components/feature-tags"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { cn } from "@/lib/class-names"
import {
  defaultCapabilityCodes,
  sortCapabilityCodes,
} from "@/lib/feature-catalog"
import { collectFieldErrors, hasFieldErrors, requireTrimmed } from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"
import {
  CURRENCIES,
  currencySymbol,
  formatPriceCents,
  type CurrencyCode,
} from "@/lib/price"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

const REMARK_MAX = 5000

type CatalogStatus = "loading" | "ready" | "failed"

/** 授权时长只接受 0（永久）或正整数天；非法输入返回 null，由表单报错。 */
function parseDurationDays(value: string): number | null {
  const trimmed = value.trim()
  if (!/^\d+$/.test(trimmed)) {
    return null
  }
  const days = Number(trimmed)
  return Number.isSafeInteger(days) ? days : null
}

/** 列表 / 详情里的时长文案：0 天就是永久。 */
function durationText(t: TFunction, days?: number): string {
  const value = days ?? 0
  return value > 0
    ? t("packageAdmin.durationValue", { count: value })
    : t("packageAdmin.permanent")
}

/** 人数上限只接受 0（不限）或正整数；非法输入返回 null，由表单报错。 */
function parseMaxMembers(value: string): number | null {
  const trimmed = value.trim()
  if (!/^\d+$/.test(trimmed)) {
    return null
  }
  const members = Number(trimmed)
  return Number.isSafeInteger(members) ? members : null
}

/**
 * 价格输入以「元」为单位，最多两位小数，提交时换算成人民币分。
 * 用字符串拆分计算，避免 Number 浮点误差把 12.30 算成 1229 分。
 */
function parsePriceCents(value: string): number | null {
  const trimmed = value.trim()
  if (!/^\d+(\.\d{1,2})?$/.test(trimmed)) {
    return null
  }
  const [yuan, fraction = ""] = trimmed.split(".")
  const cents = Number(yuan) * 100 + Number(fraction.padEnd(2, "0"))
  return Number.isSafeInteger(cents) ? cents : null
}

/** 列表 / 详情里的价格文案：0 分就是免费，其余按币种符号展示。 */
function priceText(t: TFunction, cents?: number, currency = "CNY"): string {
  const value = cents ?? 0
  if (value <= 0) {
    return t("packageAdmin.free")
  }
  return t("packageAdmin.priceValue", {
    amount: formatPriceCents(value, currency),
  })
}

// 套餐管理：平台级页面，超级管理员把功能组合成套餐，再在租户管理里分配。
// 基础套餐只有一个，名称固定、不能删除，但可以调整功能与备注。
export function PackagePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [catalog, setCatalog] = useState<FeatureItem[]>([])
  const [catalogStatus, setCatalogStatus] = useState<CatalogStatus>("loading")
  const [catalogTick, setCatalogTick] = useState(0)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<PackageItem | null>(null)
  const [name, setName] = useState("")
  const [remark, setRemark] = useState("")
  const [features, setFeatures] = useState<string[]>([])
  // 授权时长用字符串保存，空/非法输入在提交时统一校验，避免受控输入抖动。
  const [durationDays, setDurationDays] = useState("0")
  // 版本决定套餐出现在哪里：新套餐默认内测，确认要售卖后再改成公开。
  const [visibility, setVisibility] = useState<"PUBLIC" | "BETA">("BETA")
  // 人数上限 0 表示不限制；价格以「元」输入，提交时换算成分。
  const [maxMembers, setMaxMembers] = useState("0")
  const [currency, setCurrency] = useState<CurrencyCode>("CNY")
  const [priceYuan, setPriceYuan] = useState("0")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)
  const [confirming, setConfirming] = useState<PackageItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
  // 同步套餐授权：把套餐当前能力快照写回使用它的租户，支持全部或指定租户。
  const [syncing, setSyncing] = useState<PackageItem | null>(null)
  const [syncTenants, setSyncTenants] = useState<PackageNamespaceItem[]>([])
  const [syncLoading, setSyncLoading] = useState(false)
  const [syncFailed, setSyncFailed] = useState(false)
  const [syncScope, setSyncScope] = useState<"all" | "selected">("all")
  const [syncSelected, setSyncSelected] = useState<string[]>([])
  const [syncPending, setSyncPending] = useState(false)
  const [syncError, setSyncError] = useState("")
  // 只有 system:package:read 的账号（如系统管理员）走只读详情，不打开可编辑弹窗。
  const [detail, setDetail] = useState<PackageItem | null>(null)

  const canRead = session.canSystem("system:package:read")
  const canWrite = session.canSystem("system:package:write")

  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<PackageItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: query.name,
    enabled: !session.loading && canRead,
    fetch: (page, pageSize) =>
      listPackages({ page, pageSize, name: query.name }),
  })

  // 功能目录只影响弹窗里的勾选项，加载失败时用本地目录兜底，不阻塞列表。
  useEffect(() => {
    if (!canRead) {
      return
    }
    let active = true
    void listFeatures({ silent: true })
      .then((result) => {
        if (!active) {
          return
        }
        setCatalog(result.items ?? [])
        setCatalogStatus("ready")
      })
      .catch(() => {
        if (!active) {
          return
        }
        setCatalog([])
        setCatalogStatus("failed")
      })
    return () => {
      active = false
    }
  }, [catalogTick, canRead])

  if (!session.loading && !canRead) {
    return <Navigate to="/forbidden" replace />
  }

  function retryCatalog() {
    setCatalogStatus("loading")
    setCatalogTick((value) => value + 1)
  }

  // 后端目录可用时以它为准；接口未部署或失败时回落到本地能力目录。
  const catalogCodes =
    catalogStatus === "ready" && catalog.length > 0
      ? catalog.map((item) => item.code)
      : defaultCapabilityCodes()
  // 正在编辑的套餐里可能有本地目录还没有的新能力，一并展示，避免保存时被清掉。
  const dialogCodes = sortCapabilityCodes([...catalogCodes, ...features])

  function openCreate() {
    setEditing(null)
    setName("")
    setRemark("")
    setFeatures([])
    setDurationDays("0")
    setVisibility("BETA")
    setMaxMembers("0")
    setCurrency("CNY")
    setPriceYuan("0")
    setFieldErrors({})
    setFormError("")
    setDialogOpen(true)
  }

  function openEdit(item: PackageItem) {
    setEditing(item)
    setName(item.name)
    setRemark(item.remark ?? "")
    setFeatures(item.features ?? [])
    setDurationDays(String(item.durationDays ?? 0))
    setVisibility(item.visibility === "PUBLIC" ? "PUBLIC" : "BETA")
    setMaxMembers(String(item.maxMembers ?? 0))
    setCurrency(
      CURRENCIES.includes((item.currency || "CNY") as CurrencyCode)
        ? ((item.currency || "CNY") as CurrencyCode)
        : "CNY"
    )
    setPriceYuan(((item.priceCents ?? 0) / 100).toFixed(2))
    setFieldErrors({})
    setFormError("")
    setDialogOpen(true)
  }

  function closeDialog(open: boolean) {
    setDialogOpen(open)
    if (!open) {
      setEditing(null)
      setName("")
      setRemark("")
      setFeatures([])
      setDurationDays("0")
      setVisibility("BETA")
      setMaxMembers("0")
      setCurrency("CNY")
      setPriceYuan("0")
      setFieldErrors({})
      setFormError("")
    }
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const duration = parseDurationDays(durationDays)
    const members = parseMaxMembers(maxMembers)
    const priceCents = parsePriceCents(priceYuan)
    const errors = collectFieldErrors([
      // 基础套餐名称由后端固定，前端不再校验改名。
      [
        "name",
        editing?.isBase ? null : requireTrimmed(name, t("form.required")),
      ],
      [
        "durationDays",
        duration === null ? t("packageAdmin.durationInvalid") : null,
      ],
      [
        "maxMembers",
        members === null ? t("packageAdmin.maxMembersInvalid") : null,
      ],
      ["price", priceCents === null ? t("packageAdmin.priceInvalid") : null],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setFormError("")
    setFieldErrors({})
    try {
      const payload = {
        name: editing?.isBase ? editing.name : name.trim(),
        remark: remark.trim(),
        features: sortCapabilityCodes(features),
        durationDays: duration ?? 0,
        visibility,
        maxMembers: members ?? 0,
        priceCents: priceCents ?? 0,
        currency,
      }
      if (editing) {
        await updatePackage({ uid: editing.uid, ...payload })
      } else {
        await createPackage(payload)
      }
      setDialogOpen(false)
      setEditing(null)
      toast.success(t("packageAdmin.saved"))
      reload()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setFormError(failure.form)
    } finally {
      setPending(false)
    }
  }

  async function onDelete() {
    if (!confirming) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await deletePackage(confirming.uid)
      setConfirming(null)
      toast.success(t("packageAdmin.deleted"))
      reload()
    } catch (err) {
      // 被租户使用的套餐由后端拒绝，这里原样展示后端给出的原因。
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  async function openSync(item: PackageItem) {
    setSyncing(item)
    setSyncScope("all")
    setSyncSelected([])
    setSyncError("")
    setSyncFailed(false)
    setSyncTenants([])
    setSyncLoading(true)
    try {
      const result = await listPackageNamespaces(item.uid)
      const items = result.items ?? []
      setSyncTenants(items)
      setSyncSelected(items.map((tenant) => tenant.uid))
    } catch {
      setSyncFailed(true)
    } finally {
      setSyncLoading(false)
    }
  }

  async function onSync() {
    if (!syncing) {
      return
    }
    if (syncScope === "selected" && syncSelected.length === 0) {
      setSyncError(t("packageAdmin.syncNeedSelection"))
      return
    }
    setSyncPending(true)
    setSyncError("")
    try {
      const result = await syncPackage(
        syncing.uid,
        syncScope === "all" ? [] : syncSelected
      )
      toast.success(
        t("packageAdmin.syncResult", {
          synced: result.synced,
          skipped: result.skipped,
        })
      )
      setSyncing(null)
      reload()
    } catch (err) {
      setSyncError(formErrorMessage(err, t("state.failed")))
    } finally {
      setSyncPending(false)
    }
  }

  function toggleSyncTenant(uid: string, checked: boolean) {
    setSyncSelected((current) =>
      checked ? [...new Set([...current, uid])] : current.filter((item) => item !== uid)
    )
  }

  const allSyncSelected =
    syncTenants.length > 0 &&
    syncTenants.every((tenant) => syncSelected.includes(tenant.uid))

  function packageActions(item: PackageItem): RowAction[] {
    if (!canWrite) {
      return []
    }
    const actions: RowAction[] = [
      {
        key: "sync",
        label: t("packageAdmin.syncAction"),
        onSelect: () => void openSync(item),
      },
    ]
    if (!item.isBase) {
      actions.push({
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirming(item)
        },
      })
    }
    return actions
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("packageAdmin.title")}
        description={t("packageAdmin.description")}
      >
        {canWrite ? (
          <Button onClick={openCreate}>{t("action.create")}</Button>
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
          aria-label={t("packageAdmin.search")}
          placeholder={t("packageAdmin.search")}
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
          <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
        ) : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
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
            <p className="text-sm font-medium">{t("packageAdmin.noResults")}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("packageAdmin.noResultsHint", { query: query.name })}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 && !query.name ? (
          <div className="flex min-h-44 flex-col items-center justify-center rounded-lg border border-dashed border-border px-5 py-8 text-center">
            <span className="mb-3 grid size-11 place-items-center rounded-xl bg-primary/10 text-primary-ink">
              <Package aria-hidden="true" className="size-5" />
            </span>
            <p className="text-sm font-medium">{t("packageAdmin.empty")}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("packageAdmin.emptyHint")}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("packageAdmin.name")}</TableHead>
                <TableHead className="w-[18rem]">
                  {t("packageAdmin.remark")}
                </TableHead>
                <TableHead className="w-[22rem]">
                  {t("packageAdmin.features")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("packageAdmin.type")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("packageAdmin.visibility")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("packageAdmin.price")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("packageAdmin.maxMembers")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("packageAdmin.duration")}
                </TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-medium break-words">
                    {item.name}
                  </TableCell>
                  <TableCell className="max-w-0 whitespace-normal">
                    {item.remark?.trim() ? (
                      <span className="line-clamp-2 break-words text-muted-foreground">
                        {item.remark}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell className="max-w-0 whitespace-normal">
                    <span className="mb-1 block text-xs text-muted-foreground tabular-nums">
                      {t("packageAdmin.featureCount", {
                        count: item.features?.length ?? 0,
                      })}
                    </span>
                    <FeatureTags
                      codes={item.features}
                      emptyLabel={t("packageAdmin.noFeatures")}
                    />
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={item.isBase ? "BASE" : "CUSTOM"}
                      label={
                        item.isBase
                          ? t("packageAdmin.base")
                          : t("packageAdmin.custom")
                      }
                      tone={item.isBase ? "primary" : "neutral"}
                    />
                  </TableCell>
                  <TableCell
                    className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                  >
                    <StatusBadge
                      value={item.visibility}
                      label={
                        item.visibility === "PUBLIC"
                          ? t("packageAdmin.visibilityPublic")
                          : t("packageAdmin.visibilityBeta")
                      }
                      tone={item.visibility === "PUBLIC" ? "primary" : "neutral"}
                    />
                  </TableCell>
                  <TableCell
                    className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                  >
                    <span className="tabular-nums">
                      {priceText(t, item.priceCents, item.currency)}
                    </span>
                  </TableCell>
                  <TableCell
                    className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                  >
                    <span className="tabular-nums">
                      {item.maxMembers
                        ? t("packageAdmin.maxMembersValue", {
                            count: item.maxMembers,
                          })
                        : t("packageAdmin.maxMembersUnlimited")}
                    </span>
                  </TableCell>
                  <TableCell
                    className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                  >
                    {durationText(t, item.durationDays)}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={
                        canWrite ? t("action.edit") : t("action.detail")
                      }
                      moreLabel={t("action.more")}
                      onDetail={() => {
                        if (canWrite) {
                          openEdit(item)
                        } else {
                          setDetail(item)
                        }
                      }}
                      actions={packageActions(item)}
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
        title={detail?.name ?? t("action.detail")}
        onOpenChange={(open) => {
          if (!open) {
            setDetail(null)
          }
        }}
        fields={[
          { label: t("packageAdmin.name"), value: detail?.name ?? "" },
          {
            label: t("packageAdmin.type"),
            value: detail ? (
              <StatusBadge
                value={detail.isBase ? "BASE" : "CUSTOM"}
                label={
                  detail.isBase
                    ? t("packageAdmin.base")
                    : t("packageAdmin.custom")
                }
                tone={detail.isBase ? "primary" : "neutral"}
              />
            ) : (
              ""
            ),
          },
          {
            label: t("packageAdmin.remark"),
            value: detail?.remark?.trim() || "—",
          },
          {
            label: t("packageAdmin.visibility"),
            value: detail ? (
              <StatusBadge
                value={detail.visibility}
                label={
                  detail.visibility === "PUBLIC"
                    ? t("packageAdmin.visibilityPublic")
                    : t("packageAdmin.visibilityBeta")
                }
                tone={detail.visibility === "PUBLIC" ? "primary" : "neutral"}
              />
            ) : (
              ""
            ),
          },
          {
            label: t("packageAdmin.price"),
            value: priceText(t, detail?.priceCents, detail?.currency),
          },
          {
            label: t("packageAdmin.maxMembers"),
            value: detail?.maxMembers
              ? t("packageAdmin.maxMembersValue", { count: detail.maxMembers })
              : t("packageAdmin.maxMembersUnlimited"),
          },
          {
            label: t("packageAdmin.duration"),
            value: durationText(t, detail?.durationDays),
          },
          {
            label: t("packageAdmin.features"),
            span: 2,
            value: (
              <FeatureTags
                codes={detail?.features}
                emptyLabel={t("packageAdmin.noFeatures")}
              />
            ),
          },
        ]}
      />

      <Dialog open={dialogOpen} onOpenChange={closeDialog}>
        <DialogContent className="sm:max-w-3xl">
          <DialogHeader>
            <DialogTitle>
              {editing
                ? t("packageAdmin.editTitle")
                : t("packageAdmin.createTitle")}
            </DialogTitle>
          </DialogHeader>
          <form
            noValidate
            className="flex min-h-0 flex-1 flex-col gap-4"
            onSubmit={onSave}
          >
            <DialogBody className="grid gap-4">
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField
                  label={t("packageAdmin.name")}
                  htmlFor="package-name"
                  required={!editing?.isBase}
                  hint={
                    editing?.isBase
                      ? t("packageAdmin.baseNameLocked")
                      : undefined
                  }
                  error={fieldErrors.name}
                >
                  <Input
                    id="package-name"
                    value={name}
                    disabled={editing?.isBase}
                    placeholder={t("packageAdmin.namePlaceholder")}
                    maxLength={64}
                    onChange={(event) => setName(event.target.value)}
                  />
                </FormField>
                <FormField
                  label={t("packageAdmin.durationDays")}
                  htmlFor="package-duration"
                  required
                  hint={t("packageAdmin.durationHint")}
                  error={fieldErrors.durationDays}
                >
                  <Input
                    id="package-duration"
                    type="number"
                    inputMode="numeric"
                    min={0}
                    step={1}
                    value={durationDays}
                    placeholder="0"
                    onChange={(event) => setDurationDays(event.target.value)}
                  />
                </FormField>
              </div>
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField
                  label={t("packageAdmin.visibility")}
                  hint={t("packageAdmin.visibilityHint")}
                >
                  <Select
                    value={visibility}
                    onValueChange={(value) =>
                      setVisibility(value === "PUBLIC" ? "PUBLIC" : "BETA")
                    }
                  >
                    <SelectTrigger
                      className="w-full"
                      aria-label={t("packageAdmin.visibility")}
                    >
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="PUBLIC">
                        {t("packageAdmin.visibilityPublic")}
                      </SelectItem>
                      <SelectItem value="BETA">
                        {t("packageAdmin.visibilityBeta")}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </FormField>
                <FormField
                  label={t("packageAdmin.maxMembers")}
                  htmlFor="package-max-members"
                  hint={t("packageAdmin.maxMembersHint")}
                  error={fieldErrors.maxMembers}
                >
                  <Input
                    id="package-max-members"
                    type="number"
                    inputMode="numeric"
                    min={0}
                    step={1}
                    value={maxMembers}
                    placeholder="0"
                    onChange={(event) => setMaxMembers(event.target.value)}
                  />
                </FormField>
              </div>
              <FormField
                label={t("packageAdmin.price")}
                htmlFor="package-price"
                hint={t("packageAdmin.priceHint")}
                error={fieldErrors.price}
              >
                <div className="flex h-8 w-full items-center rounded-md border border-input shadow-xs focus-within:ring-3 focus-within:ring-ring/50">
                  <Select
                    value={currency}
                    onValueChange={(value) =>
                      setCurrency(value as CurrencyCode)
                    }
                  >
                    <SelectTrigger
                      className="h-8 w-32 shrink-0 rounded-r-none border-0 bg-transparent px-2 shadow-none focus-visible:ring-0"
                      aria-label={t("packageAdmin.currency")}
                    >
                      <SelectValue>
                        <span className="flex items-center gap-1 tabular-nums">
                          <span>{currencySymbol(currency)}</span>
                          <span>{currency}</span>
                        </span>
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {CURRENCIES.map((code) => (
                        <SelectItem key={code} value={code}>
                          <span className="flex items-center gap-2">
                            <span className="w-4 tabular-nums">
                              {currencySymbol(code)}
                            </span>
                            <span>{t(`currency.${code}`)}</span>
                          </span>
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <span
                    aria-hidden="true"
                    className="h-4 w-px shrink-0 bg-border"
                  />
                  <Input
                    id="package-price"
                    type="text"
                    inputMode="decimal"
                    value={priceYuan}
                    placeholder="0.00"
                    className="h-8 min-w-0 flex-1 border-0 bg-transparent shadow-none focus-visible:ring-0"
                    onChange={(event) => setPriceYuan(event.target.value)}
                  />
                </div>
              </FormField>
              <FormField
                label={t("packageAdmin.remark")}
                htmlFor="package-remark"
                error={fieldErrors.remark}
              >
                <Textarea
                  id="package-remark"
                  value={remark}
                  rows={3}
                  maxLength={REMARK_MAX}
                  placeholder={t("packageAdmin.remarkPlaceholder")}
                  className="resize-none"
                  onChange={(event) => setRemark(event.target.value)}
                />
              </FormField>
              <div className="grid min-w-0 gap-2">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm font-medium">
                    {t("packageAdmin.features")}
                  </span>
                  <span className="text-xs text-muted-foreground tabular-nums">
                    {t("packageAdmin.selectedCount", {
                      selected: features.length,
                      total: dialogCodes.length,
                    })}
                  </span>
                  {catalogStatus === "failed" ? (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={retryCatalog}
                    >
                      {t("action.retry")}
                    </Button>
                  ) : null}
                </div>
                <p className="text-xs text-muted-foreground">
                  {catalogStatus === "failed"
                    ? t("packageAdmin.featureCatalogFallback")
                    : t("packageAdmin.featureSelectHint")}
                </p>
                <FeaturePicker
                  codes={dialogCodes}
                  value={features}
                  disabled={pending}
                  onChange={setFeatures}
                />
              </div>
            </DialogBody>
            {formError ? (
              <p className="text-sm text-destructive">{formError}</p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => closeDialog(false)}
              >
                {t("action.cancel")}
              </Button>
              <Button type="submit" disabled={pending}>
                {pending ? t("action.saving") : t("action.save")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog
        open={syncing !== null}
        onOpenChange={(open) => {
          if (!open && !syncPending) {
            setSyncing(null)
          }
        }}
      >
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {t("packageAdmin.syncTitle", { name: syncing?.name ?? "" })}
            </DialogTitle>
          </DialogHeader>
          <div className="flex min-h-0 flex-1 flex-col gap-4">
            <DialogBody className="grid gap-4">
              <p className="text-sm text-muted-foreground">
                {t("packageAdmin.syncDescription")}
              </p>
              {syncLoading ? (
                <p className="text-sm text-muted-foreground">
                  {t("state.loading")}
                </p>
              ) : null}
              {!syncLoading && syncFailed ? (
                <div className="flex items-center justify-between gap-2 rounded-lg border border-destructive/25 bg-destructive/5 p-3">
                  <p className="text-sm text-destructive">
                    {t("packageAdmin.syncLoadFailed")}
                  </p>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      if (syncing) {
                        void openSync(syncing)
                      }
                    }}
                  >
                    {t("action.retry")}
                  </Button>
                </div>
              ) : null}
              {!syncLoading && !syncFailed && syncTenants.length === 0 ? (
                <p className="rounded-lg border border-dashed border-border p-3 text-sm text-muted-foreground">
                  {t("packageAdmin.syncNoTenants")}
                </p>
              ) : null}
              {!syncLoading && !syncFailed && syncTenants.length > 0 ? (
                <>
                  <div
                    aria-label={t("packageAdmin.syncScope")}
                    className="grid gap-2 sm:grid-cols-2"
                    role="radiogroup"
                  >
                    <button
                      aria-checked={syncScope === "all"}
                      className={cn(
                        "flex min-w-0 flex-col gap-1 rounded-lg border p-3 text-left transition-colors",
                        syncScope === "all"
                          ? "border-primary bg-primary/5"
                          : "border-border hover:bg-muted/40"
                      )}
                      role="radio"
                      type="button"
                      onClick={() => setSyncScope("all")}
                    >
                      <span className="text-sm font-medium">
                        {t("packageAdmin.syncScopeAll")}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {t("packageAdmin.syncAllHint", {
                          count: syncTenants.length,
                        })}
                      </span>
                    </button>
                    <button
                      aria-checked={syncScope === "selected"}
                      className={cn(
                        "flex min-w-0 flex-col gap-1 rounded-lg border p-3 text-left transition-colors",
                        syncScope === "selected"
                          ? "border-primary bg-primary/5"
                          : "border-border hover:bg-muted/40"
                      )}
                      role="radio"
                      type="button"
                      onClick={() => setSyncScope("selected")}
                    >
                      <span className="text-sm font-medium">
                        {t("packageAdmin.syncScopeSelected")}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {t("packageAdmin.syncSelectedHint")}
                      </span>
                    </button>
                  </div>
                  {syncScope === "selected" ? (
                    <div className="grid gap-2">
                      <div className="flex items-center justify-between gap-2">
                        <span className="text-sm font-medium">
                          {t("packageAdmin.syncSelectedCount", {
                            count: syncSelected.length,
                          })}
                        </span>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            setSyncSelected(
                              allSyncSelected
                                ? []
                                : syncTenants.map((tenant) => tenant.uid)
                            )
                          }
                        >
                          {allSyncSelected
                            ? t("action.cancel")
                            : t("packageAdmin.syncSelectAll")}
                        </Button>
                      </div>
                      <div className="grid max-h-64 gap-1 overflow-y-auto rounded-md border p-2 sm:grid-cols-2">
                        {syncTenants.map((tenant) => (
                          <div
                            className="flex min-h-9 min-w-0 items-center gap-2 px-1"
                            key={tenant.uid}
                          >
                            <Checkbox
                              id={`sync-tenant-${tenant.uid}`}
                              checked={syncSelected.includes(tenant.uid)}
                              onCheckedChange={(checked) =>
                                toggleSyncTenant(tenant.uid, checked === true)
                              }
                            />
                            <Label
                              htmlFor={`sync-tenant-${tenant.uid}`}
                              className="flex min-w-0 flex-1 items-center gap-2"
                            >
                              <span className="truncate text-sm">
                                {tenant.name}
                              </span>
                              <StatusBadge value={tenant.status} />
                            </Label>
                          </div>
                        ))}
                      </div>
                    </div>
                  ) : null}
                </>
              ) : null}
            </DialogBody>
            {syncError ? (
              <p className="text-sm text-destructive">{syncError}</p>
            ) : null}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                disabled={syncPending}
                onClick={() => setSyncing(null)}
              >
                {t("action.cancel")}
              </Button>
              <Button
                type="button"
                disabled={
                  syncPending ||
                  syncLoading ||
                  syncFailed ||
                  syncTenants.length === 0 ||
                  (syncScope === "selected" && syncSelected.length === 0)
                }
                onClick={() => void onSync()}
              >
                {syncPending
                  ? t("packageAdmin.syncPending")
                  : t("packageAdmin.syncSubmit")}
              </Button>
            </DialogFooter>
          </div>
        </DialogContent>
      </Dialog>

      <ConfirmAction
        open={confirming !== null}
        title={t("packageAdmin.deleteTitle")}
        description={t("packageAdmin.deleteDescription")}
        confirmLabel={t("action.confirm")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void onDelete()}
        onOpenChange={(open) => {
          if (!open) {
            setConfirming(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
