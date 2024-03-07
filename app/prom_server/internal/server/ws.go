package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/pkg/servers"
)

type WebsocketServer struct {
	*servers.WebsocketServer

	log *log.Helper
}

var sendCh = make(chan *servers.Message, 100)

func NewWebsocketServer(c *conf.Server, l log.Logger) *WebsocketServer {
	srv := servers.NewWebsocketServer(c.GetWs().GetAddr())

	srv.RegisterMessageHandler(func(msg *servers.Message) {

	})

	s := &WebsocketServer{
		log:             log.NewHelper(log.With(l, "module", "ws")),
		WebsocketServer: srv,
	}

	go func() {
		for {
			select {
			case msg := <-sendCh:
				s.SendMessage(msg)
			case <-s.StopCh:
				return
			}
		}
	}()
	return s
}
