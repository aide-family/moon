import { useState } from "react"
import { useTranslation } from "react-i18next"

import { FormField } from "@/components/form-field"
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
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"

const KINDS = ["Deployment", "StatefulSet", "Service", "ConfigMap", "Secret", "Ingress", "CronJob"] as const
type Kind = (typeof KINDS)[number]

interface ResourceFormDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSave: (manifest: string) => void
}

interface KeyValue {
  key: string
  value: string
}

function yamlValue(value: string): string {
  return JSON.stringify(value)
}

function buildManifest(input: {
  kind: Kind
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
}): string {
  const { kind, name } = input
  const metadata = `metadata:\n  name: ${name}`
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
      app: ${name}
  template:
    metadata:
      labels:
        app: ${name}
    spec:
      containers:
        - name: ${input.containerName || name}
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
    app: ${name}
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
    - host: ${input.ingressHost || "app.example.com"}
      http:
        paths:
          - path: ${input.ingressPath || "/"}
            pathType: Prefix
            backend:
              service:
                name: ${input.ingressService || name}
                port:
                  number: ${input.ingressPort || "80"}
`
    case "CronJob":
      return `apiVersion: batch/v1
kind: CronJob
${metadata}
spec:
  schedule: ${yamlValue(input.schedule || "0 * * * *")}
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: ${input.containerName || name}
              image: ${yamlValue(input.image)}
              command: ["/bin/sh", "-c", ${yamlValue(input.command || "echo hello")}]
`
  }
}

function validateName(name: string): boolean {
  return /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/.test(name)
}

export function ResourceFormDialog({ open, onOpenChange, onSave }: ResourceFormDialogProps) {
  const { t } = useTranslation()
  const [tab, setTab] = useState("form")
  const [kind, setKind] = useState<Kind>("Deployment")
  const [name, setName] = useState("app")
  const [replicas, setReplicas] = useState("1")
  const [containerName, setContainerName] = useState("app")
  const [image, setImage] = useState("nginx:1.27.2")
  const [port, setPort] = useState("80")
  const [serviceType, setServiceType] = useState("ClusterIP")
  const [servicePort, setServicePort] = useState("80")
  const [targetPort, setTargetPort] = useState("80")
  const [entries, setEntries] = useState<KeyValue[]>([{ key: "KEY", value: "value" }])
  const [ingressHost, setIngressHost] = useState("app.example.com")
  const [ingressPath, setIngressPath] = useState("/")
  const [ingressService, setIngressService] = useState("app")
  const [ingressPort, setIngressPort] = useState("80")
  const [schedule, setSchedule] = useState("0 * * * *")
  const [command, setCommand] = useState("echo hello")
  const [yaml, setYaml] = useState("")
  const [yamlOwned, setYamlOwned] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const generated = buildManifest({
    kind, name, replicas, containerName, image, port, serviceType, servicePort,
    targetPort, entries, ingressHost, ingressPath, ingressService, ingressPort,
    schedule, command,
  })
  const manifest = yamlOwned ? yaml : generated
  const needsImage = kind === "Deployment" || kind === "StatefulSet" || kind === "CronJob"

  function changeKind(next: Kind) {
    setKind(next)
    setYamlOwned(false)
    setErrors({})
  }

  function save() {
    const next: Record<string, string> = {}
    if (!validateName(name.trim())) {
      next.name = t("appStore.resourceForm.nameInvalid")
    }
    if (needsImage && !image.trim()) {
      next.image = t("form.required")
    }
    if (yamlOwned) {
      const text = yaml.trim()
      if (!/apiVersion:\s*\S+/.test(text) || !/kind:\s*\S+/.test(text) || !/name:\s*\S+/.test(text)) {
        next.yaml = t("appStore.resourceForm.yamlInvalid")
      }
    }
    setErrors(next)
    if (Object.keys(next).length > 0) return
    onSave(manifest)
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[88vh] sm:max-w-3xl">
        <DialogHeader><DialogTitle>{t("appStore.addResource")}</DialogTitle></DialogHeader>
        <DialogBody className="grid gap-4 overflow-y-auto">
          <Tabs value={tab} onValueChange={(value) => {
            if (value === "yaml" && !yamlOwned) {
              setYaml(generated)
            }
            if (value === "form") {
              setYamlOwned(false)
              setYaml("")
            }
            setTab(value)
          }}>
            <TabsList>
              <TabsTrigger value="form">{t("appStore.resourceForm.formTab")}</TabsTrigger>
              <TabsTrigger value="yaml">{t("appStore.resourceForm.yamlTab")}</TabsTrigger>
            </TabsList>
          </Tabs>

          {tab === "yaml" ? (
            <FormField label={t("appStore.resourceYaml")} htmlFor="resource-form-yaml" description={t("appStore.resourceForm.yamlHelp")} error={errors.yaml}>
              <Textarea
                id="resource-form-yaml"
                className="min-h-[360px] font-mono text-xs"
                value={yamlOwned ? yaml : generated}
                onChange={(event) => { setYamlOwned(true); setYaml(event.target.value) }}
              />
            </FormField>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField label={t("appStore.resourceKind")} htmlFor="resource-form-kind" description={t("appStore.resourceForm.kindHelp")}>
                <Select value={kind} onValueChange={(value) => changeKind(value as Kind)}>
                  <SelectTrigger id="resource-form-kind"><SelectValue /></SelectTrigger>
                  <SelectContent>{KINDS.map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}</SelectContent>
                </Select>
              </FormField>
              <FormField label={t("appStore.resourceForm.name")} htmlFor="resource-form-name" required error={errors.name} description={t("appStore.resourceForm.nameHelp")}>
                <Input id="resource-form-name" className="font-mono" value={name} placeholder="app" onChange={(event) => setName(event.target.value)} />
              </FormField>

              {kind === "Deployment" || kind === "StatefulSet" ? (
                <>
                  <FormField label={t("appStore.resourceForm.replicas")} htmlFor="resource-form-replicas" description={t("appStore.resourceForm.replicasHelp")}>
                    <Input id="resource-form-replicas" type="number" min={0} max={100} value={replicas} onChange={(event) => setReplicas(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.containerName")} htmlFor="resource-form-container">
                    <Input id="resource-form-container" className="font-mono" value={containerName} onChange={(event) => setContainerName(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.image")} htmlFor="resource-form-image" required error={errors.image} description={t("appStore.resourceForm.imageHelp")} className="sm:col-span-full">
                    <Input id="resource-form-image" className="font-mono" value={image} placeholder="192.168.10.172:5000/aide-family/demo/app:1.0.0" onChange={(event) => setImage(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.containerPort")} htmlFor="resource-form-port" description={t("appStore.resourceForm.portHelp")}>
                    <Input id="resource-form-port" type="number" min={1} max={65535} value={port} onChange={(event) => setPort(event.target.value)} />
                  </FormField>
                </>
              ) : null}

              {kind === "Service" ? (
                <>
                  <FormField label={t("appStore.resourceForm.serviceType")} htmlFor="resource-form-service-type">
                    <Select value={serviceType} onValueChange={setServiceType}>
                      <SelectTrigger id="resource-form-service-type"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        {["ClusterIP", "NodePort", "LoadBalancer"].map((item) => <SelectItem key={item} value={item}>{item}</SelectItem>)}
                      </SelectContent>
                    </Select>
                  </FormField>
                  <FormField label={t("appStore.resourceForm.servicePort")} htmlFor="resource-form-service-port">
                    <Input id="resource-form-service-port" type="number" min={1} max={65535} value={servicePort} onChange={(event) => setServicePort(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.targetPort")} htmlFor="resource-form-target-port">
                    <Input id="resource-form-target-port" type="number" min={1} max={65535} value={targetPort} onChange={(event) => setTargetPort(event.target.value)} />
                  </FormField>
                </>
              ) : null}

              {kind === "ConfigMap" || kind === "Secret" ? (
                <div className="grid gap-2 sm:col-span-full">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">{kind === "Secret" ? t("appStore.resourceForm.secretData") : t("appStore.resourceForm.configData")}</p>
                    <Button type="button" variant="outline" size="sm" onClick={() => setEntries((current) => [...current, { key: "", value: "" }])}>{t("appStore.resourceForm.addKey")}</Button>
                  </div>
                  {entries.map((entry, index) => (
                    <div key={index} className="grid gap-2 sm:grid-cols-[1fr_1.4fr_auto]">
                      <Input aria-label={t("appStore.resourceForm.key")} className="font-mono" value={entry.key} placeholder="KEY" onChange={(event) => setEntries((current) => current.map((item, position) => position === index ? { ...item, key: event.target.value } : item))} />
                      <Input aria-label={t("appStore.resourceForm.value")} className="font-mono" type={kind === "Secret" ? "password" : "text"} value={entry.value} placeholder={kind === "Secret" ? t("appStore.resourceForm.secretValuePlaceholder") : "value"} onChange={(event) => setEntries((current) => current.map((item, position) => position === index ? { ...item, value: event.target.value } : item))} />
                      <Button type="button" variant="ghost" size="sm" onClick={() => setEntries((current) => current.filter((_, position) => position !== index))}>{t("action.delete")}</Button>
                    </div>
                  ))}
                  <p className="text-xs text-muted-foreground">{kind === "Secret" ? t("appStore.resourceForm.secretHelp") : t("appStore.resourceForm.configHelp")}</p>
                </div>
              ) : null}

              {kind === "Ingress" ? (
                <>
                  <FormField label={t("appStore.resourceForm.host")} htmlFor="resource-form-host" required>
                    <Input id="resource-form-host" value={ingressHost} placeholder="app.example.com" onChange={(event) => setIngressHost(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.path")} htmlFor="resource-form-path">
                    <Input id="resource-form-path" value={ingressPath} onChange={(event) => setIngressPath(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.backendService")} htmlFor="resource-form-backend">
                    <Input id="resource-form-backend" className="font-mono" value={ingressService} onChange={(event) => setIngressService(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.backendPort")} htmlFor="resource-form-backend-port">
                    <Input id="resource-form-backend-port" type="number" min={1} max={65535} value={ingressPort} onChange={(event) => setIngressPort(event.target.value)} />
                  </FormField>
                </>
              ) : null}

              {kind === "CronJob" ? (
                <>
                  <FormField label={t("appStore.resourceForm.schedule")} htmlFor="resource-form-schedule" required description={t("appStore.resourceForm.scheduleHelp")}>
                    <Input id="resource-form-schedule" className="font-mono" value={schedule} onChange={(event) => setSchedule(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.image")} htmlFor="resource-form-cron-image" required error={errors.image}>
                    <Input id="resource-form-cron-image" className="font-mono" value={image} onChange={(event) => setImage(event.target.value)} />
                  </FormField>
                  <FormField label={t("appStore.resourceForm.command")} htmlFor="resource-form-command" className="sm:col-span-full">
                    <Input id="resource-form-command" className="font-mono" value={command} onChange={(event) => setCommand(event.target.value)} />
                  </FormField>
                </>
              ) : null}
            </div>
          )}
          <p className="text-xs text-muted-foreground">{t("appStore.resourceForm.yamlFallbackHint")}</p>
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button onClick={save}>{t("appStore.resourceForm.add")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
