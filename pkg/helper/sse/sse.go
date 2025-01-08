package sse

import (
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/go-kratos/kratos/v2/log"
	"net/http"
)

// NewSSEHandler handles the SSE connection
func NewSSEHandler(clientManager *ClientManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "token is required", http.StatusBadRequest)
			return
		}
		// 设置HTTP头部，指定这是一个SSE连接
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		claims, ok := middleware.ParseJwtClaimsFromToken(token)
		if !ok {
			http.Error(w, "token is invalid", http.StatusUnauthorized)
			return
		}

		client := NewClient(claims.GetUser())
		clientManager.AddClient(client)
		defer func() {
			clientManager.RemoveClient(client.ID)
			close(client.Send)
		}()

		go client.WriteSSE(w)
		<-r.Context().Done()
		log.Infof("client %d disconnected", client.ID)
	}
}
