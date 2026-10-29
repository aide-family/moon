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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  defaultFeatureCodes,
  sameFeatureSet,
  sortFeatureCodes,
} from "@/lib/feature-catalog"
import { collectFieldErrors, hasFieldErrors, requireSelected } from "@/lib/form"
import { failureMessage, readFormError } from "@/lib/form-error"

// 套餐列表一次取够：平台套餐是人工维护的小集合，租户授权弹窗不需要分页。
const PACKAGE_PAGE_SIZE = 100

type LoadStatus = "loading" | "ready" | "failed"

interface TenantFeaturesDialogProps {
  tenant: SystemNamespaceItem | null
  onOpenChange: (open: boolean) => void
  /** 保存成功后由页面刷新租户列表（必要时同时刷新会话里的功能）。 */
  onSaved: (uid: string) => void
}

/**
 * 租户功能授权：选择套餐 + 调整有效功能。
 * 默认勾选租户当前有效功能；切换套餐时按新套餐的默认功能重置勾选。
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
  const [features, setFeatures] = useState<string[]>([])
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
        const list = [...(packageResult.items ?? [])]
        // 当前套餐不在第一页/已停用时补一条只读选项，避免 Select 显示为空。
        if (
          currentPackageUid &&
          !list.some((item) => item.uid === currentPackageUid)
        ) {
          list.unshift({
            uid: currentPackageUid,
            name: assignment.packageName || currentPackageUid,
            features: assignment.packageFeatures ?? [],
            isBase: false,
            status: "ACTIVE",
          })
        }
        setPackages(list)
        setCatalog(featureResult.items ?? [])
        setPackageUid(currentPackageUid)
        setInitialPackageUid(currentPackageUid)
        setInitialFeatures(currentFeatures)
        setFeatures(currentFeatures)
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
      return next
    })
    // 选回当前套餐时恢复实际有效功能；换套餐时按套餐默认功能重置勾选。
    if (uid === initialPackageUid) {
      setFeatures(initialFeatures)
      return
    }
    const selected = packages.find((item) => item.uid === uid)
    setFeatures(selected?.features ?? [])
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      [
        "packageUid",
        requireSelected(packageUid, t("tenantAdmin.packageRequired")),
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
      await updateNamespaceFeatures(tenantUid, {
        packageUid,
        features: sortFeatureCodes(features),
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
      : defaultFeatureCodes()
  // 已选但目录里没有的 code 也要展示，避免保存时把它悄悄丢掉。
  const codes = sortFeatureCodes([
    ...catalogCodes,
    ...initialFeatures,
    ...features,
  ])
  const selectedPackage = packages.find((item) => item.uid === packageUid)
  const customized = selectedPackage
    ? !sameFeatureSet(selectedPackage.features ?? [], features)
    : false

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
                <div className="grid min-w-0 gap-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="text-sm font-medium">
                      {t("tenantAdmin.featuresLabel")}
                    </span>
                    <span className="text-xs text-muted-foreground tabular-nums">
                      {t("tenantAdmin.featuresCount", {
                        selected: features.length,
                        total: codes.length,
                      })}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {customized
                      ? t("tenantAdmin.featuresOverride")
                      : t("tenantAdmin.packageFeatures", {
                          count: selectedPackage?.features?.length ?? 0,
                        })}
                  </p>
                  <FeaturePicker
                    codes={codes}
                    value={features}
                    disabled={pending}
                    onChange={setFeatures}
                  />
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
