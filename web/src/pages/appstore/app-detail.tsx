import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import {
  createApplicationVersion,
  deprecateApplicationVersion,
  deleteApplicationVersion,
  getApplication,
  listApplicationVersions,
  publishApplicationVersion,
} from "@/api/application"
import type { ApplicationItem, ApplicationVersionItem } from "@/api/application/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { FormField } from "@/components/form-field"
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
import { usePagedList } from "@/hooks/use-paged-list"
import { formatUnixDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

import { DeployDialog } from "./deploy-dialog"

export function ApplicationDetailPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "" } = useParams()
  const [app, setApp] = useState<ApplicationItem | null>(null)
  const [versionDialogOpen, setVersionDialogOpen] = useState(false)
  const [version, setVersion] = useState("")
  const [changelog, setChangelog] = useState("")
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState("")
  const [deployVersion, setDeployVersion] = useState<ApplicationVersionItem | null>(null)
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

  async function createDraft() {
    if (!version.trim()) {
      setCreateError(t("form.required"))
      return
    }
    setCreating(true)
    setCreateError("")
    try {
      const result = await createApplicationVersion(uid, { applicationUid: uid, version: version.trim(), changelog: changelog.trim(), resources: [] })
      setVersionDialogOpen(false)
      toast.success(t("appStore.draftCreated"))
      navigate(`/appstore/apps/${uid}/versions/${result.version.uid}`)
    } catch (err) {
      setCreateError(formErrorMessage(err, t("state.failed")))
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

  async function handleDelete() {
    if (!deleting) return
    setActionPending(true)
    setActionError("")
    try {
      await deleteApplicationVersion(uid, deleting.uid)
      setDeleting(null)
      toast.success(t("action.delete"))
      reload()
    } catch (err) {
      setActionError(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={app?.displayName || app?.name || uid}
        description={app ? `${app.name} · ${t(`appStore.categories.${app.category}`)}${app.description ? ` · ${app.description}` : ""}` : t("state.loading")}
      >
        <Button variant="outline" onClick={() => navigate("/appstore")}>{t("action.goBack")}</Button>
        {session.can("application:write") ? <Button onClick={() => setVersionDialogOpen(true)}>{t("appStore.newVersion")}</Button> : null}
      </PageHeader>
      <ListFrame total={total} page={1} pageSize={100} showPager={false} onPrev={() => undefined} onNext={() => undefined}>
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? <p className="text-sm text-destructive">{error || t("state.failed")}</p> : null}
        {!loading && !failed && items.length === 0 ? <p className="text-sm text-muted-foreground">{t("appStore.emptyVersions")}</p> : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader><TableRow><TableHead>{t("appStore.version")}</TableHead><TableHead>{t("appStore.status")}</TableHead><TableHead>{t("appStore.resources")}</TableHead><TableHead>{t("appStore.images")}</TableHead><TableHead className={TABLE_NOWRAP}>{t("action.createdAt")}</TableHead><TableHead className="text-right">{t("action.actions")}</TableHead></TableRow></TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.uid}>
                  <TableCell className="font-mono text-xs">{item.version}</TableCell>
                  <TableCell><Badge variant={item.status === "PUBLISHED" ? "secondary" : "outline"}>{item.status}</Badge></TableCell>
                  <TableCell className="font-mono text-xs">{item.resources.length}</TableCell>
                  <TableCell className="font-mono text-xs">{item.images.length}</TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>{formatUnixDateTime(item.createdAt) || "—"}</TableCell>
                  <TableCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => navigate(`/appstore/apps/${uid}/versions/${item.uid}`)}
                      actions={[
                        ...(item.status === "PUBLISHED" ? [{ key: "deploy", label: t("appStore.deployAction"), onSelect: () => setDeployVersion(item) }] : []),
                        ...(item.status === "DRAFT" ? [{ key: "publish", label: t("appStore.publish"), onSelect: () => void handlePublish(item) }] : []),
                        ...(item.status === "PUBLISHED" ? [{ key: "deprecate", label: t("appStore.deprecate"), onSelect: () => void handleDeprecate(item) }] : []),
                        ...(item.status === "DRAFT" ? [{ key: "delete", label: t("action.delete"), destructive: true, onSelect: () => { setActionError(""); setDeleting(item) } }] : []),
                      ]}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      {app?.readme ? (
        <div className="rounded-xl border border-border/60 bg-card p-4">
          <h2 className="mb-2 font-heading text-base font-semibold">{t("appStore.readme")}</h2>
          <pre className="whitespace-pre-wrap text-sm text-muted-foreground">{app.readme}</pre>
        </div>
      ) : null}
      <Dialog open={versionDialogOpen} onOpenChange={setVersionDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader><DialogTitle>{t("appStore.newVersion")}</DialogTitle></DialogHeader>
          <DialogBody className="grid gap-4">
            <FormField label={t("appStore.version")} htmlFor="new-version" required error={createError}>
              <Input id="new-version" className="font-mono" value={version} placeholder="1.0.0" onChange={(event) => setVersion(event.target.value)} />
            </FormField>
            <FormField label={t("appStore.changelog")} htmlFor="new-changelog">
              <Textarea id="new-changelog" value={changelog} onChange={(event) => setChangelog(event.target.value)} />
            </FormField>
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setVersionDialogOpen(false)}>{t("action.cancel")}</Button>
            <Button disabled={creating} onClick={() => void createDraft()}>{creating ? t("action.saving") : t("action.create")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      {deployVersion ? (
        <DeployDialog
          open
          applicationUid={uid}
          versionUid={deployVersion.uid}
          version={deployVersion.version}
          hasClusterScoped={deployVersion.resources.some((resource) => resource.clusterScoped)}
          onOpenChange={(open) => { if (!open) setDeployVersion(null) }}
          onCreated={(deployment) => {
            setDeployVersion(null)
            toast.success(t("appStore.deployed"))
            navigate(`/appstore/deployments/${deployment.uid}`)
          }}
        />
      ) : null}
      <ConfirmAction
        open={deleting !== null}
        title={t("appStore.deleteVersionTitle")}
        description={t("appStore.deleteVersionDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={actionPending}
        error={actionError}
        onOpenChange={(open) => { if (!open) setDeleting(null) }}
        onConfirm={() => void handleDelete()}
      />
    </section>
  )
}
