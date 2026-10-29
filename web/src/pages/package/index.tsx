import { Package } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import {
  createPackage,
  deletePackage,
  listFeatures,
  listPackages,
  updatePackage,
} from "@/api/package"
import type { FeatureItem, PackageItem } from "@/api/package/types"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { cn } from "@/lib/class-names"
import { defaultFeatureCodes, sortFeatureCodes } from "@/lib/feature-catalog"
import { collectFieldErrors, hasFieldErrors, requireTrimmed } from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

const REMARK_MAX = 5000

type CatalogStatus = "loading" | "ready" | "failed"

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
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState("")
  const [pending, setPending] = useState(false)
  const [confirming, setConfirming] = useState<PackageItem | null>(null)
  const [confirmError, setConfirmError] = useState("")
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

  // 后端目录可用时以它为准；接口未部署或失败时回落到本地功能目录。
  const catalogCodes =
    catalogStatus === "ready" && catalog.length > 0
      ? catalog.map((item) => item.code)
      : defaultFeatureCodes()
  // 正在编辑的套餐里可能有本地目录还没有的新功能，一并展示，避免保存时被清掉。
  const dialogCodes = sortFeatureCodes([...catalogCodes, ...features])

  function openCreate() {
    setEditing(null)
    setName("")
    setRemark("")
    setFeatures([])
    setFieldErrors({})
    setFormError("")
    setDialogOpen(true)
  }

  function openEdit(item: PackageItem) {
    setEditing(item)
    setName(item.name)
    setRemark(item.remark ?? "")
    setFeatures(item.features ?? [])
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
      setFieldErrors({})
      setFormError("")
    }
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    const errors = collectFieldErrors([
      // 基础套餐名称由后端固定，前端不再校验改名。
      [
        "name",
        editing?.isBase ? null : requireTrimmed(name, t("form.required")),
      ],
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
        features: sortFeatureCodes(features),
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

  function packageActions(item: PackageItem): RowAction[] {
    if (!canWrite || item.isBase) {
      return []
    }
    return [
      {
        key: "delete",
        label: t("action.delete"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setConfirming(item)
        },
      },
    ]
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
          <Table className={TABLE_MIN_WIDTH.default}>
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
              <FormField
                label={t("packageAdmin.name")}
                htmlFor="package-name"
                required={!editing?.isBase}
                hint={
                  editing?.isBase ? t("packageAdmin.baseNameLocked") : undefined
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
