import { useState } from "react"
import { useTranslation } from "react-i18next"

import type {
  ApplicationResource,
  ApplicationVersionComponent,
  ApplicationVersionItem,
} from "@/api/application/types"
import { DetailSheet } from "@/components/detail-sheet"
import { Button } from "@/components/ui/button"
import { StatusBadge } from "@/components/status-badge"

import { ApplicationComponentDialog } from "./application-component-dialog"
import { ApplicationServiceCard } from "./application-service-card"

interface ApplicationServicesTabProps {
  version: ApplicationVersionItem
  components: ApplicationVersionComponent[]
  editable: boolean
  readOnlyHint?: string
  onAddResource: (component?: string) => void
  onEditService: (component: ApplicationVersionComponent) => void
  onDeleteService: (component: ApplicationVersionComponent) => void
  onOpenResource: (resource: ApplicationResource) => void
}

/**
 * 版本内容默认 Tab：把版本按服务（component）呈现成卡片网格。
 * 新增服务时先输入组件名，再在该服务下添加第一个资源；组件随资源存在。
 */
export function ApplicationServicesTab({
  version,
  components,
  editable,
  readOnlyHint,
  onAddResource,
  onEditService,
  onDeleteService,
  onOpenResource,
}: ApplicationServicesTabProps) {
  const { t } = useTranslation()
  const [sheetComponent, setSheetComponent] = useState<ApplicationVersionComponent | null>(null)
  const [createOpen, setCreateOpen] = useState(false)

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
        {editable ? (
          <div className="flex flex-wrap items-center gap-2">
            <Button variant="outline" size="sm" onClick={() => setCreateOpen(true)}>
              {t("appStore.services.createService")}
            </Button>
            <Button size="sm" onClick={() => onAddResource()}>
              {t("appStore.addResource")}
            </Button>
          </div>
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
            <div className="flex flex-wrap items-center justify-center gap-2">
              <Button variant="outline" size="sm" onClick={() => setCreateOpen(true)}>
                {t("appStore.services.createService")}
              </Button>
              <Button size="sm" onClick={() => onAddResource()}>
                {t("appStore.addResource")}
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
              onEdit={() => onEditService(component)}
              onDelete={() => onDeleteService(component)}
            />
          ))}
        </div>
      )}

      <ApplicationComponentDialog
        key={createOpen ? "create-open" : "create-closed"}
        open={createOpen}
        mode="create"
        onOpenChange={setCreateOpen}
        onSave={(name) => {
          setCreateOpen(false)
          onAddResource(name)
        }}
      />

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
                  const name = sheetComponent.name
                  setSheetComponent(null)
                  onAddResource(name)
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
                    <Button
                      variant="link"
                      size="sm"
                      className="h-auto shrink-0 px-0"
                      onClick={() => {
                        setSheetComponent(null)
                        onOpenResource(resource)
                      }}
                    >
                      {t("appStore.services.viewResource")}
                    </Button>
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
