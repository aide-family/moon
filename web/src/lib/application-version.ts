import type {
  ApplicationResource,
  ApplicationVersionAction,
  ApplicationVersionComponent,
  ApplicationVersionItem,
} from "@/api/application/types"

const WORKLOAD_KINDS = ["Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob", "Pod"]
const DEFAULT_COMPONENT = "default"

function resourceComponent(resource: ApplicationResource): string {
  return resource.component?.trim() || DEFAULT_COMPONENT
}

function compareResources(left: ApplicationResource, right: ApplicationResource): number {
  return `${left.kind}/${left.name}`.localeCompare(`${right.kind}/${right.name}`)
}

/**
 * 版本内的服务列表。服务端已给出 components 时直接使用；
 * 旧数据只有扁平 resources 时，按资源的 component 分组得到等价结构。
 */
export function buildVersionComponents(version: ApplicationVersionItem | null): ApplicationVersionComponent[] {
  if (!version) {
    return []
  }
  if (version.components && version.components.length > 0) {
    return [...version.components].sort((left, right) => left.sortOrder - right.sortOrder)
  }
  const groups = new Map<string, ApplicationResource[]>()
  for (const resource of version.resources ?? []) {
    const name = resourceComponent(resource)
    const list = groups.get(name) ?? []
    list.push(resource)
    groups.set(name, list)
  }
  return [...groups.entries()]
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([name, resources], index) => ({
      uid: name,
      name,
      displayName: "",
      sortOrder: index,
      resources: [...resources].sort(compareResources),
      resourceCount: resources.length,
    }))
}

/** 版本内全部资源：优先于服务分组读取，旧版本退化为扁平列表。 */
export function listVersionResources(version: ApplicationVersionItem | null): ApplicationResource[] {
  if (!version) {
    return []
  }
  if (version.components && version.components.length > 0) {
    return version.components.flatMap((component) => component.resources ?? [])
  }
  return version.resources ?? []
}

export function hasClusterScopedResource(version: ApplicationVersionItem | null): boolean {
  return listVersionResources(version).some((resource) => resource.clusterScoped || resource.scope === "CLUSTER")
}

/** 服务的主工作负载：优先工作负载类资源，否则取第一个资源。 */
export function primaryWorkload(component: ApplicationVersionComponent): ApplicationResource | undefined {
  const resources = component.resources ?? []
  return resources.find((resource) => WORKLOAD_KINDS.includes(resource.kind)) ?? resources[0]
}

/** 服务用到的镜像仓库标签：已接入仓库显示仓库名，外部镜像单独标记。 */
export function componentImageLabels(
  version: ApplicationVersionItem | null,
  componentName: string,
): { label: string; external: boolean }[] {
  const images = (version?.images ?? []).filter(
    (image) => (image.component || DEFAULT_COMPONENT) === componentName,
  )
  const seen = new Set<string>()
  const labels: { label: string; external: boolean }[] = []
  for (const image of images) {
    const external = image.external || image.source === "EXTERNAL"
    const label = external ? "external" : image.registryName || image.repository
    if (!label || seen.has(label)) {
      continue
    }
    seen.add(label)
    labels.push({ label, external })
  }
  return labels
}

/** 服务内引用的 Secret 键数量（每个 Secret 资源只暴露 key，不含值）。 */
export function componentSecretKeys(component: ApplicationVersionComponent): string[] {
  const keys = new Set<string>()
  for (const resource of component.resources ?? []) {
    for (const key of resource.secretKeys ?? []) {
      keys.add(key)
    }
  }
  return [...keys]
}

/** 版本是否允许某个动作：优先用服务端 allowedActions，缺失时按状态兜底。 */
export function canRunVersionAction(version: ApplicationVersionItem | null, action: ApplicationVersionAction): boolean {
  if (!version) {
    return false
  }
  const actions = version.allowedActions
  if (actions && actions.length > 0) {
    return actions.includes(action)
  }
  switch (action) {
    case "PUBLISH":
      return version.status === "DRAFT"
    case "DEPRECATE":
      return version.status === "PUBLISHED"
    case "COPY":
      return version.status !== "DRAFT"
    case "DELETE":
      return version.status === "DRAFT"
    case "RELEASE":
      return version.status === "PUBLISHED"
    default:
      return false
  }
}
