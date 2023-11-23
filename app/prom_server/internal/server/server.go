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
)

type Server struct {
	hSrv *HttpServer
	gSrv *GrpcServer
}

func NewServer(hSrv *HttpServer, gSrv *GrpcServer) *Server {
	return &Server{
		hSrv: hSrv,
		gSrv: gSrv,
	}
}

func (l *Server) List() []transport.Server {
	return []transport.Server{
		l.hSrv,
		l.gSrv,
	}
}
