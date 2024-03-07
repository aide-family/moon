package server

import (
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/google/wire"
)

// ProviderSetServer is server providers.
var ProviderSetServer = wire.NewSet(
	NewGRPCServer,
	NewHTTPServer,
	RegisterHttpServer,
	RegisterGrpcServer,
	NewServer,
	NewAlarmEvent,
	NewWebsocketServer,
)

type Server struct {
	hSrv       *HttpServer
	gSrv       *GrpcServer
	alarmEvent *AlarmEvent
	wSrv       *WebsocketServer
}

func NewServer(hSrv *HttpServer, gSrv *GrpcServer, alarmEvent *AlarmEvent, ws *WebsocketServer) *Server {
	return &Server{
		hSrv:       hSrv,
		gSrv:       gSrv,
		wSrv:       ws,
		alarmEvent: alarmEvent,
	}
}

func (l *Server) List() []transport.Server {
	return []transport.Server{
		l.hSrv,
		l.gSrv,
		l.alarmEvent,
		l.wSrv,
	}
}
