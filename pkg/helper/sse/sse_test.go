package sse

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/go-kratos/kratos/v2/log"
)

//go:embed sse.html
var sseHTML string

func TestHandleSSE(t *testing.T) {
	middleware.SetSignKey("moon-sign_key")
	clientManager := NewClientManager()
	http.HandleFunc("/events", NewSSEHandler(clientManager))
	http.HandleFunc("/msg", func(writer http.ResponseWriter, request *http.Request) {
		token := request.Header.Get("Authorization")
		if token == "" {
			http.Error(writer, "token is required", http.StatusBadRequest)
			return
		}
		claims, ok := middleware.ParseJwtClaimsFromToken(strings.TrimPrefix(token, "Bearer "))
		if !ok {
			http.Error(writer, "token is invalid", http.StatusUnauthorized)
			return
		}
		msg := request.URL.Query().Get("msg")
		log.Debugw("msg", msg)
		client, ok := clientManager.GetClient(claims.GetUser())
		if !ok {
			http.Error(writer, "client is not found", http.StatusNotFound)
		}
		if err := client.SendMessage(msg); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Debugw("msg", "send message success")
		writer.Write([]byte("ok"))
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(sseHTML))
	})

	fmt.Println("http://localhost:9999")

	http.ListenAndServe(":9999", nil)
}
