import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  createImageSyncs,
  listRegistries,
  previewImageSync,
} from "@/api/registry"
import type {
  ImageSyncSourceKind,
  PreviewImageSyncResult,
  RegistryItem,
} from "@/api/registry/types"
import { FormField } from "@/components/form-field"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
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
import { formErrorMessage } from "@/lib/form-error"


interface SyncDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  registryUid: string
  sourceRepository: string
  sourceTag: string
  onCreated: () => void
}

/** 从已接入仓库或外部镜像创建一条同步任务。 */
export function SyncDialog({
  open,
  onOpenChange,
  registryUid,
  sourceRepository,
  sourceTag,
  onCreated,
}: SyncDialogProps) {
  const { t } = useTranslation()
  const [registries, setRegistries] = useState<RegistryItem[]>([])
  const [sourceKind, setSourceKind] = useState<ImageSyncSourceKind>("CONNECTION")
  const [sourceRegistryUid, setSourceRegistryUid] = useState(registryUid)
  const [sourceRepositoryValue, setSourceRepositoryValue] = useState(sourceRepository)
  const [sourceTagValue, setSourceTagValue] = useState(sourceTag)
  const [sourceImage, setSourceImage] = useState("")
  const [sourceUsername, setSourceUsername] = useState("")
  const [sourcePassword, setSourcePassword] = useState("")
  const [sourceInsecure, setSourceInsecure] = useState(false)
  const [targetRegistryUid, setTargetRegistryUid] = useState(registryUid)
  const [targetRepository, setTargetRepository] = useState(sourceRepository)
  const [targetTag, setTargetTag] = useState(sourceTag)
  const [platforms, setPlatforms] = useState("")
  const [overwrite, setOverwrite] = useState(false)
  const [preview, setPreview] = useState<PreviewImageSyncResult | null>(null)
  const [previewing, setPreviewing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!open) return
    let active = true
    void listRegistries({ page: 1, pageSize: 100 })
      .then((result) => {
        if (!active) return
        setRegistries(result.items)
        if (!targetRegistryUid && result.items[0]) {
          setTargetRegistryUid(result.items[0].uid)
        }
      })
      .catch(() => undefined)
    return () => {
      active = false
    }
  }, [open, targetRegistryUid])

  function buildSource() {
    if (sourceKind === "EXTERNAL") {
      return {
        sourceKind,
        sourceImage: sourceImage.trim(),
        sourceUsername: sourceUsername.trim(),
        sourcePassword,
        sourceAllowInsecureTls: sourceInsecure,
      }
    }
    return {
      sourceKind,
      sourceRegistryUid,
      sourceImage: `${sourceRepositoryValue.trim()}:${sourceTagValue.trim()}`,
    }
  }

  async function handlePreview() {
    setPreviewing(true)
    setError("")
    try {
      const result = await previewImageSync({
        source: {
          ...buildSource(),
          targetRegistryUid,
          targetRepository: targetRepository.trim(),
          targetTag: targetTag.trim(),
          ...(platforms.trim() ? { platforms: platforms.split(",").map((item) => item.trim()).filter(Boolean) } : {}),
          overwrite,
        },
        targetRegistryUid,
        targetRepository: targetRepository.trim(),
        targetTag: targetTag.trim(),
      })
      setPreview(result)
    } catch (err) {
      setError(formErrorMessage(err, t("state.failed")))
      setPreview(null)
    } finally {
      setPreviewing(false)
    }
  }

  async function handleCreate() {
    setSaving(true)
    setError("")
    try {
      const result = await createImageSyncs([
        {
          ...buildSource(),
          targetRegistryUid,
          targetRepository: targetRepository.trim(),
          targetTag: targetTag.trim(),
          ...(platforms.trim() ? { platforms: platforms.split(",").map((item) => item.trim()).filter(Boolean) } : {}),
          overwrite,
        },
      ])
      if (result.failed.length > 0) {
        setError(result.failed[0].message)
        return
      }
      onCreated()
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
          <DialogTitle>{t("registry.syncTitle")}</DialogTitle>
        </DialogHeader>
        <DialogBody className="grid gap-4 sm:grid-cols-2">
          <FormField label={t("registry.sourceKind")} htmlFor="sync-source-kind" required>
            <Select value={sourceKind} onValueChange={(value) => { setSourceKind(value as ImageSyncSourceKind); setPreview(null) }}>
              <SelectTrigger id="sync-source-kind"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="CONNECTION">{t("registry.sourceConnection")}</SelectItem>
                <SelectItem value="EXTERNAL">{t("registry.sourceExternal")}</SelectItem>
              </SelectContent>
            </Select>
          </FormField>
          <div className="flex items-end pb-1">
            {preview ? (
              <Badge variant={preview.targetExists ? "destructive" : "secondary"}>
                {preview.targetExists ? t("registry.targetExists") : t("registry.previewResult")}
              </Badge>
            ) : null}
          </div>
          {sourceKind === "CONNECTION" ? (
            <>
              <FormField label={t("registry.sourceRegistry")} htmlFor="sync-source-registry" required>
                <Select value={sourceRegistryUid} onValueChange={(value) => { setSourceRegistryUid(value); setPreview(null) }}>
                  <SelectTrigger id="sync-source-registry"><SelectValue placeholder={t("registry.selectRegistry")} /></SelectTrigger>
                  <SelectContent>
                    {registries.map((item) => <SelectItem key={item.uid} value={item.uid}>{item.name}</SelectItem>)}
                  </SelectContent>
                </Select>
              </FormField>
              <div />
              <FormField label={t("registry.repository")} htmlFor="sync-source-repo" required>
                <Input id="sync-source-repo" className="font-mono" value={sourceRepositoryValue} onChange={(event) => { setSourceRepositoryValue(event.target.value); setPreview(null) }} />
              </FormField>
              <FormField label={t("registry.tag")} htmlFor="sync-source-tag" required>
                <Input id="sync-source-tag" className="font-mono" value={sourceTagValue} onChange={(event) => { setSourceTagValue(event.target.value); setPreview(null) }} />
              </FormField>
            </>
          ) : (
            <>
              <FormField label={t("registry.sourceImage")} htmlFor="sync-source-image" required className="sm:col-span-full" description={t("registry.sourceImageHelp")}>
                <Input id="sync-source-image" className="font-mono" value={sourceImage} placeholder={t("registry.sourceImagePlaceholder")} onChange={(event) => { setSourceImage(event.target.value); setPreview(null) }} />
              </FormField>
              <FormField label={t("registry.sourceUsername")} htmlFor="sync-source-username">
                <Input id="sync-source-username" value={sourceUsername} onChange={(event) => setSourceUsername(event.target.value)} />
              </FormField>
              <FormField label={t("registry.sourcePassword")} htmlFor="sync-source-password">
                <Input id="sync-source-password" type="password" value={sourcePassword} onChange={(event) => setSourcePassword(event.target.value)} />
              </FormField>
              <div className="flex items-center gap-2 sm:col-span-full">
                <Checkbox id="sync-source-insecure" checked={sourceInsecure} onCheckedChange={(value) => setSourceInsecure(value === true)} />
                <label htmlFor="sync-source-insecure" className="text-sm">{t("registry.sourceInsecureTls")}</label>
              </div>
            </>
          )}
          <FormField label={t("registry.targetRegistry")} htmlFor="sync-target-registry" required>
            <Select value={targetRegistryUid} onValueChange={(value) => { setTargetRegistryUid(value); setPreview(null) }}>
              <SelectTrigger id="sync-target-registry"><SelectValue placeholder={t("registry.selectRegistry")} /></SelectTrigger>
              <SelectContent>
                {registries.map((item) => <SelectItem key={item.uid} value={item.uid}>{item.name}</SelectItem>)}
              </SelectContent>
            </Select>
          </FormField>
          <div />
          <FormField label={t("registry.targetRepository")} htmlFor="sync-target-repo" required description={t("registry.targetRepositoryHelp")}>
            <Input id="sync-target-repo" className="font-mono" value={targetRepository} onChange={(event) => setTargetRepository(event.target.value)} />
          </FormField>
          <FormField label={t("registry.targetTag")} htmlFor="sync-target-tag" required description={t("registry.targetTagHelp")}>
            <Input id="sync-target-tag" className="font-mono" value={targetTag} onChange={(event) => setTargetTag(event.target.value)} />
          </FormField>
          <FormField label={t("registry.platforms")} htmlFor="sync-platforms" description={t("registry.platformsHelp")}>
            <Input id="sync-platforms" className="font-mono" value={platforms} placeholder="linux/amd64, linux/arm64/v8" onChange={(event) => setPlatforms(event.target.value)} />
          </FormField>
          <div className="flex items-center gap-2 pt-6">
            <Checkbox id="sync-overwrite" checked={overwrite} onCheckedChange={(value) => setOverwrite(value === true)} />
            <label htmlFor="sync-overwrite" className="text-sm">{t("registry.overwrite")}</label>
          </div>
          {preview ? (
            <div className="grid gap-1 rounded-lg border border-border/60 bg-muted/30 p-3 text-sm sm:col-span-full">
              <span className="font-mono text-xs break-all">{preview.sourceImage}</span>
              <span className="text-xs text-muted-foreground">{preview.sourceDigest}</span>
              <span className="text-xs text-muted-foreground">{preview.sourcePlatforms.join(" · ") || t("registry.allPlatforms")}</span>
            </div>
          ) : null}
          {error ? <p className="text-sm text-destructive sm:col-span-full">{error}</p> : null}
        </DialogBody>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>{t("action.cancel")}</Button>
          <Button variant="outline" disabled={previewing} onClick={() => void handlePreview()}>
            {previewing ? t("state.loading") : t("registry.preview")}
          </Button>
          <Button disabled={saving} onClick={() => void handleCreate()}>{t("registry.createSync")}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
