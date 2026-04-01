package server

import (
	"embed"
	nethttp "net/http"
	"strings"

	"github.com/aide-family/magicbox/auth/basic"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aide-family/jade_tree/internal/conf"
	"github.com/aide-family/jade_tree/internal/service"
)

//go:embed swagger
var docFS embed.FS

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, RegisterHTTPService)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, RegisterGRPCService)
)

type Server interface {
	transport.Server
	Name() string
	Instance() transport.Server
}

type server struct {
	transport.Server
	name string
}

func (s *server) Name() string               { return s.name }
func (s *server) Instance() transport.Server { return s.Server }

func newServer(name string, srv transport.Server) Server { return &server{Server: srv, name: name} }

type Servers []Server

func BindSwagger(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{Name: "Swagger", Enabled: strings.EqualFold(bc.GetEnableSwagger(), "true"), BasicAuth: bc.GetSwaggerBasicAuth(), Handler: nethttp.StripPrefix("/doc/", nethttp.FileServer(nethttp.FS(docFS))), Path: "/doc/"}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

func BindMetrics(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{Name: "Metrics", Enabled: strings.EqualFold(bc.GetEnableMetrics(), "true"), BasicAuth: bc.GetMetricsBasicAuth(), Handler: promhttp.Handler(), Path: "/metrics"}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

func RegisterService(c *conf.Bootstrap, httpSrv *http.Server, grpcSrv *grpc.Server, healthService *service.HealthService, sshCommand *service.SSHCommandService, machineInfo *service.MachineInfoService) Servers {
	var srvs Servers
	srvs = append(srvs, RegisterHTTPService(httpSrv, healthService, sshCommand, machineInfo)...)
	srvs = append(srvs, RegisterGRPCService(grpcSrv, healthService, sshCommand, machineInfo)...)
	_ = c
	return srvs
}
