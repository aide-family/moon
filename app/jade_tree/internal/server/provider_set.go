// Package server configures HTTP and gRPC servers.
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	bizcollector "github.com/aide-family/jade_tree/internal/biz/collector"
	"github.com/aide-family/jade_tree/internal/conf"
	servercron "github.com/aide-family/jade_tree/internal/server/cron"
	"github.com/aide-family/jade_tree/internal/service"
)

//go:embed swagger
var docFS embed.FS

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, servercron.NewMachineInfoReporterServer, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, servercron.NewMachineInfoReporterServer, RegisterHTTPServiceWithReporter)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, servercron.NewMachineInfoReporterServer, RegisterGRPCServiceWithReporter)
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

func BindMetrics(httpSrv *http.Server, bc *conf.Bootstrap, probeCollector *bizcollector.ProbeCollector) {
	reg := prometheus.NewRegistry()
	if probeCollector != nil && probeCollector.Enabled(bc) {
		reg.MustRegister(probeCollector)
	}
	binding := basic.HandlerBinding{Name: "Metrics", Enabled: strings.EqualFold(bc.GetEnableMetrics(), "true"), BasicAuth: bc.GetMetricsBasicAuth(), Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}), Path: "/metrics"}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

func RegisterService(c *conf.Bootstrap, httpSrv *http.Server, grpcSrv *grpc.Server, machineInfoReporter *servercron.MachineInfoReporterServer, healthService *service.HealthService, sshCommand *service.SSHCommandService, machineInfo *service.MachineInfoService, probeTask *service.ProbeTaskService) Servers {
	var srvs Servers
	srvs = append(srvs, RegisterHTTPService(httpSrv, healthService, sshCommand, machineInfo, probeTask)...)
	srvs = append(srvs, RegisterGRPCService(grpcSrv, healthService, sshCommand, machineInfo, probeTask)...)
	srvs = append(srvs, RegisterMachineInfoReporterService(machineInfoReporter)...)
	_ = c
	return srvs
}

func RegisterMachineInfoReporterService(machineInfoReporter *servercron.MachineInfoReporterServer) Servers {
	return Servers{newServer("machine-info-reporter", machineInfoReporter)}
}

func RegisterHTTPServiceWithReporter(httpSrv *http.Server, machineInfoReporter *servercron.MachineInfoReporterServer, healthService *service.HealthService, sshCommand *service.SSHCommandService, machineInfo *service.MachineInfoService, probeTask *service.ProbeTaskService) Servers {
	srvs := RegisterHTTPService(httpSrv, healthService, sshCommand, machineInfo, probeTask)
	srvs = append(srvs, RegisterMachineInfoReporterService(machineInfoReporter)...)
	return srvs
}

func RegisterGRPCServiceWithReporter(grpcSrv *grpc.Server, machineInfoReporter *servercron.MachineInfoReporterServer, healthService *service.HealthService, sshCommand *service.SSHCommandService, machineInfo *service.MachineInfoService, probeTask *service.ProbeTaskService) Servers {
	srvs := RegisterGRPCService(grpcSrv, healthService, sshCommand, machineInfo, probeTask)
	srvs = append(srvs, RegisterMachineInfoReporterService(machineInfoReporter)...)
	return srvs
}
