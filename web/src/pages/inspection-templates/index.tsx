import { BookOpen, Search } from "lucide-react"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  createInspectionTemplate,
  deleteInspectionTemplate,
  listInspectionTemplates,
  updateInspectionTemplate,
} from "@/api/inspection"
import type { InspectionTemplate, InspectionTemplateFormat } from "@/api/inspection/types"
import { useSession } from "@/auth/session"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions, type RowAction } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
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
import { TemplateCodeEditor } from "@/components/template-code-editor"
import { useListQuery } from "@/hooks/use-list-query"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { InspectionTemplateGuide } from "@/pages/inspection-templates/template-guide"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

function TemplateDialog({
  open,
  template,
  base,
  onOpenChange,
  onSaved,
}: {
  open: boolean
  template: InspectionTemplate | null
  base: InspectionTemplate | null
  onOpenChange: (open: boolean) => void
  onSaved: () => void
}) {
  const { t } = useTranslation()
  const [name, setName] = useState(() => {
    if (template) {
      return template.name
    }
    return base ? `${base.name} 副本` : ""
  })
  const [description, setDescription] = useState(
    () => template?.description ?? base?.description ?? ""
  )
  const [format, setFormat] = useState<InspectionTemplateFormat>(
    () => template?.format ?? base?.format ?? "HTML"
  )
  const [body, setBody] = useState(() => {
    const source = template ?? base
    if (!source) {
      return ""
    }
    return source.format === "MARKDOWN" ? source.markdownBody : source.htmlBody
  })
  const [pending, setPending] = useState(false)
  const [error, setError] = useState("")

  async function save(event: React.FormEvent) {
    event.preventDefault()
    if (!name.trim()) {
      setError(t("inspection.nameRequired"))
      return
    }
    setPending(true)
    setError("")
    try {
      const payload = {
        name: name.trim(),
        description: description.trim(),
        format,
        htmlBody: format === "HTML" ? body : "",
        markdownBody: format === "MARKDOWN" ? body : "",
      }
      if (template) {
        await updateInspectionTemplate(template.uid, payload)
      } else {
        await createInspectionTemplate(payload)
      }
      onSaved()
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
    } finally {
      setPending(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>
            {template ? t("inspection.templateName") : t("inspection.newTemplate")}
          </DialogTitle>
        </DialogHeader>
        <form className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={save}>
          <DialogBody className="grid gap-4">
            <Input
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder={t("inspection.templateName")}
              aria-label={t("inspection.templateName")}
            />
            <Input
              value={description}
              onChange={(event) => setDescription(event.target.value)}
              placeholder={t("inspection.templateDescription")}
              aria-label={t("inspection.templateDescription")}
            />
            <div className="grid gap-2 text-sm">
              <span>{t("inspection.templateFormat")}</span>
              {template ? (
                <StatusBadge
                  value={format}
                  label={format === "HTML" ? "HTML" : "Markdown"}
                  tone={format === "HTML" ? "info" : "violet"}
                />
              ) : (
                <div className="flex gap-2">
                  {(["HTML", "MARKDOWN"] as InspectionTemplateFormat[]).map((value) => (
                    <Button
                      key={value}
                      type="button"
                      size="sm"
                      variant={format === value ? "default" : "outline"}
                      onClick={() => {
                        setFormat(value)
                        setBody("")
                      }}
                    >
                      {value === "HTML" ? "HTML" : "Markdown"}
                    </Button>
                  ))}
                </div>
              )}
              <span className="text-xs text-muted-foreground">
                {t("inspection.templateFormatHint")}
              </span>
            </div>
            <div className="grid gap-2 text-sm">
              <span>{t("inspection.templateBody")}</span>
              <TemplateCodeEditor
                language={format === "HTML" ? "html" : "markdown"}
                value={body}
                onChange={setBody}
                height={360}
                placeholder={t("inspection.templateBodyPlaceholder")}
              />
            </div>
          </DialogBody>
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t("action.cancel")}
            </Button>
            <Button type="submit" disabled={pending}>
              {t("action.save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

export function InspectionTemplatePage() {
  const { t } = useTranslation()
  const session = useSession()
  const query = useListQuery()
  const [draftName, setDraftName] = useState(query.name)
  const [editing, setEditing] = useState<InspectionTemplate | null>(null)
  const [base, setBase] = useState<InspectionTemplate | null>(null)
  const [editorOpen, setEditorOpen] = useState(false)
  const [editorKey, setEditorKey] = useState(0)
  const [guideOpen, setGuideOpen] = useState(false)
  const canWrite = session.can("inspection_template:write")
  const canDelete = session.can("inspection_template:delete")

  const { items, total, loading, failed, error, reload } =
    usePagedList<InspectionTemplate>({
      page: query.page,
      pageSize: query.pageSize,
      filterKey: `${session.namespaceUid}|${query.name}`,
      fetch: (page, pageSize) =>
        listInspectionTemplates({ page, pageSize, name: query.name }),
    })

  function openEditor(template: InspectionTemplate | null, from: InspectionTemplate | null) {
    setEditing(template)
    setBase(from)
    setEditorKey((value) => value + 1)
    setEditorOpen(true)
  }

  async function remove(template: InspectionTemplate) {
    try {
      await deleteInspectionTemplate(template.uid)
      reload()
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
    }
  }

  function actions(template: InspectionTemplate): RowAction[] {
    const list: RowAction[] = []
    if (canWrite) {
      list.push({
        key: "clone",
        label: t("inspection.newTemplate"),
        onSelect: () => openEditor(null, template),
      })
    }
    if (canDelete) {
      list.push({
        key: "delete",
        label: t("inspection.delete"),
        destructive: true,
        onSelect: () => void remove(template),
      })
    }
    return list
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={t("inspection.templates")}
        description={t("inspection.templatesDesc")}
      >
        <Button variant="outline" onClick={() => setGuideOpen(true)}>
          <BookOpen data-icon="inline-start" />
          {t("inspection.templateGuideOpen")}
        </Button>
        {canWrite ? (
          <Button onClick={() => openEditor(null, null)}>
            {t("inspection.newTemplate")}
          </Button>
        ) : null}
      </PageHeader>
      <form
        className="flex flex-wrap items-end gap-3 rounded-xl border border-border/70 bg-card/70 p-3"
        onSubmit={(event) => {
          event.preventDefault()
          query.update({ name: draftName.trim() })
        }}
      >
        <Input
          value={draftName}
          placeholder={t("inspection.templateName")}
          aria-label={t("inspection.templateName")}
          onChange={(event) => setDraftName(event.target.value)}
        />
        <Button type="submit" variant="outline">
          <Search data-icon="inline-start" />
          {t("filter.apply")}
        </Button>
      </form>
      <ListFrame
        total={total}
        page={query.page}
        pageSize={query.pageSize}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => query.update({ page: Math.max(1, query.page - 1) })}
        onNext={() => query.update({ page: query.page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <p className="text-sm text-destructive">{error || t("inspection.failed")}</p>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("inspection.emptyTemplates")}</p>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("inspection.templateName")}</TableHead>
                <TableHead>{t("inspection.templateDescription")}</TableHead>
                <TableHead>{t("inspection.templateFormat")}</TableHead>
                <TableHead>{t("inspection.startedAt")}</TableHead>
                <TableHead className="text-right">{t("action.actions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((template) => (
                <TableRow key={template.uid}>
                  <TableCell className="font-medium">{template.name}</TableCell>
                  <TableCell className="max-w-md text-muted-foreground">
                    {template.description || "—"}
                  </TableCell>
                  <TableCell>
                    <StatusBadge
                      value={template.format}
                      label={template.format === "HTML" ? "HTML" : "Markdown"}
                      tone={template.format === "HTML" ? "info" : "violet"}
                    />
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {formatDateTime(template.updatedAt) || "—"}
                  </TableCell>
                  <TableCell className="text-right">
                    <RowActions
                      nowrap
                      detailLabel={t("action.edit")}
                      moreLabel={t("action.more")}
                      onDetail={() => openEditor(template, null)}
                      detailDisabled={!canWrite}
                      actions={actions(template)}
                    />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <InspectionTemplateGuide open={guideOpen} onOpenChange={setGuideOpen} />
      <TemplateDialog
        key={editorKey}
        open={editorOpen}
        template={editing}
        base={base}
        onOpenChange={setEditorOpen}
        onSaved={() => {
          setEditorOpen(false)
          toast.success(t("inspection.templateSaved"))
          reload()
        }}
      />
    </section>
  )
}
