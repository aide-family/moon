import { ArrowLeft, ExternalLink, Info, RefreshCw } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Link, Navigate, useParams } from "react-router-dom"

import { getBoard } from "@/api/board"
import type { BoardItem, BoardPanelItem } from "@/api/board/types"
import { ApiError } from "@/api/http"
import { useSession } from "@/auth/session"
import { PageHeader } from "@/components/page-header"
import { StatusBadge } from "@/components/status-badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/lib/class-names"
import { formatDateTime } from "@/lib/date-time"
import { failureMessage } from "@/lib/form-error"
import { buildGrafanaEmbedURL } from "@/lib/grafana-url"
import { BoardLoadHint } from "@/pages/boards/board-load-hint"
import { ServerErrorPage } from "@/pages/server-error"

// 「显示标题与描述」按大盘存在 localStorage，不改实体；缺省显示。
const SHOW_CHROME_PREFIX = "board:showChrome:"

function readShowChrome(key: string): boolean {
  try {
    return localStorage.getItem(key) !== "0"
  } catch {
    return true
  }
}

function writeShowChrome(key: string, value: boolean) {
  try {
    localStorage.setItem(key, value ? "1" : "0")
  } catch {
    // 隐私模式下 localStorage 可能不可写：本次会话内开关仍然生效。
  }
}

// Tailwind 只生成源码里出现过的类名，宽度不能拼接，必须列全 1–12 栏。
const COL_SPAN: Record<number, string> = {
  1: "md:col-span-1",
  2: "md:col-span-2",
  3: "md:col-span-3",
  4: "md:col-span-4",
  5: "md:col-span-5",
  6: "md:col-span-6",
  7: "md:col-span-7",
  8: "md:col-span-8",
  9: "md:col-span-9",
  10: "md:col-span-10",
  11: "md:col-span-11",
  12: "md:col-span-12",
}

function columnSpan(width: number): string {
  const span = Math.min(12, Math.max(1, Math.round(width)))
  return COL_SPAN[span] ?? "md:col-span-6"
}

type BoardPanelReachability = "checking" | "ok" | "failed"

/**
 * 尽最大努力判断浏览器能否访问 Grafana。
 *
 * 跨域 iframe 的 load 事件在证书、DNS、拒绝连接等错误页上也会触发，父页面
 * 读不到真实错误，所以额外发一个 no-cors HEAD 探测：只要网络层失败（DNS、
 * 证书、混合内容、连接拒绝）fetch 就会 reject；成功只能说明「能到达」，
 * 不能证明 Grafana 允许被 iframe 嵌入（X-Frame-Options/CSP 仍可能拒绝）。
 * 探测失败时页面在面板右上角显示警告图标并给出排查步骤。
 */
function useGrafanaReachability(
  url: string,
  refreshKey: string
): BoardPanelReachability {
  const bucket = `${url}|${refreshKey}`
  // 没有结果时由调用方按「探测中」渲染；effect 里只写异步结果，避免同步 setState。
  const [result, setResult] = useState<{
    bucket: string
    state: BoardPanelReachability
  }>({ bucket: "", state: "checking" })
  useEffect(() => {
    let active = true
    const controller = new AbortController()
    const timer = window.setTimeout(() => controller.abort(), 8000)
    fetch(buildGrafanaEmbedURL(url), {
      method: "HEAD",
      mode: "no-cors",
      credentials: "omit",
      cache: "no-store",
      signal: controller.signal,
    })
      .then(() => {
        if (active) {
          setResult({ bucket, state: "ok" })
        }
      })
      .catch(() => {
        if (active) {
          setResult({ bucket, state: "failed" })
        }
      })
      .finally(() => window.clearTimeout(timer))
    return () => {
      active = false
      controller.abort()
      window.clearTimeout(timer)
    }
  }, [bucket, url])
  return result.bucket === bucket ? result.state : "checking"
}

interface BoardPanelProps {
  panel: BoardPanelItem
  showChrome: boolean
  /** 占 12 栏栅格中的几列，窄屏由外层固定 col-span-12。 */
  spanClass: string
  /** 递增时重建 iframe，用于刷新面板内容。 */
  reloadKey: number
}

/** 单个面板卡片：标题/描述在 iframe 上方，隐藏时 iframe 占满卡片。 */
function BoardPanel({
  panel,
  showChrome,
  spanClass,
  reloadKey,
}: BoardPanelProps) {
  const { t } = useTranslation()
  const title = panel.title || t("board.view.untitled")
  const [frameKey, setFrameKey] = useState(0)
  const reachability = useGrafanaReachability(
    panel.grafanaUrl,
    `${reloadKey}:${frameKey}`
  )
  const failed = reachability === "failed"

  function reloadFrame() {
    setFrameKey((value) => value + 1)
  }

  return (
    <article
      className={cn(
        "relative col-span-12 flex min-w-0 flex-col",
        showChrome &&
          "overflow-hidden rounded-lg border border-border bg-card shadow-sm transition-[transform,box-shadow,border-color] duration-200 hover:-translate-y-0.5 hover:border-border hover:shadow-lg motion-reduce:transition-none motion-reduce:hover:translate-y-0",
        spanClass
      )}
    >
      {showChrome ? (
        <header className="flex flex-wrap items-start justify-between gap-2 px-3 pt-3 pb-2">
          <div className="min-w-0">
            <div className="flex items-center gap-1.5">
              <h2 className="text-sm font-medium break-words">{panel.title}</h2>
              {failed ? (
                <BoardLoadHint
                  failed
                  dense
                  openOriginalURL={panel.grafanaUrl}
                  onReload={reloadFrame}
                />
              ) : null}
            </div>
            {panel.description ? (
              <p className="text-xs leading-relaxed break-words text-muted-foreground">
                {panel.description}
              </p>
            ) : null}
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <Button
              variant="link"
              size="sm"
              className="h-auto shrink-0 px-0"
              asChild
            >
              <a
                href={panel.grafanaUrl}
                target="_blank"
                rel="noopener noreferrer"
              >
                <ExternalLink aria-hidden="true" data-icon="inline-start" />
                {t("board.view.openOriginal")}
              </a>
            </Button>
          </div>
        </header>
      ) : null}
      {!showChrome && failed ? (
        <div className="absolute top-2 right-2 z-10 rounded-md bg-background/90 shadow-sm backdrop-blur">
          <BoardLoadHint
            failed
            openOriginalURL={panel.grafanaUrl}
            onReload={reloadFrame}
          />
        </div>
      ) : null}
      <iframe
        key={`${panel.uid}-${reloadKey}-${frameKey}`}
        src={buildGrafanaEmbedURL(panel.grafanaUrl)}
        title={title}
        allowFullScreen
        referrerPolicy="no-referrer"
        loading="lazy"
        sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
        className="block w-full border-0"
        style={{ height: panel.height }}
      />
    </article>
  )
}

/**
 * 大盘查看页：12 栏栅格按面板 width 展开，窄屏统一单列。
 * iframe 固定禁止 Referer 并限制 sandbox；跨域加载失败无法被父页面读取，
 * 页面不伪造「加载成功」，只给 Grafana 侧的嵌入配置说明。
 */
export function BoardViewPage() {
  const { t } = useTranslation()
  const session = useSession()
  const { uid = "" } = useParams()
  const storageKey = `${SHOW_CHROME_PREFIX}${uid}`
  const [board, setBoard] = useState<BoardItem | null>(null)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [failedStatus, setFailedStatus] = useState(0)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)
  const [reloadKey, setReloadKey] = useState(0)
  const [showChrome, setShowChrome] = useState(() =>
    readShowChrome(`${SHOW_CHROME_PREFIX}${uid}`)
  )

  useEffect(() => {
    let active = true
    // 首屏 loading 由 useState(true) 给出；重试由事件处理器重新打开，避免在 effect 里同步 setState。
    getBoard(uid)
      .then((result) => {
        if (!active) {
          return
        }
        setBoard(result.board)
        setFailed(false)
        setFailedStatus(0)
        setError("")
      })
      .catch((err) => {
        if (!active) {
          return
        }
        setFailed(true)
        setFailedStatus(err instanceof ApiError ? err.status : 0)
        setError(failureMessage(err, ""))
      })
      .finally(() => {
        if (active) {
          setLoading(false)
        }
      })
    return () => {
      active = false
    }
  }, [uid, attempt])

  if (!session.loading && !session.can("board:read")) {
    return <Navigate to="/forbidden" replace />
  }

  function reload() {
    setLoading(true)
    setAttempt((value) => value + 1)
  }

  function toggleChrome(next: boolean) {
    setShowChrome(next)
    writeShowChrome(storageKey, next)
  }

  // 网关/后端不可用时展示标准 502/503 页，而不是把浏览器的 "Failed to fetch" 留在内容区。
  if (!loading && failed && failedStatus >= 500) {
    return (
      <ServerErrorPage
        fullPage={false}
        status={failedStatus === 503 ? 503 : 502}
        detail={error === t("error.badGatewayTitle") ? "" : error}
        onRetry={reload}
      />
    )
  }

  if (loading && !board) {
    return (
      <section className="flex min-h-0 flex-1 flex-col">
        <p className="text-sm text-muted-foreground">{t("state.loading")}</p>
      </section>
    )
  }

  if (!board) {
    return (
      <section className="flex min-h-0 flex-1 flex-col gap-4">
        <PageHeader title={t("board.title")} />
        <div className="flex flex-wrap items-center gap-2">
          <p className="text-sm text-destructive">
            {failedStatus === 404 || !error ? t("board.view.notFound") : error}
          </p>
          {failedStatus === 404 ? null : (
            <Button variant="outline" onClick={reload}>
              {t("action.retry")}
            </Button>
          )}
        </div>
        <div>
          <Button variant="outline" asChild>
            <Link to="/boards">
              <ArrowLeft aria-hidden="true" data-icon="inline-start" />
              {t("board.view.back")}
            </Link>
          </Button>
        </div>
      </section>
    )
  }

  const panelCount = board.panels.length

  return (
    <section className="flex min-h-0 flex-1 flex-col gap-4">
      <div className="flex shrink-0 flex-col gap-3">
        <PageHeader
          title={board.name}
          description={board.description || undefined}
          titleAccessory={<BoardLoadHint dense />}
        >
          <div className="flex flex-wrap items-center justify-end gap-2">
            <Button variant="outline" asChild>
              <Link to="/boards">
                <ArrowLeft aria-hidden="true" data-icon="inline-start" />
                {t("board.view.back")}
              </Link>
            </Button>
            <div className="flex h-8 items-center gap-2 rounded-lg border border-border/70 px-2.5">
              <Checkbox
                id="board-show-chrome"
                checked={showChrome}
                onCheckedChange={(value) => toggleChrome(value === true)}
              />
              <label
                htmlFor="board-show-chrome"
                className="text-sm whitespace-nowrap"
              >
                {t("board.view.showChrome")}
              </label>
            </div>
            <Button
              variant="outline"
              onClick={() => setReloadKey((v) => v + 1)}
            >
              <RefreshCw aria-hidden="true" data-icon="inline-start" />
              {t("board.view.refresh")}
            </Button>
          </div>
        </PageHeader>
        <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-muted-foreground">
          <StatusBadge value={board.status} />
          <span className="font-mono text-xs tabular-nums">
            {t("board.updatedAt")}: {formatDateTime(board.updatedAt) || "—"}
          </span>
          <span className="text-xs">
            {t("board.view.panelCount", { count: panelCount })}
          </span>
        </div>
        {board.status === "DISABLED" ? (
          <div
            role="status"
            className="flex items-start gap-2 rounded-lg border border-warning-border bg-warning-soft px-3 py-2 text-sm text-warning"
          >
            <Info aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
            <p className="min-w-0 break-words">
              {t("board.view.disabledHint")}
            </p>
          </div>
        ) : null}
      </div>
      {/* app shell 是 h-svh overflow-hidden：面板区是页面唯一的纵向滚动 owner。 */}
      <div className="flex min-h-0 min-w-0 flex-1 flex-col overflow-y-auto">
        {panelCount === 0 ? (
          <div className="grid gap-4 rounded-xl border border-dashed border-border/70 p-6">
            <div className="grid gap-1">
              <p className="text-sm font-medium">{t("board.view.empty")}</p>
              <p className="text-sm text-muted-foreground">
                {t("board.view.emptyHint")}
              </p>
            </div>
            <div className="grid gap-2">
              <p className="text-sm font-medium">{t("board.security.title")}</p>
              <ul className="grid list-disc gap-1.5 pl-5 text-sm text-muted-foreground">
                <li>{t("board.security.embedding")}</li>
                <li>{t("board.security.framing")}</li>
                <li>{t("board.security.login")}</li>
              </ul>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-12 gap-0">
            {board.panels.map((panel) => (
              <BoardPanel
                key={panel.uid}
                panel={panel}
                showChrome={showChrome}
                spanClass={columnSpan(panel.width)}
                reloadKey={reloadKey}
              />
            ))}
          </div>
        )}
      </div>
    </section>
  )
}
