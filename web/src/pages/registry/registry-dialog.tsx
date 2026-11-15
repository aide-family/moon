import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  createRegistry,
  probeRegistry,
  testRegistry,
  updateRegistry,
} from "@/api/registry"
import type { RegistryDriver, RegistryItem } from "@/api/registry/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
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
import { Textarea } from "@/components/ui/textarea"
import { formErrorMessage } from "@/lib/form-error"

const NAME_MAX = 64
const ENDPOINT_MAX = 512
const REMARK_MAX = 512

interface RegistryDialogProps {
  open: boolean
  registry: RegistryItem | null
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}

/** 接入 / 编辑镜像仓库。密码留空表示保持原值。 */
export function RegistryDialog({
  open,
  registry,
  onOpenChange,
  onSaved,
}: RegistryDialogProps) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => registry?.name ?? "")
  const [driver, setDriver] = useState<RegistryDriver>(
    () => registry?.driver ?? "HARBOR"
  )
  const [endpoint, setEndpoint] = useState(() => registry?.endpoint ?? "")
  const [username, setUsername] = useState(() => registry?.username ?? "")
  const [password, setPassword] = useState("")
  const [insecure, setInsecure] = useState(
    () => registry?.allowInsecureTls ?? false
  )
  const [remark, setRemark] = useState(() => registry?.remark ?? "")
  const [testing, setTesting] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})

  async function handleTest() {
    setTesting(true)
    setError("")
    try {
      if (registry && !password) {
        const result = await testRegistry(registry.uid)
        toast.success(t("registry.testSuccess"), {
          description: `${result.probe.endpoint} · ${result.probe.repositoryCount} repos`,
        })
      } else {
        const result = await probeRegistry({
          driver,
          endpoint: endpoint.trim(),
          username: username.trim(),
          password,
          allowInsecureTls: insecure,
        })
        toast.success(t("registry.testSuccess"), {
          description: `${result.probe.endpoint} · ${result.probe.repositoryCount} repos`,
        })
      }
    } catch (err) {
      setError(formErrorMessage(err, t("registry.testFailed")))
    } finally {
      setTesting(false)
    }
  }

  async function handleSave() {
    const next: Record<string, string> = {}
    if (!name.trim()) next.name = t("form.required")
    if (!endpoint.trim()) next.endpoint = t("form.required")
    if (name.trim().length > NAME_MAX) next.name = t("form.nameTooLong")
    if (endpoint.trim().length > ENDPOINT_MAX)
      next.endpoint = t("form.nameTooLong")
    if (remark.trim().length > REMARK_MAX)
      next.remark = t("form.nameTooLong")
    setFieldErrors(next)
    if (Object.keys(next).length > 0) return
    setSaving(true)
    setError("")
    try {
      if (registry) {
        await updateRegistry({
          uid: registry.uid,
          name: name.trim(),
          driver,
          endpoint: endpoint.trim(),
          username: username.trim(),
          password: password || undefined,
          allowInsecureTls: insecure,
          remark: remark.trim(),
        })
      } else {
        await createRegistry({
          name: name.trim(),
          driver,
          endpoint: endpoint.trim(),
          username: username.trim(),
          password,
          allowInsecureTls: insecure,
          remark: remark.trim(),
        })
      }
      onSaved()
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {registry ? t("registry.editTitle") : t("registry.createTitle")}
          </DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4 sm:grid-cols-2">
          <FormField
            label={t("registry.name")}
            htmlFor="registry-name"
            required
            error={fieldErrors.name}
            description={t("registry.nameHelp")}
            counter={{ value: name.length, max: NAME_MAX }}
          >
            <Input
              id="registry-name"
              value={name}
              maxLength={NAME_MAX}
              placeholder={t("registry.namePlaceholder")}
              onChange={(event) => setName(event.target.value)}
            />
          </FormField>
          <FormField
            label={t("registry.driver")}
            htmlFor="registry-driver"
            required
            description={t("registry.driverHelp")}
          >
            <Select
              value={driver}
              onValueChange={(value) => setDriver(value as RegistryDriver)}
            >
              <SelectTrigger id="registry-driver">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="HARBOR">{t("registry.driverHarbor")}</SelectItem>
                <SelectItem value="DOCKER_V2">{t("registry.driverDockerV2")}</SelectItem>
              </SelectContent>
            </Select>
          </FormField>
          <FormField
            label={t("registry.endpoint")}
            htmlFor="registry-endpoint"
            required
            error={fieldErrors.endpoint}
            description={t("registry.endpointHelp")}
            className="sm:col-span-full"
          >
            <Input
              id="registry-endpoint"
              className="font-mono"
              value={endpoint}
              maxLength={ENDPOINT_MAX}
              placeholder={t("registry.endpointPlaceholder")}
              onChange={(event) => setEndpoint(event.target.value)}
            />
          </FormField>
          <FormField label={t("registry.username")} htmlFor="registry-username" description={t("registry.usernameHelp")}>
            <Input
              id="registry-username"
              value={username}
              placeholder={t("registry.usernamePlaceholder")}
              onChange={(event) => setUsername(event.target.value)}
            />
          </FormField>
          <FormField
            label={t("registry.password")}
            htmlFor="registry-password"
            hint={t("registry.passwordKeepHint")}
            description={t("registry.passwordHelp")}
          >
            <Input
              id="registry-password"
              type="password"
              value={password}
              placeholder={t("registry.passwordPlaceholder")}
              onChange={(event) => setPassword(event.target.value)}
            />
          </FormField>
          <div className="flex items-start gap-2 sm:col-span-full">
            <Checkbox
              id="registry-insecure"
              checked={insecure}
              onCheckedChange={(value) => setInsecure(value === true)}
            />
            <div className="grid gap-1">
              <label htmlFor="registry-insecure" className="text-sm font-medium">
                {t("registry.allowInsecureTls")}
              </label>
              <p className="text-xs text-muted-foreground">
                {t("registry.allowInsecureTlsHint")}
              </p>
            </div>
          </div>
          <FormField
            label={t("registry.remark")}
            htmlFor="registry-remark"
            error={fieldErrors.remark}
            className="sm:col-span-full"
          >
            <Textarea
              id="registry-remark"
              value={remark}
              maxLength={REMARK_MAX}
              placeholder={t("registry.remarkPlaceholder")}
              onChange={(event) => setRemark(event.target.value)}
            />
          </FormField>
          {error ? (
            <p className="text-sm text-destructive sm:col-span-full">{error}</p>
          ) : null}
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t("action.cancel")}
          </Button>
          <Button variant="outline" disabled={testing} onClick={handleTest}>
            {testing ? t("registry.testing") : t("registry.testConnection")}
          </Button>
          <Button disabled={saving} onClick={handleSave}>
            {saving ? t("action.saving") : t("action.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
