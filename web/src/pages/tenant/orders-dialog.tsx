import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { listSystemNamespaceOrders } from "@/api/order"
import type { OrderItem } from "@/api/order/types"
import type { SystemNamespaceItem } from "@/api/system-namespace/types"
import { StatusBadge } from "@/components/status-badge"
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { cn } from "@/lib/class-names"
import { failureMessage } from "@/lib/form-error"
import { orderStatusLabel, orderStatusTone } from "@/lib/order"
import { formatPriceCents } from "@/lib/price"
import { TABLE_COLUMN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

const PAGE_SIZE = 20

type LoadStatus = "loading" | "ready" | "failed"

interface TenantOrdersDialogProps {
  tenant: SystemNamespaceItem | null
  onOpenChange: (open: boolean) => void
}

/**
 * 平台按租户查询订单：只读展示该租户的创建订单及其套餐快照、金额和支付时间。
 * 父组件用租户 uid 作为 key，切换租户时弹窗状态整体重置。
 */
export function TenantOrdersDialog({
  tenant,
  onOpenChange,
}: TenantOrdersDialogProps) {
  const { t } = useTranslation()
  const tenantUid = tenant?.uid ?? ""
  const [items, setItems] = useState<OrderItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [status, setStatus] = useState<LoadStatus>("loading")
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    if (!tenantUid) {
      return
    }
    let active = true
    listSystemNamespaceOrders(
      tenantUid,
      { page, pageSize: PAGE_SIZE },
      { silent: true }
    )
      .then((result) => {
        if (!active) {
          return
        }
        setItems(result.items ?? [])
        setTotal(result.total ?? 0)
        setStatus("ready")
      })
      .catch((err) => {
        if (!active) {
          return
        }
        setError(failureMessage(err, ""))
        setStatus("failed")
      })
    return () => {
      active = false
    }
  }, [tenantUid, page, attempt])

  if (!tenant) {
    return null
  }

  const pages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  function goPage(next: number) {
    setStatus("loading")
    setError("")
    setPage(next)
  }

  function retry() {
    setStatus("loading")
    setError("")
    setAttempt((value) => value + 1)
  }

  return (
    <Dialog open={tenant !== null} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {t("tenantAdmin.ordersTitle", { name: tenant.name })}
          </DialogTitle>
          <DialogDescription>{t("tenantAdmin.ordersDescription")}</DialogDescription>
        </DialogHeader>
        <DialogBody className="grid gap-3">
          <div className="flex items-center justify-between gap-2">
            <span className="font-mono text-sm text-muted-foreground">
              {t("list.total", { count: total })}
            </span>
            {status === "ready" ? (
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => goPage(Math.max(1, page - 1))}
                >
                  {t("action.prev")}
                </Button>
                <span className="font-mono text-sm text-muted-foreground">
                  {page} / {pages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page >= pages}
                  onClick={() => goPage(page + 1)}
                >
                  {t("action.next")}
                </Button>
              </div>
            ) : null}
          </div>
          {status === "loading" ? (
            <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
          ) : null}
          {status === "failed" ? (
            <div
              className="flex flex-col items-start gap-2 rounded-lg border border-destructive/25 bg-destructive/5 p-4 sm:flex-row sm:items-center sm:justify-between"
              role="alert"
            >
              <p className="text-sm text-destructive">
                {error || t("state.failed")}
              </p>
              <Button
                variant="outline"
                size="sm"
                onClick={retry}
              >
                {t("action.retry")}
              </Button>
            </div>
          ) : null}
          {status === "ready" && items.length === 0 ? (
            <p className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
              {t("tenantAdmin.ordersEmpty")}
            </p>
          ) : null}
          {status === "ready" && items.length > 0 ? (
            <div className="flex max-h-[50vh] min-h-0 flex-col">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("order.uid")}</TableHead>
                    <TableHead>{t("order.package")}</TableHead>
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
                      <TableCell className={cn(TABLE_NOWRAP, "tabular-nums")}>
                        {item.priceCents > 0
                          ? t("order.priceValue", {
                              amount: formatPriceCents(item.priceCents),
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
            </div>
          ) : null}
        </DialogBody>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t("action.close")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
