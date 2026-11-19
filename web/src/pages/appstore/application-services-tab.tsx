import { useState } from "react"
import { useTranslation } from "react-i18next"

import type {
  ApplicationResource,
  ApplicationVersionComponent,
  ApplicationVersionItem,
} from "@/api/application/types"
import { DetailSheet } from "@/components/detail-sheet"
import { RowActions } from "@/components/row-actions"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"

import { ApplicationServiceCard } from "./application-service-card"

interface ApplicationServicesTabProps {
  version: ApplicationVersionItem
  components: ApplicationVersionComponent[]
  editable: boolean
  readOnlyHint?: string
  onCreateService: () => void
  onEditService: (component: ApplicationVersionComponent) => void
  onDeleteService: (component: ApplicationVersionComponent) => void
  onAddResource: (component: ApplicationVersionComponent) => void
  onOpenResource: (resource: ApplicationResource) => void
  onEditResource: (resource: ApplicationResource) => void
  onDeleteResource: (resource: ApplicationResource) => void
}

/**
 * 版本内容默认 Tab：把版本按服务（component）呈现成卡片网格。
 * 服务是独立实体（可先建空服务），资源在服务卡片「更多 → 添加资源」里维护。
 */
export function ApplicationServicesTab({
  version,
  components,
  editable,
  readOnlyHint,
  onCreateService,
  onEditService,
  onDeleteService,
  onAddResource,
  onOpenResource,
  onEditResource,
  onDeleteResource,
}: ApplicationServicesTabProps) {
  const { t } = useTranslation()
  const [sheetComponent, setSheetComponent] = useState<ApplicationVersionComponent | null>(null)

  return (
    <div className="grid min-w-0 gap-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <h2 className="font-heading text-base font-semibold">{t("appStore.services.title")}</h2>
          <p className="text-sm text-muted-foreground">
            {t("appStore.services.summary", {
              components: components.length,
              resources: version.resourceCount ?? components.reduce((sum, item) => sum + (item.resourceCount || item.resources.length), 0),
            })}
          </p>
        </div>
        {editable && components.length > 0 ? (
          <Button variant="outline" size="sm" onClick={onCreateService}>
            {t("appStore.services.createService")}
          </Button>
        ) : null}
      </div>

      {!editable && readOnlyHint ? (
        <p className="rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-sm text-warning">
          {readOnlyHint}
        </p>
      ) : null}

      {components.length === 0 ? (
        <div className="grid gap-2 rounded-xl border border-dashed border-border p-6 text-center">
          <p className="font-medium">{t("appStore.services.emptyTitle")}</p>
          <p className="text-sm text-muted-foreground">{t("appStore.services.emptyHint")}</p>
          {editable ? (
            <div className="flex justify-center">
              <Button variant="outline" size="sm" onClick={onCreateService}>
                {t("appStore.services.createService")}
              </Button>
            </div>
          ) : null}
        </div>
      ) : (
        <div className="grid min-w-0 gap-4 [&>*]:shrink-0 sm:grid-cols-2">
          {components.map((component) => (
            <ApplicationServiceCard
              key={component.uid || component.name}
              component={component}
              version={version}
              editable={editable}
              onViewResources={() => setSheetComponent(component)}
              onAddResource={() => onAddResource(component)}
              onEdit={() => onEditService(component)}
              onDelete={() => onDeleteService(component)}
            />
          ))}
        </div>
      )}

      <DetailSheet
        open={sheetComponent !== null}
        size="40vw"
        title={sheetComponent ? sheetComponent.name : t("appStore.services.title")}
        description={
          sheetComponent ? (
            <span>{t("appStore.services.resourceCount", { count: sheetComponent.resourceCount || sheetComponent.resources.length })}</span>
          ) : undefined
        }
        footer={
          editable && sheetComponent ? (
            <div className="flex justify-end">
              <Button
                onClick={() => {
                  const target = sheetComponent
                  setSheetComponent(null)
                  onAddResource(target)
                }}
              >
                {t("appStore.services.addResourceToService")}
              </Button>
            </div>
          ) : undefined
        }
        onOpenChange={(open) => {
          if (!open) setSheetComponent(null)
        }}
      >
        {sheetComponent ? (
          <div className="grid min-w-0 gap-2">
            {sheetComponent.resources.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("appStore.emptyResources")}</p>
            ) : (
              <ul className="grid min-w-0 gap-2">
                {sheetComponent.resources.map((resource) => (
                  <li key={resource.uid || `${resource.kind}-${resource.name}`} className="flex min-w-0 items-center gap-2 rounded-lg border border-border/60 p-2">
                    <span className="grid min-w-0 flex-1 gap-0.5">
                      <span className="flex min-w-0 flex-wrap items-center gap-2">
                        <span className="font-mono text-xs">{resource.kind}</span>
                        <StatusBadge value={resource.scope} />
                      </span>
                      <span className="min-w-0 truncate font-mono text-xs text-muted-foreground">{resource.name}</span>
                    </span>
                    <RowActions
                      nowrap
                      detailLabel={t("action.detail")}
                      moreLabel={t("action.more")}
                      onDetail={() => {
                        setSheetComponent(null)
                        onOpenResource(resource)
                      }}
                      actions={editable ? [
                        { key: "edit", label: t("appStore.editResource"), onSelect: () => { setSheetComponent(null); onEditResource(resource) } },
                        { key: "delete", label: t("action.delete"), destructive: true, onSelect: () => { setSheetComponent(null); onDeleteResource(resource) } },
                      ] : []}
                    />
                  </li>
                ))}
              </ul>
            )}
          </div>
        ) : null}
      </DetailSheet>
    </div>
  )
}
