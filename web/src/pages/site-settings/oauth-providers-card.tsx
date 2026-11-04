import { CircleAlert, Loader2, Save } from "lucide-react"
import { useEffect, useState, type FormEvent } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { ApiError } from "@/api/http"
import { getOAuthProviders, updateOAuthProvider } from "@/api/site"
import type {
  OAuthProvider,
  OAuthProviderSettings,
  SaveOAuthProviderParams,
} from "@/api/site/types"
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
import { enumLabel } from "@/lib/enum-label"
import { failureMessage } from "@/lib/form-error"

/** 来源清单与顺序由后端固定；前端按同一顺序渲染，缺失的来源按空配置展示。 */
const providerOrder: OAuthProvider[] = ["GITHUB", "GITEE", "FEISHU", "DINGTALK"]

interface ProviderForm {
  enabled: boolean
  clientId: string
  clientSecret: string
  clearSecret: boolean
}

interface ProviderFieldErrors {
  clientId?: string
  clientSecret?: string
}

const emptyProviderForms: Record<OAuthProvider, ProviderForm> = {
  GITHUB: {
    enabled: false,
    clientId: "",
    clientSecret: "",
    clearSecret: false,
  },
  GITEE: { enabled: false, clientId: "", clientSecret: "", clearSecret: false },
  FEISHU: {
    enabled: false,
    clientId: "",
    clientSecret: "",
    clearSecret: false,
  },
  DINGTALK: {
    enabled: false,
    clientId: "",
    clientSecret: "",
    clearSecret: false,
  },
}

function toProviderForms(
  items: OAuthProviderSettings[]
): Record<OAuthProvider, ProviderForm> {
  const forms: Record<OAuthProvider, ProviderForm> = { ...emptyProviderForms }
  for (const item of items) {
    if (!providerOrder.includes(item.provider)) {
      continue
    }
    forms[item.provider] = {
      enabled: item.enabled,
      clientId: item.clientId,
      clientSecret: "",
      clearSecret: false,
    }
  }
  return forms
}

function OAuthProvidersSkeleton({ label }: { label: string }) {
  return (
    <div className="grid gap-3" aria-busy="true" aria-label={label}>
      {providerOrder.map((provider) => (
        <Skeleton key={provider} className="h-24 rounded-lg" />
      ))}
    </div>
  )
}

interface OAuthProvidersCardProps {
  canRead: boolean
  canWrite: boolean
  /** 页面顶部「刷新」递增的 key，用来把来源配置一起重新拉取。 */
  refreshKey?: number
}

/**
 * 平台「第三方登录（OAuth2）」卡片。
 *
 * 每个来源各自一个表单、各自保存：调用 PUT /oauth-providers/{provider} 只提交
 * 当前来源，所以打开一个来源不会顺带改动其他来源；密钥不回显，留空保存表示保留，
 * 「清空已保存的密钥」同样是保存后才生效的显式操作。
 */
export function OAuthProvidersCard({
  canRead,
  canWrite,
  refreshKey = 0,
}: OAuthProvidersCardProps) {
  const { t } = useTranslation()
  const [providers, setProviders] = useState<OAuthProviderSettings[] | null>(
    null
  )
  const [forms, setForms] =
    useState<Record<OAuthProvider, ProviderForm>>(emptyProviderForms)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)
  const [saving, setSaving] = useState<OAuthProvider | "">("")
  const [fieldErrors, setFieldErrors] = useState<
    Partial<Record<OAuthProvider, ProviderFieldErrors>>
  >({})
  const [formErrors, setFormErrors] = useState<
    Partial<Record<OAuthProvider, string>>
  >({})

  useEffect(() => {
    if (!canRead) {
      return
    }
    let active = true
    getOAuthProviders()
      .then((result) => {
        if (!active) {
          return
        }
        setProviders(result.items)
        setForms(toProviderForms(result.items))
        setFieldErrors({})
        setFormErrors({})
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

  function updateProviderForm(
    provider: OAuthProvider,
    patch: Partial<ProviderForm>
  ) {
    setForms((current) => ({
      ...current,
      [provider]: { ...current[provider], ...patch },
    }))
  }

  function toggleClearSecret(provider: OAuthProvider, checked: boolean) {
    // 勾选清空后密钥输入框禁用并清空，避免「既清空又提交新密钥」的歧义。
    setForms((current) => ({
      ...current,
      [provider]: {
        ...current[provider],
        clearSecret: checked,
        clientSecret: checked ? "" : current[provider].clientSecret,
      },
    }))
  }

  async function onSubmit(
    event: FormEvent<HTMLFormElement>,
    provider: OAuthProvider
  ) {
    event.preventDefault()
    if (!canWrite || saving !== "") {
      return
    }
    const form = forms[provider]
    const clientId = form.clientId.trim()
    const errors: ProviderFieldErrors = {}
    // 启用来源必须有 client id，否则登录页会出现点不动的入口。
    if (form.enabled && clientId === "") {
      errors.clientId = t("systemSite.oauth.clientIdRequired")
    }
    if (Object.keys(errors).length > 0) {
      setFieldErrors((current) => ({ ...current, [provider]: errors }))
      setFormErrors((current) => ({ ...current, [provider]: "" }))
      return
    }
    const body: SaveOAuthProviderParams = {
      enabled: form.enabled,
      clientId,
      clientSecret: form.clearSecret ? "" : form.clientSecret,
      clearSecret: form.clearSecret,
    }
    setSaving(provider)
    setFieldErrors((current) => ({ ...current, [provider]: {} }))
    setFormErrors((current) => ({ ...current, [provider]: "" }))
    try {
      const result = await updateOAuthProvider(provider, body)
      setProviders((current) => [
        ...(current ?? []).filter(
          (item) => item.provider !== result.item.provider
        ),
        result.item,
      ])
      setForms((current) => ({
        ...current,
        [provider]: {
          enabled: result.item.enabled,
          clientId: result.item.clientId,
          clientSecret: "",
          clearSecret: false,
        },
      }))
      toast.success(
        t("systemSite.oauth.saved", { name: enumLabel(t, provider) })
      )
    } catch (err) {
      const message = failureMessage(err, t("state.failed"))
      const raw = err instanceof ApiError ? err.raw : ""
      if (raw === "invalid client id") {
        setFieldErrors((current) => ({
          ...current,
          [provider]: { clientId: message },
        }))
      } else if (raw === "invalid client secret") {
        setFieldErrors((current) => ({
          ...current,
          [provider]: { clientSecret: message },
        }))
      } else {
        setFormErrors((current) => ({ ...current, [provider]: message }))
      }
    } finally {
      setSaving("")
    }
  }

  const settingsByProvider = new Map(
    (providers ?? []).map((item) => [item.provider, item] as const)
  )

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("systemSite.oauth.title")}</CardTitle>
        <CardDescription>{t("systemSite.oauth.description")}</CardDescription>
      </CardHeader>
      <CardContent>
        {!canRead ? (
          <p className="text-sm text-muted-foreground">
            {t("systemSite.oauth.noPermission")}
          </p>
        ) : loading && !providers ? (
          <OAuthProvidersSkeleton label={t("state.loading")} />
        ) : failed && !providers ? (
          <div className="flex flex-col items-start gap-3">
            <p
              role="alert"
              className="flex items-center gap-2 text-sm text-destructive"
            >
              <CircleAlert className="size-4" aria-hidden="true" />
              {t("systemSite.oauth.loadFailed")}
            </p>
            <Button type="button" variant="outline" size="sm" onClick={retry}>
              {t("action.retry")}
            </Button>
          </div>
        ) : (
          <div className="grid gap-4">
            {providerOrder.map((provider) => {
              const form = forms[provider]
              const settings = settingsByProvider.get(provider)
              const errors = fieldErrors[provider]
              const providerError = formErrors[provider]
              const providerSaving = saving === provider
              const name = enumLabel(t, provider)
              return (
                <form
                  key={provider}
                  noValidate
                  aria-busy={providerSaving || undefined}
                  className="grid min-w-0 gap-4 rounded-lg border border-border/60 p-4"
                  onSubmit={(event) => void onSubmit(event, provider)}
                >
                  <h3 className="text-sm font-medium">{name}</h3>
                  <label className="flex items-start gap-3">
                    <Checkbox
                      id={`oauth-${provider}-enabled`}
                      className="mt-0.5"
                      checked={form.enabled}
                      disabled={!canWrite || providerSaving}
                      onCheckedChange={(value) =>
                        updateProviderForm(provider, {
                          enabled: value === true,
                        })
                      }
                    />
                    <span className="grid gap-1">
                      <span className="text-sm font-medium">
                        {t("systemSite.oauth.enabled")}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {t("systemSite.oauth.enabledHint")}
                      </span>
                    </span>
                  </label>
                  <div className="grid gap-4 sm:grid-cols-2">
                    <FormField
                      label={t("systemSite.oauth.clientId")}
                      htmlFor={`oauth-${provider}-client-id`}
                      error={errors?.clientId}
                    >
                      <Input
                        id={`oauth-${provider}-client-id`}
                        value={form.clientId}
                        placeholder={t("systemSite.oauth.clientIdPlaceholder")}
                        readOnly={!canWrite || providerSaving}
                        aria-invalid={errors?.clientId ? true : undefined}
                        onChange={(event) =>
                          updateProviderForm(provider, {
                            clientId: event.target.value,
                          })
                        }
                      />
                    </FormField>
                    <FormField
                      label={t("systemSite.oauth.clientSecret")}
                      htmlFor={`oauth-${provider}-client-secret`}
                      hint={t("systemSite.oauth.clientSecretHint")}
                      error={errors?.clientSecret}
                      actions={
                        settings?.secretSet ? (
                          <StatusBadge
                            value="CONFIGURED"
                            label={t("systemSite.oauth.secretSet")}
                            tone="neutral"
                          />
                        ) : undefined
                      }
                    >
                      <Input
                        id={`oauth-${provider}-client-secret`}
                        type="password"
                        autoComplete="new-password"
                        value={form.clientSecret}
                        placeholder={
                          settings?.secretSet
                            ? t("systemSite.oauth.clientSecretPlaceholder")
                            : t("systemSite.oauth.clientSecretEmptyPlaceholder")
                        }
                        readOnly={!canWrite || providerSaving}
                        disabled={form.clearSecret}
                        aria-invalid={errors?.clientSecret ? true : undefined}
                        onChange={(event) =>
                          updateProviderForm(provider, {
                            clientSecret: event.target.value,
                          })
                        }
                      />
                    </FormField>
                  </div>
                  {settings?.secretSet ? (
                    <div className="grid gap-2 rounded-lg border border-border/60 p-3">
                      <label className="flex items-start gap-3">
                        <Checkbox
                          id={`oauth-${provider}-clear-secret`}
                          className="mt-0.5"
                          checked={form.clearSecret}
                          disabled={!canWrite || providerSaving}
                          onCheckedChange={(value) =>
                            toggleClearSecret(provider, value === true)
                          }
                        />
                        <span className="grid gap-1">
                          <span className="text-sm font-medium">
                            {t("systemSite.oauth.clearSecret")}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {t("systemSite.oauth.clearSecretHint")}
                          </span>
                        </span>
                      </label>
                      {form.clearSecret ? (
                        <p className="text-xs text-warning">
                          {t("systemSite.oauth.clearSecretNotice")}
                        </p>
                      ) : null}
                    </div>
                  ) : null}
                  {providerError ? (
                    <p role="alert" className="text-sm text-destructive">
                      {providerError}
                    </p>
                  ) : null}
                  {canWrite ? (
                    <div className="flex justify-end">
                      <Button
                        type="submit"
                        disabled={saving !== ""}
                        aria-label={`${t("systemSite.oauth.save")} ${name}`}
                      >
                        {providerSaving ? (
                          <Loader2
                            className="size-4 animate-spin"
                            aria-hidden="true"
                          />
                        ) : (
                          <Save aria-hidden="true" />
                        )}
                        {providerSaving
                          ? t("systemSite.saving")
                          : t("systemSite.oauth.save")}
                      </Button>
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">
                      {t("systemSite.readOnly")}
                    </p>
                  )}
                </form>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
