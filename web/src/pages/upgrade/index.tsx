import {
  CircleAlert,
  DatabaseBackup,
  Download,
  RefreshCw,
  Upload,
} from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate } from "react-router-dom"
import { toast } from "sonner"

import {
  downloadArchive,
  getMaintenanceStatus,
  importArchive,
  runSchemaMigration,
} from "@/api/maintenance"
import type {
  ArchiveImportResult,
  MaintenanceStatus,
  SchemaMigrationResult,
} from "@/api/maintenance/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/class-names"

/** proto 的 int64 在 JSON 里是字符串，展示前统一转成数字。 */
function toNumber(value: number | string | undefined): number {
  const parsed = typeof value === "string" ? Number(value) : (value ?? 0)
  return Number.isFinite(parsed) ? parsed : 0
}

function formatNumber(value: number | string | undefined): string {
  return new Intl.NumberFormat().format(toNumber(value))
}

function formatDuration(value: number | string | undefined): string {
  const ms = toNumber(value)
  if (ms < 0) return "—"
  if (ms < 1000) return `${Math.round(ms)} ms`
  return `${(ms / 1000).toFixed(1)} s`
}

function errorMessage(err: unknown, fallback: string): string {
  return err instanceof Error && err.message ? err.message : fallback
}

function Stat({
  label,
  value,
  hint,
  tone = "default",
}: {
  label: string
  value: string
  hint?: string
  tone?: "default" | "warning"
}) {
  return (
    <div className="min-w-0 rounded-lg border border-border/60 bg-background/50 px-3 py-2">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd
        className={cn(
          "mt-0.5 truncate font-medium tabular-nums",
          tone === "warning" && "text-warning"
        )}
        title={value}
      >
        {value}
      </dd>
      {hint ? (
        <p className="mt-0.5 text-xs text-muted-foreground">{hint}</p>
      ) : null}
    </div>
  )
}

function TableChips({
  items,
  emptyLabel,
}: {
  items: MaintenanceStatus["tables"]
  emptyLabel: string
}) {
  const { t } = useTranslation()
  if (items.length === 0) {
    return <p className="text-xs text-muted-foreground">{emptyLabel}</p>
  }
  return (
    <div className="flex flex-wrap gap-1.5">
      {items.map((item) => (
        <span
          key={item.name}
          className={cn(
            "inline-flex min-w-0 items-center gap-1.5 rounded-md border border-border/60 bg-muted/40 px-2 py-1 font-mono text-xs",
            !item.exists && "border-warning-border text-muted-foreground"
          )}
        >
          <span className="truncate">{item.name}</span>
          <span className="text-muted-foreground">
            {item.exists
              ? t("maintenance.tableRows", { count: formatNumber(item.rows) })
              : t("maintenance.tableMissing")}
          </span>
        </span>
      ))}
    </div>
  )
}

// 升级与迁移：平台级页面，超级管理员在这里导出迁移包、到新环境导入，
// 并在服务端补跑数据库迁移。只读角色可以查看环境状态。
export function UpgradePage() {
  const { t } = useTranslation()
  const session = useSession()
  const [status, setStatus] = useState<MaintenanceStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [exporting, setExporting] = useState(false)
  const [importing, setImporting] = useState(false)
  const [migrating, setMigrating] = useState(false)
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [importOpen, setImportOpen] = useState(false)
  const [migrateOpen, setMigrateOpen] = useState(false)
  const [importError, setImportError] = useState("")
  const [migrateError, setMigrateError] = useState("")
  const [imported, setImported] = useState<ArchiveImportResult | null>(null)
  const [migrated, setMigrated] = useState<SchemaMigrationResult | null>(null)
  const fileInput = useRef<HTMLInputElement>(null)

  const canWrite = session.canSystem("system:migration:write")

  useEffect(() => {
    let active = true
    getMaintenanceStatus({ silent: true })
      .then((result) => {
        if (!active) return
        setStatus(result)
        setFailed(false)
      })
      .catch(() => {
        if (active) setFailed(true)
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
    }
  }, [attempt])

  if (!session.loading && !session.canSystem("system:migration:read")) {
    return <Navigate to="/forbidden" replace />
  }

  function reload() {
    setLoading(true)
    setAttempt((value) => value + 1)
  }

  async function handleExport() {
    setExporting(true)
    try {
      const { blob, filename } = await downloadArchive()
      const url = URL.createObjectURL(blob)
      const link = document.createElement("a")
      link.href = url
      link.download = filename
      document.body.appendChild(link)
      link.click()
      link.remove()
      URL.revokeObjectURL(url)
      toast.success(
        t("maintenance.exportDone", {
          tables: status?.archiveTableCount ?? 0,
          rows: formatNumber(status?.archiveRowCount ?? 0),
        })
      )
    } catch (err) {
      toast.error(errorMessage(err, t("maintenance.exportFailed")))
    } finally {
      setExporting(false)
    }
  }

  async function handleImport() {
    if (!selectedFile) return
    setImporting(true)
    setImportError("")
    try {
      const result = await importArchive(selectedFile)
      setImported(result)
      setImportOpen(false)
      setSelectedFile(null)
      toast.success(t("maintenance.importDoneTitle"))
    } catch (err) {
      setImportError(errorMessage(err, t("maintenance.importFailed")))
    } finally {
      setImporting(false)
    }
  }

  async function handleMigrate() {
    setMigrating(true)
    setMigrateError("")
    try {
      const result = await runSchemaMigration()
      setMigrated(result)
      setMigrateOpen(false)
      setAttempt((value) => value + 1)
      toast.success(
        t("maintenance.migrateDone", {
          tables: result.tableCount,
          duration: formatDuration(result.durationMs),
        })
      )
    } catch (err) {
      setMigrateError(errorMessage(err, t("maintenance.migrateFailed")))
    } finally {
      setMigrating(false)
    }
  }

  function relogin() {
    localStorage.removeItem("token")
    localStorage.removeItem("refreshToken")
    window.location.assign("/login")
  }

  const archivedTables = (status?.tables ?? []).filter((item) => item.archived)
  const runtimeTables = (status?.tables ?? []).filter((item) => !item.archived)
  const steps = [1, 2, 3, 4].map((index) => t(`maintenance.step${index}`))

  return (
    <div className="flex min-h-0 flex-col gap-5 overflow-y-auto">
      <PageHeader
        title={t("maintenance.title")}
        description={t("maintenance.description")}
      >
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={reload}
          disabled={loading}
        >
          <RefreshCw
            className={loading ? "motion-safe:animate-spin" : undefined}
            aria-hidden="true"
          />
          {t("maintenance.refresh")}
        </Button>
      </PageHeader>

      {!canWrite ? (
        <p className="rounded-md border border-border/60 bg-muted/30 px-3 py-2 text-sm text-muted-foreground">
          {t("maintenance.readOnly")}
        </p>
      ) : null}

      {loading && !status ? (
        <div
          className="grid gap-4 lg:grid-cols-2"
          aria-busy="true"
          aria-label={t("state.loading")}
        >
          <Skeleton className="h-56 rounded-xl" />
          <Skeleton className="h-56 rounded-xl" />
          <Skeleton className="h-44 rounded-xl" />
          <Skeleton className="h-44 rounded-xl" />
        </div>
      ) : failed && !status ? (
        <Card>
          <CardContent className="flex flex-col items-center gap-3 py-10 text-center">
            <CircleAlert className="size-6 text-destructive" aria-hidden="true" />
            <p className="font-medium">{t("maintenance.loadFailed")}</p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={reload}
            >
              {t("action.retry")}
            </Button>
          </CardContent>
        </Card>
      ) : status ? (
        <>
          <div className="grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <DatabaseBackup className="size-4" aria-hidden="true" />
                  {t("maintenance.statusTitle")}
                </CardTitle>
                <CardDescription>
                  {t("maintenance.statusDescription")}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <dl className="grid grid-cols-2 gap-3">
                  <Stat
                    label={t("maintenance.appVersion")}
                    value={status.appVersion || "—"}
                  />
                  <Stat
                    label={t("maintenance.archiveFormat")}
                    value={t("maintenance.archiveFormatValue", {
                      version: status.archiveFormatVersion,
                    })}
                  />
                  <Stat
                    label={t("maintenance.postgresVersion")}
                    value={status.postgresVersion || "—"}
                  />
                  <Stat
                    label={t("maintenance.schemaReady")}
                    value={
                      status.schemaReady
                        ? t("maintenance.schemaReadyValue")
                        : t("maintenance.schemaNotReady")
                    }
                    tone={status.schemaReady ? "default" : "warning"}
                  />
                  <Stat
                    label={t("maintenance.archiveTables")}
                    value={formatNumber(status.archiveTableCount)}
                  />
                  <Stat
                    label={t("maintenance.archiveRows")}
                    value={formatNumber(status.archiveRowCount)}
                  />
                  <div className="col-span-2">
                    <Stat
                      label={t("maintenance.encryptKey")}
                      value={status.encryptKeyCheck || "—"}
                    />
                  </div>
                </dl>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Download className="size-4" aria-hidden="true" />
                  {t("maintenance.exportTitle")}
                </CardTitle>
                <CardDescription>
                  {t("maintenance.exportDescription")}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-wrap items-center gap-3">
                <Button
                  type="button"
                  onClick={handleExport}
                  disabled={!canWrite || exporting || importing || migrating || !status.schemaReady}
                >
                  <Download
                    className={exporting ? "motion-safe:animate-pulse" : undefined}
                    aria-hidden="true"
                  />
                  {exporting
                    ? t("maintenance.exporting")
                    : t("maintenance.exportAction")}
                </Button>
                {!status.schemaReady ? (
                  <p className="text-sm text-warning">
                    {t("maintenance.schemaNotReady")}
                  </p>
                ) : null}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Upload className="size-4" aria-hidden="true" />
                  {t("maintenance.importTitle")}
                </CardTitle>
                <CardDescription>
                  {t("maintenance.importDescription")}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                {imported ? (
                  <div className="rounded-lg border border-success-border bg-success-soft px-3 py-2 text-sm">
                    <p className="font-medium">
                      {t("maintenance.importDoneTitle")}
                    </p>
                    <p className="mt-1 text-muted-foreground">
                      {t("maintenance.importDoneBody", {
                        tables: imported.tables,
                        rows: formatNumber(imported.rows),
                        version: imported.sourceVersion || "—",
                      })}
                    </p>
                    <Button
                      type="button"
                      size="sm"
                      className="mt-2"
                      onClick={relogin}
                    >
                      {t("maintenance.importRelogin")}
                    </Button>
                  </div>
                ) : null}
                <div className="flex flex-wrap items-center gap-3">
                  <input
                    ref={fileInput}
                    type="file"
                    className="hidden"
                    accept=".alioth-backup,application/gzip,application/x-gzip"
                    onChange={(event) => {
                      const file = event.target.files?.[0] ?? null
                      event.target.value = ""
                      if (!file) return
                      setImportError("")
                      setSelectedFile(file)
                      setImportOpen(true)
                    }}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => fileInput.current?.click()}
                    disabled={!canWrite || importing || exporting || migrating}
                  >
                    <Upload aria-hidden="true" />
                    {importing
                      ? t("maintenance.importing")
                      : t("maintenance.importAction")}
                  </Button>
                  <p className="text-xs text-muted-foreground">
                    {t("maintenance.importHint")}
                  </p>
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <RefreshCw className="size-4" aria-hidden="true" />
                  {t("maintenance.migrateTitle")}
                </CardTitle>
                <CardDescription>
                  {t("maintenance.migrateDescription")}
                </CardDescription>
              </CardHeader>
              <CardContent className="flex flex-col gap-3">
                {migrated ? (
                  <p className="rounded-lg border border-success-border bg-success-soft px-3 py-2 text-sm">
                    {t("maintenance.migrateDone", {
                      tables: migrated.tableCount,
                      duration: formatDuration(migrated.durationMs),
                    })}
                  </p>
                ) : null}
                <div className="flex flex-wrap items-center gap-3">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => {
                      setMigrateError("")
                      setMigrateOpen(true)
                    }}
                    disabled={!canWrite || migrating || exporting || importing}
                  >
                    <RefreshCw
                      className={
                        migrating ? "motion-safe:animate-spin" : undefined
                      }
                      aria-hidden="true"
                    />
                    {migrating
                      ? t("maintenance.migrating")
                      : t("maintenance.migrateAction")}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>

          <Card>
            <CardHeader>
              <CardTitle>{t("maintenance.tablesTitle")}</CardTitle>
              <CardDescription>
                {t("maintenance.tablesDescription")}
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              <div className="min-w-0">
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {t("maintenance.archivedTables", {
                    count: archivedTables.length,
                  })}
                </p>
                <TableChips
                  items={archivedTables}
                  emptyLabel={t("state.empty")}
                />
              </div>
              <div className="min-w-0">
                <p className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {t("maintenance.runtimeTables", {
                    count: runtimeTables.length,
                  })}
                </p>
                <TableChips items={runtimeTables} emptyLabel={t("state.empty")} />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t("maintenance.stepsTitle")}</CardTitle>
              <CardDescription>{t("maintenance.stepsNote")}</CardDescription>
            </CardHeader>
            <CardContent>
              <ol className="list-decimal space-y-2 pl-5 text-sm">
                {steps.map((step) => (
                  <li key={step}>{step}</li>
                ))}
              </ol>
            </CardContent>
          </Card>
        </>
      ) : null}

      <ConfirmAction
        open={importOpen}
        title={t("maintenance.importConfirmTitle")}
        description={t("maintenance.importConfirmBody")}
        confirmLabel={t("maintenance.importConfirmAction")}
        pendingLabel={t("maintenance.importing")}
        cancelLabel={t("action.cancel")}
        pending={importing}
        error={importError}
        onConfirm={handleImport}
        onOpenChange={(open) => {
          if (importing) return
          setImportOpen(open)
          if (!open) {
            setSelectedFile(null)
            setImportError("")
          }
        }}
      />

      <ConfirmAction
        open={migrateOpen}
        title={t("maintenance.migrateConfirmTitle")}
        description={t("maintenance.migrateConfirmBody")}
        confirmLabel={t("maintenance.migrateConfirmAction")}
        pendingLabel={t("maintenance.migrating")}
        cancelLabel={t("action.cancel")}
        pending={migrating}
        error={migrateError}
        onConfirm={handleMigrate}
        onOpenChange={(open) => {
          if (migrating) return
          setMigrateOpen(open)
          if (!open) setMigrateError("")
        }}
      />
    </div>
  )
}
