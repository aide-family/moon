import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getApplicationVersionResource } from "@/api/application"
import type {
  ApplicationResource,
  ApplicationVersionImage,
} from "@/api/application/types"
import { CodeBlock } from "@/components/code-block"
import { DetailSheet } from "@/components/detail-sheet"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { enumLabel } from "@/lib/enum-label"
import { failureMessage } from "@/lib/form-error"

interface ApplicationResourceDetailSheetProps {
  open: boolean
  applicationUid: string
  versionUid: string
  resource: ApplicationResource | null
  images: ApplicationVersionImage[]
  editable: boolean
  onEdit: () => void
  onDelete: () => void
  onOpenChange: (open: boolean) => void
}

/** 资源详情抽屉：YAML 原文、镜像与密钥摘要，以及编辑/删除动作。 */
export function ApplicationResourceDetailSheet({
  open,
  applicationUid,
  versionUid,
  resource,
  images,
  editable,
  onEdit,
  onDelete,
  onOpenChange,
}: ApplicationResourceDetailSheetProps) {
  const { t } = useTranslation()
  const [detail, setDetail] = useState<ApplicationResource | null>(resource)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!open || !resource) {
      return
    }
    let active = true
    void getApplicationVersionResource(applicationUid, versionUid, resource.uid)
      .then((result) => {
        if (!active) return
        setError("")
        setDetail(result.resource)
      })
      .catch((err) => {
        if (active) setError(failureMessage(err, t("state.failed")))
      })
    return () => {
      active = false
    }
    // resource.uid 变化或抽屉重新打开时刷新详情。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, applicationUid, versionUid, resource?.uid])

  const current = detail ?? resource
  if (!current) {
    return <DetailSheet open={open} title={t("appStore.resources")} onOpenChange={onOpenChange} />
  }
  const relatedImages = images.filter(
    (image) => image.resourceUid === current.uid || image.resourceItem === current.item,
  )

  return (
    <DetailSheet
      open={open}
      size="40vw"
      title={`${current.kind} ${current.name}`}
      titleCopy={{ value: current.name, label: t("appStore.resourceName") }}
      description={
        <span className="flex flex-wrap items-center gap-2">
          <StatusBadge value={current.scope} label={enumLabel(t, current.scope)} />
          <span className="font-mono text-xs">{current.apiVersion}</span>
          <span className="font-mono text-xs">{current.component || t("appStore.services.defaultComponent")}</span>
        </span>
      }
      footer={
        editable ? (
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={onEdit}>{t("appStore.editResource")}</Button>
            <Button variant="destructive" onClick={onDelete}>{t("action.delete")}</Button>
          </div>
        ) : undefined
      }
      onOpenChange={onOpenChange}
    >
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <section className="grid min-w-0 gap-2">
        <h3 className="font-heading text-xs font-semibold tracking-wide">{t("appStore.resourceYaml")}</h3>
        <CodeBlock
          code={current.manifest}
          language="yaml"
          label="YAML"
          copyable
          copyLabel={t("action.copy")}
          copiedLabel={t("action.copied")}
          copyErrorLabel={t("action.copyFailed")}
          maxHeightClass="max-h-[46vh]"
        />
      </section>
      <section className="grid min-w-0 gap-2">
        <h3 className="font-heading text-xs font-semibold tracking-wide">{t("appStore.images")}</h3>
        {relatedImages.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("appStore.noImages")}</p>
        ) : (
          <ul className="grid gap-2">
            {relatedImages.map((image) => (
              <li key={`${image.containerPath}-${image.image}`} className="grid min-w-0 gap-1 rounded-lg border border-border/60 p-2">
                <div className="flex min-w-0 flex-wrap items-center gap-2">
                  <span className="min-w-0 font-mono text-xs break-all">{image.image}</span>
                  <StatusBadge
                    value={image.external || image.source === "EXTERNAL" ? "EXTERNAL" : "REGISTRY"}
                    label={image.external || image.source === "EXTERNAL" ? t("appStore.services.externalImage") : image.registryName || t("enum.REGISTRY")}
                  />
                </div>
                <p className="font-mono text-xs text-muted-foreground break-all">
                  {image.resolvedImage || image.image}
                </p>
                <p className="font-mono text-xs text-muted-foreground">
                  {image.digest || t("appStore.imagePicker.digestUnpinned")} · {image.platforms.join(" · ") || "—"}
                </p>
              </li>
            ))}
          </ul>
        )}
      </section>
      <section className="grid min-w-0 gap-2">
        <h3 className="font-heading text-xs font-semibold tracking-wide">{t("appStore.secretKeys")}</h3>
        {current.secretKeys.length === 0 ? (
          <p className="text-sm text-muted-foreground">—</p>
        ) : (
          <div className="flex flex-wrap gap-2">
            {current.secretKeys.map((key) => (
              <span key={key} className="rounded-md bg-muted px-2 py-1 font-mono text-xs">{key}</span>
            ))}
          </div>
        )}
        <p className="text-xs text-muted-foreground">{t("appStore.secretValueWriteOnly")}</p>
      </section>
    </DetailSheet>
  )
}
