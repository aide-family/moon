import { ArrowLeft, RefreshCw } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, Navigate, useParams, useSearchParams } from "react-router-dom"

import {
  getCluster,
  issueKubernetesPodExecTicket,
  kubernetesPodExecSocketURL,
  listKubernetesPods,
} from "@/api/cluster"
import { useSession } from "@/auth/session"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Field, FieldLabel } from "@/components/ui/field"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  XtermTerminal,
  type XtermSessionState,
} from "@/components/xterm-terminal"
import { formErrorMessage } from "@/lib/form-error"

// 服务端通过 {"type":"exit"} 返回的关闭原因是固定英文串；已知原因映射成 zh/en 文案，
// 未知原因保留原文，便于排查真实故障。
const POD_EXEC_REASON_KEYS: Record<string, string> = {
  exited: "kube.podExecExited",
  closed: "kube.podExecClosed",
  "pod or container not found": "kube.podExecTargetMissing",
  "cluster credential is not allowed to exec into this pod":
    "kube.podExecForbidden",
  "pod exec session ended": "kube.podExecFailed",
}

/**
 * Pod 终端：集群详情下的子路由。
 * URL 是 /clusters/{cluster_uid}/pods/{namespace}/{name}/exec?container={容器}，
 * 容器切换会重建会话，返回按钮回到 Pod 视图并保留命名空间。
 */
export function PodExecPage() {
  const { t } = useTranslation()
  const { uid = "", namespace = "", name = "" } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const session = useSession()
  const [clusterName, setClusterName] = useState("")
  const [containers, setContainers] = useState<string[]>([])
  const [podLoading, setPodLoading] = useState(true)
  const [podFailed, setPodFailed] = useState(false)
  const [podError, setPodError] = useState("")
  const [reloadKey, setReloadKey] = useState(0)
  const [attempt, setAttempt] = useState(0)
  // 会话状态带上它对应的 sessionKey：容器切换或重连时旧状态立即失效，不会闪一下上次的错误。
  const [execState, setExecState] = useState<{
    key: string
    value: XtermSessionState
  }>({ key: "", value: { state: "connecting" } })

  const canExec = session.can("k8s:pod:exec")
  const requestedContainer = searchParams.get("container") ?? ""
  const container = useMemo(() => {
    if (containers.length === 0) {
      return ""
    }
    return containers.includes(requestedContainer)
      ? requestedContainer
      : containers[0]
  }, [containers, requestedContainer])
  const sessionKey = `${container}|${attempt}`
  const state: XtermSessionState =
    execState.key === sessionKey ? execState.value : { state: "connecting" }

  useEffect(() => {
    if (!uid) {
      return
    }
    let active = true
    getCluster(uid)
      .then((result) => {
        if (active) {
          setClusterName(result.cluster.name)
        }
      })
      .catch(() => {
        // 集群名只用于页头展示，取不到就回落到 UID。
      })
    return () => {
      active = false
    }
  }, [uid])

  useEffect(() => {
    if (!uid || !namespace || !name) {
      return
    }
    let active = true
    async function loadPod() {
      setPodLoading(true)
      setPodFailed(false)
      setPodError("")
      try {
        const result = await listKubernetesPods(uid, {
          namespace,
          name,
          pageSize: 100,
        })
        if (!active) {
          return
        }
        const pod = (result.items ?? []).find((item) => item.name === name)
        if (!pod) {
          // 只有 k8s:pod:exec、没有 k8s:pod:read 的账号拿不到 Pod 列表；
          // URL 已带容器时仍然允许直接开终端，只是没有容器切换器。
          if (requestedContainer) {
            setContainers([requestedContainer])
            return
          }
          setContainers([])
          setPodFailed(true)
          setPodError(t("kube.podExecPodMissing"))
          return
        }
        setContainers(
          (pod.containers ?? [])
            .map((item) => item.name)
            .filter((item): item is string => Boolean(item))
        )
      } catch (err) {
        if (!active) {
          return
        }
        if (requestedContainer) {
          setContainers([requestedContainer])
          setPodLoading(false)
          return
        }
        setContainers([])
        setPodFailed(true)
        setPodError(formErrorMessage(err, t("state.failed")))
      } finally {
        if (active) {
          setPodLoading(false)
        }
      }
    }
    void loadPod()
    return () => {
      active = false
    }
  }, [uid, namespace, name, reloadKey, requestedContainer, t])

  if (!session.loading && !canExec) {
    return <Navigate to="/forbidden" replace />
  }

  const backSearch = new URLSearchParams({ tab: "pods" })
  if (namespace) {
    backSearch.set("namespace", namespace)
  }
  const backTo = `/clusters/${encodeURIComponent(uid)}?${backSearch.toString()}`

  function selectContainer(next: string) {
    const params = new URLSearchParams(searchParams)
    params.set("container", next)
    setSearchParams(params, { replace: true })
  }

  function reconnect() {
    setAttempt((value) => value + 1)
  }

  const closedReasonKey =
    state.state === "closed" ? POD_EXEC_REASON_KEYS[state.reason] : undefined
  const statusMessage =
    state.state === "failed"
      ? state.message || t("kube.podExecFailed")
      : state.state === "closed"
        ? closedReasonKey
          ? t(closedReasonKey)
          : state.reason || t("kube.podExecClosed")
        : ""
  const statusHint =
    state.state === "connecting"
      ? t("kube.podExecConnecting")
      : state.state === "open"
        ? t("kube.podExecConnected")
        : ""

  const podPath = `${namespace}/${name}`

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-4">
      <PageHeader
        title={t("kube.podExecTitle")}
        description={`${clusterName || uid} · ${podPath}${container ? ` · ${container}` : ""}`}
      >
        <Button variant="outline" asChild>
          <Link to={backTo}>
            <ArrowLeft aria-hidden="true" data-icon="inline-start" />
            {t("kube.backToPods")}
          </Link>
        </Button>
      </PageHeader>
      <div className="flex flex-wrap items-end justify-between gap-3">
        <Field className="w-56 shrink-0">
          <FieldLabel htmlFor="pod-exec-container">
            {t("kube.container")}
          </FieldLabel>
          <Select
            value={container}
            disabled={podLoading || containers.length <= 1}
            onValueChange={selectContainer}
          >
            <SelectTrigger id="pod-exec-container" className="w-full">
              <SelectValue placeholder={t("kube.container")} />
            </SelectTrigger>
            <SelectContent>
              {containers.map((item) => (
                <SelectItem key={item} value={item}>
                  {item}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </Field>
        <div className="flex flex-wrap items-center gap-2">
          {statusHint ? (
            <span className="text-xs text-muted-foreground" role="status">
              {statusHint}
            </span>
          ) : null}
          <Button
            variant="outline"
            size="icon"
            aria-label={t("action.refresh")}
            onClick={() => setReloadKey((value) => value + 1)}
          >
            <RefreshCw aria-hidden="true" />
          </Button>
        </div>
      </div>
      {podLoading ? (
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      ) : null}
      {!podLoading && podFailed ? (
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-sm text-destructive">
            {podError || t("state.failed")}
          </p>
          <Button
            variant="outline"
            onClick={() => setReloadKey((value) => value + 1)}
          >
            {t("action.retry")}
          </Button>
        </div>
      ) : null}
      {!podLoading && !podFailed && !container ? (
        <p className="text-sm text-muted-foreground">
          {t("kube.podExecNoContainer")}
        </p>
      ) : null}
      {statusMessage ? (
        <div className="flex flex-wrap items-center gap-2 rounded-lg border border-border/70 bg-card/70 px-3 py-2">
          <p className="text-sm text-destructive">{statusMessage}</p>
          <Button variant="outline" size="sm" onClick={reconnect}>
            {t("kube.podExecReconnect")}
          </Button>
        </div>
      ) : null}
      {!podLoading && !podFailed && container ? (
        <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
          <XtermTerminal
            sessionKey={sessionKey}
            issueTicket={async ({ cols, rows }) =>
              (
                await issueKubernetesPodExecTicket(uid, namespace, name, {
                  container,
                  cols,
                  rows,
                })
              ).ticket
            }
            socketUrl={(ticket) =>
              kubernetesPodExecSocketURL(uid, namespace, name, ticket)
            }
            onStateChange={(value) => setExecState({ key: sessionKey, value })}
            className="flex-1"
          />
        </div>
      ) : null}
    </section>
  )
}
