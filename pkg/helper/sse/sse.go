package sse

import (
	"errors"
	"net/http"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/go-kratos/kratos/v2/log"
)

var ErrClientNotFound = errors.New("client not found")

var clientManager *ClientManager

func init() {
	clientManager = NewClientManager()
}

// HandleSSE handles the SSE connection
func HandleSSE(w http.ResponseWriter, r *http.Request) {
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

	client := NewClient(claims.UserID)
	clientManager.AddClient(client)
	defer func() {
		clientManager.RemoveClient(client.ID)
		close(client.Send)
	}()

	go client.WriteSSE(w)
	<-r.Context().Done()
	log.Infof("client %d disconnected", client.ID)
}

// SendMessage sends a message to the client
func SendMessage(id uint32, message string) error {
	client, ok := clientManager.GetClient(id)
	if !ok {
		return ErrClientNotFound
	}
	client.Send <- []byte(message)
	return nil
}
