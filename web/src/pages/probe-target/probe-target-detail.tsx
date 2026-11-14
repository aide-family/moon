import { ChevronDown, ChevronRight } from "lucide-react"
import { Fragment, useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import {
  getProbeTargetGroup,
  listProbeTargetResults,
  probeTargetGroupNow,
} from "@/api/probe-target"
import type {
  ProbeResultItem,
  ProbeTargetGroupItem,
} from "@/api/probe-target/types"
import { DetailSheet, type DetailSection } from "@/components/detail-sheet"
import { LabelTags } from "@/components/label-tags"
import { StatusBadge } from "@/components/status-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { formatDateTime } from "@/lib/date-time"
import { readFormError } from "@/lib/form-error"
import {
  TABLE_COLUMN_WIDTH,
  TABLE_MIN_WIDTH,
  TABLE_NOWRAP,
} from "@/lib/table-layout"

import { PROBE_HTTP, ProbeDetailSummary } from "./probe-target-shared"

interface TargetRecordState {
  items: ProbeResultItem[]
  total: number
  loading: boolean
}

// 抽屉内按 uid 取组详情：加载态用初始 state 表达，effect 里不做同步 setState。
type DetailState =
  | { status: "loading" }
  | { status: "ready"; group: ProbeTargetGroupItem }
  | { status: "failed" }

interface ProbeTargetDetailSheetProps {
  // 空字符串表示抽屉关闭；保留 URL 上的 uid，深链直接落到详情。
  uid: string
  // 列表行里的名称，详情返回前先占住抽屉标题，避免打开瞬间只剩页面名。
  name?: string
  // 组变更（探测、启停、编辑）后自增，触发抽屉内的组详情重新拉取。
  version: number
  canWrite: boolean
  onOpenChange: (open: boolean) => void
  // 探测成功：页面刷新列表并按需递增 version。
  onChanged: () => void
  // 页面级错误提示（与列表共用 usePagedList 的 fail）。
  onError: (message: string) => void
}

/**
 * 探测目标组详情抽屉：目标组概览、目标列表，以及按目标展开的探测记录。
 * 记录按目标缓存，只有第一次展开时才请求该目标的记录。
 */
export function ProbeTargetDetailSheet({
  uid,
  name = "",
  version,
  canWrite,
  onOpenChange,
  onChanged,
  onError,
}: ProbeTargetDetailSheetProps) {
  const { t } = useTranslation()
  const [state, setState] = useState<DetailState>({ status: "loading" })
  // 探测记录按目标展开：一行目标点开后只拉这个目标的记录，按目标缓存。
  // 页面用 key={uid} 挂载抽屉，切换目标组时展开状态与缓存随之清空。
  const [expandedTarget, setExpandedTarget] = useState("")
  const [targetRecords, setTargetRecords] = useState<
    Record<string, TargetRecordState>
  >({})
  const [note, setNote] = useState("")
  const [probing, setProbing] = useState(false)

  // 只在 uid / version 变化时重新取组详情；version 由页面在启停、编辑后递增。
  // 加载态由初始 state 给出，effect 里不做同步 setState。
  useEffect(() => {
    if (!uid) {
      return
    }
    let active = true
    getProbeTargetGroup(uid)
      .then((result) => {
        if (active) {
          setState({ status: "ready", group: result.group })
        }
      })
      .catch(() => {
        if (active) {
          setState({ status: "failed" })
        }
      })
    return () => {
      active = false
    }
  }, [uid, version])

  const group = state.status === "ready" ? state.group : null

  // loadTargetRecords 拉一个目标的探测记录；force 用于「刷新」按钮。
  async function loadTargetRecords(targetUid: string, force = false) {
    if (!uid) {
      return
    }
    if (
      !force &&
      targetRecords[targetUid] &&
      !targetRecords[targetUid].loading
    ) {
      return
    }
    setTargetRecords((prev) => ({
      ...prev,
      [targetUid]: {
        items: prev[targetUid]?.items ?? [],
        total: prev[targetUid]?.total ?? 0,
        loading: true,
      },
    }))
    try {
      const result = await listProbeTargetResults(uid, {
        targetUid,
        page: 1,
        pageSize: 20,
      })
      setTargetRecords((prev) => ({
        ...prev,
        [targetUid]: {
          items: result.items ?? [],
          total: result.total ?? 0,
          loading: false,
        },
      }))
    } catch {
      setTargetRecords((prev) => ({
        ...prev,
        [targetUid]: { items: [], total: 0, loading: false },
      }))
    }
  }

  // toggleTargetRecords 展开当前目标的记录行，第一次展开时拉数据。
  function toggleTargetRecords(targetUid: string) {
    if (expandedTarget === targetUid) {
      setExpandedTarget("")
      return
    }
    setExpandedTarget(targetUid)
    void loadTargetRecords(targetUid)
  }

  async function probeNow() {
    if (!group) {
      return
    }
    setProbing(true)
    setNote("")
    try {
      const result = await probeTargetGroupNow(group.uid)
      setNote(
        t("probeTarget.probeNowDone", { count: result.results?.length ?? 0 })
      )
      onChanged()
    } catch (err) {
      const failure = readFormError(err, t("state.failed"))
      if (failure.form) {
        onError(failure.form)
      }
    } finally {
      setProbing(false)
    }
  }

  const targets = group?.targets ?? []
  // HTTP 专属列（状态码、证书剩余天数）只对 HTTP 目标组展示。
  const isHttp = group?.driver === PROBE_HTTP

  // 组详情非空时才构造分组；用非空局部变量，闭包里不再依赖外层收窄。
  let sections: DetailSection[] = []
  if (group) {
    const current = group
    sections = [
      {
        title: t("probeTarget.basicInfo"),
        fields: [
          {
            label: t("probeTarget.type"),
            value: <StatusBadge value={current.driver} />,
          },
          {
            label: t("probeTarget.targetCountLabel"),
            value: (
              <span className="tabular-nums">
                {t("probeTarget.targetCountValue", {
                  total: current.targetCount,
                  active: current.activeTargetCount,
                })}
              </span>
            ),
          },
          {
            label: t("probeTarget.executor"),
            value: (
              <span className="grid gap-1">
                <span>{t("probeTarget.executorCluster")}</span>
                <span className="text-xs text-muted-foreground">
                  {t("probeTarget.executorHint")}
                </span>
              </span>
            ),
          },
          {
            label: t("probeTarget.lastProbedAt"),
            value: (
              <span className="tabular-nums">
                {current.lastProbedAt
                  ? formatDateTime(current.lastProbedAt)
                  : t("probeTarget.neverProbed")}
              </span>
            ),
          },
          {
            label: t("probeTarget.labels"),
            value: <LabelTags labels={current.labels} />,
            span: 2,
          },
          {
            label: t("probeTarget.remark"),
            value: current.remark,
            span: 2,
          },
        ],
      },
      {
        title: t("probeTarget.targetsSection"),
        content: (
          <div className="flex max-h-[24rem] min-h-0 flex-col">
            <Table className={TABLE_MIN_WIDTH.default}>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-10" />
                  <TableHead>{t("probeTarget.targetUrl")}</TableHead>
                  <TableHead className={TABLE_COLUMN_WIDTH.status}>
                    {t("namespace.status")}
                  </TableHead>
                  <TableHead className={TABLE_COLUMN_WIDTH.status}>
                    {t("probeTarget.lastResult")}
                  </TableHead>
                  {isHttp ? (
                    <TableHead className={TABLE_COLUMN_WIDTH.type}>
                      {t("probeTarget.statusCode")}
                    </TableHead>
                  ) : (
                    <TableHead className="min-w-[10rem]">
                      {t("probeTarget.detailColumn")}
                    </TableHead>
                  )}
                  <TableHead className={TABLE_COLUMN_WIDTH.type}>
                    {t("probeTarget.duration")}
                  </TableHead>
                  {isHttp ? (
                    <TableHead className={TABLE_COLUMN_WIDTH.type}>
                      {t("probeTarget.certDays")}
                    </TableHead>
                  ) : null}
                </TableRow>
              </TableHeader>
              <TableBody>
                {targets.map((target) => {
                  const targetKey = target.uid || target.spec
                  const expanded = expandedTarget === targetKey
                  const records = targetRecords[targetKey]
                  return (
                    <Fragment key={targetKey}>
                      <TableRow>
                        <TableCell className={TABLE_NOWRAP}>
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon-sm"
                            aria-expanded={expanded}
                            aria-controls={`probe-records-${targetKey}`}
                            aria-label={
                              expanded
                                ? t("probeTarget.recordsCollapse")
                                : t("probeTarget.recordsExpand")
                            }
                            onClick={() => toggleTargetRecords(targetKey)}
                          >
                            {expanded ? <ChevronDown /> : <ChevronRight />}
                          </Button>
                        </TableCell>
                        <TableCell className="min-w-0 break-all">
                          {target.spec}
                        </TableCell>
                        <TableCell className={TABLE_NOWRAP}>
                          <StatusBadge value={target.status} />
                        </TableCell>
                        <TableCell className={TABLE_NOWRAP}>
                          {target.lastResult ? (
                            <StatusBadge
                              value={target.lastResult.ok ? "OK" : "FAILED"}
                              label={
                                target.lastResult.ok
                                  ? t("probeTarget.ok")
                                  : t("probeTarget.failed")
                              }
                            />
                          ) : (
                            "—"
                          )}
                        </TableCell>
                        {isHttp ? (
                          <TableCell className={TABLE_NOWRAP}>
                            {target.lastResult?.statusCode || "—"}
                          </TableCell>
                        ) : (
                          <TableCell className="max-w-[16rem] min-w-0 align-top">
                            <ProbeDetailSummary
                              driver={current.driver}
                              result={target.lastResult}
                            />
                          </TableCell>
                        )}
                        <TableCell className={TABLE_NOWRAP}>
                          {target.lastResult
                            ? t("probeTarget.durationValue", {
                                ms: target.lastResult.durationMs,
                              })
                            : "—"}
                        </TableCell>
                        {isHttp ? (
                          <TableCell className={TABLE_NOWRAP}>
                            {target.lastResult?.certDays
                              ? t("probeTarget.certDaysValue", {
                                  days: target.lastResult.certDays,
                                })
                              : "—"}
                          </TableCell>
                        ) : null}
                      </TableRow>
                      {expanded ? (
                        <TableRow
                          id={`probe-records-${targetKey}`}
                          className="bg-muted/20 hover:bg-muted/20"
                        >
                          <TableCell
                            colSpan={isHttp ? 7 : 6}
                            className="min-w-0"
                          >
                            <div className="grid min-w-0 gap-2 py-1">
                              <div className="flex flex-wrap items-center justify-between gap-2">
                                <span className="text-xs text-muted-foreground">
                                  {t("probeTarget.recordsTotal", {
                                    count: records?.total ?? 0,
                                  })}
                                </span>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  disabled={records?.loading}
                                  onClick={() =>
                                    void loadTargetRecords(targetKey, true)
                                  }
                                >
                                  {t("action.refresh")}
                                </Button>
                              </div>
                              {records?.loading && !records.items.length ? (
                                <p className="text-sm text-muted-foreground">
                                  {t("state.loading")}
                                </p>
                              ) : null}
                              {!records?.loading &&
                              (records?.items.length ?? 0) === 0 ? (
                                <p className="text-sm text-muted-foreground">
                                  {t("probeTarget.recordsEmpty")}
                                </p>
                              ) : null}
                              {(records?.items.length ?? 0) > 0 ? (
                                <div className="flex max-h-64 min-h-0 flex-col">
                                  <Table className={TABLE_MIN_WIDTH.wide}>
                                    <TableHeader>
                                      <TableRow>
                                        <TableHead
                                          className={TABLE_COLUMN_WIDTH.time}
                                        >
                                          {t("probeTarget.resultTime")}
                                        </TableHead>
                                        <TableHead>
                                          {t("probeTarget.rule")}
                                        </TableHead>
                                        <TableHead
                                          className={TABLE_COLUMN_WIDTH.status}
                                        >
                                          {t("probeTarget.resultStatus")}
                                        </TableHead>
                                        {isHttp ? (
                                          <TableHead
                                            className={TABLE_COLUMN_WIDTH.type}
                                          >
                                            {t("probeTarget.statusCode")}
                                          </TableHead>
                                        ) : (
                                          <TableHead className="min-w-[10rem]">
                                            {t("probeTarget.detailColumn")}
                                          </TableHead>
                                        )}
                                        <TableHead
                                          className={TABLE_COLUMN_WIDTH.type}
                                        >
                                          {t("probeTarget.duration")}
                                        </TableHead>
                                        <TableHead>
                                          {t("probeTarget.errorMessage")}
                                        </TableHead>
                                      </TableRow>
                                    </TableHeader>
                                    <TableBody>
                                      {records?.items.map((item) => (
                                        <TableRow
                                          key={`${item.targetUid}-${item.slot || item.probedAt}`}
                                        >
                                          <TableCell
                                            className={`${TABLE_COLUMN_WIDTH.time} ${TABLE_NOWRAP} tabular-nums`}
                                          >
                                            {formatDateTime(
                                              item.probedAt || item.slot
                                            )}
                                          </TableCell>
                                          <TableCell className="min-w-0 break-words">
                                            {item.ruleName ||
                                              (item.ruleUid
                                                ? "—"
                                                : t("probeTarget.manualProbe"))}
                                          </TableCell>
                                          <TableCell className={TABLE_NOWRAP}>
                                            <StatusBadge
                                              value={item.ok ? "OK" : "FAILED"}
                                              label={
                                                item.ok
                                                  ? t("probeTarget.ok")
                                                  : t("probeTarget.failed")
                                              }
                                            />
                                          </TableCell>
                                          {isHttp ? (
                                            <TableCell
                                              className={TABLE_NOWRAP}
                                            >
                                              {item.statusCode || "—"}
                                            </TableCell>
                                          ) : (
                                            <TableCell className="max-w-[16rem] min-w-0 align-top">
                                              <ProbeDetailSummary
                                                driver={current.driver}
                                                result={item}
                                              />
                                            </TableCell>
                                          )}
                                          <TableCell className={TABLE_NOWRAP}>
                                            {t("probeTarget.durationValue", {
                                              ms: item.durationMs,
                                            })}
                                          </TableCell>
                                          <TableCell className="min-w-0 break-words">
                                            {item.ok
                                              ? item.errorMessage || "—"
                                              : `${item.errorKind ? `${item.errorKind}: ` : ""}${item.errorMessage || "—"}`}
                                          </TableCell>
                                        </TableRow>
                                      ))}
                                    </TableBody>
                                  </Table>
                                </div>
                              ) : null}
                            </div>
                          </TableCell>
                        </TableRow>
                      ) : null}
                    </Fragment>
                  )
                })}
              </TableBody>
            </Table>
          </div>
        ),
      },
    ]
  }

  return (
    <DetailSheet
      open={uid !== ""}
      size="40vw"
      title={group?.name || name || t("probeTarget.title")}
      description={
        group ? (
          <>
            <StatusBadge value={group.status} />
            <Badge variant="secondary">
              {t("probeTarget.targetCountBadge", { count: group.targetCount })}
            </Badge>
          </>
        ) : undefined
      }
      sections={sections}
      onOpenChange={onOpenChange}
      footer={
        group && canWrite ? (
          <div className="flex flex-wrap items-center justify-end gap-2">
            {note ? (
              <p className="mr-auto text-sm text-muted-foreground">{note}</p>
            ) : null}
            <Button
              disabled={probing || group.activeTargetCount === 0}
              onClick={() => void probeNow()}
            >
              {probing ? t("probeTarget.probing") : t("probeTarget.probeNow")}
            </Button>
          </div>
        ) : undefined
      }
    >
      {group ? null : (
        <p
          className={
            state.status === "failed"
              ? "text-sm text-destructive"
              : "text-sm text-muted-foreground"
          }
        >
          {state.status === "failed"
            ? t("state.failed")
            : t("state.loading")}
        </p>
      )}
    </DetailSheet>
  )
}
