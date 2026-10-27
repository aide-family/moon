import { useSearchParams } from "react-router-dom"

import {
  isClusterSection,
  WORKLOAD_KINDS,
  type ClusterSection,
} from "./sections"

export interface ClusterDetailPatch {
  section?: ClusterSection
  /** 仅 workloads 生效：Deployment / StatefulSet / DaemonSet；空串表示全部。 */
  kind?: string
  namespace?: string
  name?: string
  page?: number
  pageSize?: number
}

function boundedInt(
  value: string | null,
  fallback: number,
  min: number,
  max: number
): number {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < min || parsed > max) {
    return fallback
  }
  return parsed
}

/**
 * 详情页的 URL 状态：tab / kind / namespace / name / page / pageSize。
 * 用 tab + kind 表达资源视图，保留旧链接（?tab=workloads&kind=Deployment）的语义：
 * 没有 tab 视为集群概况（从列表进入详情时的默认视图），kind 只对工作负载生效。
 */
export function useClusterDetailQuery() {
  const [params, setParams] = useSearchParams()
  const rawTab = params.get("tab")
  const section: ClusterSection = isClusterSection(rawTab) ? rawTab : "overview"
  const rawKind = params.get("kind") ?? ""
  const kind =
    section === "workloads" &&
    (WORKLOAD_KINDS as readonly string[]).includes(rawKind)
      ? rawKind
      : ""
  const namespace = params.get("namespace") ?? ""
  const name = params.get("name") ?? ""
  const page = boundedInt(params.get("page"), 1, 1, 1_000_000)
  const pageSize = boundedInt(params.get("pageSize"), 20, 1, 100)

  function update(patch: ClusterDetailPatch) {
    const next = new URLSearchParams(params)
    const setText = (key: string, value: string) => {
      if (value) {
        next.set(key, value)
      } else {
        next.delete(key)
      }
    }
    let resetPage = false

    if (patch.section !== undefined) {
      // 概况是默认视图，URL 里不写 tab；其他视图必须显式写入，否则会被读成概况。
      if (patch.section === "overview") {
        next.delete("tab")
      } else {
        next.set("tab", patch.section)
      }
      // 名称和类型只对当前资源视图有意义，切换视图时清空，避免条件被静默带走。
      next.delete("name")
      if (patch.kind === undefined) {
        next.delete("kind")
      }
      resetPage = true
    }

    if (patch.kind !== undefined) {
      const nextSection = patch.section ?? section
      if (nextSection === "workloads") {
        setText("kind", patch.kind)
      } else {
        next.delete("kind")
      }
      resetPage = true
    }

    if (patch.namespace !== undefined) {
      setText("namespace", patch.namespace)
      resetPage = true
    }
    if (patch.name !== undefined) {
      setText("name", patch.name.trim())
      resetPage = true
    }
    if (patch.page !== undefined) {
      next.set("page", String(Math.max(1, patch.page)))
    } else if (resetPage) {
      next.set("page", "1")
    }
    if (patch.pageSize !== undefined) {
      next.set("pageSize", String(patch.pageSize))
    }
    setParams(next)
  }

  return { section, kind, namespace, name, page, pageSize, update }
}
