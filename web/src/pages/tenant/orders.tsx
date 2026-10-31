import { ArrowLeft } from "lucide-react"
import { useTranslation } from "react-i18next"
import { useLocation, useNavigate, useParams } from "react-router-dom"

import { listSystemNamespaceOrders } from "@/api/order"
import type { OrderItem } from "@/api/order/types"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
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
import { orderKindLabel, orderStatusLabel, orderStatusTone } from "@/lib/order"
import { formatPriceCents } from "@/lib/price"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

// 平台按租户查询订单。独立路由而不是弹窗：订单是长列表，需要分页、可分享的地址
// 和明确的返回路径；从租户管理页进入时带上租户名称，直接打开时退回显示 uid。
export function TenantOrdersPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const params = useParams()
  const tenantUid = params.uid ?? ""
  const state = location.state as { name?: string } | null
  const tenantName = state?.name || tenantUid
  const query = useListQuery()

  const {
    items,
    total,
    loading,
    failed,
    error: loadError,
    reload,
  } = usePagedList<OrderItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: tenantUid,
    enabled: tenantUid !== "",
    fetch: (page, pageSize) =>
      listSystemNamespaceOrders(tenantUid, { page, pageSize }),
  })

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-5">
      <PageHeader
        title={t("tenantAdmin.ordersTitle", { name: tenantName })}
        description={t("tenantAdmin.ordersDescription")}
      >
        <Button
          className="shrink-0"
          variant="outline"
          onClick={() => navigate("/tenants")}
        >
          <ArrowLeft aria-hidden="true" className="mr-2 size-4" />
          {t("action.goBack")}
        </Button>
      </PageHeader>

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
        {!loading && !failed && items.length === 0 ? (
          <div className="flex min-h-44 flex-col items-center justify-center rounded-lg border border-dashed border-border px-5 py-8 text-center">
            <p className="text-sm font-medium">{t("tenantAdmin.ordersEmpty")}</p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("order.uid")}</TableHead>
                <TableHead>{t("order.package")}</TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("order.kind")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.type, TABLE_NOWRAP)}
                >
                  {t("order.amount")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.status, TABLE_NOWRAP)}
                >
                  {t("order.status")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.time, TABLE_NOWRAP)}
                >
                  {t("order.createdAt")}
                </TableHead>
                <TableHead
                  className={cn(TABLE_COLUMN_WIDTH.time, TABLE_NOWRAP)}
                >
                  {t("order.paidAt")}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-mono text-xs break-all">
                    {item.uid}
                  </TableCell>
                  <TableCell className="break-words">
                    {item.packageName}
                  </TableCell>
                  <TableCell className={cn(TABLE_NOWRAP)}>
                    {orderKindLabel(t, item.kind)}
                  </TableCell>
                  <TableCell className={cn(TABLE_NOWRAP, "tabular-nums")}>
                    {item.priceCents > 0
                      ? t("order.priceValue", {
                          amount: formatPriceCents(
                            item.priceCents,
                            item.currency
                          ),
                        })
                      : t("order.free")}
                  </TableCell>
                  <TableCell className={cn(TABLE_NOWRAP)}>
                    <StatusBadge
                      value={item.status}
                      label={orderStatusLabel(t, item.status)}
                      tone={orderStatusTone(item.status)}
                    />
                  </TableCell>
                  <TableCell
                    className={cn(
                      TABLE_COLUMN_WIDTH.time,
                      TABLE_NOWRAP,
                      "text-muted-foreground tabular-nums"
                    )}
                  >
                    {item.createdAt && item.createdAt > 0
                      ? new Date(item.createdAt * 1000).toLocaleString()
                      : "—"}
                  </TableCell>
                  <TableCell
                    className={cn(
                      TABLE_COLUMN_WIDTH.time,
                      TABLE_NOWRAP,
                      "text-muted-foreground tabular-nums"
                    )}
                  >
                    {item.paidAt && item.paidAt > 0
                      ? new Date(item.paidAt * 1000).toLocaleString()
                      : "—"}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
    </section>
  )
}
