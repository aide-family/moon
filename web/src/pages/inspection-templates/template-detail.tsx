import { Pencil } from "lucide-react"
import { useTranslation } from "react-i18next"

import type { InspectionTemplate } from "@/api/inspection/types"
import { CodeBlock } from "@/components/code-block"
import { DetailSheet } from "@/components/detail-sheet"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { formatDateTime } from "@/lib/date-time"

/** 报告模板详情：只读展示元信息与模板正文，可一键转入编辑。 */
export function InspectionTemplateDetail({
  open,
  template,
  canWrite,
  onEdit,
  onOpenChange,
}: {
  open: boolean
  template: InspectionTemplate | null
  canWrite: boolean
  onEdit: () => void
  onOpenChange: (open: boolean) => void
}) {
  const { t } = useTranslation()
  if (!template) {
    return null
  }
  const body = template.format === "MARKDOWN" ? template.markdownBody : template.htmlBody

  return (
    <DetailSheet
      open={open}
      onOpenChange={onOpenChange}
      size="wide"
      title={template.name}
      description={
        <>
          <StatusBadge
            value={template.format}
            label={template.format === "HTML" ? "HTML" : "Markdown"}
            tone={template.format === "HTML" ? "info" : "violet"}
          />
          <span>{t("inspection.startedAt")} {formatDateTime(template.updatedAt) || "—"}</span>
        </>
      }
      actions={
        canWrite ? (
          <Button type="button" size="sm" variant="outline" onClick={onEdit}>
            <Pencil data-icon="inline-start" />
            {t("action.edit")}
          </Button>
        ) : null
      }
      contentMode="fill"
    >
      <div className="flex min-h-0 flex-1 flex-col gap-4 px-4 py-4 sm:px-5">
        <section className="grid shrink-0 gap-2">
          <h3 className="border-b pb-1.5 font-heading text-xs font-semibold tracking-wide">
            {t("inspection.stepBasic")}
          </h3>
          <dl className="grid gap-x-6 gap-y-3 @md:grid-cols-2">
            {[
              [t("inspection.templateName"), template.name],
              [t("inspection.templateDescription"), template.description || "—"],
              [t("inspection.templateFormat"), template.format],
              [
                t("inspection.startedAt"),
                formatDateTime(template.updatedAt) || "—",
              ],
            ].map(([label, value]) => (
              <div key={label} className="grid min-w-0 gap-1">
                <dt className="text-xs text-muted-foreground">{label}</dt>
                <dd className="text-sm break-words">{value}</dd>
              </div>
            ))}
          </dl>
        </section>
        <section className="flex min-h-0 flex-1 flex-col gap-2">
          <h3 className="shrink-0 border-b pb-1.5 font-heading text-xs font-semibold tracking-wide">
            {t("inspection.templateBody")}
          </h3>
          <CodeBlock
            code={body}
            language={template.format === "MARKDOWN" ? "markdown" : "html"}
            label={template.format === "MARKDOWN" ? "Markdown" : "HTML"}
            copyable
            wrap
            fill
            className="min-h-0 flex-1"
          />
        </section>
      </div>
    </DetailSheet>
  )
}
