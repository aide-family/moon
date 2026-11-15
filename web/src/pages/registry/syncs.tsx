import { useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate } from "react-router-dom"
import { toast } from "sonner"

import { cancelImageSync, listImageSyncs, retryImageSync } from "@/api/registry"
import type { ImageSyncItem, ImageSyncStatus } from "@/api/registry/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatUnixDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

import { SyncDialog } from "./sync-dialog"

const ALL = "ALL"
const STATUS_VALUES = new Set(["PENDING", "RUNNING", "SUCCEEDED", "FAILED", "CANCELED"])

function normalizeStatus(value: string): ImageSyncStatus | "" {
  return STATUS_VALUES.has(value) ? (value as ImageSyncStatus) : ""
}

/** 同步任务列表：已接入源与外部源、阶段进度、重试与取消。 */
export function RegistrySyncPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const query = useListQuery()
  const status = normalizeStatus(query.status)
  const sourceKind = query.type === "EXTERNAL" ? "EXTERNAL" : query.type === "CONNECTION" ? "CONNECTION" : ""
  const [createOpen, setCreateOpen] = useState(false)
  const { items, total, loading, failed, error, reload } = usePagedList<ImageSyncItem>({
    page: query.page,
    pageSize: query.pageSize,
    filterKey: `${session.namespaceUid}|${status}|${sourceKind}`,
    fetch: (page, pageSize) => listImageSyncs({ page, pageSize, status, sourceKind }),
  })

  if (!session.loading && !session.can("registry_sync:read")) {
    return <Navigate to="/forbidden" replace />
  }

  async function handleRetry(item: ImageSyncItem) {
    try {
      await retryImageSync(item.uid)
      toast.success(t("registry.retry"))
      reload()
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) return
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  async function handleCancel(item: ImageSyncItem) {
    try {
      await cancelImageSync(item.uid)
      toast.success(t("registry.cancel"))
      reload()
    } catch (err) {
      if (err instanceof ApiError && err.status === 403) return
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  function badge(item: ImageSyncItem) {
    if (item.status === "SUCCEEDED") return <Badge variant="secondary">{item.status}</Badge>
    if (item.status === "FAILED") return <Badge variant="destructive">{item.status}</Badge>
    if (item.status === "CANCELED") return <Badge variant="outline">{item.status}</Badge>
    return <Badge variant="default">{item.status}</Badge>
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("registry.syncTitle")}
        description={t("registry.syncDescription")}
      >
        <Button variant="outline" onClick={() => navigate("/registries")}>
          {t("action.goBack")}
        </Button>
        {session.can("registry_sync:write") ? (
          <Button onClick={() => setCreateOpen(true)}>{t("registry.createSync")}</Button>
        ) : null}
      </PageHeader>
      <div className="flex flex-wrap items-center gap-2">
        <Select
          value={status || ALL}
          onValueChange={(value) => query.update({ status: value === ALL ? "" : value })}
        >
          <SelectTrigger className="w-40"><SelectValue>{status || t("registry.status")}</SelectValue></SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>{t("registry.status")}</SelectItem>
            {["PENDING", "RUNNING", "SUCCEEDED", "FAILED", "CANCELED"].map((value) => (
              <SelectItem key={value} value={value}>{value}</SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select
          value={sourceKind || ALL}
          onValueChange={(value) => query.update({ type: value === ALL ? "" : value })}
        >
          <SelectTrigger className="w-44"><SelectValue>{sourceKind || t("registry.sourceKind")}</SelectValue></SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL}>{t("registry.sourceKind")}</SelectItem>
            <SelectItem value="CONNECTION">{t("registry.sourceConnection")}</SelectItem>
            <SelectItem value="EXTERNAL">{t("registry.sourceExternal")}</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{error || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>{t("action.retry")}</Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <div className="grid gap-2">
            <p className="text-sm text-muted-foreground">{t("registry.syncEmpty")}</p>
            <p className="text-sm text-muted-foreground">{t("registry.syncEmptyHint")}</p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("registry.sourceImage")}</TableHead>
                <TableHead>{t("registry.targetRepository")}</TableHead>
                <TableHead>{t("registry.architecture")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("registry.status")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("registry.createdAt")}</TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell>
                    <div className="grid gap-1">
                      <span className="font-mono text-xs break-all">{item.sourceImage || "—"}</span>
                      <Badge variant="outline">{item.sourceKind}</Badge>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="grid gap-1">
                      <span className="font-mono text-xs break-all">{item.targetRepository}:{item.targetTag}</span>
                      <span className="text-xs text-muted-foreground">{item.targetRegistryName}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {item.sourcePlatforms.length === 0 ? "—" : item.sourcePlatforms.map((platform) => (
                        <Badge key={platform} variant="outline">{platform}</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <div className="grid gap-1">
                      {badge(item)}
                      {item.totalBlobs > 0 ? (
                        <span className="text-xs text-muted-foreground">
                          {item.stage} · {item.copiedBlobs}/{item.totalBlobs}
                        </span>
                      ) : null}
                      {item.message ? <span className="max-w-[18rem] truncate text-xs text-destructive" title={item.message}>{item.message}</span> : null}
                    </div>
                  </TableCell>
                  <TableCell className={TABLE_NOWRAP}>
                    <div className="grid gap-1 text-xs text-muted-foreground">
                      <span>{item.operatorName || "—"}</span>
                      <span>{formatUnixDateTime(item.createdAt) || "—"}</span>
                    </div>
                  </TableCell>
                  <TableActionsCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => toast.info(item.uid, { description: item.stage || item.status })}
                      actions={[
                        ...(session.can("registry_sync:write") && (item.status === "FAILED" || item.status === "CANCELED")
                          ? [{ key: "retry", label: t("registry.retry"), onSelect: () => void handleRetry(item) }]
                          : []),
                        ...(session.can("registry_sync:write") && (item.status === "PENDING" || item.status === "RUNNING")
                          ? [{ key: "cancel", label: t("registry.cancel"), onSelect: () => void handleCancel(item) }]
                          : []),
                      ]}
                    />
                  </TableActionsCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <SyncDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        registryUid=""
        sourceRepository=""
        sourceTag=""
        onCreated={() => {
          setCreateOpen(false)
          toast.success(t("registry.syncCreated"))
          reload()
        }}
      />
    </section>
  )
}
