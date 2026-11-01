import { FitAddon } from "@xterm/addon-fit"
import { Terminal } from "@xterm/xterm"
import { useEffect, useEffectEvent, useRef } from "react"
import "@xterm/xterm/css/xterm.css"

import { cn } from "@/lib/class-names"

/**
 * WebSocket 握手超时。网关没转发 Upgrade、票据校验后连接被挂起、或地址拼错时，
 * 浏览器可能永远停在 CONNECTING；超时后必须给出可操作的失败态，不能一直转圈。
 */
const socketConnectTimeoutMs = 10_000

export type XtermSessionState =
  | { state: "connecting" }
  | { state: "open" }
  /** 服务端主动结束；reason 是后端给出的原因，空串表示没有说明。 */
  | { state: "closed"; reason: string }
  /** 票据申请或连接失败；message 是已本地化的错误，空串表示没有说明。 */
  | { state: "failed"; message: string; reason?: "timeout" }

interface XtermTerminalProps {
  /** 会话标识：变化时销毁旧终端与连接，按新参数重建。 */
  sessionKey: string
  /** 申请一次性票据；窗口大小取终端当前行列数。 */
  issueTicket: (size: { cols: number; rows: number }) => Promise<string>
  /** 用票据拼 WebSocket 地址。 */
  socketUrl: (ticket: string) => string
  /** 连接生命周期变化时回调；页面据此展示退出原因与重连入口。 */
  onStateChange: (state: XtermSessionState) => void
  className?: string
}

/**
 * xterm 会话容器：申请票据 → 建立 WebSocket → 双向转发字节。
 * 二进制帧是终端字节，文本帧按 JSON 解析，服务端用 {"type":"exit"} 结束会话。
 * 仅负责终端本身；标题、错误条和重连按钮由页面渲染。
 */
export function XtermTerminal({
  sessionKey,
  issueTicket,
  socketUrl,
  onStateChange,
  className,
}: XtermTerminalProps) {
  const host = useRef<HTMLDivElement>(null)
  const emitState = useEffectEvent(onStateChange)
  const requestTicket = useEffectEvent(issueTicket)
  const buildSocketUrl = useEffectEvent(socketUrl)

  useEffect(() => {
    const el = host.current
    if (!el) {
      return
    }
    // 嵌套函数里要用到元素，单独用非空常量接住，避免闭包丢失收窄。
    const container: HTMLDivElement = el
    let active = true
    let socket: WebSocket | null = null
    let connectTimer: number | undefined
    // 服务端主动结束或连接失败后，socket 的 close/error 事件不再覆盖已有原因。
    let settled = false
    let lastCols = 0
    let lastRows = 0

    emitState({ state: "connecting" })
    const term = new Terminal({
      cursorBlink: true,
      fontFamily: "JetBrains Mono, ui-monospace, monospace",
      fontSize: 14,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(container)
    fit.fit()
    lastCols = term.cols
    lastRows = term.rows

    function sendResize(force = false) {
      if (container.clientWidth === 0 || container.clientHeight === 0) {
        return
      }
      fit.fit()
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        return
      }
      if (!force && term.cols === lastCols && term.rows === lastRows) {
        return
      }
      lastCols = term.cols
      lastRows = term.rows
      socket.send(
        JSON.stringify({ type: "resize", cols: term.cols, rows: term.rows })
      )
    }

    function handleWindowResize() {
      sendResize()
    }

    term.onData((data) => {
      if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(new TextEncoder().encode(data))
      }
    })

    void requestTicket({ cols: term.cols, rows: term.rows })
      .then((ticket) => {
        if (!active) {
          return
        }
        const url = buildSocketUrl(ticket)
        // 地址必须是绝对 ws/wss；相对地址会被浏览器按当前页面解析，
        // 落到前端的 SPA fallback 上就会一直 pending（例如网关没有转发升级请求）。
        if (!/^wss?:\/\//i.test(url)) {
          settled = true
          emitState({ state: "failed", message: "" })
          return
        }
        try {
          socket = new WebSocket(url)
        } catch {
          settled = true
          emitState({ state: "failed", message: "" })
          return
        }
        socket.binaryType = "arraybuffer"
        connectTimer = window.setTimeout(() => {
          if (!active || settled) {
            return
          }
          settled = true
          const pending = socket
          socket = null
          pending?.close()
          emitState({ state: "failed", message: "", reason: "timeout" })
        }, socketConnectTimeoutMs)
        socket.onopen = () => {
          if (!active) {
            return
          }
          window.clearTimeout(connectTimer)
          emitState({ state: "open" })
          sendResize(true)
        }
        socket.onmessage = (event) => {
          if (typeof event.data === "string") {
            try {
              const payload = JSON.parse(event.data) as {
                type?: string
                reason?: string
              }
              if (payload.type === "exit") {
                settled = true
                emitState({ state: "closed", reason: payload.reason ?? "" })
              }
            } catch {
              term.write(event.data)
            }
            return
          }
          term.write(new Uint8Array(event.data as ArrayBuffer))
        }
        socket.onerror = () => {
          if (!active || settled) {
            return
          }
          window.clearTimeout(connectTimer)
          settled = true
          emitState({ state: "failed", message: "" })
        }
        socket.onclose = () => {
          if (!active || settled) {
            return
          }
          window.clearTimeout(connectTimer)
          settled = true
          emitState({ state: "closed", reason: "" })
        }
      })
      .catch((err: unknown) => {
        if (!active || settled) {
          return
        }
        settled = true
        emitState({
          state: "failed",
          message: err instanceof Error ? err.message : "",
        })
      })

    // 终端尺寸也会随布局变化（不仅是窗口 resize），用 ResizeObserver 跟随并同步 stty size。
    const observer = new ResizeObserver(() => sendResize())
    observer.observe(container)
    window.addEventListener("resize", handleWindowResize)

    return () => {
      active = false
      window.clearTimeout(connectTimer)
      observer.disconnect()
      window.removeEventListener("resize", handleWindowResize)
      socket?.close()
      socket = null
      term.dispose()
    }
  }, [sessionKey])

  return (
    <div
      ref={host}
      className={cn(
        "min-h-0 min-w-0 overflow-hidden rounded-lg border border-border/70 bg-[var(--code-bg)] p-2",
        className
      )}
    />
  )
}
