import type { ReactNode } from "react"
import { useTranslation } from "react-i18next"

import type { KubernetesResourceKind } from "@/api/cluster"
import {
  DetailSheet,
  DetailSectionList,
  type DetailSection,
  type DetailSheetProps,
} from "@/components/detail-sheet"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

import { ClusterYamlPreview } from "./yaml-preview"

export interface ClusterDetailYamlTarget {
  clusterUid: string
  kind: KubernetesResourceKind
  /** 集群级资源（Node、PersistentVolume）不传。 */
  namespace?: string
  name: string
}

interface ClusterDetailSheetProps extends Omit<
  DetailSheetProps,
  "children" | "sections" | "fields"
> {
  /** 「详细信息」Tab 的结构化字段。 */
  info: DetailSection[]
  /** 「YAML」Tab 的对象标识；为空时 YAML 区块显示占位。 */
  yaml: ClusterDetailYamlTarget | null
  /** 抽屉头部下方的附加内容（如 Secret 的只读说明）。 */
  children?: ReactNode
}

/**
 * 集群资源详情抽屉：用「详细信息 / YAML」两个 Tab 组织内容。
 *
 * 内容区不再整体滚动（contentMode="fill"）：Tab 列表固定在顶部，
 * 「详细信息」自己滚动，「YAML」铺满剩余高度、只在代码内容区滚动，
 * 这样长 YAML 不会把 Tab 顶出视口，也不会让外层抽屉变成第二个滚动容器。
 * YAML 按对象懒加载，切换资源时 Tabs 通过 key 重置回详细信息。
 */
export function ClusterDetailSheet({
  info,
  yaml,
  children,
  ...rest
}: ClusterDetailSheetProps) {
  const { t } = useTranslation()
  const key = yaml
    ? `${yaml.kind}:${yaml.namespace ?? ""}:${yaml.name}`
    : "empty"

  return (
    <DetailSheet {...rest} contentMode="fill">
      <Tabs
        key={key}
        defaultValue="info"
        className="flex min-h-0 flex-1 flex-col gap-3"
      >
        <div className="flex shrink-0 flex-wrap items-center gap-2 px-4 pt-3 sm:px-5">
          <TabsList className="w-fit shrink-0">
            <TabsTrigger value="info">{t("kube.detailInfo")}</TabsTrigger>
            <TabsTrigger value="yaml">{t("kube.yaml")}</TabsTrigger>
          </TabsList>
        </div>
        {children ? (
          <div className="shrink-0 px-4 sm:px-5">{children}</div>
        ) : null}
        <TabsContent
          value="info"
          className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pb-4 sm:px-5"
        >
          <DetailSectionList sections={info} />
        </TabsContent>
        <TabsContent
          value="yaml"
          className="flex min-h-0 flex-1 flex-col overflow-hidden px-4 pb-4 sm:px-5"
        >
          {yaml ? (
            <ClusterYamlPreview {...yaml} />
          ) : (
            <p className="text-sm text-muted-foreground">
              {t("kube.yamlUnavailable")}
            </p>
          )}
        </TabsContent>
      </Tabs>
    </DetailSheet>
  )
}
