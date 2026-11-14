import { Pencil } from "lucide-react"
import { useTranslation } from "react-i18next"

import type { InspectionTemplate } from "@/api/inspection/types"
import { CodeBlock } from "@/components/code-block"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
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
  const sections: DetailSection[] = [
    {
      title: t("inspection.templateBody"),
      content: (
        <CodeBlock
          code={body}
          language={template.format === "MARKDOWN" ? "markdown" : "html"}
          label={template.format === "MARKDOWN" ? "Markdown" : "HTML"}
          copyable
          wrap
          maxHeightClass="max-h-[60vh]"
        />
      ),
    },
  ]

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
      sections={[
        {
          title: t("inspection.stepBasic"),
          fields: [
            { label: t("inspection.templateName"), value: template.name },
            {
              label: t("inspection.templateDescription"),
              value: template.description || "—",
              span: 2,
            },
            { label: t("inspection.templateFormat"), value: template.format },
            {
              label: t("inspection.startedAt"),
              value: formatDateTime(template.updatedAt) || "—",
            },
          ],
        },
        ...sections,
      ]}
    />
  )
}
