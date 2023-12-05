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
)

type Server struct {
	hSrv       *HttpServer
	gSrv       *GrpcServer
	alarmEvent *AlarmEvent
}

func NewServer(hSrv *HttpServer, gSrv *GrpcServer, alarmEvent *AlarmEvent) *Server {
	return &Server{
		hSrv:       hSrv,
		gSrv:       gSrv,
		alarmEvent: alarmEvent,
	}
}

func (l *Server) List() []transport.Server {
	return []transport.Server{
		l.hSrv,
		l.gSrv,
		l.alarmEvent,
	}
}
