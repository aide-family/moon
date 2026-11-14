import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { cancelOrder, listOrders, payOrder } from "@/api/order"
import type { OrderItem } from "@/api/order/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { EnumIcon } from "@/components/enum-icon"
import type { FilterChip } from "@/components/filter-chips"
import { ListFilter } from "@/components/list-filter"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
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
import { useFilterDraft } from "@/hooks/use-filter-draft"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { cn } from "@/lib/class-names"
import { formErrorMessage } from "@/lib/form-error"
import {
  orderKindLabel,
  orderStatusLabel,
  orderStatusTone,
} from "@/lib/order"
import { formatPriceCents } from "@/lib/price"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

const ALL_STATUS = "ALL"

// 订单页只处理当前用户自己的订单：待支付订单可以继续支付或取消，
// 支付成功后后端在同一个事务里创建租户并绑定套餐。平台按租户查订单在租户管理页。
export function OrderPage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [filtersOpen, setFiltersOpen] = useState(false)
  const { draft, update, dirty } = useFilterDraft({ status: query.status })
  const [cancelling, setCancelling] = useState<OrderItem | null>(null)
  const [pending, setPending] = useState(false)
  const [actionError, setActionError] = useState("")
  const [confirmError, setConfirmError] = useState("")
  const [detail, setDetail] = useState<OrderItem | null>(null)

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
    filterKey: query.status,
    fetch: (page, pageSize) =>
      listOrders({
        page,
        pageSize,
        status: query.status || undefined,
      }),
  })

  async function onPay(item: OrderItem) {
    setActionError("")
    try {
      const result = await payOrder(item.uid)
      toast.success(
        t("order.paidSuccess", {
          name: result.namespace?.name || item.namespaceName,
        })
      )
      // 支付会新增一个“我加入的租户”，刷新会话让切换器和权限立即看到它。
      session.reload()
      reload()
    } catch (err) {
      setActionError(formErrorMessage(err, t("state.failed")))
    }
  }

  async function onCancel() {
    if (!cancelling) {
      return
    }
    setPending(true)
    setConfirmError("")
    try {
      await cancelOrder(cancelling.uid)
      toast.success(t("order.cancelledToast"))
      setCancelling(null)
      reload()
    } catch (err) {
      setConfirmError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  function orderActions(item: OrderItem): RowAction[] {
    if (item.status !== "PENDING") {
      return []
    }
    return [
      {
        key: "pay",
        label: t("order.pay"),
        onSelect: () => void onPay(item),
      },
      {
        key: "cancel",
        label: t("action.cancel"),
        destructive: true,
        onSelect: () => {
          setConfirmError("")
          setCancelling(item)
        },
      },
    ]
  }

  const chips: FilterChip[] = query.status
    ? [
        {
          key: "status",
          label: `${t("order.status")}: ${orderStatusLabel(t, query.status)}`,
          onRemove: clearFilters,
        },
      ]
    : []

  function applyFilters() {
    query.update({ status: draft.status })
    setFiltersOpen(false)
  }

  function clearFilters() {
    query.update({ status: "" })
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-5">
      <PageHeader title={t("order.title")} description={t("order.description")} />

      <ListFilter
        id="order-filters"
        open={filtersOpen}
        onOpenChange={setFiltersOpen}
        chips={chips}
        onClearAll={clearFilters}
        onClear={clearFilters}
        appliedCount={chips.length}
        dirty={dirty}
        onApply={applyFilters}
      >
        <Field className="w-44 shrink-0">
          <FieldLabel
            htmlFor="order-filter-status"
            className="text-xs text-muted-foreground"
          >
            {t("order.status")}
          </FieldLabel>
          <Select
            value={draft.status || ALL_STATUS}
            onValueChange={(value) =>
              update("status", value === ALL_STATUS ? "" : value)
            }
          >
            <SelectTrigger id="order-filter-status" className="w-full">
              <SelectValue placeholder={t("order.allStatus")}>
                <span className="flex items-center gap-2">
                  <EnumIcon value={draft.status} />
                  <span>
                    {draft.status
                      ? orderStatusLabel(t, draft.status)
                      : t("order.allStatus")}
                  </span>
                </span>
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={ALL_STATUS}>{t("order.allStatus")}</SelectItem>
              <SelectItem value="PENDING">
                <span className="flex items-center gap-2">
                  <EnumIcon value="PENDING" />
                  <span>{t("order.pending")}</span>
                </span>
              </SelectItem>
              <SelectItem value="PAID">
                <span className="flex items-center gap-2">
                  <EnumIcon value="PAID" />
                  <span>{t("order.paid")}</span>
                </span>
              </SelectItem>
              <SelectItem value="CANCELLED">
                <span className="flex items-center gap-2">
                  <EnumIcon value="CANCELLED" />
                  <span>{t("order.cancelled")}</span>
                </span>
              </SelectItem>
            </SelectContent>
          </Select>
        </Field>
      </ListFilter>

      {actionError ? (
        <div
          className="flex flex-col items-start gap-2 rounded-lg border border-destructive/25 bg-destructive/5 p-4 sm:flex-row sm:items-center sm:justify-between"
          role="alert"
        >
          <p className="text-sm text-destructive">{actionError}</p>
          <Button variant="outline" onClick={() => setActionError("")}>
            {t("action.close")}
          </Button>
        </div>
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
        {!loading && !failed && items.length === 0 ? (
          <div className="flex min-h-44 flex-col items-center justify-center rounded-lg border border-dashed border-border px-5 py-8 text-center">
            <p className="text-sm font-medium">{t("order.empty")}</p>
            <p className="mt-1 text-sm text-muted-foreground">
              {t("order.emptyHint")}
            </p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("order.namespace")}</TableHead>
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
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-medium break-words">
                    {item.namespaceName}
                    <span className="mt-0.5 block font-mono text-xs text-muted-foreground">
                      {item.uid}
                    </span>
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
                  <TableCell>
                    <RowActions
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      nowrap
                      onDetail={() => setDetail(item)}
                      actions={orderActions(item)}
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
        title={detail?.namespaceName || t("action.detail")}
        onOpenChange={(value) => {
          if (!value) {
            setDetail(null)
          }
        }}
        fields={[
          { label: t("detail.uid"), value: detail?.uid ?? "" },
          { label: t("order.namespace"), value: detail?.namespaceName ?? "" },
          {
            label: t("order.namespaceUid"),
            value: detail?.namespaceUid || "—",
          },
          { label: t("order.package"), value: detail?.packageName ?? "" },
          {
            label: t("order.kind"),
            value: detail ? orderKindLabel(t, detail.kind) : "—",
          },
          {
            label: t("order.amount"),
            value: detail
              ? detail.priceCents > 0
                ? t("order.priceValue", {
                    amount: formatPriceCents(
                      detail.priceCents,
                      detail.currency
                    ),
                  })
                : t("order.free")
              : "—",
          },
          {
            label: t("order.duration"),
            value: detail
              ? t("order.durationValue", { count: detail.durationDays ?? 0 })
              : "—",
          },
          {
            label: t("order.maxMembers"),
            value: detail?.maxMembers
              ? t("order.membersValue", { count: detail.maxMembers })
              : t("order.membersUnlimited"),
          },
          {
            label: t("order.status"),
            value: detail ? (
              <StatusBadge
                value={detail.status}
                label={orderStatusLabel(t, detail.status)}
                tone={orderStatusTone(detail.status)}
              />
            ) : (
              ""
            ),
          },
          {
            label: t("order.createdAt"),
            value:
              detail?.createdAt && detail.createdAt > 0
                ? new Date(detail.createdAt * 1000).toLocaleString()
                : "—",
          },
          {
            label: t("order.paidAt"),
            value:
              detail?.paidAt && detail.paidAt > 0
                ? new Date(detail.paidAt * 1000).toLocaleString()
                : "—",
          },
        ]}
      />

      <ConfirmAction
        open={cancelling !== null}
        title={t("order.cancelTitle")}
        description={t("order.cancelDescription", {
          name: cancelling?.namespaceName ?? "",
        })}
        confirmLabel={t("action.confirm")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={pending}
        error={confirmError}
        onConfirm={() => void onCancel()}
        onOpenChange={(value) => {
          if (!value) {
            setCancelling(null)
            setConfirmError("")
          }
        }}
      />
    </section>
  )
}
