import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getNamespaceFeatures,
  listFeatures,
  listPackages,
  updateNamespaceFeatures,
} from "@/api/package"
import type { FeatureItem, PackageItem } from "@/api/package/types"
import type { SystemNamespaceItem } from "@/api/system-namespace/types"
import { FeaturePicker } from "@/components/feature-picker"
import { FeatureTags } from "@/components/feature-tags"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
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
  defaultCapabilityCodes,
  sameCapabilitySet,
  sortCapabilityCodes,
} from "@/lib/feature-catalog"
import { collectFieldErrors, hasFieldErrors, requireSelected } from "@/lib/form"
import { unixSeconds } from "@/lib/date-time"
import { failureMessage, readFormError } from "@/lib/form-error"

// 套餐列表一次取够：平台套餐是人工维护的小集合，租户授权弹窗不需要分页。
const PACKAGE_PAGE_SIZE = 100

type LoadStatus = "loading" | "ready" | "failed"

/** Unix 秒 → <input type="date"> 的本地日期值。 */
function dateValue(seconds?: number | string): string {
  const value = unixSeconds(seconds)
  if (value <= 0) {
    return ""
  }
  const date = new Date(value * 1000)
  const pad = (value: number) => String(value).padStart(2, "0")
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`
}

/** <input type="date"> 的本地日期值 → 当天 23:59:59 的 Unix 秒。 */
function endOfDaySeconds(value: string): number {
  const [year, month, day] = value.split("-").map(Number)
  if (!year || !month || !day) {
    return 0
  }
  return Math.floor(new Date(year, month - 1, day, 23, 59, 59).getTime() / 1000)
}

function todayValue(): string {
  return dateValue(Math.floor(Date.now() / 1000))
}

/** 日期输入值（当天 23:59:59）是否已经过去。 */
function expiryInPast(value: string): boolean {
  return endOfDaySeconds(value) * 1000 <= Date.now()
}

interface TenantFeaturesDialogProps {
  tenant: SystemNamespaceItem | null
  onOpenChange: (open: boolean) => void
  /** 保存成功后由页面刷新租户列表（必要时同时刷新会话里的能力）。 */
  onSaved: (uid: string) => void
}

/**
 * 租户功能授权：选择套餐 + 调整有效能力 + 设置到期时间。
 * 套餐自带能力锁定（只增不减），只能额外追加；到期时间留空则按套餐默认时长。
 */
export function TenantFeaturesDialog({
  tenant,
  onOpenChange,
  onSaved,
}: TenantFeaturesDialogProps) {
  const { t } = useTranslation()
  const tenantUid = tenant?.uid ?? ""
  const [status, setStatus] = useState<LoadStatus>("loading")
  const [loadError, setLoadError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const [packages, setPackages] = useState<PackageItem[]>([])
  const [catalog, setCatalog] = useState<FeatureItem[]>([])
  const [packageUid, setPackageUid] = useState("")
  const [initialPackageUid, setInitialPackageUid] = useState("")
  const [initialFeatures, setInitialFeatures] = useState<string[]>([])
  // 套餐自带能力：锁定，不能取消；切套餐时换成新套餐的默认能力。
  // 切换前要保留初始套餐的能力/时长/到期状态，切回来时原样恢复。
  const [packageFeatures, setPackageFeatures] = useState<string[]>([])
  const [initialPackageFeatures, setInitialPackageFeatures] = useState<string[]>([])
  const [initialDurationDays, setInitialDurationDays] = useState(0)
  const [initialExpired, setInitialExpired] = useState(false)
  const [features, setFeatures] = useState<string[]>([])
  const [expiry, setExpiry] = useState("")
  // 初始到期时间：切回原套餐时恢复成租户当前的到期时间。
  const [initialExpiry, setInitialExpiry] = useState("")
  const [expired, setExpired] = useState(false)
  const [durationDays, setDurationDays] = useState(0)
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [error, setError] = useState("")
  const [pending, setPending] = useState(false)

  useEffect(() => {
    if (!tenantUid) {
      return
    }
    let active = true
    Promise.all([
      getNamespaceFeatures(tenantUid, { silent: true }),
      listPackages({ page: 1, pageSize: PACKAGE_PAGE_SIZE }, { silent: true }),
      listFeatures({ silent: true }),
    ])
      .then(([assignment, packageResult, featureResult]) => {
        if (!active) {
          return
        }
        const currentPackageUid = assignment.packageUid ?? ""
        const currentFeatures = assignment.features ?? []
        const baseFeatures = assignment.packageFeatures ?? []
        const list = [...(packageResult.items ?? [])]
        // 当前套餐不在第一页/已停用时补一条只读选项，避免 Select 显示为空。
        if (
          currentPackageUid &&
          !list.some((item) => item.uid === currentPackageUid)
        ) {
          list.unshift({
            uid: currentPackageUid,
            name: assignment.packageName || currentPackageUid,
            features: baseFeatures,
            isBase: false,
            status: "ACTIVE",
            visibility: "BETA",
            maxMembers: 0,
            priceCents: 0,
            currency: "CNY",
            durationDays: assignment.durationDays ?? 0,
          })
        }
        setPackages(list)
        setCatalog(featureResult.items ?? [])
        setPackageUid(currentPackageUid)
        setInitialPackageUid(currentPackageUid)
        setInitialFeatures(currentFeatures)
        setPackageFeatures(baseFeatures)
        setInitialPackageFeatures(baseFeatures)
        setFeatures(currentFeatures)
        const currentExpiry = dateValue(assignment.expiresAt)
        setExpiry(currentExpiry)
        setInitialExpiry(currentExpiry)
        setExpired(Boolean(assignment.expired))
        setInitialExpired(Boolean(assignment.expired))
        setDurationDays(assignment.durationDays ?? 0)
        setInitialDurationDays(assignment.durationDays ?? 0)
        setStatus("ready")
      })
      .catch((err) => {
        if (!active) {
          return
        }
        setLoadError(failureMessage(err, ""))
        setStatus("failed")
      })
    return () => {
      active = false
    }
  }, [tenantUid, attempt])

  if (!tenant) {
    return null
  }

  function retry() {
    setStatus("loading")
    setLoadError("")
    setAttempt((value) => value + 1)
  }

  function onPackageChange(uid: string) {
    setPackageUid(uid)
    setFieldErrors((current) => {
      const next = { ...current }
      delete next.packageUid
      delete next.expiresAt
      return next
    })
    // 换套餐按新套餐的默认能力重置；到期时间回到「按套餐时长」。
    // 切回原套餐时恢复租户当前的授权（含额外能力与到期时间）。
    const selected = packages.find((item) => item.uid === uid)
    const isInitial = uid === initialPackageUid
    setFeatures(isInitial ? initialFeatures : selected?.features ?? [])
    setPackageFeatures(
      isInitial ? initialPackageFeatures : selected?.features ?? []
    )
    setDurationDays(
      isInitial ? initialDurationDays : selected?.durationDays ?? 0
    )
    setExpiry(isInitial ? initialExpiry : "")
    setExpired(isInitial ? initialExpired : false)
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      [
        "packageUid",
        requireSelected(packageUid, t("tenantAdmin.packageRequired")),
      ],
      [
        "expiresAt",
        expiry !== "" && expiryInPast(expiry)
          ? t("tenantAdmin.expiryInvalid")
          : null,
      ],
    ])
    if (hasFieldErrors(errors)) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      // 目标能力必须包含套餐能力（后端也会校验），这里始终把锁定项并进去。
      await updateNamespaceFeatures(tenantUid, {
        packageUid,
        features: targetFeatures,
        expiresAt: expiry !== "" ? endOfDaySeconds(expiry) : 0,
      })
      toast.success(t("tenantAdmin.featuresSaved"))
      onSaved(tenantUid)
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  const catalogCodes =
    catalog.length > 0
      ? catalog.map((item) => item.code)
      : defaultCapabilityCodes()
  const selectedPackage = packages.find((item) => item.uid === packageUid)
  // 锁定的套餐能力：当前套餐优先用接口返回的 packageFeatures，切到别的套餐时用它自己的能力。
  const lockedCodes = sortCapabilityCodes(
    packageUid === initialPackageUid && packageFeatures.length > 0
      ? packageFeatures
      : selectedPackage?.features ?? []
  )
  const lockedSet = new Set(lockedCodes)
  // 已选但目录里没有的 code 也要展示，避免保存时把它悄悄丢掉。
  const codes = sortCapabilityCodes([
    ...catalogCodes,
    ...initialFeatures,
    ...packageFeatures,
    ...features,
    ...lockedCodes,
  ])
  const targetFeatures = sortCapabilityCodes([...features, ...lockedCodes])
  const extraFeatures = targetFeatures.filter((code) => !lockedSet.has(code))
  // 套餐时长：原套餐用接口返回的时长（0 就是永久），切套餐后取新套餐的默认时长。
  const effectiveDays =
    packageUid === initialPackageUid
      ? durationDays
      : selectedPackage?.durationDays ?? 0
  const permanent = effectiveDays === 0
  // 只在校验结果是「已过期」且日期没被改到未来时提示。
  const expiredNow = expired && expiry !== "" && expiryInPast(expiry)
  const customized = !sameCapabilitySet(lockedCodes, targetFeatures)

  return (
    <Dialog open={tenant !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t("tenantAdmin.featuresTitle")}</DialogTitle>
          <DialogDescription>
            {t("tenantAdmin.featuresDescription", { name: tenant.name })}
          </DialogDescription>
        </DialogHeader>
        <form
          noValidate
          className="flex min-h-0 flex-1 flex-col gap-4"
          onSubmit={onSave}
        >
          <DialogBody className="grid gap-4">
            {status === "loading" ? (
              <p className="text-sm text-muted-foreground">
                {t("tenantAdmin.featuresLoading")}
              </p>
            ) : null}
            {status === "failed" ? (
              <div className="flex items-center gap-2">
                <p className="text-sm text-destructive">
                  {loadError || t("tenantAdmin.featuresLoadFailed")}
                </p>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={retry}
                >
                  {t("action.retry")}
                </Button>
              </div>
            ) : null}
            {status === "ready" ? (
              <>
                <FormField
                  label={t("tenantAdmin.package")}
                  htmlFor="tenant-package"
                  required
                  hint={t("tenantAdmin.packageHint")}
                  error={fieldErrors.packageUid}
                >
                  <Select value={packageUid} onValueChange={onPackageChange}>
                    <SelectTrigger id="tenant-package">
                      <SelectValue
                        placeholder={t("tenantAdmin.packagePlaceholder")}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      {packages.map((item) => (
                        <SelectItem key={item.uid} value={item.uid}>
                          {item.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </FormField>
                <div className="flex min-w-0 flex-wrap items-center gap-2 rounded-lg border border-border/70 bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
                  <span>
                    {t("tenantAdmin.packageDuration")}：
                    <span className="font-medium text-foreground">
                      {permanent
                        ? t("tenantAdmin.permanent")
                        : t("tenantAdmin.durationDays", {
                            count: effectiveDays,
                          })}
                    </span>
                  </span>
                  <span aria-hidden="true">·</span>
                  <span>
                    {t("tenantAdmin.packageFeaturesCount", {
                      count: lockedCodes.length,
                    })}
                  </span>
                </div>
                <FormField
                  label={t("tenantAdmin.expiresAt")}
                  htmlFor="tenant-expiry"
                  hint={
                    permanent
                      ? t("tenantAdmin.expiresAtPermanentHint")
                      : t("tenantAdmin.expiresAtHint", {
                          count: effectiveDays,
                        })
                  }
                  error={fieldErrors.expiresAt}
                >
                  <div className="grid min-w-0 gap-1.5">
                    <Input
                      id="tenant-expiry"
                      type="date"
                      value={expiry}
                      min={todayValue()}
                      disabled={permanent}
                      onChange={(event) => setExpiry(event.target.value)}
                    />
                    <p className="text-xs text-muted-foreground">
                      {permanent
                        ? t("tenantAdmin.expiresAtPermanent")
                        : expiry
                          ? t("tenantAdmin.expiresAtValue", {
                              date: expiry,
                            })
                          : t("tenantAdmin.expiresAtDefault")}
                      {expiredNow ? (
                        <span className="ml-2 text-destructive">
                          {t("tenantAdmin.expired")}
                        </span>
                      ) : null}
                    </p>
                  </div>
                </FormField>
                <div className="grid min-w-0 gap-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-medium">
                      {t("tenantAdmin.featuresLabel")}
                    </span>
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {t("tenantAdmin.featuresCount", {
                        selected: targetFeatures.length,
                        total: codes.length,
                      })}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {customized
                      ? t("tenantAdmin.featuresOverride")
                      : t("tenantAdmin.packageFeatures", {
                          count: lockedCodes.length,
                        })}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {t("tenantAdmin.packageLockedHint")}
                  </p>
                  <FeaturePicker
                    codes={codes}
                    value={features}
                    lockedCodes={lockedCodes}
                    lockedLabel={t("tenantAdmin.packageLocked")}
                    disabled={pending}
                    onChange={setFeatures}
                  />
                </div>
                <div className="grid min-w-0 gap-1.5 rounded-lg border border-border/70 bg-muted/40 px-3 py-2">
                  <span className="text-xs font-medium">
                    {t("tenantAdmin.extraFeatures")}
                  </span>
                  {extraFeatures.length > 0 ? (
                    <FeatureTags codes={extraFeatures} emptyLabel="—" />
                  ) : (
                    <span className="text-xs text-muted-foreground">
                      {t("tenantAdmin.noExtraFeatures")}
                    </span>
                  )}
                  <span className="text-xs text-muted-foreground">
                    {t("tenantAdmin.extraFeaturesHint")}
                  </span>
                </div>
              </>
            ) : null}
          </DialogBody>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending || status !== "ready"}>
              {pending ? t("action.saving") : t("action.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
