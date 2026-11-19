import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import {
  deleteApplicationVersion,
  deleteApplicationVersionComponent,
  deleteApplicationVersionResource,
  deprecateApplicationVersion,
  getApplicationVersion,
  publishApplicationVersion,
  saveApplicationVersionComponent,
  validateApplicationVersion,
} from "@/api/application"
import type {
  ApplicationDeploymentItem,
  ApplicationResource,
  ApplicationVersionComponent,
  ApplicationVersionItem,
} from "@/api/application/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { HeaderActionsMenu, type HeaderAction } from "@/components/header-actions-menu"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  buildVersionComponents,
  canRunVersionAction,
  listVersionResources,
} from "@/lib/application-version"
import { formatUnixDateTime } from "@/lib/date-time"
import { enumLabel } from "@/lib/enum-label"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

import { ApplicationComponentDialog } from "./application-component-dialog"
import { ApplicationDuplicateVersionDialog } from "./duplicate-version-dialog"
import { ApplicationReleaseDialog } from "./release-dialog"
import { ApplicationResourceDetailSheet } from "./application-resource-detail-sheet"
import { ApplicationResourceFormDialog } from "./application-resource-form-dialog"
import { ApplicationServicesTab } from "./application-services-tab"

const TABS = ["services", "resources", "images"] as const
type VersionTab = (typeof TABS)[number]

const RESOURCE_TABLE_COLUMNS = 7

interface ResourceDialogState {
  mode: "create" | "edit"
  component: string
  resource: ApplicationResource | null
}

export function ApplicationVersionPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "", versionUid = "" } = useParams()
  const [params, setParams] = useSearchParams()
  const tabParam = params.get("tab")
  const tab: VersionTab = TABS.includes(tabParam as VersionTab) ? (tabParam as VersionTab) : "services"
  const [version, setVersion] = useState<ApplicationVersionItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [actionPending, setActionPending] = useState(false)
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [validationWarnings, setValidationWarnings] = useState<string[]>([])
  const [resourceDialog, setResourceDialog] = useState<ResourceDialogState | null>(null)
  const [detailResource, setDetailResource] = useState<ApplicationResource | null>(null)
  const [deletingResource, setDeletingResource] = useState<ApplicationResource | null>(null)
  const [deleteResourceError, setDeleteResourceError] = useState("")
  const [deleteVersionOpen, setDeleteVersionOpen] = useState(false)
  const [duplicateOpen, setDuplicateOpen] = useState(false)
  const [releaseOpen, setReleaseOpen] = useState(false)
  const [componentDialog, setComponentDialog] = useState<{
    mode: "create" | "edit"
    component: ApplicationVersionComponent | null
  } | null>(null)
  const [deleteComponent, setDeleteComponent] = useState<ApplicationVersionComponent | null>(null)
  const [componentPending, setComponentPending] = useState(false)
  const [componentError, setComponentError] = useState("")

  useEffect(() => {
    let active = true
    void getApplicationVersion(uid, versionUid)
      .then((result) => {
        if (active) setVersion(result.version)
      })
      .catch((err) => {
        if (active) setError(formErrorMessage(err, t("state.failed")))
      })
      .finally(() => {
        if (active) setLoading(false)
      })
    return () => {
      active = false
    }
  }, [uid, versionUid, t])

  const components = useMemo(() => buildVersionComponents(version), [version])
  const resources = useMemo(() => listVersionResources(version), [version])

  if (!session.loading && !session.can("application:read")) {
    return <Navigate to="/forbidden" replace />
  }
  if (!version) {
    return (
      <section className="flex min-h-0 flex-1 flex-col gap-6">
        <PageHeader title={t("appStore.version")} description={loading ? t("state.loading") : error} />
      </section>
    )
  }

  const writable = session.can("application:write")
  const releasable = session.can("release:write")
  const editable = version.status === "DRAFT" && writable
  const copyable = writable && canRunVersionAction(version, "COPY")
  const headerActions: HeaderAction[] = [
    ...(writable && canRunVersionAction(version, "DEPRECATE")
      ? [{
          key: "deprecate",
          label: t("appStore.deprecate"),
          onSelect: () => {
            setActionPending(true)
            void deprecateApplicationVersion(uid, versionUid)
              .then((result) => {
                setVersion(result.version)
                toast.success(t("appStore.deprecated"))
              })
              .catch((err) => {
                setError(formErrorMessage(err, t("state.failed")))
              })
              .finally(() => setActionPending(false))
          },
        }]
      : []),
    ...(copyable ? [{ key: "copy", label: t("appStore.copyVersion"), onSelect: () => setDuplicateOpen(true) }] : []),
    ...(version.status === "DRAFT" ? [{ key: "validate", label: t("appStore.validate"), onSelect: () => void validate() }] : []),
    ...(writable && canRunVersionAction(version, "DELETE")
      ? [{ key: "delete", label: t("action.delete"), destructive: true, onSelect: () => setDeleteVersionOpen(true) }]
      : []),
  ]
  const readOnlyHint = editable
    ? undefined
    : version.status === "DRAFT"
      ? t("appStore.readOnlyNoPermission")
      : t("appStore.readOnlyVersionHint")

  function setTab(next: string) {
    const query = new URLSearchParams(params)
    query.set("tab", next)
    setParams(query)
  }

  function openCreateResource(component?: string) {
    setResourceDialog({ mode: "create", component: component ?? "", resource: null })
  }

  function openEditResource(resource: ApplicationResource) {
    setDetailResource(null)
    setResourceDialog({ mode: "edit", component: resource.component || "", resource })
  }

  function applyVersion(next: ApplicationVersionItem) {
    setVersion(next)
  }

  async function publish() {
    setActionPending(true)
    setError("")
    try {
      const result = await publishApplicationVersion(uid, versionUid)
      setVersion(result.version)
      toast.success(t("appStore.published"))
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  async function validate() {
    setActionPending(true)
    setValidationErrors([])
    setValidationWarnings([])
    try {
      const result = await validateApplicationVersion(uid, versionUid)
      setValidationErrors(result.errors)
      setValidationWarnings(result.warnings)
      if (result.version) setVersion(result.version)
      if (result.errors.length === 0) toast.success(t("appStore.validationPassed"))
    } catch (err) {
      setValidationErrors([formErrorMessage(err, t("state.failed"))])
    } finally {
      setActionPending(false)
    }
  }

  async function removeVersion() {
    setActionPending(true)
    setError("")
    try {
      await deleteApplicationVersion(uid, versionUid)
      toast.success(t("appStore.versionDeleted"))
      navigate(`/appstore/apps/${uid}`)
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
      setActionPending(false)
    }
  }

  async function removeResource() {
    if (!deletingResource || !version) return
    setActionPending(true)
    setDeleteResourceError("")
    try {
      const result = await deleteApplicationVersionResource({
        applicationUid: uid,
        uid: versionUid,
        resourceUid: deletingResource.uid,
        expectedRevision: version.revision,
      })
      setVersion(result.version)
      setDeletingResource(null)
      toast.success(t("appStore.resourceDeleted"))
    } catch (err) {
      setDeleteResourceError(formErrorMessage(err, t("state.failed")))
    } finally {
      setActionPending(false)
    }
  }

  async function saveComponent(name: string, description: string) {
    if (!componentDialog || !version) return
    const target = componentDialog.component
    setComponentPending(true)
    setComponentError("")
    try {
      const result = await saveApplicationVersionComponent({
        applicationUid: uid,
        uid: versionUid,
        componentUid: target?.uid ?? "",
        name,
        description,
        expectedRevision: version.revision,
      })
      setVersion(result.version)
      setComponentDialog(null)
      toast.success(target ? t("appStore.services.renamed") : t("appStore.services.created"))
    } catch (err) {
      setComponentError(formErrorMessage(err, t("state.failed")))
    } finally {
      setComponentPending(false)
    }
  }

  async function deleteService() {
    if (!deleteComponent || !version) return
    const target = deleteComponent
    setComponentPending(true)
    setComponentError("")
    try {
      const result = await deleteApplicationVersionComponent({
        applicationUid: uid,
        uid: versionUid,
        componentUid: target.uid,
        expectedRevision: version.revision,
      })
      setVersion(result.version)
      setDeleteComponent(null)
      toast.success(t("appStore.services.deleted"))
    } catch (err) {
      setComponentError(formErrorMessage(err, t("state.failed")))
    } finally {
      setComponentPending(false)
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={version.version}
        description={t("appStore.versionDescription", {
          status: enumLabel(t, version.status),
          components: version.componentCount ?? components.length,
          resources: version.resourceCount ?? resources.length,
          images: version.images.length,
        })}
        titleAccessory={<StatusBadge value={version.status} />}
      >
        <Button variant="outline" onClick={() => navigate(`/appstore/apps/${uid}`)}>{t("action.goBack")}</Button>
        {editable && canRunVersionAction(version, "PUBLISH") ? (
          <Button disabled={actionPending} onClick={() => void publish()}>{t("appStore.publish")}</Button>
        ) : null}
        {releasable && canRunVersionAction(version, "RELEASE") ? (
          <Button variant="secondary" onClick={() => setReleaseOpen(true)}>{t("appStore.releaseAction")}</Button>
        ) : null}
        {headerActions.length > 0 ? (
          <HeaderActionsMenu label={t("action.more")} actions={headerActions} />
        ) : null}
      </PageHeader>

      <div className="grid gap-2">
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        {validationErrors.length > 0 ? (
          <div role="alert" className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
            <p className="mb-1 font-medium">{t("appStore.validationErrors")}</p>
            <ul className="list-disc pl-5 break-words">{validationErrors.map((item) => <li key={item}>{item}</li>)}</ul>
          </div>
        ) : null}
        {validationWarnings.length > 0 ? (
          <div className="rounded-lg border border-warning-border bg-warning-soft p-3 text-sm text-warning">
            <p className="mb-1 font-medium">{t("appStore.validationWarnings")}</p>
            <ul className="list-disc pl-5 break-words">{validationWarnings.map((item) => <li key={item}>{item}</li>)}</ul>
          </div>
        ) : null}
      </div>

      <Tabs value={tab} onValueChange={setTab} className="min-h-0 flex-1">
        <TabsList className="shrink-0">
          <TabsTrigger value="services">{t("appStore.tabs.services")}</TabsTrigger>
          <TabsTrigger value="resources">{t("appStore.tabs.resources")}</TabsTrigger>
          <TabsTrigger value="images">{t("appStore.tabs.images")}</TabsTrigger>
        </TabsList>

        <TabsContent value="services" className="min-h-0 flex-1 overflow-y-auto pb-2">
          <ApplicationServicesTab
            version={version}
            components={components}
            editable={editable}
            readOnlyHint={readOnlyHint}
            onCreateService={() => {
              setComponentError("")
              setComponentDialog({ mode: "create", component: null })
            }}
            onEditService={(component) => {
              setComponentError("")
              setComponentDialog({ mode: "edit", component })
            }}
            onDeleteService={(component) => {
              setComponentError("")
              setDeleteComponent(component)
            }}
            onAddResource={(component) => openCreateResource(component.name)}
            onOpenResource={(resource) => setDetailResource(resource)}
            onEditResource={openEditResource}
            onDeleteResource={(resource) => {
              setDeleteResourceError("")
              setDeletingResource(resource)
            }}
          />
        </TabsContent>

        <TabsContent value="resources" className="min-h-0 flex-1 overflow-y-auto pb-2">
          <div className="grid gap-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h2 className="font-heading text-base font-semibold">{t("appStore.resourceList")}</h2>
              {editable ? (
                <Button
                  variant="outline"
                  size="sm"
                  disabled={components.length === 0}
                  title={components.length === 0 ? t("appStore.services.createFirstHint") : undefined}
                  onClick={() => openCreateResource(components[0]?.name)}
                >
                  {t("appStore.addResource")}
                </Button>
              ) : null}
            </div>
            <div className="overflow-hidden rounded-xl border border-border/60 bg-card">
              <Table className={TABLE_MIN_WIDTH.wide}>
                <TableHeader>
                  <TableRow>
                    <TableHead>{t("appStore.resourceKind")}</TableHead>
                    <TableHead>{t("appStore.resourceName")}</TableHead>
                    <TableHead>{t("appStore.services.component")}</TableHead>
                    <TableHead>{t("appStore.resourceScope")}</TableHead>
                    <TableHead>{t("appStore.images")}</TableHead>
                    <TableHead>{t("appStore.secretKeys")}</TableHead>
                    <TableActionsHead>{t("action.actions")}</TableActionsHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {resources.map((resource) => (
                    <TableRow key={resource.uid || `${resource.kind}-${resource.name}`}>
                      <TableCell className="font-mono text-xs">{resource.kind || "—"}</TableCell>
                      <TableCell className="font-mono text-xs">{resource.name || "—"}</TableCell>
                      <TableCell className="font-mono text-xs">{resource.component || t("appStore.services.defaultComponent")}</TableCell>
                      <TableCell><StatusBadge value={resource.scope} /></TableCell>
                      <TableCell className="font-mono text-xs">{resource.images?.length ?? 0}</TableCell>
                      <TableCell className="font-mono text-xs">{resource.secretKeys?.join(", ") || "—"}</TableCell>
                      <TableActionsCell className={TABLE_NOWRAP}>
                        <RowActions
                          nowrap
                          detailLabel={t("action.detail")}
                          moreLabel={t("action.more")}
                          onDetail={() => setDetailResource(resource)}
                          actions={editable ? [
                            { key: "edit", label: t("appStore.editResource"), onSelect: () => openEditResource(resource) },
                            { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => { setDeleteResourceError(""); setDeletingResource(resource) } },
                          ] : []}
                        />
                      </TableActionsCell>
                    </TableRow>
                  ))}
                  {resources.length === 0 ? (
                    <TableRow><TableCell colSpan={RESOURCE_TABLE_COLUMNS} className="text-sm text-muted-foreground">{t("appStore.emptyResources")}</TableCell></TableRow>
                  ) : null}
                </TableBody>
              </Table>
            </div>
            <div className="grid gap-2 rounded-xl border border-border/60 bg-card p-4">
              <h2 className="font-heading text-base font-semibold">{t("appStore.changelog")}</h2>
              <p className="text-sm whitespace-pre-wrap text-foreground/90">{version.changelog || "—"}</p>
              <p className="text-xs text-muted-foreground">
                {t("appStore.versionMeta", {
                  revision: version.revision,
                  createdAt: formatUnixDateTime(version.createdAt) || "—",
                  publishedAt: formatUnixDateTime(version.publishedAt) || "—",
                })}
              </p>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="images" className="min-h-0 flex-1 overflow-y-auto pb-2">
          <div className="grid gap-3">
            {components.map((component) => {
              const images = version.images.filter(
                (image) => (image.component || "default") === component.name,
              )
              return (
                <div key={component.uid || component.name} className="grid gap-2 rounded-xl border border-border/60 bg-card p-4">
                  <h3 className="font-heading text-sm font-semibold">
                    {component.name}
                    <span className="ml-2 text-xs font-normal text-muted-foreground">
                      {t("appStore.services.resourceCount", { count: component.resourceCount || component.resources.length })}
                    </span>
                  </h3>
                  {images.length === 0 ? <p className="text-sm text-muted-foreground">—</p> : (
                    <ul className="grid gap-2">
                      {images.map((image) => (
                        <li key={`${image.containerPath}-${image.image}`} className="grid gap-1 rounded-lg border border-border/60 p-2">
                          <div className="flex min-w-0 flex-wrap items-center gap-2">
                            <span className="min-w-0 font-mono text-xs break-all">{image.image}</span>
                            <StatusBadge
                              value={image.external || image.source === "EXTERNAL" ? "EXTERNAL" : "REGISTRY"}
                              label={image.external || image.source === "EXTERNAL" ? t("appStore.services.externalImage") : image.registryName || [image.project, image.repository].filter(Boolean).join("/") || t("enum.REGISTRY")}
                            />
                          </div>
                          {image.resolvedImage && image.resolvedImage !== image.image ? (
                            <p className="font-mono text-xs text-muted-foreground break-all">
                              {t("appStore.actualImage")}：{image.resolvedImage}
                            </p>
                          ) : null}
                          <p className="font-mono text-xs text-muted-foreground">
                            {image.digest || t("appStore.imagePicker.digestUnpinned")} · {image.platforms.join(" · ") || "—"}
                          </p>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              )
            })}
            {components.length === 0 ? <p className="text-sm text-muted-foreground">{t("appStore.noImages")}</p> : null}
          </div>
        </TabsContent>
      </Tabs>

      {resourceDialog ? (
        <ApplicationResourceFormDialog
          key={`${resourceDialog.mode}-${resourceDialog.resource?.uid ?? "new"}-${version.revision}`}
          open
          mode={resourceDialog.mode}
          applicationUid={uid}
          versionUid={versionUid}
          revision={version.revision}
          initialComponent={resourceDialog.component}
          components={components.map((component) => component.name)}
          resource={resourceDialog.resource}
          onOpenChange={(open) => { if (!open) setResourceDialog(null) }}
          onSaved={(next) => {
            applyVersion(next)
            setResourceDialog(null)
            toast.success(t("state.saved"))
          }}
        />
      ) : null}

      <ApplicationResourceDetailSheet
        open={detailResource !== null}
        applicationUid={uid}
        versionUid={versionUid}
        resource={detailResource}
        images={version.images}
        editable={editable}
        onEdit={() => { if (detailResource) openEditResource(detailResource) }}
        onDelete={() => {
          if (!detailResource) return
          setDetailResource(null)
          setDeleteResourceError("")
          setDeletingResource(detailResource)
        }}
        onOpenChange={(open) => { if (!open) setDetailResource(null) }}
      />

      <ApplicationDuplicateVersionDialog
        open={duplicateOpen}
        applicationUid={uid}
        source={version}
        onOpenChange={setDuplicateOpen}
        onDuplicated={(next) => {
          setDuplicateOpen(false)
          toast.success(t("appStore.versionCopied", { version: next.version }))
          navigate(`/appstore/apps/${uid}/versions/${next.uid}`)
        }}
      />

      <ApplicationReleaseDialog
        key={`release-${version.uid}-${releaseOpen ? "open" : "closed"}`}
        open={releaseOpen}
        applicationUid={uid}
        initialVersionUid={version.uid}
        onOpenChange={setReleaseOpen}
        onReleased={(deployment: ApplicationDeploymentItem | null) => {
          setReleaseOpen(false)
          toast.success(t("appStore.operationSubmitted"))
          if (deployment) {
            navigate(`/appstore/deployments/${deployment.uid}`)
          }
        }}
      />

      <ApplicationComponentDialog
        key={componentDialog ? `${componentDialog.mode}-${componentDialog.component?.uid ?? "new"}` : "component-closed"}
        open={componentDialog !== null}
        mode={componentDialog?.mode ?? "create"}
        initialName={componentDialog?.component?.name ?? ""}
        initialDescription={componentDialog?.component?.description ?? ""}
        pending={componentPending}
        error={componentError}
        onOpenChange={(open) => { if (!open) setComponentDialog(null) }}
        onSave={(name, description) => void saveComponent(name, description)}
      />

      <ConfirmAction
        open={deleteVersionOpen}
        title={t("appStore.deleteVersionTitle")}
        description={t("appStore.deleteVersionDescription")}
        confirmLabel={t("action.delete")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={actionPending}
        onOpenChange={setDeleteVersionOpen}
        onConfirm={() => void removeVersion()}
      />

      <ConfirmAction
        open={deletingResource !== null}
        title={t("appStore.deleteResourceTitle")}
        description={t("appStore.deleteResourceDescription", { name: deletingResource?.name ?? "" })}
        confirmLabel={t("action.delete")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={actionPending}
        error={deleteResourceError}
        onOpenChange={(open) => { if (!open) setDeletingResource(null) }}
        onConfirm={() => void removeResource()}
      />

      <ConfirmAction
        open={deleteComponent !== null}
        title={t("appStore.services.deleteTitle")}
        description={t("appStore.services.deleteDescription", {
          name: deleteComponent?.name ?? "",
          count: deleteComponent?.resources.length ?? 0,
        })}
        confirmLabel={t("action.delete")}
        pendingLabel={t("action.saving")}
        cancelLabel={t("action.cancel")}
        pending={componentPending}
        error={componentError}
        onOpenChange={(open) => { if (!open) setDeleteComponent(null) }}
        onConfirm={() => void deleteService()}
      />
    </section>
  )
}
