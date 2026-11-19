import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import {
  createApplicationVersion,
  deleteApplication,
  deleteApplicationVersion,
  deprecateApplicationVersion,
  getApplication,
  listApplicationVersions,
  publishApplicationVersion,
} from "@/api/application"
import type { ApplicationItem, ApplicationVersionItem } from "@/api/application/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { FormField } from "@/components/form-field"
import { HeaderActionsMenu } from "@/components/header-actions-menu"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Textarea } from "@/components/ui/textarea"
import { usePagedList } from "@/hooks/use-paged-list"
import { applicationCategoryLabel } from "@/lib/application-category"
import { canRunVersionAction } from "@/lib/application-version"
import { formatUnixDateTime } from "@/lib/date-time"
import { checkField } from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

import { ApplicationDialog } from "./app-dialog"
import { ApplicationDuplicateVersionDialog } from "./duplicate-version-dialog"
import { ApplicationReleaseDialog } from "./release-dialog"

export function ApplicationDetailPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "" } = useParams()
  const [app, setApp] = useState<ApplicationItem | null>(null)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteAppOpen, setDeleteAppOpen] = useState(false)
  const [deleteAppPending, setDeleteAppPending] = useState(false)
  const [deleteAppError, setDeleteAppError] = useState("")
  const [versionDialogOpen, setVersionDialogOpen] = useState(false)
  const [version, setVersion] = useState("")
  const [changelog, setChangelog] = useState("")
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState("")
  const [createFormError, setCreateFormError] = useState("")
  const [duplicateSource, setDuplicateSource] = useState<ApplicationVersionItem | null>(null)
  const [releaseVersionUid, setReleaseVersionUid] = useState<string | undefined>(undefined)
  const [releaseOpen, setReleaseOpen] = useState(false)
  const [deleting, setDeleting] = useState<ApplicationVersionItem | null>(null)
  const [actionPending, setActionPending] = useState(false)
  const [actionError, setActionError] = useState("")
  const { items, total, loading, failed, error, reload } = usePagedList<ApplicationVersionItem>({
    page: 1,
    pageSize: 100,
    filterKey: `${session.namespaceUid}|${uid}`,
    fetch: (page, pageSize) => listApplicationVersions(uid, { page, pageSize }),
  })

  useEffect(() => {
    if (!uid) return
    let active = true
    void getApplication(uid)
      .then((result) => {
        if (active) setApp(result.application)
      })
      .catch(() => undefined)
    return () => {
      active = false
    }
  }, [uid])

  if (!session.loading && !session.can("application:read")) {
    return <Navigate to="/forbidden" replace />
  }

  const writable = session.can("application:write")
  const releasable = session.can("release:write")

  async function createDraft() {
    const versionError = checkField({
      label: t("appStore.version"),
      value: version,
      required: true,
      max: 32,
    })
    if (versionError) {
      setCreateError(versionError)
      return
    }
    setCreating(true)
    setCreateError("")
    setCreateFormError("")
    try {
      const result = await createApplicationVersion({
        applicationUid: uid,
        version: version.trim(),
        changelog: changelog.trim(),
        resources: [],
      })
      setVersionDialogOpen(false)
      setVersion("")
      setChangelog("")
      toast.success(t("appStore.draftCreated"))
      navigate(`/appstore/apps/${uid}/versions/${result.version.uid}`)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setCreateError(failure.fields.version ?? failure.fields.name ?? "")
      setCreateFormError(failure.fields.version ? "" : failure.form)
    } finally {
      setCreating(false)
    }
  }

  async function handlePublish(item: ApplicationVersionItem) {
    setActionPending(true)
    try {
      await publishApplicationVersion(uid, item.uid)
      toast.success(t("appStore.published"))
      reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  async function handleDeprecate(item: ApplicationVersionItem) {
    setActionPending(true)
    try {
      await deprecateApplicationVersion(uid, item.uid)
      toast.success(t("appStore.deprecated"))
      reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  async function handleDeleteVersion() {
    if (!deleting) return
    setActionPending(true)
    setActionError("")
    try {
      await deleteApplicationVersion(uid, deleting.uid)
      setDeleting(null)
      toast.success(t("appStore.versionDeleted"))
      reload()
    } catch (err) {
      setActionError(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  async function handleDeleteApp() {
    setDeleteAppPending(true)
    setDeleteAppError("")
    try {
      await deleteApplication(uid)
      toast.success(t("appStore.appDeleted"))
      navigate("/appstore")
    } catch (err) {
      setDeleteAppError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeleteAppPending(false)
    }
  }

  function versionActions(item: ApplicationVersionItem) {
    return [
      ...(releasable && canRunVersionAction(item, "RELEASE")
        ? [{ key: "release", label: t("appStore.releaseAction"), onSelect: () => { setReleaseVersionUid(item.uid); setReleaseOpen(true) } }]
        : []),
      ...(writable && canRunVersionAction(item, "PUBLISH")
        ? [{ key: "publish", label: t("appStore.publish"), onSelect: () => void handlePublish(item) }]
        : []),
      ...(writable && canRunVersionAction(item, "DEPRECATE")
        ? [{ key: "deprecate", label: t("appStore.deprecate"), onSelect: () => void handleDeprecate(item) }]
        : []),
      ...(writable && canRunVersionAction(item, "COPY")
        ? [{ key: "copy", label: t("appStore.copyVersion"), onSelect: () => setDuplicateSource(item) }]
        : []),
      ...(writable && canRunVersionAction(item, "DELETE")
        ? [{ key: "delete", label: t("action.delete"), destructive: true, onSelect: () => { setActionError(""); setDeleting(item) } }]
        : []),
    ]
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={app?.displayName || app?.name || uid}
        description={app ? `${app.name} · ${applicationCategoryLabel(t, app.category)}${app.description ? ` · ${app.description}` : ""}` : t("state.loading")}
      >
        <Button variant="outline" onClick={() => navigate("/appstore")}>{t("action.goBack")}</Button>
        {releasable ? (
          <Button variant="secondary" onClick={() => { setReleaseVersionUid(undefined); setReleaseOpen(true) }}>
            {t("appStore.releaseAction")}
          </Button>
        ) : null}
        {writable ? <Button onClick={() => setVersionDialogOpen(true)}>{t("appStore.newVersion")}</Button> : null}
        {writable || session.can("application:delete") ? (
          <HeaderActionsMenu
            label={t("action.more")}
            actions={[
              ...(writable ? [{ key: "edit", label: t("appStore.editApp"), onSelect: () => setEditOpen(true) }] : []),
              ...(session.can("application:delete")
                ? [{ key: "delete", label: t("action.delete"), destructive: true, onSelect: () => { setDeleteAppError(""); setDeleteAppOpen(true) } }]
                : []),
            ]}
          />
        ) : null}
      </PageHeader>
      <div className="min-h-0 flex-1 overflow-y-auto">
        <div className="grid gap-6 pb-2">
          <ListFrame total={total} page={1} pageSize={100} showPager={false} onPrev={() => undefined} onNext={() => undefined}>
            {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
            {!loading && failed ? <p className="text-sm text-destructive">{error || t("state.failed")}</p> : null}
            {!loading && !failed && items.length === 0 ? (
              <div className="grid gap-2">
                <p className="text-sm text-muted-foreground">{t("appStore.emptyVersions")}</p>
                <p className="text-sm text-muted-foreground">{t("appStore.emptyVersionsHint")}</p>
              </div>
            ) : null}
            {!loading && !failed && items.length > 0 ? (
              <Table className={TABLE_MIN_WIDTH.default}>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("appStore.version")}</TableHead>
                    <TableHead>{t("appStore.status")}</TableHead>
                    <TableHead className={TABLE_NOWRAP}>{t("appStore.services.title")}</TableHead>
                    <TableHead className={TABLE_NOWRAP}>{t("appStore.resources")}</TableHead>
                    <TableHead className={TABLE_NOWRAP}>{t("action.createdAt")}</TableHead>
                    <TableActionsHead>{t("action.actions")}</TableActionsHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {items.map((item) => (
                    <TableRow key={item.uid}>
                      <TableCell>
                        <button
                          type="button"
                          className="text-left font-mono text-xs font-medium underline-offset-4 hover:underline"
                          onClick={() => navigate(`/appstore/apps/${uid}/versions/${item.uid}`)}
                        >
                          {item.version}
                        </button>
                        {item.changelog ? <p className="max-w-[28rem] truncate text-xs text-muted-foreground">{item.changelog}</p> : null}
                      </TableCell>
                      <TableCell><StatusBadge value={item.status} /></TableCell>
                      <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{item.componentCount || item.components?.length || 0}</TableCell>
                      <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{item.resourceCount || item.resources?.length || 0}</TableCell>
                      <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(item.createdAt) || "—"}</TableCell>
                      <TableActionsCell>
                        <RowActions
                          nowrap
                          detailLabel={t("action.detail")}
                          moreLabel={t("action.more")}
                          onDetail={() => navigate(`/appstore/apps/${uid}/versions/${item.uid}`)}
                          actions={versionActions(item)}
                        />
                      </TableActionsCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            ) : null}
          </ListFrame>
          {app?.readme ? (
            <div className="rounded-xl border border-border/60 bg-card p-4">
              <h2 className="mb-2 font-heading text-base font-semibold">{t("appStore.readme")}</h2>
              <pre className="whitespace-pre-wrap text-sm text-foreground/90">{app.readme}</pre>
            </div>
          ) : null}
        </div>
      </div>

      <ApplicationDialog
        key={editOpen ? `edit-${app?.uid ?? ""}` : "edit-closed"}
        open={editOpen}
        application={app}
        onOpenChange={setEditOpen}
        onSaved={(item) => {
          setEditOpen(false)
          setApp(item)
          toast.success(t("state.saved"))
        }}
      />

      <Dialog open={versionDialogOpen} onOpenChange={setVersionDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{t("appStore.newVersion")}</DialogTitle></DialogHeader>
          <DialogBody className="grid gap-4">
            <FormField label={t("appStore.version")} htmlFor="new-version" required error={createError} hint={t("appStore.versionHelp")}>
              <Input id="new-version" className="font-mono" value={version} placeholder={t("appStore.versionPlaceholder")} onChange={(event) => setVersion(event.target.value)} />
            </FormField>
            <FormField label={t("appStore.changelog")} htmlFor="new-changelog" hint={t("appStore.changelogHelp")}>
              <Textarea id="new-changelog" value={changelog} placeholder={t("appStore.changelogPlaceholder")} onChange={(event) => setChangelog(event.target.value)} />
            </FormField>
          </DialogBody>
          {createFormError ? <p className="text-sm text-destructive">{createFormError}</p> : null}
          <DialogFooter>
            <Button variant="ghost" onClick={() => setVersionDialogOpen(false)}>{t("action.cancel")}</Button>
            <Button disabled={creating} onClick={() => void createDraft()}>{creating ? t("action.saving") : t("action.create")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ApplicationDuplicateVersionDialog
        open={duplicateSource !== null}
        applicationUid={uid}
        source={duplicateSource}
        onOpenChange={(open) => { if (!open) setDuplicateSource(null) }}
        onDuplicated={(item) => {
          setDuplicateSource(null)
          toast.success(t("appStore.versionCopied", { version: item.version }))
          navigate(`/appstore/apps/${uid}/versions/${item.uid}`)
        }}
      />

      <ApplicationReleaseDialog
        key={`release-${releaseVersionUid ?? "any"}-${releaseOpen ? "open" : "closed"}`}
        open={releaseOpen}
        applicationUid={uid}
        initialVersionUid={releaseVersionUid}
        onOpenChange={setReleaseOpen}
        onReleased={(deployment) => {
          setReleaseOpen(false)
          toast.success(t("appStore.operationSubmitted"))
          if (deployment) {
            navigate(`/appstore/deployments/${deployment.uid}`)
          } else {
            reload()
          }
        }}
      />

      <ConfirmAction
        open={deleting !== null}
        title={t("appStore.deleteVersionTitle")}
        description={t("appStore.deleteVersionDescription")}
        confirmLabel={t("action.delete")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={actionPending}
        error={actionError}
        onOpenChange={(open) => { if (!open) setDeleting(null) }}
        onConfirm={() => void handleDeleteVersion()}
      />

      <ConfirmAction
        open={deleteAppOpen}
        title={t("appStore.deleteAppTitle")}
        description={t("appStore.deleteAppDescription", { name: app?.name ?? "" })}
        confirmLabel={t("action.delete")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={deleteAppPending}
        error={deleteAppError}
        onOpenChange={setDeleteAppOpen}
        onConfirm={() => void handleDeleteApp()}
      />
    </section>
  )
}
