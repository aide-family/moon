import { Boxes, KeyRound, Layers } from "lucide-react"
import { useTranslation } from "react-i18next"

import type {
  ApplicationVersionComponent,
  ApplicationVersionItem,
} from "@/api/application/types"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  componentImageLabels,
  componentSecretKeys,
  primaryWorkload,
} from "@/lib/application-version"

interface ApplicationServiceCardProps {
  component: ApplicationVersionComponent
  version: ApplicationVersionItem
  editable: boolean
  onViewResources: () => void
  onEdit: () => void
  onDelete: () => void
}

/** 版本内的一个服务卡片：主工作负载、镜像仓库、资源/密钥数量与操作入口。 */
export function ApplicationServiceCard({
  component,
  version,
  editable,
  onViewResources,
  onEdit,
  onDelete,
}: ApplicationServiceCardProps) {
  const { t } = useTranslation()
  const workload = primaryWorkload(component)
  const images = componentImageLabels(version, component.name)
  const secrets = componentSecretKeys(component)
  const clusterScoped = (component.resources ?? []).some(
    (resource) => resource.clusterScoped || resource.scope === "CLUSTER",
  )
  const resources = component.resources ?? []

  return (
    <Card className="min-w-0">
      <CardHeader className="border-b border-border/60">
        <CardTitle className="flex min-w-0 flex-wrap items-center gap-2">
          <span className="min-w-0 truncate font-mono">{component.name}</span>
          {component.displayName ? (
            <span className="min-w-0 truncate text-sm font-normal text-muted-foreground">
              {component.displayName}
            </span>
          ) : null}
        </CardTitle>
      </CardHeader>
      <CardContent className="grid min-w-0 gap-3 text-sm">
        <div className="grid min-w-0 gap-1">
          <span className="text-xs text-muted-foreground">{t("appStore.services.primaryWorkload")}</span>
          {workload ? (
            <span className="flex min-w-0 flex-wrap items-center gap-2">
              <Badge variant="outline" className="font-mono">{workload.kind}</Badge>
              <span className="min-w-0 truncate font-mono text-xs">{workload.name || "—"}</span>
            </span>
          ) : (
            <span className="text-muted-foreground">{t("appStore.services.noWorkload")}</span>
          )}
        </div>
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
            <Layers className="size-3.5" aria-hidden />
            {t("appStore.services.images")}
          </span>
          {images.length === 0 ? (
            <span className="text-xs text-muted-foreground">—</span>
          ) : (
            images.map((image) => (
              <Badge
                key={`${image.external ? "external" : "registry"}-${image.label}`}
                variant={image.external ? "outline" : "secondary"}
                className="max-w-full truncate"
                title={image.label}
              >
                {image.external ? t("appStore.services.externalImage") : image.label}
              </Badge>
            ))
          )}
        </div>
        <div className="flex min-w-0 flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
          <span>{t("appStore.services.resourceCount", { count: component.resourceCount || resources.length })}</span>
          <span className="inline-flex items-center gap-1">
            <KeyRound className="size-3.5" aria-hidden />
            {t("appStore.services.secretCount", { count: secrets.length })}
          </span>
          <span className="inline-flex items-center gap-1">
            <Boxes className="size-3.5" aria-hidden />
            {t("appStore.services.componentHint")}
          </span>
        </div>
        {clusterScoped ? (
          <p className="rounded-md border border-warning-border bg-warning-soft px-2 py-1.5 text-xs text-warning">
            {t("appStore.services.clusterScoped")}
          </p>
        ) : null}
      </CardContent>
      <CardFooter className="flex flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" onClick={onViewResources}>
          {t("appStore.services.viewResources")}
        </Button>
        {editable ? (
          <>
            <Button variant="ghost" size="sm" onClick={onEdit}>
              {t("appStore.services.editService")}
            </Button>
            <Button variant="ghost" size="sm" className="text-destructive" onClick={onDelete}>
              {t("action.delete")}
            </Button>
          </>
        ) : null}
      </CardFooter>
    </Card>
  )
}
