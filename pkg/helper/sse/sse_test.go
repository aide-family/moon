package sse

import (
	_ "embed"
	"net/http"
	"testing"

	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/go-kratos/kratos/v2/log"
)

//go:embed sse.html
var sseHtml string

func TestHandleSSE(t *testing.T) {
	middleware.SetSignKey("moon-sign_key")
	http.HandleFunc("/events", HandleSSE)
	http.HandleFunc("/msg", func(writer http.ResponseWriter, request *http.Request) {
		msg := request.URL.Query().Get("msg")
		log.Debugw("msg", msg)
		SendMessage(2, msg)
	})
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(sseHtml))
	})

	http.ListenAndServe(":9999", nil)
}
