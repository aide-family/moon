import { useTranslation } from "react-i18next"

import type { DatasourceItem } from "@/api/datasource/types"
import { BrandIcon } from "@/components/brand-icon"
import { DetailSheet } from "@/components/detail-sheet"
import { LabelTags } from "@/components/label-tags"
import { StatusBadge } from "@/components/status-badge"

interface DatasourceDetailSheetProps {
  item: DatasourceItem | null
  onOpenChange: (open: boolean) => void
}

/**
 * 数据源详情抽屉：列表页不再保留右侧常驻详情，改为从操作栏的「详情」进入。
 * 内容与原来的右侧详情面板一致；通知组名称取自列表响应，只用于展示。
 */
export function DatasourceDetailSheet({
  item,
  onOpenChange,
}: DatasourceDetailSheetProps) {
  const { t } = useTranslation()
  // 空数组表示未绑定通知组；启用可用性告警却未绑定时要提示事件无人接收。
  const groupNames = item?.notificationGroupNames ?? []

  return (
    <DetailSheet
      open={item !== null}
      title={item?.name ?? ""}
      size="40vw"
      onOpenChange={onOpenChange}
    >
      {item ? (
        <div className="grid min-w-0 gap-4">
          <div className="grid gap-2">
            <p className="flex min-w-0 items-center gap-2 text-sm">
              <BrandIcon
                driver={item.driver}
                className="size-5 shrink-0 rounded-md border bg-background p-0.5"
              />
              <span className="min-w-0 break-all text-muted-foreground">
                {item.endpoint}
              </span>
            </p>
            <div className="flex flex-wrap items-center gap-1.5">
              <StatusBadge value={item.type} />
              <StatusBadge value={item.driver} />
              <StatusBadge value={item.status} />
            </div>
          </div>
          <dl className="grid min-w-0 gap-4 @md:grid-cols-2">
            <div className="grid min-w-0 gap-1">
              <dt className="text-sm text-muted-foreground">
                {t("datasource.level")}
              </dt>
              <dd className="min-w-0 break-words">
                {item.levelName || item.levelUid || "—"}
              </dd>
            </div>
            <div className="grid min-w-0 gap-1">
              <dt className="text-sm text-muted-foreground">
                {t("datasource.secretSet")}
              </dt>
              <dd>
                {item.secretSet ? t("datasource.yes") : t("datasource.no")}
              </dd>
            </div>
            <div className="grid min-w-0 gap-1 @md:col-span-2">
              <dt className="text-sm text-muted-foreground">
                {t("datasource.labels")}
              </dt>
              <dd className="min-w-0">
                <LabelTags labels={item.labels} />
              </dd>
            </div>
            <div className="grid min-w-0 gap-1 @md:col-span-2">
              <dt className="text-sm text-muted-foreground">
                {t("datasource.remark")}
              </dt>
              <dd className="min-w-0 break-words whitespace-pre-wrap">
                {item.remark || "—"}
              </dd>
            </div>
          </dl>
          <section className="grid min-w-0 gap-3 border-t pt-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-sm font-medium">
                {t("datasource.availabilityAlertTitle")}
              </h3>
              <span className="text-sm text-muted-foreground">
                {item.availabilityAlertEnabled
                  ? t("datasource.availabilityAlertOn")
                  : t("datasource.availabilityAlertOff")}
              </span>
            </div>
            <dl className="grid min-w-0 gap-4 @md:grid-cols-2">
              <div className="grid min-w-0 gap-1">
                <dt className="text-sm text-muted-foreground">
                  {t("datasource.failureWindow")}
                </dt>
                <dd className="tabular-nums">
                  {t("datasource.failureWindowValue", {
                    count: item.failureWindowMinutes,
                  })}
                </dd>
              </div>
              <div className="grid min-w-0 gap-1">
                <dt className="text-sm text-muted-foreground">
                  {t("datasource.failureThreshold")}
                </dt>
                <dd className="tabular-nums">
                  {t("datasource.failureThresholdValue", {
                    count: item.failureThreshold,
                  })}
                </dd>
              </div>
              <div className="grid min-w-0 gap-1 @md:col-span-2">
                <dt className="text-sm text-muted-foreground">
                  {t("datasource.notificationGroups")}
                </dt>
                <dd className="min-w-0 break-words">
                  {groupNames.length > 0
                    ? groupNames.join(", ")
                    : t("datasource.notificationGroupsUnbound")}
                </dd>
              </div>
            </dl>
            {item.availabilityAlertEnabled && groupNames.length === 0 ? (
              <p className="rounded-md border border-warning-border bg-warning-soft px-2.5 py-1.5 text-xs text-warning">
                {t("datasource.noNotificationGroupsHint")}
              </p>
            ) : null}
          </section>
        </div>
      ) : null}
    </DetailSheet>
  )
}
