package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// podExecWriteTimeout 是单帧写超时。浏览器断网后写操作不会无限阻塞，
// 处理器可以在超时后回收远端会话。
const podExecWriteTimeout = 10 * time.Second

var podExecUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

// ServePodExec 把浏览器 WebSocket 接到 Pod 容器的终端。
// 协议与 Agent Shell 一致：二进制帧是终端字节，文本帧 {"type":"resize","cols":..,"rows":..}
// 调整窗口，服务端关闭前发 {"type":"exit","reason":".."}。
// 建连凭证是 IssueKubernetesPodExecTicket 签发的票据，容器和窗口大小由票据带入。
func (s *KubernetesService) ServePodExec(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// 终端生命周期不受 HTTP 请求上下文约束：升级后的连接不再有请求 deadline，
	// 关闭时机由 WebSocket 断开或 Terminal.CloseAll 决定。
	shell, err := s.terminal.OpenShell(
		context.WithoutCancel(r.Context()),
		vars["cluster_uid"], vars["namespace"], vars["name"],
		r.URL.Query().Get("container"), r.URL.Query().Get("ticket"),
		0, 0,
	)
	if err != nil {
		// 票据问题按未认证返回 401；集群或凭据问题沿用业务错误码，但不回显细节。
		status := int(kerrors.FromError(err).Code)
		if status < http.StatusBadRequest || status > 599 {
			status = http.StatusInternalServerError
		}
		http.Error(w, "pod exec unavailable", status)
		return
	}
	conn, err := podExecUpgrader.Upgrade(w, r, nil)
	if err != nil {
		shell.Stop()
		return
	}
	defer conn.Close()
	defer shell.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			kind, payload, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if kind == websocket.TextMessage {
				var resize struct {
					Type string `json:"type"`
					Cols int32  `json:"cols"`
					Rows int32  `json:"rows"`
				}
				if json.Unmarshal(payload, &resize) == nil && resize.Type == "resize" {
					_ = shell.Resize(resize.Cols, resize.Rows)
				}
				continue
			}
			if err := shell.Write(payload); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case reason := <-shell.Closed:
			writePodExecExit(conn, reason)
			return
		case data, ok := <-shell.Data:
			if !ok {
				// Data 关闭前 Closed 已经写入原因，这里只等它送达再发 exit 帧。
				select {
				case reason := <-shell.Closed:
					writePodExecExit(conn, reason)
				case <-done:
				}
				return
			}
			if err := writePodExecFrame(conn, websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}
}

// writePodExecExit 在关闭前把结束原因告诉浏览器。reason 为空时只说明会话已结束。
func writePodExecExit(conn *websocket.Conn, reason string) {
	if reason == "" {
		reason = "closed"
	}
	raw, _ := json.Marshal(map[string]string{"type": "exit", "reason": reason})
	_ = writePodExecFrame(conn, websocket.TextMessage, raw)
}

// writePodExecFrame 带写超时地发一帧。写失败说明连接已不可用，调用方直接结束会话。
func writePodExecFrame(conn *websocket.Conn, kind int, payload []byte) error {
	_ = conn.SetWriteDeadline(time.Now().Add(podExecWriteTimeout))
	return conn.WriteMessage(kind, payload)
}
