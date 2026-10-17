package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
)

func (s *AgentService) IssueShellTicket(ctx context.Context, req *v1.IssueShellTicketRequest) (*v1.IssueShellTicketResponse, error) {
	ticket, err := s.agents.IssueShellTicket(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.IssueShellTicketResponse{Ticket: ticket}, nil
}

var shellUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

// ServeShell 把浏览器 WebSocket 接到这台机器的终端。
func (s *AgentService) ServeShell(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]
	link, err := s.agents.OpenShell(context.WithoutCancel(r.Context()), uid, r.URL.Query().Get("ticket"), 80, 24)
	if err != nil {
		http.Error(w, "shell unavailable", http.StatusUnauthorized)
		return
	}
	conn, err := shellUpgrader.Upgrade(w, r, nil)
	if err != nil {
		link.Stop()
		return
	}
	defer conn.Close()
	defer link.Stop()

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
					_ = link.Resize(resize.Cols, resize.Rows)
				}
				continue
			}
			if err := link.Write(payload); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case data, ok := <-link.Data:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		case reason := <-link.Closed:
			raw, _ := json.Marshal(map[string]string{"type": "exit", "reason": reason})
			_ = conn.WriteMessage(websocket.TextMessage, raw)
			return
		}
	}
}
