import { useState } from "react"
import { useTranslation } from "react-i18next"
import { parse as parseYaml } from "yaml"

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
import { checkField } from "@/lib/form"
import { formErrorMessage, readFormError } from "@/lib/form-error"

const FORM_KINDS = [
  "Deployment",
  "StatefulSet",
  "Service",
  "ConfigMap",
  "Secret",
  "Ingress",
  "CronJob",
] as const
type FormKind = (typeof FORM_KINDS)[number]

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

interface FormValues {
  kind: FormKind
  name: string
  replicas: string
  containerName: string
  image: string
  port: string
  serviceType: string
  servicePort: string
  targetPort: string
  entries: KeyValue[]
  ingressHost: string
  ingressPath: string
  ingressService: string
  ingressPort: string
  schedule: string
  command: string
}

// 新增时保持空值，只有结构化默认项（副本数、Service 类型、Ingress 路径）。
const EMPTY_VALUES: FormValues = {
  kind: "Deployment",
  name: "",
  replicas: "1",
  containerName: "",
  image: "",
  port: "",
  serviceType: "ClusterIP",
  servicePort: "",
  targetPort: "",
  entries: [],
  ingressHost: "",
  ingressPath: "/",
  ingressService: "",
  ingressPort: "",
  schedule: "",
  command: "",
}

function yamlValue(value: string): string {
  return JSON.stringify(value)
}

function buildManifest(input: FormValues): string {
  const { kind, name } = input
  const metadata = `metadata:\n  name: ${yamlValue(name)}`
  switch (kind) {
    case "Deployment":
    case "StatefulSet":
      return `apiVersion: apps/v1
kind: ${kind}
${metadata}
spec:
  replicas: ${input.replicas || "1"}
  selector:
    matchLabels:
      app: ${yamlValue(name)}
  template:
    metadata:
      labels:
        app: ${yamlValue(name)}
    spec:
      containers:
        - name: ${yamlValue(input.containerName || name)}
          image: ${yamlValue(input.image)}
          ports:
            - containerPort: ${input.port || "80"}
`
    case "Service":
      return `apiVersion: v1
kind: Service
${metadata}
spec:
  type: ${input.serviceType || "ClusterIP"}
  selector:
    app: ${yamlValue(name)}
  ports:
    - port: ${input.servicePort || "80"}
      targetPort: ${input.targetPort || input.servicePort || "80"}
`
    case "ConfigMap":
    case "Secret": {
      const body = input.entries
        .filter((entry) => entry.key.trim() !== "")
        .map((entry) => `  ${entry.key.trim()}: ${yamlValue(entry.value)}`)
        .join("\n")
      return `apiVersion: v1
kind: ${kind}
${metadata}
${kind === "Secret" ? "stringData" : "data"}:
${body || "  {}"}
`
    }
    case "Ingress":
      return `apiVersion: networking.k8s.io/v1
kind: Ingress
${metadata}
spec:
  ingressClassName: nginx
  rules:
    - host: ${yamlValue(input.ingressHost)}
      http:
        paths:
          - path: ${input.ingressPath || "/"}
            pathType: Prefix
            backend:
              service:
                name: ${yamlValue(input.ingressService || name)}
                port:
                  number: ${input.ingressPort || "80"}
`
    case "CronJob":
      return `apiVersion: batch/v1
kind: CronJob
${metadata}
spec:
  schedule: ${yamlValue(input.schedule)}
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: ${yamlValue(input.containerName || name)}
              image: ${yamlValue(input.image)}
              command: ["/bin/sh", "-c", ${yamlValue(input.command)}]
`
  }
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

function parseEntries(value: unknown): KeyValue[] {
  const record = asRecord(value)
  if (!record) return []
  return Object.entries(record).map(([key, item]) => ({ key, value: readText(item) }))
}

/**
 * 用 yaml 包解析清单并回填表单字段。返回 null 表示无法解析或缺少 kind，
 * 此时界面保持 YAML 编辑并提示。
 */
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
  const values: FormValues = {
    ...EMPTY_VALUES,
    kind: isFormKind(kind) ? kind : "Deployment",
    name: readText(asRecord(root.metadata)?.name),
  }
  const spec = asRecord(root.spec)

  if (kind === "Deployment" || kind === "StatefulSet") {
    values.replicas = readText(spec?.replicas) || "1"
    const podSpec = asRecord(asRecord(spec?.template)?.spec)
    const container = firstRecord(podSpec?.containers)
    values.containerName = readText(container?.name)
    values.image = readText(container?.image)
    values.port = readText(firstRecord(container?.ports)?.containerPort)
  } else if (kind === "Service") {
    values.serviceType = readText(spec?.type) || "ClusterIP"
    const port = firstRecord(spec?.ports)
    values.servicePort = readText(port?.port)
    values.targetPort = readText(port?.targetPort)
  } else if (kind === "ConfigMap" || kind === "Secret") {
    values.entries = parseEntries(root.stringData ?? root.data)
  } else if (kind === "Ingress") {
    const rule = firstRecord(spec?.rules)
    values.ingressHost = readText(rule?.host)
    const path = firstRecord(asRecord(rule?.http)?.paths)
    values.ingressPath = readText(path?.path) || "/"
    const backend = asRecord(asRecord(path?.backend)?.service)
    values.ingressService = readText(backend?.name)
    const backendPort = backend?.port
    values.ingressPort =
      readText(asRecord(backendPort)?.number) || readText(backendPort)
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
  const [kind, setKind] = useState<FormKind>(initialValues.kind)
  const [name, setName] = useState(initialValues.name)
  const [replicas, setReplicas] = useState(initialValues.replicas)
  const [containerName, setContainerName] = useState(initialValues.containerName)
  const [image, setImage] = useState(initialValues.image)
  const [port, setPort] = useState(initialValues.port)
  const [serviceType, setServiceType] = useState(initialValues.serviceType)
  const [servicePort, setServicePort] = useState(initialValues.servicePort)
  const [targetPort, setTargetPort] = useState(initialValues.targetPort)
  const [entries, setEntries] = useState<KeyValue[]>(initialValues.entries)
  const [ingressHost, setIngressHost] = useState(initialValues.ingressHost)
  const [ingressPath, setIngressPath] = useState(initialValues.ingressPath)
  const [ingressService, setIngressService] = useState(initialValues.ingressService)
  const [ingressPort, setIngressPort] = useState(initialValues.ingressPort)
  const [schedule, setSchedule] = useState(initialValues.schedule)
  const [command, setCommand] = useState(initialValues.command)
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

  const values: FormValues = {
    kind, name, replicas, containerName, image, port, serviceType, servicePort,
    targetPort, entries, ingressHost, ingressPath, ingressService, ingressPort,
    schedule, command,
  }
  const generated = buildManifest(values)
  const manifest = yamlOwned && !formTouched ? yaml : generated
  const needsImage = kind === "Deployment" || kind === "StatefulSet" || kind === "CronJob"

  function touch<T>(setter: (value: T) => void): (value: T) => void {
    return (value: T) => {
      setFormTouched(true)
      setter(value)
    }
  }

  function applyValues(next: FormValues) {
    setKind(next.kind)
    setName(next.name)
    setReplicas(next.replicas)
    setContainerName(next.containerName)
    setImage(next.image)
    setPort(next.port)
    setServiceType(next.serviceType)
    setServicePort(next.servicePort)
    setTargetPort(next.targetPort)
    setEntries(next.entries)
    setIngressHost(next.ingressHost)
    setIngressPath(next.ingressPath)
    setIngressService(next.ingressService)
    setIngressPort(next.ingressPort)
    setSchedule(next.schedule)
    setCommand(next.command)
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
    setKind(next)
    setFormTouched(true)
    setErrors({})
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
        value: name,
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
        value: image,
        required: needsImage,
      })
      if (imageError) {
        next.image = imageError
      }
      const portError = (label: string, value: string) =>
        checkField({ label, value, required: true, integerRange: [1, 65535] })
      if (kind === "Deployment" || kind === "StatefulSet") {
        const error = portError(t("appStore.resourceForm.containerPort"), port)
        if (error) next.port = error
      }
      if (kind === "Service") {
        const servicePortError = portError(t("appStore.resourceForm.servicePort"), servicePort)
        if (servicePortError) next.servicePort = servicePortError
        const targetPortError = portError(t("appStore.resourceForm.targetPort"), targetPort)
        if (targetPortError) next.targetPort = targetPortError
      }
      if (kind === "Ingress") {
        const hostError = checkField({
          label: t("appStore.resourceForm.host"),
          value: ingressHost,
          required: true,
        })
        if (hostError) next.ingressHost = hostError
        const ingressPortError = portError(t("appStore.resourceForm.backendPort"), ingressPort)
        if (ingressPortError) next.ingressPort = ingressPortError
      }
      if (kind === "CronJob") {
        const scheduleError = checkField({
          label: t("appStore.resourceForm.schedule"),
          value: schedule,
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
              item: name.trim() || componentName.trim(),
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
                  <Select value={kind} onValueChange={(value) => changeKind(value as FormKind)}>
                    <SelectTrigger id="resource-form-kind"><SelectValue /></SelectTrigger>
                    <SelectContent>{FORM_KINDS.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}</SelectContent>
                  </Select>
                </FormField>
                <FormField label={t("appStore.resourceForm.name")} htmlFor="resource-form-name" required error={errors.name} hint={t("appStore.resourceForm.nameHelp")}>
                  <Input id="resource-form-name" className="font-mono" value={name} placeholder="app" onChange={(event) => touch(setName)(event.target.value)} />
                </FormField>

                {kind === "Deployment" || kind === "StatefulSet" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.replicas")} htmlFor="resource-form-replicas" hint={t("appStore.resourceForm.replicasHelp")}>
                      <Input id="resource-form-replicas" type="number" min={0} max={100} value={replicas} onChange={(event) => touch(setReplicas)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.containerName")} htmlFor="resource-form-container">
                      <Input id="resource-form-container" className="font-mono" value={containerName} placeholder="app" onChange={(event) => touch(setContainerName)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.image")} htmlFor="resource-form-image" required error={errors.image} hint={t("appStore.resourceForm.imageHelp")} className="sm:col-span-full">
                      <ImageReferencePicker
                        id="resource-form-image"
                        value={image}
                        onChange={touch(setImage)}
                        invalid={Boolean(errors.image)}
                      />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.containerPort")} htmlFor="resource-form-port" hint={t("appStore.resourceForm.portHelp")} error={errors.port}>
                      <Input id="resource-form-port" type="number" min={1} max={65535} value={port} onChange={(event) => touch(setPort)(event.target.value)} />
                    </FormField>
                  </>
                ) : null}

                {kind === "Service" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.serviceType")} htmlFor="resource-form-service-type">
                      <Select value={serviceType} onValueChange={touch(setServiceType)}>
                        <SelectTrigger id="resource-form-service-type"><SelectValue /></SelectTrigger>
                        <SelectContent>
                          {["ClusterIP", "NodePort", "LoadBalancer"].map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
                        </SelectContent>
                      </Select>
                    </FormField>
                    <FormField label={t("appStore.resourceForm.servicePort")} htmlFor="resource-form-service-port" error={errors.servicePort}>
                      <Input id="resource-form-service-port" type="number" min={1} max={65535} value={servicePort} onChange={(event) => touch(setServicePort)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.targetPort")} htmlFor="resource-form-target-port" error={errors.targetPort}>
                      <Input id="resource-form-target-port" type="number" min={1} max={65535} value={targetPort} onChange={(event) => touch(setTargetPort)(event.target.value)} />
                    </FormField>
                  </>
                ) : null}

                {kind === "ConfigMap" || kind === "Secret" ? (
                  <FormField
                    label={kind === "Secret" ? t("appStore.resourceForm.secretData") : t("appStore.resourceForm.configData")}
                    hint={kind === "Secret" ? t("appStore.resourceForm.secretHelp") : t("appStore.resourceForm.configHelp")}
                    className="sm:col-span-full"
                    actions={
                      <Button type="button" variant="outline" size="sm" onClick={() => touch(setEntries)([...entries, { key: "", value: "" }])}>{t("appStore.resourceForm.addKey")}</Button>
                    }
                  >
                    <KeyValueRowLabels
                      keyLabel={t("appStore.resourceForm.key")}
                      valueLabel={t("appStore.resourceForm.value")}
                      className="sm:grid-cols-[1fr_1.4fr_auto]"
                      trailing={<span aria-hidden="true" className="hidden w-[4.5rem] sm:block" />}
                    />
                    {entries.map((entry, index) => (
                      <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                        <Input aria-label={t("appStore.resourceForm.key")} className="font-mono" value={entry.key} placeholder="KEY" onChange={(event) => touch(setEntries)(entries.map((item, position) => position === index ? { ...item, key: event.target.value } : item))} />
                        <Input aria-label={t("appStore.resourceForm.value")} className="font-mono" type={kind === "Secret" ? "password" : "text"} value={entry.value} placeholder={kind === "Secret" ? t("appStore.resourceForm.secretValuePlaceholder") : "value"} onChange={(event) => touch(setEntries)(entries.map((item, position) => position === index ? { ...item, value: event.target.value } : item))} />
                        <Button type="button" variant="ghost" size="sm" className="w-[4.5rem]" onClick={() => touch(setEntries)(entries.filter((_, position) => position !== index))}>{t("action.delete")}</Button>
                      </div>
                    ))}
                  </FormField>
                ) : null}

                {kind === "Ingress" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.host")} htmlFor="resource-form-host" required error={errors.ingressHost}>
                      <Input id="resource-form-host" value={ingressHost} placeholder="app.example.com" onChange={(event) => touch(setIngressHost)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.path")} htmlFor="resource-form-path">
                      <Input id="resource-form-path" value={ingressPath} placeholder="/" onChange={(event) => touch(setIngressPath)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.backendService")} htmlFor="resource-form-backend">
                      <Input id="resource-form-backend" className="font-mono" value={ingressService} placeholder="app" onChange={(event) => touch(setIngressService)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.backendPort")} htmlFor="resource-form-backend-port" error={errors.ingressPort}>
                      <Input id="resource-form-backend-port" type="number" min={1} max={65535} value={ingressPort} onChange={(event) => touch(setIngressPort)(event.target.value)} />
                    </FormField>
                  </>
                ) : null}

                {kind === "CronJob" ? (
                  <>
                    <FormField label={t("appStore.resourceForm.schedule")} htmlFor="resource-form-schedule" required hint={t("appStore.resourceForm.scheduleHelp")} error={errors.schedule}>
                      <Input id="resource-form-schedule" className="font-mono" value={schedule} placeholder="0 * * * *" onChange={(event) => touch(setSchedule)(event.target.value)} />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.image")} htmlFor="resource-form-cron-image" required error={errors.image}>
                      <ImageReferencePicker
                        id="resource-form-cron-image"
                        value={image}
                        onChange={touch(setImage)}
                        invalid={Boolean(errors.image)}
                      />
                    </FormField>
                    <FormField label={t("appStore.resourceForm.command")} htmlFor="resource-form-command" className="sm:col-span-full">
                      <Input id="resource-form-command" className="font-mono" value={command} placeholder="/bin/sh -c echo hello" onChange={(event) => touch(setCommand)(event.target.value)} />
                    </FormField>
                  </>
                ) : null}
              </div>
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
