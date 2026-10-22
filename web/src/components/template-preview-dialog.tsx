import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { previewNotificationTemplate } from "@/api/notification-template"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { failureText } from "@/lib/form-error"

interface TemplatePreviewDialogProps {
  open: boolean
  onClose: () => void
  type: string
  app?: string
  subject: string
  body: string
}

/**
 * 用固定 mock 事件渲染模板。渲染在服务端完成，和 merak 实际投递共用 pkg/alerttmpl，
 * 所以这里看到的结果就是真实发送时的字段展开结果（query 函数除外）。
 */
export function TemplatePreviewDialog({ open, onClose, type, app, subject, body }: TemplatePreviewDialogProps) {
  const { t } = useTranslation()
  const [renderedSubject, setRenderedSubject] = useState("")
  const [renderedBody, setRenderedBody] = useState("")
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!open) {
      return
    }
    let active = true
    async function load() {
      setLoading(true)
      setError("")
      try {
        const result = await previewNotificationTemplate({ type, app, subject, body })
        if (active) {
          setRenderedSubject(result.subject)
          setRenderedBody(result.body)
        }
      } catch (err) {
        if (active) {
          setError(failureText(err, t("state.failed")))
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
  }, [open, type, app, subject, body, t])

  return (
    <Dialog open={open} onOpenChange={(value) => !value && onClose()}>
      <DialogContent className="max-h-[80vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("template.previewTitle")}</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">{t("template.previewHint")}</p>
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {error ? <p className="text-sm text-destructive">{error}</p> : null}
        {!loading && !error ? (
          <div className="grid gap-4">
            {type === "EMAIL" ? (
              <div className="grid gap-1">
                <p className="text-sm font-medium">{t("template.previewSubject")}</p>
                <p className="rounded-md border bg-muted/40 p-3 break-words">{renderedSubject || t("messageLog.notProvided")}</p>
              </div>
            ) : null}
            <div className="grid gap-1">
              <p className="text-sm font-medium">{t("template.previewBody")}</p>
              <pre className="max-h-[40vh] overflow-auto whitespace-pre-wrap break-words rounded-md border bg-muted/40 p-3 font-mono text-sm">
                {renderedBody || t("messageLog.notProvided")}
              </pre>
            </div>
          </div>
        ) : null}
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            {t("action.cancel")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
