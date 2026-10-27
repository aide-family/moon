import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react"

import type {
  AlertAnalysisFingerprints,
  AlertAnalysisLevel,
  AlertAnalysisRecovery,
  AlertAnalysisStatus,
  AlertAnalysisSummary,
  AlertAnalysisTrend,
} from "@/api/alert-event"

/**
 * 历史告警-统计分析的分块串行加载状态机。
 *
 * 每个图表块一次独立请求，按 ANALYSIS_BLOCK_ORDER 串行发送：前一个返回后才发
 * 下一个。每个块各自维护 idle / loading / ready / error，互不阻塞；某块失败时
 * 只标记该块，串行链继续跑后面的块。
 *
 * 条件变化（from/to/status/interval/聚焦窗口）通过每个块自己的 signature 表达：
 * 只有 signature 变化的块会重新请求。聚焦窗口不改变 trend 的 signature，所以
 * 点击时间桶时只重跑其余 5 个块；筛选条件变化时 6 个块的 signature 一起变化，
 * 整体重置后重新串行。enabled 为假（分析 tab 未激活）时完全不发请求；切回同一
 * 条件时已 ready/error 的块保持原状，不会重复请求。
 */

export type AnalysisBlockName =
  "summary" | "status" | "level" | "trend" | "fingerprints" | "recovery"

/** 串行顺序：先出最轻的 KPI，再按后端接口耗时递增。 */
export const ANALYSIS_BLOCK_ORDER: readonly AnalysisBlockName[] = [
  "summary",
  "status",
  "level",
  "trend",
  "fingerprints",
  "recovery",
]

export interface AnalysisRequests {
  summary: () => Promise<AlertAnalysisSummary>
  status: () => Promise<AlertAnalysisStatus>
  level: () => Promise<AlertAnalysisLevel>
  trend: () => Promise<AlertAnalysisTrend>
  fingerprints: () => Promise<AlertAnalysisFingerprints>
  recovery: () => Promise<AlertAnalysisRecovery>
}

/** 每个块自己的输入签名；signature 相同就代表这个块不需要重跑。 */
export type AnalysisSignatures = Record<AnalysisBlockName, string>

export type AnalysisBlockStatus = "idle" | "loading" | "ready" | "error"

export interface AnalysisBlockState<T> {
  /** 这份数据属于哪一组输入条件。 */
  signature: string
  status: AnalysisBlockStatus
  data: T | null
}

export interface AnalysisBlocks {
  summary: AnalysisBlockState<AlertAnalysisSummary>
  status: AnalysisBlockState<AlertAnalysisStatus>
  level: AnalysisBlockState<AlertAnalysisLevel>
  trend: AnalysisBlockState<AlertAnalysisTrend>
  fingerprints: AnalysisBlockState<AlertAnalysisFingerprints>
  recovery: AnalysisBlockState<AlertAnalysisRecovery>
}

export interface AlertAnalysisBlocks {
  /** 已按当前 signature 对齐的块状态；条件刚变化、请求还没发时是 idle。 */
  blocks: AnalysisBlocks
  /** 重试单个失败的块：只重发这一块，不会重跑其他块。 */
  retry: (block: AnalysisBlockName) => void
}

interface UseAlertAnalysisBlocksOptions {
  /** 分析 tab 是否激活；为假时一个请求都不发。 */
  enabled: boolean
  signatures: AnalysisSignatures
  requests: AnalysisRequests
}

function idleBlock<T>(): AnalysisBlockState<T> {
  return { signature: "", status: "idle", data: null }
}

function idleBlocks(): AnalysisBlocks {
  return {
    summary: idleBlock(),
    status: idleBlock(),
    level: idleBlock(),
    trend: idleBlock(),
    fingerprints: idleBlock(),
    recovery: idleBlock(),
  }
}

/** 已拿到当前条件下的最终结果（成功或失败）：切回同一条件时不再请求。 */
function isSettled(
  blocks: AnalysisBlocks,
  block: AnalysisBlockName,
  signature: string
): boolean {
  const state: AnalysisBlockState<unknown> = blocks[block]
  if (state.signature !== signature) {
    return false
  }
  return state.status === "ready" || state.status === "error"
}

/** 读取一个块的数据；没 ready 时返回 null，避免把上一组条件的数据画出来。 */
export function analysisBlockData<T>(block: AnalysisBlockState<T>): T | null {
  return block.status === "ready" ? block.data : null
}

/** 降级提示只看 complete 标记，与块的具体数据类型无关。 */
export function analysisBlockIncomplete(
  block: AnalysisBlockState<unknown>
): boolean {
  if (block.status !== "ready" || block.data === null) {
    return false
  }
  return (block.data as { complete?: boolean }).complete === false
}

/**
 * 运行时由调用点保证 data 与 name 对应；逐块泛型的联合类型在这里收窄不了，
 * 只能在这里断言一次，读取侧仍按 AnalysisBlocks 拿到精确类型。
 */
function withBlock(
  blocks: AnalysisBlocks,
  name: AnalysisBlockName,
  next: AnalysisBlockState<unknown>
): AnalysisBlocks {
  return { ...blocks, [name]: next } as AnalysisBlocks
}

export function useAlertAnalysisBlocks({
  enabled,
  signatures,
  requests,
}: UseAlertAnalysisBlocksOptions): AlertAnalysisBlocks {
  const [rawBlocks, setRawBlocks] = useState<AnalysisBlocks>(idleBlocks)
  const blocksRef = useRef(rawBlocks)
  const requestsRef = useRef(requests)
  const signaturesRef = useRef(signatures)
  const enabledRef = useRef(enabled)

  // 最新 props 同步放在 layout effect 里：串行链的续跑与重试要读最新请求/条件，
  // 但又不能在渲染期间写 ref。
  useLayoutEffect(() => {
    blocksRef.current = rawBlocks
    requestsRef.current = requests
    signaturesRef.current = signatures
    enabledRef.current = enabled
  })

  // 请求序号：条件变化、卸载、重试都会作废旧串行链与过期响应。
  const tokenRef = useRef(0)
  // 同块同条件的在途请求复用，避免 StrictMode 双次执行 effect 或重复触发时重发。
  const inflightRef = useRef(new Map<string, Promise<unknown>>())

  const fetchBlock = useCallback(
    async (token: number, block: AnalysisBlockName) => {
      const signature = signaturesRef.current[block]
      const key = `${block}|${signature}`
      let request: Promise<unknown> | undefined = inflightRef.current.get(key)
      if (!request) {
        request = Promise.resolve().then(
          () => requestsRef.current[block]() as Promise<unknown>
        )
        inflightRef.current.set(key, request)
        const clear = () => {
          if (inflightRef.current.get(key) === request) {
            inflightRef.current.delete(key)
          }
        }
        void request.then(clear, clear)
      }
      setRawBlocks((prev) =>
        withBlock(prev, block, { signature, status: "loading", data: null })
      )
      try {
        const data = await request
        if (tokenRef.current !== token) {
          return
        }
        setRawBlocks((prev) =>
          withBlock(prev, block, { signature, status: "ready", data })
        )
      } catch {
        if (tokenRef.current !== token) {
          return
        }
        setRawBlocks((prev) =>
          withBlock(prev, block, { signature, status: "error", data: null })
        )
      }
    },
    []
  )

  const runChain = useCallback(
    async (token: number, targets: AnalysisBlockName[]) => {
      for (const block of targets) {
        // 条件已经变化 / 组件卸载 / 被新的串行链取消：后面的块不再发请求。
        if (tokenRef.current !== token || !enabledRef.current) {
          return
        }
        await fetchBlock(token, block)
      }
    },
    [fetchBlock]
  )

  useEffect(() => {
    // 每次 enabled 或任一 signature 变化，都作废上一条串行链的后续请求。
    tokenRef.current += 1
    if (!enabled) {
      return
    }
    const token = tokenRef.current
    // blocksRef 是最近一次渲染的块状态：只补发还没拿到当前条件结果的块。
    const current = signaturesRef.current
    const targets = ANALYSIS_BLOCK_ORDER.filter(
      (block) => !isSettled(blocksRef.current, block, current[block])
    )
    if (targets.length > 0) {
      void runChain(token, targets)
    }
    return () => {
      tokenRef.current += 1
    }
  }, [
    enabled,
    signatures.summary,
    signatures.status,
    signatures.level,
    signatures.trend,
    signatures.fingerprints,
    signatures.recovery,
    runChain,
  ])

  const retry = useCallback(
    (block: AnalysisBlockName) => {
      if (!enabledRef.current) {
        return
      }
      // 只重发这一块：不取消正在跑的串行链（否则链上还没完成的块会被一并重发），
      // 也不动其他块；同块同条件的在途请求会被 fetchBlock 复用。
      void fetchBlock(tokenRef.current, block)
    },
    [fetchBlock]
  )

  // 条件刚变化、effect 还没跑时，按 signature 返回 idle，避免把上一组条件的
  // 数据画在新筛选下面。
  const blocks = useMemo(() => {
    const aligned: AnalysisBlocks = {
      summary: idleBlock(),
      status: idleBlock(),
      level: idleBlock(),
      trend: idleBlock(),
      fingerprints: idleBlock(),
      recovery: idleBlock(),
    }
    const align = (name: AnalysisBlockName) => {
      const state = rawBlocks[name]
      if (state.signature === signatures[name]) {
        ;(aligned as Record<AnalysisBlockName, AnalysisBlockState<unknown>>)[
          name
        ] = state
      }
    }
    for (const name of ANALYSIS_BLOCK_ORDER) {
      align(name)
    }
    return aligned
  }, [rawBlocks, signatures])

  return { blocks, retry }
}
