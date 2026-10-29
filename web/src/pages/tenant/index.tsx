import { Building2 } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate } from "react-router-dom"

import { listSystemNamespaces } from "@/api/system-namespace"
import type { SystemNamespaceItem } from "@/api/system-namespace/types"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead, TableStatusHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
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
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"
import { TenantFeaturesDialog } from "./features-dialog"

// 系统管理员租户列表：平台级只读入口。点「只读查看」把会话切到目标租户的
// 只读视角（permissions 只有 :read）并跳到概览，写按钮随之自动隐藏。
export function TenantPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const [draft, setDraft] = useState(query.name)
  const [entering, setEntering] = useState("")
  const [enterError, setEnterError] = useState("")
  // 功能授权弹窗只对超级管理员开放；系统管理员只读视角不显示该操作。
  const [featuresTenant, setFeaturesTenant] =
    useState<SystemNamespaceItem | null>(null)
  const canManage = session.canSystem("system:namespace:read")
  const canAssignFeatures =
    session.canSystem("system:package:write") && !session.readOnly

  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<SystemNamespaceItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: query.name,
    enabled: !session.loading && canManage,
    fetch: (page, pageSize) =>
      listSystemNamespaces({ page, pageSize, name: query.name }),
  })

  async function onView(item: SystemNamespaceItem) {
    if (entering) {
      return
    }
    // 已经停在目标租户（成员视角或只读视角）时直接回概览，不重复切会话。
    if (item.uid === session.namespaceUid) {
      navigate("/overview")
      return
    }
    setEnterError("")
    setEntering(item.uid)
    const ok = await session.enterNamespace(item.uid)
    setEntering("")
    if (ok) {
      navigate("/overview")
      return
    }
    setEnterError(t("tenantAdmin.enterFailed"))
  }

  function featureActions(item: SystemNamespaceItem): RowAction[] {
    if (!canAssignFeatures) {
      return []
    }
    return [
      {
        key: "features",
        label: t("tenantAdmin.featuresAction"),
        onSelect: () => setFeaturesTenant(item),
      },
    ]
  }

  function onFeaturesSaved(uid: string) {
    reload()
    // 当前正停在这个租户时，菜单功能要立刻跟随新的授权刷新。
    if (uid === session.namespaceUid) {
      session.reload()
    }
  }

  if (!session.loading && !canManage) {
    return <Navigate to="/forbidden" replace />
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-5">
      <PageHeader
        title={t("tenantAdmin.title")}
        description={t("tenantAdmin.description")}
      />
      <form
        className="flex gap-2"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draft })
        }}
      >
        <Input
          className="max-w-sm"
          aria-label={t("tenantAdmin.search")}
          placeholder={t("tenantAdmin.search")}
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
        />
        <Button type="submit" variant="outline">
          {t("action.search")}
        </Button>
      </form>
      {entering ? (
        <p aria-live="polite" className="text-sm text-muted-foreground">
          {t("tenantAdmin.entering")}
        </p>
      ) : null}
      {enterError ? (
        <p className="text-sm text-destructive" role="alert">
          {enterError}
        </p>
      ) : null}
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
            <p className="text-sm font-medium">{t("tenantAdmin.noResults")}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("tenantAdmin.noResultsHint", { query: query.name })}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 && !query.name ? (
          <div className="flex min-h-44 flex-col items-center justify-center rounded-lg border border-dashed border-border px-5 py-8 text-center">
            <span className="mb-3 grid size-11 place-items-center rounded-xl bg-primary/10 text-primary-ink">
              <Building2 aria-hidden="true" className="size-5" />
            </span>
            <p className="text-sm font-medium">{t("tenantAdmin.empty")}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("tenantAdmin.emptyHint")}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("tenantAdmin.name")}</TableHead>
                <TableHead className={TABLE_COLUMN_WIDTH.identity}>
                  {t("tenantAdmin.owner")}
                </TableHead>
                <TableStatusHead>{t("tenantAdmin.status")}</TableStatusHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("tenantAdmin.memberCount")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.time, TABLE_NOWRAP)}
                >
                  {t("tenantAdmin.createdAt")}
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
                  <TableCell
                    className={cn(TABLE_COLUMN_WIDTH.identity, "break-all")}
                  >
                    <span className="flex min-w-0 flex-col">
                      <span>
                        {item.ownerName || item.ownerEmail || item.ownerUid}
                      </span>
                      {item.ownerName && item.ownerEmail ? (
                        <span className="text-xs text-muted-foreground">
                          {item.ownerEmail}
                        </span>
                      ) : null}
                    </span>
                  </TableCell>
                  <TableCell>
                    <StatusBadge value={item.status} />
                  </TableCell>
                  <TableCell
                    className={cn(
                      TABLE_COLUMN_WIDTH.type,
                      TABLE_NOWRAP,
                      "font-mono text-xs tabular-nums"
                    )}
                  >
                    {item.memberCount}
                  </TableCell>
                  <TableCell
                    className={cn(
                      TABLE_COLUMN_WIDTH.time,
                      TABLE_NOWRAP,
                      "font-mono text-xs tabular-nums"
                    )}
                  >
                    {item.createdAt
                      ? new Date(item.createdAt * 1000).toLocaleString()
                      : "—"}
                  </TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("tenantAdmin.view")}
                      moreLabel={t("action.more")}
                      onDetail={() => void onView(item)}
                      actions={featureActions(item)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <TenantFeaturesDialog
        key={featuresTenant?.uid ?? "none"}
        tenant={featuresTenant}
        onOpenChange={(open) => {
          if (!open) {
            setFeaturesTenant(null)
          }
        }}
        onSaved={onFeaturesSaved}
      />
    </section>
  )
}
