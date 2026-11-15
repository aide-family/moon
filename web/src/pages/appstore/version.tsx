import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams } from "react-router-dom"
import { toast } from "sonner"

import {
  deleteApplicationVersion,
  getApplicationVersion,
  publishApplicationVersion,
  updateApplicationVersion,
  validateApplicationVersion,
} from "@/api/application"
import type { ApplicationVersionItem } from "@/api/application/types"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { PageHeader } from "@/components/page-header"
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
import { Textarea } from "@/components/ui/textarea"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

const KINDS = ["Deployment", "StatefulSet", "DaemonSet", "CronJob", "Service", "Ingress", "ConfigMap", "Secret", "HorizontalPodAutoscaler", "NetworkPolicy", "ClusterRole", "CustomResourceDefinition", "CustomResource"]

const TEMPLATES: Record<string, string> = {
  Deployment: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 1
  selector:
    matchLabels: {app: app}
  template:
    metadata:
      labels: {app: app}
    spec:
      containers:
        - name: app
          image: nginx:1.27.2
          ports: [{containerPort: 80}]`,
  Service: `apiVersion: v1
kind: Service
metadata:
  name: app
spec:
  selector: {app: app}
  ports: [{port: 80, targetPort: 80}]`,
  ConfigMap: `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  KEY: value`,
  Secret: `apiVersion: v1
kind: Secret
metadata:
  name: app-secret
stringData:
  PASSWORD: change-me`,
  Ingress: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
spec:
  ingressClassName: nginx
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: app
                port: {number: 80}`,
  ClusterRole: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: app-reader
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list"]`,
  CustomResourceDefinition: `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  names:
    kind: Example
    plural: examples
  scope: Namespaced
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object`,
  CustomResource: `apiVersion: example.com/v1
kind: Example
metadata:
  name: example`,
}

interface ResourceDraft {
  item: string
  manifest: string
  kind?: string
  name?: string
  scope?: string
}

function inspectManifest(manifest: string): { kind?: string; name?: string; scope?: string } {
  const kind = manifest.match(/^kind:\s*([A-Za-z0-9]+)/m)?.[1]
  const name = manifest.match(/^\s*name:\s*([A-Za-z0-9.-]+)/m)?.[1]
  const clusterScoped = ["ClusterRole", "ClusterRoleBinding", "CustomResourceDefinition", "PersistentVolume", "StorageClass", "PriorityClass", "IngressClass"].includes(kind || "")
  return { kind, name, scope: clusterScoped ? "CLUSTER" : kind ? "NAMESPACED" : undefined }
}

interface YamlDialogProps {
  open: boolean
  title: string
  initialKind: string
  initialManifest: string
  onOpenChange: (open: boolean) => void
  onSave: (manifest: string) => void
}

function YamlResourceDialog({ open, title, initialKind, initialManifest, onOpenChange, onSave }: YamlDialogProps) {
  const { t } = useTranslation()
  const [kind, setKind] = useState(initialKind || "Deployment")
  const [manifest, setManifest] = useState(initialManifest || TEMPLATES[initialKind] || TEMPLATES.Deployment)
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader><DialogTitle>{title}</DialogTitle></DialogHeader>
        <DialogBody className="grid gap-4">
          <FormField label={t("appStore.resourceKind")} htmlFor="resource-kind">
            <Select value={kind} onValueChange={(value) => { setKind(value); setManifest(TEMPLATES[value] || TEMPLATES.CustomResource) }}>
              <SelectTrigger id="resource-kind"><SelectValue /></SelectTrigger>
              <SelectContent>{KINDS.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}</SelectContent>
            </Select>
          </FormField>
          <FormField label={t("appStore.resourceYaml")} htmlFor="resource-yaml">
            <Textarea id="resource-yaml" className="min-h-[360px] font-mono text-xs" value={manifest} onChange={(event) => setManifest(event.target.value)} />
          </FormField>
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button onClick={() => { onSave(manifest); onOpenChange(false) }}>{t("action.save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function ApplicationVersionPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "", versionUid = "" } = useParams()
  const [version, setVersion] = useState<ApplicationVersionItem | null>(null)
  const [resources, setResources] = useState<ResourceDraft[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [validationWarnings, setValidationWarnings] = useState<string[]>([])
  const [yamlOpen, setYamlOpen] = useState(false)
  const [yamlIndex, setYamlIndex] = useState(-1)
  const [yamlKind, setYamlKind] = useState("Deployment")
  const [yamlManifest, setYamlManifest] = useState("")
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deletePending, setDeletePending] = useState(false)

  useEffect(() => {
    let active = true
    void getApplicationVersion(uid, versionUid)
      .then((result) => {
        if (!active) return
        setVersion(result.version)
        setResources(result.version.resources.map((item) => ({ item: item.item, manifest: item.manifest, kind: item.kind, name: item.name, scope: item.scope })))
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

  if (!session.loading && !session.can("application:read")) {
    return <Navigate to="/forbidden" replace />
  }
  if (!version) {
    return <section className="flex min-h-0 flex-1 flex-col gap-6"><PageHeader title={t("appStore.version")} description={loading ? t("state.loading") : error} /></section>
  }
  const editable = version.status === "DRAFT" && session.can("application:write")

  function localizedWarning(warning: string): string {
    const external = warning.match(/^image (.+) is external and is not digest-pinned$/)
    if (external) return t("appStore.externalImageWarning", { image: external[1] })
    const unresolved = warning.match(/^image (.+) could not be resolved and is treated as external$/)
    if (unresolved) return t("appStore.externalImageUnresolved", { image: unresolved[1] })
    if (warning.includes("image registry list is unavailable")) return t("appStore.registryListUnavailable")
    return warning
  }

  async function saveDraft() {
    if (!version) return false
    setSaving(true)
    setError("")
    try {
      const result = await updateApplicationVersion(uid, version.uid, {
        applicationUid: uid,
        changelog: version?.changelog ?? "",
        resources: resources.map((item) => ({ item: item.item, manifest: item.manifest })),
      })
      setVersion(result.version)
      setResources(result.version.resources.map((item) => ({ item: item.item, manifest: item.manifest, kind: item.kind, name: item.name, scope: item.scope })))
      toast.success(t("state.saved"))
      return true
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
      return false
    } finally {
      setSaving(false)
    }
  }

  async function validate() {
    setValidationErrors([])
    setValidationWarnings([])
    if (editable) {
      const saved = await saveDraft()
      if (!saved) return
    }
    try {
      const result = await validateApplicationVersion(uid, versionUid)
      setValidationErrors(result.errors)
      setValidationWarnings(result.warnings.map(localizedWarning))
      if (result.version) setVersion(result.version)
      if (result.errors.length === 0) toast.success(t("appStore.validationPassed"))
    } catch (err) {
      setValidationErrors([formErrorMessage(err, t("state.failed"))])
    }
  }

  async function publish() {
    setSaving(true)
    setError("")
    try {
      if (editable) {
        await updateApplicationVersion(uid, versionUid, {
          applicationUid: uid,
          changelog: version?.changelog ?? "",
          resources: resources.map((item) => ({ item: item.item, manifest: item.manifest })),
        })
      }
      const result = await publishApplicationVersion(uid, versionUid)
      setVersion(result.version)
      toast.success(t("appStore.published"))
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setSaving(false)
    }
  }

  async function removeDraft() {
    setDeletePending(true)
    try {
      await deleteApplicationVersion(uid, versionUid)
      navigate(`/appstore/apps/${uid}`)
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

  function editResource(index: number) {
    setYamlIndex(index)
    setYamlKind(resources[index]?.kind || "Deployment")
    setYamlManifest(resources[index]?.manifest || "")
    setYamlOpen(true)
  }

  function addResource() {
    setYamlIndex(-1)
    setYamlKind("Deployment")
    setYamlManifest(TEMPLATES.Deployment)
    setYamlOpen(true)
  }

  function saveResource(manifest: string) {
    const inspected = inspectManifest(manifest)
    if (yamlIndex >= 0) {
      setResources((current) => current.map((item, index) => index === yamlIndex ? { ...item, manifest, ...inspected } : item))
    } else {
      setResources((current) => [...current, { item: `r-${Date.now()}`, manifest, ...inspected }])
    }
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={`${version.version}`}
        description={`${version.status} · ${t("appStore.resources")}: ${resources.length} · ${t("appStore.images")}: ${version.images.length}`}
      >
        <Button variant="outline" onClick={() => navigate(`/appstore/apps/${uid}`)}>{t("action.goBack")}</Button>
        {editable ? <Button variant="outline" onClick={() => void validate()}>{t("appStore.validate")}</Button> : null}
        {editable ? <Button variant="outline" disabled={saving} onClick={() => void saveDraft()}>{t("appStore.saveDraft")}</Button> : null}
        {editable ? <Button disabled={saving} onClick={() => void publish()}>{t("appStore.publish")}</Button> : null}
        {editable ? <Button variant="destructive" onClick={() => setDeleteOpen(true)}>{t("action.delete")}</Button> : null}
      </PageHeader>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      {validationErrors.length > 0 ? <div className="rounded-lg border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">{validationErrors.join("; ")}</div> : null}
      {validationWarnings.length > 0 ? <div className="rounded-lg border border-warning-border bg-warning-soft p-3 text-sm text-warning">{validationWarnings.join("; ")}</div> : null}
      <div className="rounded-xl border border-border/60 bg-card">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <h2 className="font-heading text-base font-semibold">{t("appStore.resourceList")}</h2>
          {editable ? <Button variant="outline" size="sm" onClick={addResource}>{t("appStore.addResource")}</Button> : null}
        </div>
        <div className="overflow-auto p-4">
          <Table className={TABLE_MIN_WIDTH.wide}>
            <TableHeader><TableRow><TableHead>{t("appStore.resourceKind")}</TableHead><TableHead>{t("appStore.resourceName")}</TableHead><TableHead>{t("appStore.resourceScope")}</TableHead><TableHead>{t("appStore.images")}</TableHead><TableHead>{t("appStore.secretKeys")}</TableHead><TableHead className="text-right">{t("action.actions")}</TableHead></TableRow></TableHeader>
            <TableBody>
              {resources.map((resource, index) => (
                <TableRow key={resource.item}>
                  <TableCell className="font-mono text-xs">{resource.kind || "—"}</TableCell>
                  <TableCell className="font-mono text-xs">{resource.name || "—"}</TableCell>
                  <TableCell><Badge variant={resource.scope === "CLUSTER" ? "destructive" : "outline"}>{resource.scope || "UNKNOWN"}</Badge></TableCell>
                  <TableCell className="font-mono text-xs">{version.resources[index]?.images.length ?? 0}</TableCell>
                  <TableCell className="font-mono text-xs">{version.resources[index]?.secretKeys.join(", ") || "—"}</TableCell>
                  <TableCell className={TABLE_NOWRAP}>{editable ? <Button variant="link" size="sm" onClick={() => editResource(index)}>{t("appStore.editYaml")}</Button> : "—"}</TableCell>
                </TableRow>
              ))}
              {resources.length === 0 ? <TableRow><TableCell colSpan={6} className="text-sm text-muted-foreground">{t("appStore.emptyResources")}</TableCell></TableRow> : null}
            </TableBody>
          </Table>
        </div>
      </div>
      <div className="rounded-xl border border-border/60 bg-card p-4">
        <h2 className="mb-2 font-heading text-base font-semibold">{t("appStore.images")}</h2>
        {version.images.length === 0 ? <p className="text-sm text-muted-foreground">—</p> : (
          <div className="grid gap-2">
            {version.images.map((image) => (
              <div key={`${image.resourceItem}-${image.image}`} className="rounded-lg border border-border/60 p-2 text-sm">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-mono text-xs break-all">{image.image}</span>
                  <Badge variant={image.external ? "destructive" : "secondary"}>{image.external ? "EXTERNAL" : image.registryName}</Badge>
                </div>
                <p className="mt-1 font-mono text-xs text-muted-foreground">{image.digest || "—"} · {image.platforms.join(" · ") || "—"}</p>
              </div>
            ))}
          </div>
        )}
      </div>
      <YamlResourceDialog
        key={`${yamlOpen}-${yamlIndex}-${yamlKind}`}
        open={yamlOpen}
        title={yamlIndex >= 0 ? t("appStore.editYaml") : t("appStore.addResource")}
        initialKind={yamlKind}
        initialManifest={yamlManifest}
        onOpenChange={setYamlOpen}
        onSave={saveResource}
      />
      <ConfirmAction
        open={deleteOpen}
        title={t("appStore.deleteVersionTitle")}
        description={t("appStore.deleteVersionDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={deletePending}
        onOpenChange={setDeleteOpen}
        onConfirm={() => void removeDraft()}
      />
    </section>
  )
}
