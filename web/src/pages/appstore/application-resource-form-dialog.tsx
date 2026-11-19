import { ChevronDown } from "lucide-react"
import { useState, type ReactNode } from "react"
import { useTranslation } from "react-i18next"
import { parse as parseYaml, stringify as stringifyYaml } from "yaml"

import {
  createApplicationVersionResource,
  saveApplicationVersionResource,
  validateApplicationVersionResource,
} from "@/api/application"
import type { ApplicationResource, ApplicationVersionItem } from "@/api/application/types"
import { SearchableSelect } from "@/components/searchable-select"
import { ImageReferencePicker } from "@/components/image-reference-picker"
import { FormField, KeyValueRowLabels } from "@/components/form-field"
import { YamlEditor } from "@/components/yaml-editor"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible"
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { cn } from "@/lib/class-names"
import { checkField } from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"

const FORM_KINDS = [
  "Deployment",
  "StatefulSet",
  "DaemonSet",
  "Service",
  "ConfigMap",
  "Secret",
  "Ingress",
  "CronJob",
] as const
type FormKind = (typeof FORM_KINDS)[number]

const WORKLOAD_KINDS: readonly FormKind[] = ["Deployment", "StatefulSet", "DaemonSet"]
const IMAGE_KINDS: readonly FormKind[] = ["Deployment", "StatefulSet", "DaemonSet", "CronJob"]
const VOLUME_TYPES = ["emptyDir", "configMap", "secret", "persistentVolumeClaim", "hostPath"] as const
type VolumeType = (typeof VOLUME_TYPES)[number]
const PROBE_TYPES = ["httpGet", "tcpSocket", "exec"] as const
type ProbeType = (typeof PROBE_TYPES)[number]
const PATH_TYPES = ["Prefix", "Exact", "ImplementationSpecific"] as const
type PathType = (typeof PATH_TYPES)[number]
const IMAGE_PULL_POLICIES = ["IfNotPresent", "Always", "Never"] as const
const INGRESS_CLASS_PRESETS = ["nginx", "traefik", "alb"] as const

interface ApplicationResourceFormDialogProps {
  open: boolean
  mode: "create" | "edit"
  applicationUid: string
  versionUid: string
  /** 版本当前 revision，作为乐观锁提交。 */
  revision: number
  /** 新增/编辑时归属的服务（组件）名。 */
  initialComponent?: string
  /** 版本内已有的服务名，用于组件选择。 */
  components?: string[]
  /** 编辑时的资源详情。 */
  resource?: ApplicationResource | null
  onOpenChange: (open: boolean) => void
  onSaved: (version: ApplicationVersionItem) => void
}

interface KeyValue {
  key: string
  value: string
}

interface EnvRow {
  name: string
  value: string
}

interface VolumeRow {
  name: string
  type: VolumeType
  source: string
  mountPath: string
}

interface ProbeRow {
  enabled: boolean
  type: ProbeType
  path: string
  port: string
  command: string
  initialDelaySeconds: string
  periodSeconds: string
}

interface ResourceLimits {
  requestsCpu: string
  requestsMemory: string
  limitsCpu: string
  limitsMemory: string
}

interface IngressPathRow {
  path: string
  pathType: PathType
  service: string
  port: string
}

interface IngressRuleRow {
  host: string
  paths: IngressPathRow[]
}

interface FormValues {
  kind: FormKind
  name: string
  replicas: string
  containerName: string
  image: string
  imagePullPolicy: string
  port: string
  serviceType: string
  servicePort: string
  targetPort: string
  entries: KeyValue[]
  schedule: string
  command: string
  containerCommand: string
  containerArgs: string
  labels: KeyValue[]
  annotations: KeyValue[]
  env: EnvRow[]
  volumes: VolumeRow[]
  probes: {
    liveness: ProbeRow
    readiness: ProbeRow
    startup: ProbeRow
  }
  resources: ResourceLimits
  ingressClassName: string
  ingressAnnotations: KeyValue[]
  tlsEnabled: boolean
  tlsSecretName: string
  rules: IngressRuleRow[]
}

function emptyProbe(): ProbeRow {
  return {
    enabled: false,
    type: "httpGet",
    path: "/",
    port: "",
    command: "",
    initialDelaySeconds: "",
    periodSeconds: "",
  }
}

function emptyPath(): IngressPathRow {
  return { path: "/", pathType: "Prefix", service: "", port: "" }
}

// 新增时保持空值，只有结构化默认项：副本数、Service 类型、探针/规则的空壳。
const EMPTY_VALUES: FormValues = {
  kind: "Deployment",
  name: "",
  replicas: "1",
  containerName: "",
  image: "",
  imagePullPolicy: "IfNotPresent",
  port: "",
  serviceType: "ClusterIP",
  servicePort: "",
  targetPort: "",
  entries: [],
  schedule: "",
  command: "",
  containerCommand: "",
  containerArgs: "",
  labels: [],
  annotations: [],
  env: [],
  volumes: [],
  probes: { liveness: emptyProbe(), readiness: emptyProbe(), startup: emptyProbe() },
  resources: { requestsCpu: "", requestsMemory: "", limitsCpu: "", limitsMemory: "" },
  ingressClassName: "",
  ingressAnnotations: [],
  tlsEnabled: false,
  tlsSecretName: "",
  rules: [{ host: "", paths: [emptyPath()] }],
}

function kvToObject(rows: KeyValue[]): Record<string, string> {
  const out: Record<string, string> = {}
  for (const row of rows) {
    const key = row.key.trim()
    if (key) {
      out[key] = row.value
    }
  }
  return out
}

function splitTokens(value: string): string[] {
  return value.trim() === "" ? [] : value.trim().split(/\s+/)
}

function numberOrString(value: string): number | string {
  const trimmed = value.trim()
  const parsed = Number(trimmed)
  return trimmed !== "" && Number.isInteger(parsed) ? parsed : trimmed
}

function buildProbe(row: ProbeRow): Record<string, unknown> | undefined {
  if (!row.enabled) {
    return undefined
  }
  const probe: Record<string, unknown> = {}
  if (row.type === "httpGet") {
    probe.httpGet = { path: row.path.trim() || "/", port: numberOrString(row.port) }
  } else if (row.type === "tcpSocket") {
    probe.tcpSocket = { port: numberOrString(row.port) }
  } else {
    probe.exec = { command: splitTokens(row.command) }
  }
  if (row.initialDelaySeconds.trim() !== "") {
    probe.initialDelaySeconds = Number(row.initialDelaySeconds)
  }
  if (row.periodSeconds.trim() !== "") {
    probe.periodSeconds = Number(row.periodSeconds)
  }
  return probe
}

function buildResources(limits: ResourceLimits): Record<string, unknown> | undefined {
  const requests: Record<string, string> = {}
  const limitsMap: Record<string, string> = {}
  if (limits.requestsCpu.trim()) requests.cpu = limits.requestsCpu.trim()
  if (limits.requestsMemory.trim()) requests.memory = limits.requestsMemory.trim()
  if (limits.limitsCpu.trim()) limitsMap.cpu = limits.limitsCpu.trim()
  if (limits.limitsMemory.trim()) limitsMap.memory = limits.limitsMemory.trim()
  if (Object.keys(requests).length === 0 && Object.keys(limitsMap).length === 0) {
    return undefined
  }
  const resources: Record<string, unknown> = {}
  if (Object.keys(requests).length > 0) resources.requests = requests
  if (Object.keys(limitsMap).length > 0) resources.limits = limitsMap
  return resources
}

function buildVolume(row: VolumeRow): Record<string, unknown> {
  const name = row.name.trim()
  const source = row.source.trim()
  switch (row.type) {
    case "emptyDir":
      return { name, emptyDir: {} }
    case "configMap":
      return { name, configMap: { name: source } }
    case "secret":
      return { name, secret: { secretName: source } }
    case "persistentVolumeClaim":
      return { name, persistentVolumeClaim: { claimName: source } }
    case "hostPath":
      return { name, hostPath: { path: source } }
  }
}

function buildWorkloadContainer(values: FormValues): Record<string, unknown> {
  const container: Record<string, unknown> = {
    name: values.containerName.trim() || values.name.trim(),
    image: values.image.trim(),
    imagePullPolicy: values.imagePullPolicy,
  }
  if (values.port.trim() !== "") {
    container.ports = [{ containerPort: numberOrString(values.port) }]
  }
  const env = values.env
    .filter((row) => row.name.trim() !== "")
    .map((row) => ({ name: row.name.trim(), value: row.value }))
  if (env.length > 0) {
    container.env = env
  }
  const volumeMounts = values.volumes
    .filter((row) => row.name.trim() !== "" && row.mountPath.trim() !== "")
    .map((row) => ({ name: row.name.trim(), mountPath: row.mountPath.trim() }))
  if (volumeMounts.length > 0) {
    container.volumeMounts = volumeMounts
  }
  const command = splitTokens(values.containerCommand)
  if (command.length > 0) {
    container.command = command
  }
  const args = splitTokens(values.containerArgs)
  if (args.length > 0) {
    container.args = args
  }
  const resources = buildResources(values.resources)
  if (resources) {
    container.resources = resources
  }
  const livenessProbe = buildProbe(values.probes.liveness)
  if (livenessProbe) container.livenessProbe = livenessProbe
  const readinessProbe = buildProbe(values.probes.readiness)
  if (readinessProbe) container.readinessProbe = readinessProbe
  const startupProbe = buildProbe(values.probes.startup)
  if (startupProbe) container.startupProbe = startupProbe
  return container
}

function buildManifest(values: FormValues): string {
  const kind = values.kind
  const name = values.name.trim()
  const labels = kvToObject(values.labels)
  const annotations = kvToObject(values.annotations)
  const metadata: Record<string, unknown> = { name }
  if (Object.keys(labels).length > 0) metadata.labels = labels
  if (Object.keys(annotations).length > 0) metadata.annotations = annotations

  if (WORKLOAD_KINDS.includes(kind)) {
    const selector = Object.keys(labels).length > 0 ? labels : { app: name }
    const spec: Record<string, unknown> = {
      selector: { matchLabels: selector },
      template: {
        metadata: { labels: selector },
        spec: {
          containers: [buildWorkloadContainer(values)],
        },
      },
    }
    if (kind !== "DaemonSet") {
      spec.replicas = Number(values.replicas || "1")
    }
    const volumes = values.volumes
      .filter((row) => row.name.trim() !== "")
      .map(buildVolume)
    if (volumes.length > 0) {
      ;(spec.template as { spec: Record<string, unknown> }).spec.volumes = volumes
    }
    return stringifyYaml({ apiVersion: "apps/v1", kind, metadata, spec })
  }

  if (kind === "Service") {
    return stringifyYaml({
      apiVersion: "v1",
      kind,
      metadata,
      spec: {
        type: values.serviceType || "ClusterIP",
        selector: Object.keys(labels).length > 0 ? labels : { app: name },
        ports: [
          {
            port: numberOrString(values.servicePort) || 80,
            targetPort: numberOrString(values.targetPort || values.servicePort) || 80,
          },
        ],
      },
    })
  }

  if (kind === "ConfigMap" || kind === "Secret") {
    const data = kvToObject(values.entries)
    return stringifyYaml({
      apiVersion: "v1",
      kind,
      metadata,
      ...(kind === "Secret" ? { stringData: data } : { data }),
    })
  }

  if (kind === "Ingress") {
    const ingressAnnotations = kvToObject(values.ingressAnnotations)
    if (Object.keys(ingressAnnotations).length > 0) {
      metadata.annotations = ingressAnnotations
    }
    const spec: Record<string, unknown> = {}
    if (values.ingressClassName.trim() !== "") {
      spec.ingressClassName = values.ingressClassName.trim()
    }
    if (values.tlsEnabled) {
      const hosts = values.rules.map((rule) => rule.host.trim()).filter((host) => host !== "")
      const tls: Record<string, unknown> = {}
      if (hosts.length > 0) tls.hosts = hosts
      if (values.tlsSecretName.trim() !== "") tls.secretName = values.tlsSecretName.trim()
      spec.tls = [tls]
    }
    spec.rules = values.rules.map((rule) => ({
      host: rule.host.trim(),
      http: {
        paths: rule.paths.map((path) => ({
          path: path.path.trim() || "/",
          pathType: path.pathType,
          backend: {
            service: {
              name: path.service.trim(),
              port: { number: numberOrString(path.port) || 80 },
            },
          },
        })),
      },
    }))
    return stringifyYaml({
      apiVersion: "networking.k8s.io/v1",
      kind,
      metadata,
      spec,
    })
  }

  // CronJob
  return stringifyYaml({
    apiVersion: "batch/v1",
    kind,
    metadata,
    spec: {
      schedule: values.schedule,
      jobTemplate: {
        spec: {
          template: {
            spec: {
              restartPolicy: "OnFailure",
              containers: [
                {
                  name: values.containerName.trim() || name,
                  image: values.image.trim(),
                  command: ["/bin/sh", "-c", values.command],
                },
              ],
            },
          },
        },
      },
    },
  })
}

function isFormKind(kind: string): kind is FormKind {
  return (FORM_KINDS as readonly string[]).includes(kind)
}

function asRecord(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null
}

function firstRecord(value: unknown): Record<string, unknown> | null {
  return Array.isArray(value) ? asRecord(value[0]) : null
}

function readText(value: unknown): string {
  if (typeof value === "string") return value
  if (typeof value === "number" || typeof value === "boolean") return String(value)
  return ""
}

function readCommand(value: unknown): string {
  if (!Array.isArray(value)) return ""
  const parts = value.map((item) => readText(item)).filter((item) => item !== "")
  if (parts.length === 3 && /(?:^|\/)sh$/.test(parts[0]) && parts[1] === "-c") {
    return parts[2]
  }
  return parts.join(" ")
}

function parseRows(value: unknown): KeyValue[] {
  const record = asRecord(value)
  if (!record) return []
  return Object.entries(record).map(([key, item]) => ({ key, value: readText(item) }))
}

function parseProbe(value: unknown): ProbeRow {
  const record = asRecord(value)
  if (!record) {
    return emptyProbe()
  }
  const common = {
    enabled: true,
    initialDelaySeconds: readText(record.initialDelaySeconds),
    periodSeconds: readText(record.periodSeconds),
  }
  const httpGet = asRecord(record.httpGet)
  if (httpGet) {
    return { ...common, type: "httpGet", path: readText(httpGet.path) || "/", port: readText(httpGet.port), command: "" }
  }
  const tcpSocket = asRecord(record.tcpSocket)
  if (tcpSocket) {
    return { ...common, type: "tcpSocket", path: "", port: readText(tcpSocket.port), command: "" }
  }
  const exec = asRecord(record.exec)
  return {
    ...common,
    type: "exec",
    path: "",
    port: "",
    command: Array.isArray(exec?.command) ? exec.command.map((item) => readText(item)).join(" ") : "",
  }
}

function parseVolumes(podSpec: Record<string, unknown> | null, container: Record<string, unknown> | null): VolumeRow[] {
  if (!Array.isArray(podSpec?.volumes)) {
    return []
  }
  const mounts = new Map<string, string>()
  if (Array.isArray(container?.volumeMounts)) {
    for (const item of container.volumeMounts) {
      const mount = asRecord(item)
      if (mount) mounts.set(readText(mount.name), readText(mount.mountPath))
    }
  }
  const rows: VolumeRow[] = []
  for (const item of podSpec.volumes) {
    const volume = asRecord(item)
    if (!volume) continue
    const name = readText(volume.name)
    const mountPath = mounts.get(name) ?? ""
    const emptyDir = asRecord(volume.emptyDir)
    if (emptyDir) {
      rows.push({ name, type: "emptyDir", source: "", mountPath })
      continue
    }
    const configMap = asRecord(volume.configMap)
    if (configMap) {
      rows.push({ name, type: "configMap", source: readText(configMap.name), mountPath })
      continue
    }
    const secret = asRecord(volume.secret)
    if (secret) {
      rows.push({ name, type: "secret", source: readText(secret.secretName), mountPath })
      continue
    }
    const pvc = asRecord(volume.persistentVolumeClaim)
    if (pvc) {
      rows.push({ name, type: "persistentVolumeClaim", source: readText(pvc.claimName), mountPath })
      continue
    }
    const hostPath = asRecord(volume.hostPath)
    if (hostPath) {
      rows.push({ name, type: "hostPath", source: readText(hostPath.path), mountPath })
    }
  }
  return rows
}

function parseRules(spec: Record<string, unknown> | null): IngressRuleRow[] {
  if (!Array.isArray(spec?.rules)) {
    return []
  }
  return spec.rules.map((item) => {
    const rule = asRecord(item)
    const paths = Array.isArray(asRecord(rule?.http)?.paths) ? (asRecord(rule?.http)?.paths as unknown[]) : []
    return {
      host: readText(rule?.host),
      paths: paths.map((pathItem) => {
        const path = asRecord(pathItem)
        const backend = asRecord(asRecord(path?.backend)?.service)
        const port = backend?.port
        return {
          path: readText(path?.path) || "/",
          pathType: (PATH_TYPES as readonly string[]).includes(readText(path?.pathType))
            ? (readText(path?.pathType) as PathType)
            : "Prefix",
          service: readText(backend?.name),
          port: readText(asRecord(port)?.number) || readText(port),
        }
      }),
    }
  })
}

/** 用 yaml 包解析清单并回填表单字段；无法解析或缺少 kind 时返回 null。 */
function parseManifest(manifest: string): { kind: string; values: FormValues } | null {
  let document: unknown
  try {
    document = parseYaml(manifest)
  } catch {
    return null
  }
  const root = asRecord(document)
  if (!root) {
    return null
  }
  const kind = readText(root.kind)
  if (!kind) {
    return null
  }
  const metadata = asRecord(root.metadata)
  const values: FormValues = {
    ...EMPTY_VALUES,
    kind: isFormKind(kind) ? kind : "Deployment",
    name: readText(metadata?.name),
    labels: parseRows(metadata?.labels),
    annotations: parseRows(metadata?.annotations),
    ingressAnnotations: parseRows(metadata?.annotations),
  }
  const spec = asRecord(root.spec)

  if (WORKLOAD_KINDS.includes(kind as FormKind)) {
    values.replicas = readText(spec?.replicas) || "1"
    const podSpec = asRecord(asRecord(spec?.template)?.spec)
    const container = firstRecord(podSpec?.containers)
    values.containerName = readText(container?.name)
    values.image = readText(container?.image)
    values.imagePullPolicy = readText(container?.imagePullPolicy) || "IfNotPresent"
    values.port = readText(firstRecord(container?.ports)?.containerPort)
    values.containerCommand = Array.isArray(container?.command)
      ? container.command.map((item) => readText(item)).join(" ")
      : ""
    values.containerArgs = Array.isArray(container?.args)
      ? container.args.map((item) => readText(item)).join(" ")
      : ""
    values.env = Array.isArray(container?.env)
      ? container.env.flatMap((item) => {
          const env = asRecord(item)
          return env ? [{ name: readText(env.name), value: readText(env.value) }] : []
        })
      : []
    values.volumes = parseVolumes(podSpec, container)
    values.probes = {
      liveness: parseProbe(container?.livenessProbe),
      readiness: parseProbe(container?.readinessProbe),
      startup: parseProbe(container?.startupProbe),
    }
    const resources = asRecord(container?.resources)
    const requests = asRecord(resources?.requests)
    const limits = asRecord(resources?.limits)
    values.resources = {
      requestsCpu: readText(requests?.cpu),
      requestsMemory: readText(requests?.memory),
      limitsCpu: readText(limits?.cpu),
      limitsMemory: readText(limits?.memory),
    }
  } else if (kind === "Service") {
    values.serviceType = readText(spec?.type) || "ClusterIP"
    const port = firstRecord(spec?.ports)
    values.servicePort = readText(port?.port)
    values.targetPort = readText(port?.targetPort)
  } else if (kind === "ConfigMap" || kind === "Secret") {
    values.entries = parseRows(root.stringData ?? root.data)
  } else if (kind === "Ingress") {
    values.ingressClassName = readText(spec?.ingressClassName)
    const tls = firstRecord(spec?.tls)
    values.tlsEnabled = Boolean(tls)
    values.tlsSecretName = readText(tls?.secretName)
    const rules = parseRules(spec)
    if (rules.length > 0) {
      values.rules = rules
    }
  } else if (kind === "CronJob") {
    values.schedule = readText(spec?.schedule)
    const jobSpec = asRecord(asRecord(spec?.jobTemplate)?.spec)
    const podSpec = asRecord(asRecord(jobSpec?.template)?.spec)
    const container = firstRecord(podSpec?.containers)
    values.containerName = readText(container?.name)
    values.image = readText(container?.image)
    values.command = readCommand(container?.command)
  }

  return { kind, values }
}

/** 折叠区：长表单按领域分组，默认折叠，编辑时按已有数据展开。 */
function FormSection({
  title,
  hint,
  defaultOpen = false,
  children,
}: {
  title: string
  hint?: string
  defaultOpen?: boolean
  children: ReactNode
}) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <Collapsible open={open} onOpenChange={setOpen} className="min-w-0 rounded-lg border border-border/60">
      <CollapsibleTrigger asChild>
        <button
          type="button"
          className="flex w-full items-center justify-between gap-2 rounded-lg px-3 py-2 text-left text-sm font-medium hover:bg-muted/40"
        >
          <span className="min-w-0 truncate">{title}</span>
          <ChevronDown
            aria-hidden
            className={cn("size-4 shrink-0 text-muted-foreground transition-transform", open && "rotate-180")}
          />
        </button>
      </CollapsibleTrigger>
      <CollapsibleContent className="grid min-w-0 gap-3 border-t border-border/60 p-3">
        {hint ? <p className="text-xs text-muted-foreground">{hint}</p> : null}
        {children}
      </CollapsibleContent>
    </Collapsible>
  )
}

function RowHeader({ labels }: { labels: string[] }) {
  return (
    <div
      aria-hidden
      className="hidden gap-2 px-0.5 text-xs leading-5 text-muted-foreground sm:grid"
      style={{ gridTemplateColumns: `repeat(${labels.length}, minmax(0, 1fr)) auto` }}
    >
      {labels.map((label) => (
        <span key={label} className="min-w-0 truncate">{label}</span>
      ))}
      <span className="w-[4.5rem]" />
    </div>
  )
}

interface ProbeEditorProps {
  label: string
  value: ProbeRow
  onChange: (value: ProbeRow) => void
}

function ProbeEditor({ label, value, onChange }: ProbeEditorProps) {
  const { t } = useTranslation()
  return (
    <div className="grid min-w-0 gap-2 rounded-lg border border-border/60 p-3">
      <label className="flex items-center gap-2 text-sm font-medium">
        <Checkbox
          checked={value.enabled}
          onCheckedChange={(checked) => onChange({ ...value, enabled: checked === true })}
        />
        {label}
      </label>
      {value.enabled ? (
        <div className="grid gap-2 sm:grid-cols-2">
          <FormField label={t("appStore.resourceForm.probeType")} hint={t("appStore.resourceForm.probeHint")}>
            <Select value={value.type} onValueChange={(next) => onChange({ ...value, type: next as ProbeType })}>
              <SelectTrigger aria-label={t("appStore.resourceForm.probeType")}><SelectValue /></SelectTrigger>
              <SelectContent>
                {PROBE_TYPES.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
              </SelectContent>
            </Select>
          </FormField>
          {value.type === "httpGet" ? (
            <FormField label={t("appStore.resourceForm.probePath")} hint={t("appStore.resourceForm.probePathHint")}>
              <Input aria-label={t("appStore.resourceForm.probePath")} className="font-mono" value={value.path} placeholder="/healthz" onChange={(event) => onChange({ ...value, path: event.target.value })} />
            </FormField>
          ) : null}
          {value.type !== "exec" ? (
            <FormField label={t("appStore.resourceForm.probePort")} hint={t("appStore.resourceForm.probePortHint")}>
              <Input aria-label={t("appStore.resourceForm.probePort")} className="font-mono" type="number" min={1} max={65535} value={value.port} placeholder="8080" onChange={(event) => onChange({ ...value, port: event.target.value })} />
            </FormField>
          ) : (
            <FormField label={t("appStore.resourceForm.probeCommand")} hint={t("appStore.resourceForm.probeCommandHint")} className="sm:col-span-full">
              <Input aria-label={t("appStore.resourceForm.probeCommand")} className="font-mono" value={value.command} placeholder="/bin/sh -c true" onChange={(event) => onChange({ ...value, command: event.target.value })} />
            </FormField>
          )}
          <FormField label={t("appStore.resourceForm.initialDelaySeconds")} hint={t("appStore.resourceForm.probeSecondsHint")}>
            <Input aria-label={t("appStore.resourceForm.initialDelaySeconds")} type="number" min={0} value={value.initialDelaySeconds} placeholder="10" onChange={(event) => onChange({ ...value, initialDelaySeconds: event.target.value })} />
          </FormField>
          <FormField label={t("appStore.resourceForm.periodSeconds")} hint={t("appStore.resourceForm.probeSecondsHint")}>
            <Input aria-label={t("appStore.resourceForm.periodSeconds")} type="number" min={0} value={value.periodSeconds} placeholder="10" onChange={(event) => onChange({ ...value, periodSeconds: event.target.value })} />
          </FormField>
        </div>
      ) : null}
    </div>
  )
}

/**
 * 资源新增/编辑对话框：表单（常用 kind）与 YAML 双 tab，YAML 使用 monaco-yaml
 * 的 schema 补全与校验，并可用 ValidateApplicationVersionResource 做服务端校验。
 * 编辑时默认按 manifest 回填表单；未知 kind 或解析失败则停留在 YAML 并提示。
 */
export function ApplicationResourceFormDialog({
  open,
  mode,
  applicationUid,
  versionUid,
  revision,
  initialComponent,
  components = [],
  resource,
  onOpenChange,
  onSaved,
}: ApplicationResourceFormDialogProps) {
  const { t } = useTranslation()
  const [initial] = useState(() => {
    const parsed = mode === "edit" && resource ? parseManifest(resource.manifest) : null
    const supported = mode === "edit" ? Boolean(parsed && isFormKind(parsed.kind)) : true
    return { parsed, supported }
  })
  const initialValues = initial.parsed?.values ?? EMPTY_VALUES
  const initialNotice =
    mode !== "edit" || !resource
      ? ""
      : initial.parsed
        ? initial.supported
          ? ""
          : t("appStore.resourceForm.unsupportedKindHint", { kind: initial.parsed.kind })
        : t("appStore.resourceForm.parseFailedHint")

  const [tab, setTab] = useState(mode === "edit" && !initial.supported ? "yaml" : "form")
  const [componentName, setComponentName] = useState(initialComponent || resource?.component || "")
  const [values, setValues] = useState<FormValues>(initialValues)
  const [yaml, setYaml] = useState(resource?.manifest ?? "")
  const [yamlOwned, setYamlOwned] = useState(mode === "edit")
  const [formTouched, setFormTouched] = useState(false)
  const [formNotice, setFormNotice] = useState(initialNotice)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState("")
  const [saving, setSaving] = useState(false)
  const [validating, setValidating] = useState(false)
  const [validationErrors, setValidationErrors] = useState<string[]>([])
  const [validationWarnings, setValidationWarnings] = useState<string[]>([])

  const generated = buildManifest(values)
  const manifest = yamlOwned && !formTouched ? yaml : generated
  const isWorkload = WORKLOAD_KINDS.includes(values.kind)
  const needsImage = IMAGE_KINDS.includes(values.kind)

  function update(patch: Partial<FormValues>) {
    setFormTouched(true)
    setValues((current) => ({ ...current, ...patch }))
  }

  function applyValues(next: FormValues) {
    setValues(next)
  }

  function changeTab(next: string) {
    if (next === "yaml") {
      setYaml(manifest)
      setYamlOwned(true)
      setFormTouched(false)
      setFormNotice("")
      setTab("yaml")
      return
    }
    // 从 YAML 切回表单：以当前 YAML 为准重新解析并回填。
    const parsed = parseManifest(manifest)
    if (!parsed) {
      setFormNotice(t("appStore.resourceForm.parseFailedHint"))
      return
    }
    if (!isFormKind(parsed.kind)) {
      setFormNotice(t("appStore.resourceForm.unsupportedKindHint", { kind: parsed.kind }))
      return
    }
    applyValues(parsed.values)
    setFormTouched(false)
    setFormNotice("")
    setTab("form")
  }

  function changeKind(next: FormKind) {
    update({ kind: next })
    setErrors({})
  }

  function changeClassName(next: string) {
    const preset = (INGRESS_CLASS_PRESETS as readonly string[]).includes(next)
    const presetAnnotations: Record<string, string> = {
      nginx: { "kubernetes.io/ingress.class": "nginx" },
      traefik: { "kubernetes.io/ingress.class": "traefik" },
      alb: { "alb.ingress.kubernetes.io/scheme": "internet-facing" },
    }[next] ?? {}
    const emptyAnnotations = values.ingressAnnotations.every(
      (row) => row.key.trim() === "" && row.value.trim() === "",
    )
    update({
      ingressClassName: next,
      ingressAnnotations:
        preset && emptyAnnotations
          ? Object.entries(presetAnnotations).map(([key, value]) => ({ key, value }))
          : values.ingressAnnotations,
    })
  }

  const componentOptions = [...new Set([componentName, ...components].filter(Boolean))].map((item) => ({
    value: item,
    label: item,
  }))

  function validateLocally(): boolean {
    const next: Record<string, string> = {}
    const componentError = checkField({
      label: t("appStore.resourceForm.component"),
      value: componentName,
      required: true,
      max: 63,
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
      hint: t("appStore.resourceForm.componentHelp"),
    })
    if (componentError) {
      next.component = componentError
    }
    if (tab === "form") {
      const nameError = checkField({
        label: t("appStore.resourceForm.name"),
        value: values.name,
        required: true,
        max: 253,
        pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
        hint: t("appStore.resourceForm.nameInvalid"),
      })
      if (nameError) {
        next.name = nameError
      }
      const imageError = checkField({
        label: t("appStore.resourceForm.image"),
        value: values.image,
        required: needsImage,
      })
      if (imageError) {
        next.image = imageError
      }
      const portError = (label: string, value: string, required = true) =>
        checkField({ label, value, required, integerRange: [1, 65535] })
      if (isWorkload) {
        const error = portError(t("appStore.resourceForm.containerPort"), values.port, false)
        if (error) next.port = error
        for (const [index, row] of values.volumes.entries()) {
          const hasAny = row.name.trim() || row.source.trim() || row.mountPath.trim()
          if (!hasAny) continue
          if (!row.name.trim()) {
            next.volumes = t("appStore.resourceForm.volumeNameRequired", { index: index + 1 })
            break
          }
          if (row.type !== "emptyDir" && !row.source.trim()) {
            next.volumes = t("appStore.resourceForm.volumeSourceRequired", { index: index + 1 })
            break
          }
        }
        for (const key of ["liveness", "readiness", "startup"] as const) {
          const probe = values.probes[key]
          if (!probe.enabled) continue
          const probeLabel = t(`appStore.resourceForm.probe.${key}`)
          if (probe.type === "exec") {
            if (!probe.command.trim()) {
              next[`probe-${key}`] = t("appStore.resourceForm.probeCommandRequired", { name: probeLabel })
              break
            }
          } else {
            const probePortError = checkField({
              label: t("appStore.resourceForm.probePort"),
              value: probe.port,
              required: true,
              integerRange: [1, 65535],
            })
            if (probePortError) {
              next[`probe-${key}`] = t("appStore.resourceForm.probePortRequired", { name: probeLabel })
              break
            }
            if (probe.type === "httpGet" && !probe.path.trim().startsWith("/")) {
              next[`probe-${key}`] = t("appStore.resourceForm.probePathInvalid", { name: probeLabel })
              break
            }
          }
        }
      }
      if (values.kind === "Service") {
        const servicePortError = portError(t("appStore.resourceForm.servicePort"), values.servicePort)
        if (servicePortError) next.servicePort = servicePortError
        const targetPortError = portError(t("appStore.resourceForm.targetPort"), values.targetPort)
        if (targetPortError) next.targetPort = targetPortError
      }
      if (values.kind === "Ingress") {
        if (values.tlsEnabled && !values.tlsSecretName.trim()) {
          next.tlsSecretName = t("appStore.resourceForm.tlsSecretRequired")
        }
        if (values.rules.length === 0) {
          next.ingressRules = t("appStore.resourceForm.ingressNoRules")
        } else {
          ruleLoop: for (const [ruleIndex, rule] of values.rules.entries()) {
            if (!rule.host.trim()) {
              next.ingressRules = t("appStore.resourceForm.ingressHostRequired", { index: ruleIndex + 1 })
              break
            }
            for (const path of rule.paths) {
              if (!path.path.trim().startsWith("/")) {
                next.ingressRules = t("appStore.resourceForm.ingressPathInvalid", { index: ruleIndex + 1 })
                break ruleLoop
              }
              if (!path.service.trim()) {
                next.ingressRules = t("appStore.resourceForm.ingressServiceRequired", { index: ruleIndex + 1 })
                break ruleLoop
              }
              const pathPortError = portError(t("appStore.resourceForm.rulePort"), path.port)
              if (pathPortError) {
                next.ingressRules = t("appStore.resourceForm.ingressPortInvalid", { index: ruleIndex + 1 })
                break ruleLoop
              }
            }
          }
        }
      }
      if (values.kind === "CronJob") {
        const scheduleError = checkField({
          label: t("appStore.resourceForm.schedule"),
          value: values.schedule,
          required: true,
          pattern: /^[0-9A-Za-z*/,-]+(?:\s+[0-9A-Za-z*/,-]+){4}$/,
          hint: t("field.hint.cron"),
        })
        if (scheduleError) next.schedule = scheduleError
      }
    } else {
      const text = yaml.trim()
      if (!/apiVersion:\s*\S+/.test(text) || !/kind:\s*\S+/.test(text) || !/name:\s*\S+/.test(text)) {
        next.yaml = t("appStore.resourceForm.yamlInvalid")
      }
    }
    setErrors(next)
    return Object.keys(next).length === 0
  }

  async function validateOnServer() {
    if (!validateLocally()) {
      return
    }
    setValidating(true)
    setValidationErrors([])
    setValidationWarnings([])
    try {
      const result = await validateApplicationVersionResource({
        applicationUid,
        uid: versionUid,
        component: componentName.trim(),
        manifest,
      })
      setValidationErrors(result.errors)
      setValidationWarnings(result.warnings)
      if (result.errors.length === 0) {
        setFormError("")
      }
    } catch (err) {
      setValidationErrors([formErrorMessage(err, t("state.failed"))])
    } finally {
      setValidating(false)
    }
  }

  async function save() {
    if (!validateLocally()) {
      return
    }
    setSaving(true)
    setFormError("")
    try {
      const result = mode === "edit" && resource
        ? await saveApplicationVersionResource({
            applicationUid,
            uid: versionUid,
            resourceUid: resource.uid,
            component: componentName.trim(),
            manifest,
            expectedRevision: revision,
          })
        : await createApplicationVersionResource({
            applicationUid,
            uid: versionUid,
            resource: {
              // item 是版本内的稳定标识；带上 kind 前缀，避免同名 Deployment 与
              // Service 共用 metadata.name 时被后端判成重复资源。
              item: values.name.trim() ? `${values.kind.toLowerCase()}-${values.name.trim()}` : componentName.trim(),
              manifest,
              component: componentName.trim(),
            },
            expectedRevision: revision,
          })
      onSaved(result.version)
      onOpenChange(false)
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setErrors((current) => ({ ...current, ...failure.fields }))
      setFormError(failure.form)
    } finally {
      setSaving(false)
    }
  }

  const hasInitialMetadata = initialValues.labels.length > 0 || initialValues.annotations.length > 0
  const hasInitialEnv = initialValues.env.length > 0
  const hasInitialVolumes = initialValues.volumes.length > 0
  const hasInitialProbes = Object.values(initialValues.probes).some((probe) => probe.enabled)
  const hasInitialResources = Object.values(initialValues.resources).some((value) => value !== "")
  const hasInitialIngressAnnotations = initialValues.ingressAnnotations.length > 0

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{mode === "edit" ? t("appStore.resourceForm.editTitle") : t("appStore.addResource")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4">
          <FormField
            label={t("appStore.resourceForm.component")}
            htmlFor="resource-component"
            required
            error={errors.component}
            hint={t("appStore.resourceForm.componentHelp")}
          >
            <SearchableSelect
              id="resource-component"
              value={componentName}
              options={componentOptions}
              onChange={(next) => setComponentName(next)}
              placeholder={t("appStore.resourceForm.componentPlaceholder")}
              searchPlaceholder={t("appStore.resourceForm.componentSearch")}
              emptyText={t("appStore.resourceForm.componentEmpty")}
              invalid={Boolean(errors.component)}
            />
          </FormField>

          {formNotice ? (
            <p className="rounded-md border border-warning-border bg-warning-soft px-2 py-1.5 text-xs text-warning">
              {formNotice}
            </p>
          ) : null}

          <Tabs value={tab} onValueChange={changeTab}>
            <TabsList>
              <TabsTrigger value="form">{t("appStore.resourceForm.formTab")}</TabsTrigger>
              <TabsTrigger value="yaml">{t("appStore.resourceForm.yamlTab")}</TabsTrigger>
            </TabsList>
            <TabsContent value="form" className="grid gap-4 pt-2">
              {yamlOwned ? (
                <p className="rounded-md border border-warning-border bg-warning-soft px-2 py-1.5 text-xs text-warning">
                  {t("appStore.resourceForm.formRebuildHint")}
                </p>
              ) : null}
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField label={t("appStore.resourceKind")} htmlFor="resource-form-kind" hint={t("appStore.resourceForm.kindHelp")}>
                  <Select value={values.kind} onValueChange={(value) => changeKind(value as FormKind)}>
                    <SelectTrigger id="resource-form-kind"><SelectValue /></SelectTrigger>
                    <SelectContent>{FORM_KINDS.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}</SelectContent>
                  </Select>
                </FormField>
                <FormField label={t("appStore.resourceForm.name")} htmlFor="resource-form-name" required error={errors.name} hint={t("appStore.resourceForm.nameHelp")}>
                  <Input id="resource-form-name" className="font-mono" value={values.name} placeholder="app" onChange={(event) => update({ name: event.target.value })} />
                </FormField>

                {isWorkload ? (
                  <>
                    {values.kind !== "DaemonSet" ? (
                      <FormField label={t("appStore.resourceForm.replicas")} htmlFor="resource-form-replicas" hint={t("appStore.resourceForm.replicasHelp")}>
                        <Input id="resource-form-replicas" type="number" min={0} max={200} value={values.replicas} placeholder="2" onChange={(event) => update({ replicas: event.target.value })} />
                      </FormField>
                    ) : null}
                    <FormField label={t("appStore.resourceForm.containerName")} htmlFor="resource-form-container" hint={t("appStore.resourceForm.containerNameHint")}>
                      <Input id="resource-form-container" className="font-mono" value={values.containerName} placeholder="app" onChange={(event) => update({ containerName: event.target.value })} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.image")} htmlFor="resource-form-image" required error={errors.image} hint={t("appStore.resourceForm.imageHelp")} className="sm:col-span-full">
                      <ImageReferencePicker
                        id="resource-form-image"
                        value={values.image}
                        onChange={(next) => update({ image: next })}
                        invalid={Boolean(errors.image)}
                      />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.imagePullPolicy")} hint={t("appStore.resourceForm.pullPolicyHint")}>
                      <Select value={values.imagePullPolicy} onValueChange={(next) => update({ imagePullPolicy: next })}>
                        <SelectTrigger><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {IMAGE_PULL_POLICIES.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
                        </SelectContent>
                      </Select>
                    </FormField>
                    <FormField label={t("appStore.resourceForm.containerPort")} htmlFor="resource-form-port" hint={t("appStore.resourceForm.portHelp")} error={errors.port}>
                      <Input id="resource-form-port" type="number" min={1} max={65535} value={values.port} placeholder="8080" onChange={(event) => update({ port: event.target.value })} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.containerCommand")} htmlFor="resource-form-command" hint={t("appStore.resourceForm.tokensHint")}>
                      <Input id="resource-form-command" className="font-mono" value={values.containerCommand} placeholder="/app/server --config /etc/app" onChange={(event) => update({ containerCommand: event.target.value })} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.containerArgs")} htmlFor="resource-form-args" hint={t("appStore.resourceForm.tokensHint")}>
                      <Input id="resource-form-args" className="font-mono" value={values.containerArgs} placeholder="--verbose" onChange={(event) => update({ containerArgs: event.target.value })} />
                    </FormField>
                  </>
                ) : null}

                {values.kind === "Service" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.serviceType")} htmlFor="resource-form-service-type">
                      <Select value={values.serviceType} onValueChange={(next) => update({ serviceType: next })}>
                        <SelectTrigger id="resource-form-service-type"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {["ClusterIP", "NodePort", "LoadBalancer"].map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
                        </SelectContent>
                      </Select>
                    </FormField>
                    <FormField label={t("appStore.resourceForm.servicePort")} htmlFor="resource-form-service-port" error={errors.servicePort}>
                      <Input id="resource-form-service-port" type="number" min={1} max={65535} value={values.servicePort} placeholder="80" onChange={(event) => update({ servicePort: event.target.value })} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.targetPort")} htmlFor="resource-form-target-port" error={errors.targetPort}>
                      <Input id="resource-form-target-port" type="number" min={1} max={65535} value={values.targetPort} placeholder="80" onChange={(event) => update({ targetPort: event.target.value })} />
                    </FormField>
                  </>
                ) : null}

                {values.kind === "ConfigMap" || values.kind === "Secret" ? (
                  <FormField
                    label={values.kind === "Secret" ? t("appStore.resourceForm.secretData") : t("appStore.resourceForm.configData")}
                    hint={values.kind === "Secret" ? t("appStore.resourceForm.secretHelp") : t("appStore.resourceForm.configHelp")}
                    className="sm:col-span-full"
                    actions={
                      <Button type="button" variant="outline" size="sm" onClick={() => update({ entries: [...values.entries, { key: "", value: "" }] })}>{t("appStore.resourceForm.addKey")}</Button>
                    }
                  >
                    <KeyValueRowLabels
                      keyLabel={t("appStore.resourceForm.key")}
                      valueLabel={t("appStore.resourceForm.value")}
                      className="sm:grid-cols-[1fr_1.4fr_auto]"
                      trailing={<span aria-hidden="true" className="hidden w-[4.5rem] sm:block" />}
                    />
                    {values.entries.map((entry, index) => (
                      <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                        <Input aria-label={t("appStore.resourceForm.key")} className="font-mono" value={entry.key} placeholder="KEY" onChange={(event) => update({ entries: values.entries.map((item, position) => position === index ? { ...item, key: event.target.value } : item) })} />
                        <Input aria-label={t("appStore.resourceForm.value")} className="font-mono" type={values.kind === "Secret" ? "password" : "text"} value={entry.value} placeholder={values.kind === "Secret" ? t("appStore.resourceForm.secretValuePlaceholder") : "value"} onChange={(event) => update({ entries: values.entries.map((item, position) => position === index ? { ...item, value: event.target.value } : item) })} />
                        <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => update({ entries: values.entries.filter((_, position) => position !== index) })}>{t("action.delete")}</Button>
                      </div>
                    ))}
                  </FormField>
                ) : null}

                {values.kind === "Ingress" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.ingressClassName")} htmlFor="resource-form-ingress-class" hint={t("appStore.resourceForm.ingressClassNameHint")}>
                      <Input
                        id="resource-form-ingress-class"
                        className="font-mono"
                        value={values.ingressClassName}
                        list="ingress-class-presets"
                        placeholder="nginx"
                        onChange={(event) => changeClassName(event.target.value)}
                      />
                      <datalist id="ingress-class-presets">
                        {INGRESS_CLASS_PRESETS.map((item) => <option key={item} value={item} />)}
                      </datalist>
                    </FormField>
                    <div className="grid gap-2">
                      <label className="flex items-center gap-2 self-end pb-1 text-sm">
                        <Checkbox
                          checked={values.tlsEnabled}
                          onCheckedChange={(checked) => update({ tlsEnabled: checked === true })}
                        />
                        {t("appStore.resourceForm.tlsEnabled")}
                      </label>
                    </div>
                    {values.tlsEnabled ? (
                      <FormField label={t("appStore.resourceForm.tlsSecretName")} htmlFor="resource-form-tls-secret" error={errors.tlsSecretName} hint={t("appStore.resourceForm.tlsSecretHint")} className="sm:col-span-full">
                        <Input id="resource-form-tls-secret" className="font-mono" value={values.tlsSecretName} placeholder="app-tls" onChange={(event) => update({ tlsSecretName: event.target.value })} />
                      </FormField>
                    ) : null}
                  </>
                ) : null}

                {values.kind === "CronJob" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.schedule")} htmlFor="resource-form-schedule" required hint={t("appStore.resourceForm.scheduleHelp")} error={errors.schedule}>
                      <Input id="resource-form-schedule" className="font-mono" value={values.schedule} placeholder="0 * * * *" onChange={(event) => update({ schedule: event.target.value })} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.image")} htmlFor="resource-form-cron-image" required error={errors.image}>
                      <ImageReferencePicker
                        id="resource-form-cron-image"
                        value={values.image}
                        onChange={(next) => update({ image: next })}
                        invalid={Boolean(errors.image)}
                      />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.command")} htmlFor="resource-form-command" className="sm:col-span-full">
                      <Input id="resource-form-command" className="font-mono" value={values.command} placeholder="/bin/sh -c echo hello" onChange={(event) => update({ command: event.target.value })} />
                    </FormField>
                  </>
                ) : null}
              </div>

              {isWorkload ? (
                <>
                  <FormSection
                    title={t("appStore.resourceForm.metadataSection")}
                    hint={t("appStore.resourceForm.metadataHint")}
                    defaultOpen={hasInitialMetadata}
                  >
                    <FormField
                      label={t("appStore.resourceForm.labels")}
                      hint={t("appStore.resourceForm.labelsHint")}
                      actions={
                        <Button type="button" variant="outline" size="sm" onClick={() => update({ labels: [...values.labels, { key: "", value: "" }] })}>
                          {t("appStore.resourceForm.addLabel")}
                        </Button>
                      }
                    >
                      <KeyValueRowLabels
                        keyLabel={t("appStore.resourceForm.key")}
                        valueLabel={t("appStore.resourceForm.value")}
                        className="sm:grid-cols-[1fr_1.4fr_auto]"
                        trailing={<span aria-hidden="true" className="hidden w-[4.5rem] sm:block" />}
                      />
                      {values.labels.map((row, index) => (
                        <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                          <Input aria-label={t("appStore.resourceForm.labelKey")} className="font-mono" value={row.key} placeholder="app.kubernetes.io/name" onChange={(event) => update({ labels: values.labels.map((item, position) => position === index ? { ...item, key: event.target.value } : item) })} />
                          <Input aria-label={t("appStore.resourceForm.labelValue")} className="font-mono" value={row.value} placeholder="app" onChange={(event) => update({ labels: values.labels.map((item, position) => position === index ? { ...item, value: event.target.value } : item) })} />
                          <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => update({ labels: values.labels.filter((_, position) => position !== index) })}>{t("action.delete")}</Button>
                        </div>
                      ))}
                    </FormField>
                    <FormField
                      label={t("appStore.resourceForm.annotations")}
                      hint={t("appStore.resourceForm.annotationsHint")}
                      actions={
                        <Button type="button" variant="outline" size="sm" onClick={() => update({ annotations: [...values.annotations, { key: "", value: "" }] })}>
                          {t("appStore.resourceForm.addAnnotation")}
                        </Button>
                      }
                    >
                      <KeyValueRowLabels
                        keyLabel={t("appStore.resourceForm.key")}
                        valueLabel={t("appStore.resourceForm.value")}
                        className="sm:grid-cols-[1fr_1.4fr_auto]"
                        trailing={<span aria-hidden="true" className="hidden w-[4.5rem] sm:block" />}
                      />
                      {values.annotations.map((row, index) => (
                        <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                          <Input aria-label={t("appStore.resourceForm.annotationKey")} className="font-mono" value={row.key} placeholder="prometheus.io/scrape" onChange={(event) => update({ annotations: values.annotations.map((item, position) => position === index ? { ...item, key: event.target.value } : item) })} />
                          <Input aria-label={t("appStore.resourceForm.annotationValue")} className="font-mono" value={row.value} placeholder="true" onChange={(event) => update({ annotations: values.annotations.map((item, position) => position === index ? { ...item, value: event.target.value } : item) })} />
                          <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => update({ annotations: values.annotations.filter((_, position) => position !== index) })}>{t("action.delete")}</Button>
                        </div>
                      ))}
                    </FormField>
                  </FormSection>

                  <FormSection
                    title={t("appStore.resourceForm.env")}
                    hint={t("appStore.resourceForm.envHint")}
                    defaultOpen={hasInitialEnv}
                  >
                    <FormField
                      label={t("appStore.resourceForm.env")}
                      hint={t("appStore.resourceForm.envHint")}
                      actions={
                        <Button type="button" variant="outline" size="sm" onClick={() => update({ env: [...values.env, { name: "", value: "" }] })}>
                          {t("appStore.resourceForm.addEnv")}
                        </Button>
                      }
                    >
                      <KeyValueRowLabels
                        keyLabel={t("appStore.resourceForm.envName")}
                        valueLabel={t("appStore.resourceForm.envValue")}
                        className="sm:grid-cols-[1fr_1.4fr_auto]"
                        trailing={<span aria-hidden="true" className="hidden w-[4.5rem] sm:block" />}
                      />
                      {values.env.map((row, index) => (
                        <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                          <Input aria-label={t("appStore.resourceForm.envName")} className="font-mono" value={row.name} placeholder="LOG_LEVEL" onChange={(event) => update({ env: values.env.map((item, position) => position === index ? { ...item, name: event.target.value } : item) })} />
                          <Input aria-label={t("appStore.resourceForm.envValue")} className="font-mono" value={row.value} placeholder="info" onChange={(event) => update({ env: values.env.map((item, position) => position === index ? { ...item, value: event.target.value } : item) })} />
                          <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => update({ env: values.env.filter((_, position) => position !== index) })}>{t("action.delete")}</Button>
                        </div>
                      ))}
                    </FormField>
                  </FormSection>

                  <FormSection
                    title={t("appStore.resourceForm.volumes")}
                    hint={t("appStore.resourceForm.volumesHint")}
                    defaultOpen={hasInitialVolumes}
                  >
                    <FormField
                      label={t("appStore.resourceForm.volumes")}
                      hint={t("appStore.resourceForm.volumesHint")}
                      error={errors.volumes}
                      actions={
                        <Button type="button" variant="outline" size="sm" onClick={() => update({ volumes: [...values.volumes, { name: "", type: "emptyDir", source: "", mountPath: "" }] })}>
                          {t("appStore.resourceForm.addVolume")}
                        </Button>
                      }
                    >
                      <RowHeader labels={[
                        t("appStore.resourceForm.volumeName"),
                        t("appStore.resourceForm.volumeType"),
                        t("appStore.resourceForm.volumeSource"),
                        t("appStore.resourceForm.mountPath"),
                      ]} />
                      {values.volumes.map((row, index) => (
                        <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1fr_1.4fr_1.4fr_auto]">
                          <Input aria-label={t("appStore.resourceForm.volumeName")} className="font-mono" value={row.name} placeholder="data" onChange={(event) => update({ volumes: values.volumes.map((item, position) => position === index ? { ...item, name: event.target.value } : item) })} />
                          <Select value={row.type} onValueChange={(next) => update({ volumes: values.volumes.map((item, position) => position === index ? { ...item, type: next as VolumeType } : item) })}>
                            <SelectTrigger aria-label={t("appStore.resourceForm.volumeType")}><SelectValue /></SelectTrigger>
                            <SelectContent>
                              {VOLUME_TYPES.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
                            </SelectContent>
                          </Select>
                          <Input aria-label={t("appStore.resourceForm.volumeSource")} className="font-mono" value={row.source} placeholder={row.type === "hostPath" ? "/data" : "app-config"} disabled={row.type === "emptyDir"} onChange={(event) => update({ volumes: values.volumes.map((item, position) => position === index ? { ...item, source: event.target.value } : item) })} />
                          <Input aria-label={t("appStore.resourceForm.mountPath")} className="font-mono" value={row.mountPath} placeholder="/data" onChange={(event) => update({ volumes: values.volumes.map((item, position) => position === index ? { ...item, mountPath: event.target.value } : item) })} />
                          <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => update({ volumes: values.volumes.filter((_, position) => position !== index) })}>{t("action.delete")}</Button>
                        </div>
                      ))}
                    </FormField>
                  </FormSection>

                  <FormSection
                    title={t("appStore.resourceForm.probes")}
                    hint={t("appStore.resourceForm.probesHint")}
                    defaultOpen={hasInitialProbes}
                  >
                    {(["liveness", "readiness", "startup"] as const).map((key) => (
                      <div key={key} className="grid gap-1">
                        <ProbeEditor
                          label={t(`appStore.resourceForm.probe.${key}`)}
                          value={values.probes[key]}
                          onChange={(next) => update({ probes: { ...values.probes, [key]: next } })}
                        />
                        {errors[`probe-${key}`] ? (
                          <p role="alert" className="text-xs text-destructive">{errors[`probe-${key}`]}</p>
                        ) : null}
                      </div>
                    ))}
                  </FormSection>

                  <FormSection
                    title={t("appStore.resourceForm.resources")}
                    hint={t("appStore.resourceForm.resourcesHint")}
                    defaultOpen={hasInitialResources}
                  >
                    <FormField label={t("appStore.resourceForm.requests")} hint={t("appStore.resourceForm.resourcesHint")}>
                      <div className="grid gap-2 sm:grid-cols-2">
                        <Input aria-label={`${t("appStore.resourceForm.requests")} ${t("appStore.resourceForm.cpu")}`} className="font-mono" value={values.resources.requestsCpu} placeholder="100m" onChange={(event) => update({ resources: { ...values.resources, requestsCpu: event.target.value } })} />
                        <Input aria-label={`${t("appStore.resourceForm.requests")} ${t("appStore.resourceForm.memory")}`} className="font-mono" value={values.resources.requestsMemory} placeholder="128Mi" onChange={(event) => update({ resources: { ...values.resources, requestsMemory: event.target.value } })} />
                      </div>
                    </FormField>
                    <FormField label={t("appStore.resourceForm.limits")} hint={t("appStore.resourceForm.resourcesHint")}>
                      <div className="grid gap-2 sm:grid-cols-2">
                        <Input aria-label={`${t("appStore.resourceForm.limits")} ${t("appStore.resourceForm.cpu")}`} className="font-mono" value={values.resources.limitsCpu} placeholder="500m" onChange={(event) => update({ resources: { ...values.resources, limitsCpu: event.target.value } })} />
                        <Input aria-label={`${t("appStore.resourceForm.limits")} ${t("appStore.resourceForm.memory")}`} className="font-mono" value={values.resources.limitsMemory} placeholder="512Mi" onChange={(event) => update({ resources: { ...values.resources, limitsMemory: event.target.value } })} />
                      </div>
                    </FormField>
                  </FormSection>
                </>
              ) : null}

              {values.kind === "Ingress" ? (
                <>
                  <FormSection
                    title={t("appStore.resourceForm.ingressAnnotations")}
                    hint={t("appStore.resourceForm.ingressAnnotationsHint")}
                    defaultOpen={hasInitialIngressAnnotations}
                  >
                    <FormField
                      label={t("appStore.resourceForm.ingressAnnotations")}
                      hint={t("appStore.resourceForm.ingressAnnotationsHint")}
                      actions={
                        <Button type="button" variant="outline" size="sm" onClick={() => update({ ingressAnnotations: [...values.ingressAnnotations, { key: "", value: "" }] })}>
                          {t("appStore.resourceForm.addAnnotation")}
                        </Button>
                      }
                    >
                      <KeyValueRowLabels
                        keyLabel={t("appStore.resourceForm.key")}
                        valueLabel={t("appStore.resourceForm.value")}
                        className="sm:grid-cols-[1fr_1.4fr_auto]"
                        trailing={<span aria-hidden="true" className="hidden w-[4.5rem] sm:block" />}
                      />
                      {values.ingressAnnotations.map((row, index) => (
                        <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                          <Input aria-label={t("appStore.resourceForm.annotationKey")} className="font-mono" value={row.key} placeholder="nginx.ingress.kubernetes.io/proxy-body-size" onChange={(event) => update({ ingressAnnotations: values.ingressAnnotations.map((item, position) => position === index ? { ...item, key: event.target.value } : item) })} />
                          <Input aria-label={t("appStore.resourceForm.annotationValue")} className="font-mono" value={row.value} placeholder="10m" onChange={(event) => update({ ingressAnnotations: values.ingressAnnotations.map((item, position) => position === index ? { ...item, value: event.target.value } : item) })} />
                          <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => update({ ingressAnnotations: values.ingressAnnotations.filter((_, position) => position !== index) })}>{t("action.delete")}</Button>
                        </div>
                      ))}
                    </FormField>
                  </FormSection>

                  <FormField
                    label={t("appStore.resourceForm.ingressRules")}
                    hint={t("appStore.resourceForm.ingressRulesHint")}
                    error={errors.ingressRules}
                    actions={
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => update({ rules: [...values.rules, { host: "", paths: [emptyPath()] }] })}
                      >
                        {t("appStore.resourceForm.addRule")}
                      </Button>
                    }
                  >
                    <div className="grid gap-3">
                      {values.rules.map((rule, ruleIndex) => (
                        <div key={ruleIndex} className="grid gap-3 rounded-lg border border-border/60 p-3">
                          <div className="flex items-center justify-between gap-2">
                            <span className="text-sm font-medium">{t("appStore.resourceForm.rule", { index: ruleIndex + 1 })}</span>
                            {values.rules.length > 1 ? (
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                className="text-destructive"
                                onClick={() => update({ rules: values.rules.filter((_, position) => position !== ruleIndex) })}
                              >
                                {t("appStore.resourceForm.removeRule")}
                              </Button>
                            ) : null}
                          </div>
                          <Input
                            aria-label={t("appStore.resourceForm.ruleHost")}
                            className="font-mono"
                            value={rule.host}
                            placeholder="app.example.com"
                            onChange={(event) => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, host: event.target.value } : item) })}
                          />
                          <RowHeader labels={[
                            t("appStore.resourceForm.rulePath"),
                            t("appStore.resourceForm.pathType"),
                            t("appStore.resourceForm.ruleService"),
                            t("appStore.resourceForm.rulePort"),
                          ]} />
                          {rule.paths.map((path, pathIndex) => (
                            <div key={pathIndex} className="grid gap-2 sm:grid-cols-[1.2fr_0.9fr_1.2fr_0.7fr_auto]">
                              <Input
                                aria-label={t("appStore.resourceForm.rulePath")}
                                className="font-mono"
                                value={path.path}
                                placeholder="/api"
                                onChange={(event) => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, paths: item.paths.map((entry, pathPosition) => pathPosition === pathIndex ? { ...entry, path: event.target.value } : entry) } : item) })}
                              />
                              <Select
                                value={path.pathType}
                                onValueChange={(next) => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, paths: item.paths.map((entry, pathPosition) => pathPosition === pathIndex ? { ...entry, pathType: next as PathType } : entry) } : item) })}
                              >
                                <SelectTrigger aria-label={t("appStore.resourceForm.pathType")}><SelectValue /></SelectTrigger>
                                <SelectContent>
                                  {PATH_TYPES.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
                                </SelectContent>
                              </Select>
                              <Input
                                aria-label={t("appStore.resourceForm.ruleService")}
                                className="font-mono"
                                value={path.service}
                                placeholder="app"
                                onChange={(event) => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, paths: item.paths.map((entry, pathPosition) => pathPosition === pathIndex ? { ...entry, service: event.target.value } : entry) } : item) })}
                              />
                              <Input
                                aria-label={t("appStore.resourceForm.rulePort")}
                                className="font-mono"
                                type="number"
                                min={1}
                                max={65535}
                                value={path.port}
                                placeholder="80"
                                onChange={(event) => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, paths: item.paths.map((entry, pathPosition) => pathPosition === pathIndex ? { ...entry, port: event.target.value } : entry) } : item) })}
                              />
                              <Button
                                type="button"
                                variant="ghost"
                                size="sm"
                                className="w-[4.5rem]"
                                disabled={rule.paths.length <= 1}
                                onClick={() => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, paths: item.paths.filter((_, pathPosition) => pathPosition !== pathIndex) } : item) })}
                              >
                                {t("appStore.resourceForm.removePath")}
                              </Button>
                            </div>
                          ))}
                          <div>
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              onClick={() => update({ rules: values.rules.map((item, position) => position === ruleIndex ? { ...item, paths: [...item.paths, emptyPath()] } : item) })}
                            >
                              {t("appStore.resourceForm.addPath")}
                            </Button>
                          </div>
                        </div>
                      ))}
                    </div>
                  </FormField>
                </>
              ) : null}

              <p className="text-xs text-muted-foreground">{t("appStore.resourceForm.yamlFallbackHint")}</p>
            </TabsContent>
            <TabsContent value="yaml" className="grid gap-2 pt-2">
              <FormField label={t("appStore.resourceYaml")} hint={t("appStore.resourceForm.yamlEditorHelp")} error={errors.yaml}>
                <YamlEditor
                  id="resource-form-yaml"
                  value={manifest}
                  onChange={(next) => {
                    setYaml(next)
                    setYamlOwned(true)
                    setFormTouched(false)
                  }}
                  height={360}
                  ariaLabel={t("appStore.resourceYaml")}
                />
              </FormField>
            </TabsContent>
          </Tabs>

          {validationErrors.length > 0 ? (
            <div role="alert" className="rounded-md border border-destructive/40 bg-destructive/10 p-2 text-xs text-destructive">
              <p className="mb-1 font-medium">{t("appStore.validationErrors")}</p>
              <ul className="list-disc pl-5 break-words">{validationErrors.map((item) => <li key={item}>{item}</li>)}</ul>
            </div>
          ) : null}
          {validationWarnings.length > 0 ? (
            <div className="rounded-md border border-warning-border bg-warning-soft p-2 text-xs text-warning">
              <p className="mb-1 font-medium">{t("appStore.validationWarnings")}</p>
              <ul className="list-disc pl-5 break-words">{validationWarnings.map((item) => <li key={item}>{item}</li>)}</ul>
            </div>
          ) : null}
        </DialogBody>
        {formError ? <p className="text-sm text-destructive">{formError}</p> : null}
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button variant="outline" disabled={validating || saving} onClick={() => void validateOnServer()}>
            {validating ? t("appStore.validating") : t("appStore.validate")}
          </Button>
          <Button disabled={saving} onClick={() => void save()}>{saving ? t("action.saving") : t("action.save")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
