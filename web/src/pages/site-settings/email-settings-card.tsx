import { CircleAlert, Loader2, Save } from "lucide-react"
import { useEffect, useState, type FormEvent } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import { getEmailSettings, updateEmailSettings } from "@/api/site"
import type { EmailSettings, SaveEmailSettingsParams } from "@/api/site/types"
import { FormField } from "@/components/form-field"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Skeleton } from "@/components/ui/skeleton"
import { failureMessage } from "@/lib/form-error"

type EmailField = "host" | "port" | "from"
type EmailFieldErrors = Partial<Record<EmailField, string>>

/** 端口在表单里保持字符串：空串表示沿用服务器配置，提交时再转成数字。 */
interface EmailForm {
  enabled: boolean
  host: string
  port: string
  username: string
  password: string
  from: string
  clearPassword: boolean
}

const emptyEmailForm: EmailForm = {
  enabled: false,
  host: "",
  port: "",
  username: "",
  password: "",
  from: "",
  clearPassword: false,
}

// 后端校验只返回一条英文原文，没有字段名；按原文把错误放回对应输入框。
const messageFields: Record<string, EmailField> = {
  "invalid smtp host": "host",
  "invalid smtp port": "port",
  "invalid smtp from": "from",
}

function toEmailForm(settings: EmailSettings): EmailForm {
  return {
    enabled: settings.enabled,
    host: settings.host,
    // 0 表示没有单独保存端口，输入框按空值展示并提示沿用服务器配置。
    port: settings.port > 0 ? String(settings.port) : "",
    username: settings.username,
    password: "",
    from: settings.from,
    clearPassword: false,
  }
}

function EmailSettingsSkeleton({ label }: { label: string }) {
  return (
    <div className="grid gap-3" aria-busy="true" aria-label={label}>
      <Skeleton className="h-16 rounded-lg" />
      <div className="grid gap-3 sm:grid-cols-2">
        <Skeleton className="h-16 rounded-lg" />
        <Skeleton className="h-16 rounded-lg" />
      </div>
      <Skeleton className="h-16 rounded-lg" />
    </div>
  )
}

interface EmailSettingsCardProps {
  canRead: boolean
  canWrite: boolean
  /** 页面顶部「刷新」递增的 key，用来把邮件配置一起重新拉取。 */
  refreshKey?: number
}

/**
 * 平台「邮件服务（SMTP）」卡片：登录验证码与邀请邮件共用这一份配置。
 * 密码不回显，留空保存表示保留；「清空已保存的密码」是显式操作，保存后才生效。
 */
export function EmailSettingsCard({
  canRead,
  canWrite,
  refreshKey = 0,
}: EmailSettingsCardProps) {
  const { t } = useTranslation()
  const [settings, setSettings] = useState<EmailSettings | null>(null)
  const [form, setForm] = useState<EmailForm>(emptyEmailForm)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [saving, setSaving] = useState(false)
  const [fieldErrors, setFieldErrors] = useState<EmailFieldErrors>({})
  const [formError, setFormError] = useState("")

  useEffect(() => {
    if (!canRead) {
      return
    }
    let active = true
    getEmailSettings()
      .then((result) => {
        if (!active) {
          return
        }
        setSettings(result.email)
        setForm(toEmailForm(result.email))
        setFieldErrors({})
        setFormError("")
      })
      .catch(() => {
        if (active) {
          setFailed(true)
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [attempt, canRead, refreshKey])

  function retry() {
    setLoading(true)
    setFailed(false)
    setAttempt((value) => value + 1)
  }

  function updateField(field: keyof EmailForm, value: string | boolean) {
    setForm((current) => ({ ...current, [field]: value }))
  }

  function toggleClearPassword(checked: boolean) {
    // 勾选清空后密码输入框禁用并清空，避免「既清空又提交新密码」的歧义。
    setForm((current) => ({
      ...current,
      clearPassword: checked,
      password: checked ? "" : current.password,
    }))
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!canWrite || saving) {
      return
    }
    const errors: EmailFieldErrors = {}
    const portText = form.port.trim()
    if (portText !== "") {
      const port = Number(portText)
      if (!Number.isInteger(port) || port < 1 || port > 65535) {
        errors.port = t("systemSite.email.portInvalid")
      }
    }
    // 启用后后端会立刻用它发验证码与邀请邮件，缺字段必然发信失败，这里先拦一层。
    if (form.enabled) {
      if (form.host.trim() === "") {
        errors.host = t("systemSite.email.hostRequired")
      }
      if (portText === "") {
        errors.port = t("systemSite.email.portRequired")
      }
      if (form.from.trim() === "") {
        errors.from = t("systemSite.email.fromRequired")
      }
    }
    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors)
      setFormError("")
      return
    }
    const body: SaveEmailSettingsParams = {
      enabled: form.enabled,
      host: form.host.trim(),
      port: portText === "" ? 0 : Number(portText),
      username: form.username.trim(),
      password: form.password,
      from: form.from.trim(),
      clearPassword: form.clearPassword,
    }
    setSaving(true)
    setFieldErrors({})
    setFormError("")
    try {
      const result = await updateEmailSettings(body)
      setSettings(result.email)
      setForm(toEmailForm(result.email))
      toast.success(t("systemSite.email.saved"))
    } catch (err) {
      const message = failureMessage(err, t("state.failed"))
      const raw = err instanceof ApiError ? err.raw : ""
      const field = messageFields[raw]
      if (field) {
        setFieldErrors({ [field]: message })
      } else {
        setFormError(message)
      }
    } finally {
      setSaving(false)
    }
  }

  const passwordPlaceholder = settings?.passwordSet
    ? t("systemSite.email.passwordPlaceholder")
    : t("systemSite.email.passwordEmptyPlaceholder")

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("systemSite.email.title")}</CardTitle>
        <CardDescription>{t("systemSite.email.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        {!canRead ? (
          <p className="text-sm text-muted-foreground">
            {t("systemSite.email.noPermission")}
          </p>
        ) : loading && !settings ? (
          <EmailSettingsSkeleton label={t("state.loading")} />
        ) : failed && !settings ? (
          <div className="flex flex-col items-start gap-3">
            <p
              role="alert"
              className="flex items-center gap-2 text-sm text-destructive"
            >
              <CircleAlert className="size-4" aria-hidden="true" />
              {t("systemSite.email.loadFailed")}
            </p>
            <Button type="button" variant="outline" size="sm" onClick={retry}>
              {t("action.retry")}
            </Button>
          </div>
        ) : (
          <form noValidate className="grid gap-4" onSubmit={onSubmit}>
            <label className="flex items-start gap-3 rounded-lg border border-border/60 p-4">
              <Checkbox
                id="email-settings-enabled"
                className="mt-0.5"
                checked={form.enabled}
                disabled={!canWrite}
                onCheckedChange={(value) =>
                  updateField("enabled", value === true)
                }
              />
              <span className="grid gap-1">
                <span className="text-sm font-medium">
                  {t("systemSite.email.enabled")}
                </span>
                <span className="text-xs text-muted-foreground">
                  {t("systemSite.email.enabledHint")}
                </span>
              </span>
            </label>

            <div className="grid gap-4 sm:grid-cols-2">
              <FormField
                label={t("systemSite.email.host")}
                htmlFor="email-settings-host"
                error={fieldErrors.host}
              >
                <Input
                  id="email-settings-host"
                  value={form.host}
                  placeholder={t("systemSite.email.hostPlaceholder")}
                  readOnly={!canWrite}
                  aria-invalid={fieldErrors.host ? true : undefined}
                  onChange={(event) => updateField("host", event.target.value)}
                />
              </FormField>
              <FormField
                label={t("systemSite.email.port")}
                htmlFor="email-settings-port"
                hint={t("systemSite.email.portHint")}
                error={fieldErrors.port}
              >
                <Input
                  id="email-settings-port"
                  type="number"
                  inputMode="numeric"
                  min={1}
                  max={65535}
                  value={form.port}
                  placeholder={t("systemSite.email.portPlaceholder")}
                  readOnly={!canWrite}
                  aria-invalid={fieldErrors.port ? true : undefined}
                  onChange={(event) => updateField("port", event.target.value)}
                />
              </FormField>
              <FormField
                label={t("systemSite.email.username")}
                htmlFor="email-settings-username"
              >
                <Input
                  id="email-settings-username"
                  value={form.username}
                  placeholder={t("systemSite.email.usernamePlaceholder")}
                  readOnly={!canWrite}
                  onChange={(event) =>
                    updateField("username", event.target.value)
                  }
                />
              </FormField>
              <FormField
                label={t("systemSite.email.from")}
                htmlFor="email-settings-from"
                hint={t("systemSite.email.fromHint")}
                error={fieldErrors.from}
              >
                <Input
                  id="email-settings-from"
                  value={form.from}
                  placeholder={t("systemSite.email.fromPlaceholder")}
                  readOnly={!canWrite}
                  aria-invalid={fieldErrors.from ? true : undefined}
                  onChange={(event) => updateField("from", event.target.value)}
                />
              </FormField>
              <FormField
                label={t("systemSite.email.password")}
                htmlFor="email-settings-password"
                hint={t("systemSite.email.passwordHint")}
                className="sm:col-span-full"
                actions={
                  settings?.passwordSet ? (
                    <StatusBadge
                      value="CONFIGURED"
                      label={t("systemSite.email.passwordSet")}
                      tone="neutral"
                    />
                  ) : undefined
                }
              >
                <Input
                  id="email-settings-password"
                  type="password"
                  autoComplete="new-password"
                  value={form.password}
                  placeholder={passwordPlaceholder}
                  readOnly={!canWrite}
                  disabled={form.clearPassword}
                  onChange={(event) =>
                    updateField("password", event.target.value)
                  }
                />
              </FormField>
            </div>

            {settings?.passwordSet ? (
              <div className="grid gap-2 rounded-lg border border-border/60 p-4">
                <label className="flex items-start gap-3">
                  <Checkbox
                    id="email-settings-clear-password"
                    className="mt-0.5"
                    checked={form.clearPassword}
                    disabled={!canWrite}
                    onCheckedChange={(value) =>
                      toggleClearPassword(value === true)
                    }
                  />
                  <span className="grid gap-1">
                    <span className="text-sm font-medium">
                      {t("systemSite.email.clearPassword")}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {t("systemSite.email.clearPasswordHint")}
                    </span>
                  </span>
                </label>
                {form.clearPassword ? (
                  <p className="text-xs text-warning">
                    {t("systemSite.email.clearPasswordNotice")}
                  </p>
                ) : null}
              </div>
            ) : null}

            {formError ? (
              <p role="alert" className="text-sm text-destructive">
                {formError}
              </p>
            ) : null}
            {canWrite ? (
              <div className="flex justify-end">
                <Button type="submit" disabled={saving}>
                  {saving ? (
                    <Loader2
                      className="size-4 animate-spin"
                      aria-hidden="true"
                    />
                  ) : (
                    <Save aria-hidden="true" />
                  )}
                  {saving ? t("systemSite.saving") : t("systemSite.email.save")}
                </Button>
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">
                {t("systemSite.readOnly")}
              </p>
            )}
          </form>
        )}
      </CardContent>
    </Card>
  )
}
