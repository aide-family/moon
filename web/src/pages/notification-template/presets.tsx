import { useEffect, useMemo, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { useNavigate, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import {
  createNotificationTemplate,
  listNotificationPresets,
} from "@/api/notification-template"
import type { NotificationPresetItem } from "@/api/notification-template/types"
import { ChannelLogo } from "@/components/channel-logo"
import { FormField } from "@/components/form-field"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { TemplateEditor } from "@/components/template-editor"
import { TemplatePreviewDialog } from "@/components/template-preview-dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { failureMessage, readFormError } from "@/lib/form-error"
import { cn } from "@/lib/class-names"

/**
 * 预设模板主从页：左侧预设名称列表，右侧编辑内容。
 * 预设本身只读，保存会创建当前 Namespace 的自定义模板副本。
 */
export function NotificationPresetPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [params, setParams] = useSearchParams()
  const [presets, setPresets] = useState<NotificationPresetItem[]>([])
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [loadError, setLoadError] = useState("")
  const [name, setName] = useState("")
  const [subject, setSubject] = useState("")
  const [body, setBody] = useState("")
  const [remark, setRemark] = useState("")
  const [error, setError] = useState("")
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [pending, setPending] = useState(false)
  const [previewing, setPreviewing] = useState(false)

  const initialUID = useRef(params.get("uid") ?? "")
  const selectedUID = params.get("uid") ?? ""
  const selected = presets.find((item) => item.uid === selectedUID) ?? presets[0]

  useEffect(() => {
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      try {
        const result = await listNotificationPresets()
        if (!active) {
          return
        }
        const items = result.items ?? []
        setPresets(items)
        const chosen = items.find((item) => item.uid === initialUID.current) ?? items[0]
        if (chosen) {
          setName(`${chosen.name}${t("template.copySuffix")}`)
          setSubject(chosen.subject)
          setBody(chosen.body)
          setRemark(chosen.description)
          setError("")
          setFieldErrors({})
        }
      } catch (err) {
        if (active) {
          setFailed(true)
          setLoadError(failureMessage(err, t("state.failed")))
        }
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [t])

  function selectPreset(item: NotificationPresetItem) {
    const next = new URLSearchParams(params)
    next.set("uid", item.uid)
    setParams(next)
    setName(`${item.name}${t("template.copySuffix")}`)
    setSubject(item.subject)
    setBody(item.body)
    setRemark(item.description)
    setError("")
    setFieldErrors({})
  }

  async function onSave(event: React.FormEvent) {
    event.preventDefault()
    if (!selected) {
      return
    }
    const errors: Record<string, string> = {}
    if (!name.trim()) {
      errors.name = t("form.required")
    }
    if (selected.type === "EMAIL" && !subject.trim()) {
      errors.subject = t("form.required")
    }
    if (!body.trim()) {
      errors.body = t("form.required")
    }
    if (Object.keys(errors).length > 0) {
      setFieldErrors(errors)
      return
    }
    setPending(true)
    setError("")
    setFieldErrors({})
    try {
      await createNotificationTemplate({
        name: name.trim(),
        type: selected.type,
        app: selected.type === "WEBHOOK" ? selected.app : "OTHER",
        subject: selected.type === "EMAIL" ? subject : "",
        body,
        remark,
      })
      toast.success(t("template.presetSaved"))
      navigate("/notification-templates")
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      setFieldErrors(failure.fields)
      setError(failure.form)
    } finally {
      setPending(false)
    }
  }

  const previewDialog = useMemo(
    () =>
      selected ? (
        <TemplatePreviewDialog
          open={previewing}
          onClose={() => setPreviewing(false)}
          type={selected.type}
          app={selected.app}
          subject={subject}
          body={body}
        />
      ) : null,
    [previewing, selected, subject, body],
  )

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader title={t("template.presetPageTitle")} description={t("template.presetPageDescription")}>
        <Button variant="outline" onClick={() => navigate("/notification-templates")}>
          {t("template.backToList")}
        </Button>
      </PageHeader>
      {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
      {!loading && failed ? (
        <div className="flex items-center gap-2">
          <p className="text-sm text-destructive">{loadError || t("state.failed")}</p>
          <Button variant="outline" onClick={() => navigate("/notification-templates")}>
            {t("template.backToList")}
          </Button>
        </div>
      ) : null}
      {!loading && !failed && presets.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("template.presetEmpty")}</p>
      ) : null}
      {!loading && !failed && presets.length > 0 ? (
        <div className="grid min-h-0 flex-1 gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
          <div className="glass-card min-h-0 overflow-y-auto rounded-xl p-2">
            <div className="grid gap-1">
              {presets.map((item) => (
                <button
                  key={item.uid}
                  type="button"
                  onClick={() => selectPreset(item)}
                  className={cn(
                    "flex min-w-0 items-start gap-2 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-accent",
                    selected?.uid === item.uid ? "bg-accent" : "",
                  )}
                >
                  <ChannelLogo type={item.type} app={item.app} className="mt-0.5 size-4" />
                  <span className="min-w-0">
                    <span className="block truncate font-medium">{item.name}</span>
                    <span className="block truncate text-xs text-muted-foreground">{item.description}</span>
                  </span>
                </button>
              ))}
            </div>
          </div>
          <div className="glass-card min-h-0 overflow-y-auto rounded-xl p-4">
            {selected ? (
              <form noValidate className="grid gap-4" onSubmit={onSave}>
                <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                  <span className="flex items-center gap-2">
                    <ChannelLogo type={selected.type} app={selected.app} className="size-4" />
                    <StatusBadge value={selected.type} />
                  </span>
                  {selected.type === "WEBHOOK" ? <StatusBadge value={selected.app} /> : null}
                  <span>· {t("template.presetReadonlyHint")}</span>
                </div>
                <FormField label={t("template.name")} htmlFor="preset-name" required error={fieldErrors.name}>
                  <Input
                    id="preset-name"
                    value={name}
                    placeholder={t("template.presetNamePlaceholder")}
                    onChange={(event) => setName(event.target.value)}
                    required
                  />
                </FormField>
                {selected.type === "EMAIL" ? (
                  <FormField
                    label={t("template.subject")}
                    htmlFor="preset-subject"
                    hint={t("template.subjectHint")}
                    required
                    error={fieldErrors.subject}
                  >
                    <TemplateEditor
                      value={subject}
                      onChange={setSubject}
                      singleLine
                      placeholder={t("template.subjectPlaceholder")}
                      ariaLabel={t("template.subject")}
                    />
                  </FormField>
                ) : null}
                <FormField
                  label={t("template.body")}
                  htmlFor="preset-body"
                  hint={t("template.bodyHint")}
                  required
                  error={fieldErrors.body}
                >
                  <TemplateEditor
                    value={body}
                    onChange={setBody}
                    height={300}
                    placeholder={t("template.bodyPlaceholder")}
                    ariaLabel={t("template.body")}
                  />
                </FormField>
                <FormField label={t("template.remark")} htmlFor="preset-remark">
                  <Input
                    id="preset-remark"
                    value={remark}
                    placeholder={t("template.remarkPlaceholder")}
                    onChange={(event) => setRemark(event.target.value)}
                  />
                </FormField>
                {error ? <p className="text-sm text-destructive">{error}</p> : null}
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" onClick={() => setPreviewing(true)}>
                    {t("template.preview")}
                  </Button>
                  <Button type="submit" disabled={pending}>
                    {t("template.saveAsMine")}
                  </Button>
                </div>
              </form>
            ) : null}
          </div>
        </div>
      ) : null}
      {previewDialog}
    </section>
  )
}
