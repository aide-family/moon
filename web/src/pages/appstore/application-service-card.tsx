import { Boxes, Info, KeyRound, Layers } from "lucide-react"
import ReactMarkdown from "react-markdown"
import { useTranslation } from "react-i18next"
import remarkGfm from "remark-gfm"

import type {
  ApplicationVersionComponent,
  ApplicationVersionItem,
} from "@/api/application/types"
import { HeaderActionsMenu } from "@/components/header-actions-menu"
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
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
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
  onAddResource: () => void
  onEdit: () => void
  onDelete: () => void
}

/** 版本内的一个服务卡片：主工作负载、镜像仓库、资源/密钥数量与操作入口。 */
export function ApplicationServiceCard({
  component,
  version,
  editable,
  onViewResources,
  onAddResource,
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
  const hasDescription = Boolean(component.description?.trim())

  return (
    <Card className="min-w-0">
      <CardHeader className="border-b border-border/60">
        <CardTitle className="flex min-w-0 flex-wrap items-center gap-2">
          <span className="min-w-0 truncate font-mono">{component.name}</span>
          {component.displayName && component.displayName !== component.name ? (
            <span className="min-w-0 truncate text-sm font-normal text-muted-foreground">
              {component.displayName}
            </span>
          ) : null}
          {hasDescription ? (
            <Popover>
              <PopoverTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  className="ml-auto size-6 shrink-0 text-muted-foreground"
                  aria-label={t("appStore.services.descriptionLabel")}
                >
                  <Info className="size-3.5" aria-hidden />
                </Button>
              </PopoverTrigger>
              <PopoverContent align="end" className="max-h-72 w-80 overflow-y-auto text-sm">
                <ServiceDescriptionMarkdown content={component.description} />
              </PopoverContent>
            </Popover>
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
          <HeaderActionsMenu
            label={t("action.more")}
            actions={[
              { key: "add-resource", label: t("appStore.addResource"), onSelect: onAddResource },
              { key: "edit-service", label: t("appStore.services.editService"), onSelect: onEdit },
              { key: "delete-service", label: t("action.delete"), destructive: true, onSelect: onDelete },
            ]}
          />
        ) : null}
      </CardFooter>
    </Card>
  )
}

/** 服务备注以 Markdown 渲染（不启用 raw HTML），与通知模板的 markdown 口径一致。 */
function ServiceDescriptionMarkdown({ content }: { content: string }) {
  return (
    <div className="grid min-w-0 gap-2 break-words [&_a]:text-primary-ink [&_a]:underline [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:font-mono [&_code]:text-xs [&_h1]:font-heading [&_h1]:text-base [&_h1]:font-semibold [&_h2]:font-heading [&_h2]:text-sm [&_h2]:font-semibold [&_h3]:font-semibold [&_li]:ml-4 [&_li]:list-disc [&_ol_li]:list-decimal [&_p]:m-0 [&_pre]:overflow-x-auto [&_pre]:rounded [&_pre]:bg-muted [&_pre]:p-2">
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
    </div>
  )
}
