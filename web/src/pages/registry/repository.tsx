import { Search } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Navigate, useNavigate, useParams, useSearchParams } from "react-router-dom"
import { toast } from "sonner"

import {
  copyRegistryTag,
  deleteRegistryTag,
  getRegistry,
  getRegistryTag,
  listRegistryTags,
} from "@/api/registry"
import type {
  RegistryItem,
  RegistryTagDetail,
  RegistryTagItem,
} from "@/api/registry/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { ConfirmAction } from "@/components/confirm-action"
import { DetailSheet } from "@/components/detail-sheet"
import { FormField } from "@/components/form-field"
import { ListFrame } from "@/components/list-frame"
import { PageHeader } from "@/components/page-header"
import { RowActions } from "@/components/row-actions"
import { TableActionsCell, TableActionsHead } from "@/components/table-heads"
import { Badge } from "@/components/ui/badge"
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { usePagedList } from "@/hooks/use-paged-list"
import { formatUnixDateTime } from "@/lib/date-time"
import { formErrorMessage } from "@/lib/form-error"
import { TABLE_MIN_WIDTH, TABLE_NOWRAP } from "@/lib/table-layout"

import { SyncDialog } from "./sync-dialog"

const ALL_ARCH = "ALL"

function formatBytes(value: number | string): string {
  const bytes = Number(value)
  if (!Number.isFinite(bytes) || bytes <= 0) return "—"
  const units = ["B", "KB", "MB", "GB", "TB"]
  let size = bytes
  let unit = 0
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024
    unit += 1
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`
}

/** 镜像详情：Tag、多架构、digest 与引用；同步/复制/删除 Tag。 */
export function RegistryRepositoryPage() {
  const { t } = useTranslation()
  const session = useSession()
  const navigate = useNavigate()
  const { uid = "" } = useParams()
  const [params, setParams] = useSearchParams()
  const project = params.get("project") ?? ""
  const repository = params.get("repository") ?? ""
  const architecture = params.get("architecture") ?? ""
  const name = params.get("name") ?? ""
  const page = Math.max(1, Number(params.get("page")) || 1)
  const [registry, setRegistry] = useState<RegistryItem | null>(null)
  const [detail, setDetail] = useState<RegistryTagDetail | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)
  const [copying, setCopying] = useState<RegistryTagItem | null>(null)
  const [copyTargetRepository, setCopyTargetRepository] = useState("")
  const [copyTargetTag, setCopyTargetTag] = useState("")
  const [copyPending, setCopyPending] = useState(false)
  const [copyError, setCopyError] = useState("")
  const [deleting, setDeleting] = useState<RegistryTagItem | null>(null)
  const [deletePending, setDeletePending] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [syncOpen, setSyncOpen] = useState(false)
  const [syncSource, setSyncSource] = useState<{ repository: string; tag: string } | null>(null)
  const { items, total, loading, failed, error, reload } = usePagedList<RegistryTagItem>({
    page,
    pageSize: 20,
    filterKey: `${uid}|${project}|${repository}|${architecture}|${name}`,
    fetch: (pageNo, pageSize) =>
      listRegistryTags(uid, {
        project,
        repository,
        architecture,
        name,
        page: pageNo,
        pageSize,
      }),
  })

  useEffect(() => {
    if (!uid) return
    let active = true
    void getRegistry(uid)
      .then((result) => {
        if (active) setRegistry(result.registry)
      })
      .catch((err) => {
        if (err instanceof ApiError && err.status === 403) return
      })
    return () => {
      active = false
    }
  }, [uid])

  if (!session.loading && !session.can("registry:read")) {
    return <Navigate to="/forbidden" replace />
  }
  if (!repository) {
    return <Navigate to={`/registries/${uid}`} replace />
  }

  function updateParams(next: { architecture?: string; name?: string; page?: number }) {
    const query = new URLSearchParams(params)
    if (next.architecture !== undefined) {
      if (next.architecture) query.set("architecture", next.architecture)
      else query.delete("architecture")
      query.set("page", "1")
    }
    if (next.name !== undefined) {
      if (next.name) query.set("name", next.name)
      else query.delete("name")
      query.set("page", "1")
    }
    if (next.page !== undefined) query.set("page", String(next.page))
    setParams(query)
  }

  async function openDetail(item: RegistryTagItem) {
    setDetailOpen(true)
    setDetailLoading(true)
    setDetail(null)
    try {
      const result = await getRegistryTag(uid, { project, repository, tag: item.tag })
      setDetail(result.tag)
    } catch (err) {
      toast.error(formErrorMessage(err, t("state.failed")))
      setDetailOpen(false)
    } finally {
      setDetailLoading(false)
    }
  }

  function openCopy(item: RegistryTagItem) {
    setCopying(item)
    setCopyTargetRepository(repository)
    setCopyTargetTag(`${item.tag}-copy`)
    setCopyError("")
  }

  async function submitCopy() {
    if (!copying) return
    if (!copyTargetRepository.trim()) {
      setCopyError(t("field.required", { label: t("registry.targetRepository") }))
      return
    }
    if (!copyTargetTag.trim()) {
      setCopyError(t("field.required", { label: t("registry.targetTag") }))
      return
    }
    setCopyPending(true)
    try {
      await copyRegistryTag(uid, {
        project,
        repository,
        tag: copying.tag,
        digest: copying.digest,
        targetRepository: copyTargetRepository.trim(),
        targetTag: copyTargetTag.trim(),
      })
      setCopying(null)
      toast.success(t("state.saved"))
      reload()
    } catch (err) {
      setCopyError(formErrorMessage(err, t("state.failed")))
    } finally {
      setCopyPending(false)
    }
  }

  async function handleDelete() {
    if (!deleting) return
    setDeletePending(true)
    try {
      await deleteRegistryTag(uid, {
        project,
        repository,
        tag: deleting.tag,
        digest: deleting.digest,
      })
      setDeleting(null)
      toast.success(t("registry.deleted"))
      reload()
    } catch (err) {
      setDeleteError(formErrorMessage(err, t("state.failed")))
    } finally {
      setDeletePending(false)
    }
  }

  const detailSections = detail
    ? [
        {
          title: t("registry.tagDetail"),
          fields: [
            { label: t("registry.digest"), value: detail.digest },
            { label: t("registry.mediaType"), value: detail.mediaType },
            { label: t("registry.size"), value: formatBytes(detail.size) },
            {
              label: t("registry.pushedAt"),
              value: formatUnixDateTime(detail.createdAt) || "—",
            },
            { label: t("registry.labels"), value: detail.labels.join(", ") || "—", span: 2 as const },
          ],
        },
        {
          title: t("registry.platforms"),
          content: (
            <div className="grid gap-2">
              {detail.platforms.map((platform) => (
                <div key={`${platform.os}/${platform.architecture}/${platform.variant}`} className="flex flex-wrap items-center gap-2 text-sm">
                  <Badge variant="outline">
                    {[platform.os, platform.architecture, platform.variant].filter(Boolean).join("/")}
                  </Badge>
                  <span className="font-mono text-xs text-muted-foreground break-all">{platform.digest}</span>
                  <span className="text-xs text-muted-foreground">{formatBytes(platform.size)}</span>
                </div>
              ))}
            </div>
          ),
        },
        {
          title: t("registry.references"),
          content:
            detail.references.length === 0 ? (
              <p className="text-sm text-muted-foreground">—</p>
            ) : (
              <div className="grid gap-2">
                {detail.references.map((reference) => (
                  <div key={reference.versionUid} className="text-sm">
                    {reference.applicationName} · {reference.version}{" "}
                    <Badge variant="outline">{reference.status}</Badge>
                  </div>
                ))}
              </div>
            ),
        },
      ]
    : []

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-6">
      <PageHeader
        title={repository}
        description={
          registry
            ? `${registry.endpoint} · ${registry.driver}`
            : t("state.loading")
        }
      >
        <Button variant="outline" onClick={() => navigate(`/registries/${uid}`)}>
          {t("action.goBack")}
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            setSyncSource({ repository: repository, tag: "latest" })
            setSyncOpen(true)
          }}
        >
          {t("registry.syncTitle")}
        </Button>
      </PageHeader>
      <div className="flex flex-wrap items-center gap-2">
        <Select
          value={architecture || ALL_ARCH}
          onValueChange={(value) => updateParams({ architecture: value === ALL_ARCH ? "" : value })}
        >
          <SelectTrigger className="w-44">
            <SelectValue>{architecture || t("registry.allArchitectures")}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={ALL_ARCH}>{t("registry.allArchitectures")}</SelectItem>
            <SelectItem value="amd64">amd64</SelectItem>
            <SelectItem value="arm64">arm64</SelectItem>
          </SelectContent>
        </Select>
        <div className="relative min-w-0 w-full max-w-[320px]">
          <Search aria-hidden="true" className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="pl-9"
            value={name}
            placeholder={t("registry.searchTag")}
            onChange={(event) => updateParams({ name: event.target.value })}
          />
        </div>
      </div>
      <ListFrame
        total={total}
        page={page}
        pageSize={20}
        showPager={!loading && !failed && items.length > 0}
        onPrev={() => updateParams({ page: Math.max(1, page - 1) })}
        onNext={() => updateParams({ page: page + 1 })}
      >
        {loading ? <p className="text-sm text-muted-foreground">{t("state.loading")}</p> : null}
        {!loading && failed ? (
          <div className="flex items-center gap-2">
            <p className="text-sm text-destructive">{error || t("state.failed")}</p>
            <Button variant="outline" onClick={() => reload()}>{t("action.retry")}</Button>
          </div>
        ) : null}
        {!loading && !failed && items.length === 0 ? (
          <div className="grid gap-2">
            <p className="text-sm text-muted-foreground">{t("registry.tagsEmpty")}</p>
            <p className="text-sm text-muted-foreground">{t("registry.tagsEmptyHint")}</p>
          </div>
        ) : null}
        {!loading && !failed && items.length > 0 ? (
          <Table className={TABLE_MIN_WIDTH.default}>
            <TableHeader>
              <TableRow>
                <TableHead>{t("registry.tag")}</TableHead>
                <TableHead>{t("registry.architecture")}</TableHead>
                <TableHead>{t("registry.digest")}</TableHead>
                <TableHead className={TABLE_NOWRAP}>{t("registry.size")}</TableHead>
                <TableActionsHead>{t("action.actions")}</TableActionsHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.tag}>
                  <TableCell className="font-mono text-xs">{item.tag}</TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {item.platforms.length === 0 ? "—" : item.platforms.map((platform) => (
                        <Badge key={platform} variant="outline">{platform}</Badge>
                      ))}
                    </div>
                  </TableCell>
                  <TableCell className="font-mono text-xs break-all">
                    {item.digest ? item.digest.slice(0, 19) + "…" : "—"}
                  </TableCell>
                  <TableCell className={`font-mono text-xs ${TABLE_NOWRAP}`}>
                    {formatBytes(item.size)}
                  </TableCell>
                  <TableActionsCell>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => void openDetail(item)}
                      actions={[
                        {
                          key: "sync",
                          label: t("registry.syncTitle"),
                          onSelect: () => {
                            setSyncSource({ repository, tag: item.tag })
                            setSyncOpen(true)
                          },
                        },
                        ...(session.can("registry_image:write")
                          ? [{
                              key: "copy",
                              label: t("registry.copyTag"),
                              onSelect: () => openCopy(item),
                            }]
                          : []),
                        ...(session.can("registry_image:delete")
                          ? [{
                              key: "delete",
                              label: t("action.delete"),
                              destructive: true,
                              onSelect: () => {
                                setDeleteError("")
                                setDeleting(item)
                              },
                            }]
                          : []),
                      ]}
                    />
                  </TableActionsCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : null}
      </ListFrame>
      <DetailSheet
        open={detailOpen}
        onOpenChange={setDetailOpen}
        title={detail ? `${repository}:${detail.tag}` : repository}
        description={detailLoading ? t("state.loading") : undefined}
        sections={detailSections}
        wide
      />
      <Dialog open={copying !== null} onOpenChange={(open) => { if (!open) setCopying(null) }}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("registry.copyTagTitle")}</DialogTitle>
          </DialogHeader>
          <DialogBody className="grid gap-4">
            <p className="text-sm text-muted-foreground">{t("registry.copyTagDescription")}</p>
            <FormField label={t("registry.targetRepository")} htmlFor="copy-target-repo" required description={t("registry.copyTargetRepositoryHelp")}>
              <Input id="copy-target-repo" className="font-mono" value={copyTargetRepository} placeholder={t("registry.targetRepositoryPlaceholder")} onChange={(event) => setCopyTargetRepository(event.target.value)} />
            </FormField>
            <FormField label={t("registry.targetTag")} htmlFor="copy-target-tag" required description={t("registry.copyTargetTagHelp")}>
              <Input id="copy-target-tag" className="font-mono" value={copyTargetTag} placeholder={t("registry.targetTagPlaceholder")} onChange={(event) => setCopyTargetTag(event.target.value)} />
            </FormField>
            {copyError ? <p className="text-sm text-destructive">{copyError}</p> : null}
          </DialogBody>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setCopying(null)}>{t("action.cancel")}</Button>
            <Button disabled={copyPending} onClick={() => void submitCopy()}>{t("action.save")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
      <SyncDialog
        open={syncOpen}
        onOpenChange={setSyncOpen}
        registryUid={uid}
        sourceRepository={syncSource?.repository ?? repository}
        sourceTag={syncSource?.tag ?? ""}
        onCreated={() => {
          setSyncOpen(false)
          toast.success(t("registry.syncCreated"), { description: t("registry.syncCreatedHint") })
        }}
      />
      <ConfirmAction
        open={deleting !== null}
        title={t("registry.deleteTagTitle")}
        description={t("registry.deleteTagDescription")}
        confirmLabel={t("action.delete")}
        cancelLabel={t("action.cancel")}
        pending={deletePending}
        error={deleteError}
        onOpenChange={(open) => { if (!open) setDeleting(null) }}
        onConfirm={() => void handleDelete()}
      />
    </section>
  )
}
