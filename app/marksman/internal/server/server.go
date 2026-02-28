// Package server is a server package for kratos.
package server

import (
	"embed"
	nethttp "net/http"
	"strings"

	"buf.build/go/protoyaml"
	magicboxapiv1 "github.com/aide-family/magicbox/api/v1"
	"github.com/aide-family/magicbox/domain/auth/basic"
	"github.com/aide-family/magicbox/oauth"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.yaml.in/yaml/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/service"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

//go:embed swagger
var docFS embed.FS

type protoYAMLCodec struct {
	marshalOptions   protoyaml.MarshalOptions
	unmarshalOptions protoyaml.UnmarshalOptions
}

func newProtoYAMLCodec() *protoYAMLCodec {
	return &protoYAMLCodec{
		marshalOptions: protoyaml.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: false, // 过滤 0 值和空值
			Indent:          2,
		},
		unmarshalOptions: protoyaml.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
}

// Marshal implements encoding.Codec.
func (c *protoYAMLCodec) Marshal(v any) ([]byte, error) {
	switch m := v.(type) {
	case protoreflect.ProtoMessage:
		return c.marshalOptions.Marshal(m)
	default:
		return yaml.Marshal(m)
	}
}

// Unmarshal implements encoding.Codec.
func (c *protoYAMLCodec) Unmarshal(data []byte, v any) error {
	switch m := v.(type) {
	case protoreflect.ProtoMessage:
		return c.unmarshalOptions.Unmarshal(data, m)
	default:
		return yaml.Unmarshal(data, m)
	}
}

// Name implements encoding.Codec.
func (c *protoYAMLCodec) Name() string {
	return "yaml"
}

var (
	ProviderSetServerAll  = wire.NewSet(NewHTTPServer, NewGRPCServer, RegisterService)
	ProviderSetServerHTTP = wire.NewSet(NewHTTPServer, RegisterHTTPService)
	ProviderSetServerGRPC = wire.NewSet(NewGRPCServer, RegisterGRPCService)
)

// init initializes the json.MarshalOptions.
func init() {
	json.MarshalOptions = protojson.MarshalOptions{
		// UseEnumNumbers:  true, // Emit enum values as numbers instead of their string representation (default is string).
		UseProtoNames:   true, // Use the field names defined in the proto file as the output field names.
		EmitUnpopulated: true, // Emit fields even if they are unset or empty.
	}
	encoding.RegisterCodec(newProtoYAMLCodec())
}

type Server interface {
	transport.Server
	Name() string
	Instance() transport.Server
}

type server struct {
	transport.Server
	name string
}

func (s *server) Name() string {
	return s.name
}

func (s *server) Instance() transport.Server {
	return s.Server
}

func newServer(name string, srv transport.Server) Server {
	return &server{
		Server: srv,
		name:   name,
	}
}

type Servers []Server

func BindSwagger(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{
		Name:      "Swagger",
		Enabled:   strings.EqualFold(bc.GetEnableSwagger(), "true"),
		BasicAuth: bc.GetSwaggerBasicAuth(),
		Handler:   nethttp.StripPrefix("/doc/", nethttp.FileServer(nethttp.FS(docFS))),
		Path:      "/doc/",
	}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

func BindMetrics(httpSrv *http.Server, bc *conf.Bootstrap) {
	binding := basic.HandlerBinding{
		Name:      "Metrics",
		Enabled:   strings.EqualFold(bc.GetEnableMetrics(), "true"),
		BasicAuth: bc.GetMetricsBasicAuth(),
		Handler:   promhttp.Handler(),
		Path:      "/metrics",
	}
	basic.BindHandlerWithAuth(httpSrv, binding)
}

// RegisterService registers the service.
func RegisterService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	grpcSrv *grpc.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	levelService *service.LevelService,
	datasourceService *service.DatasourceService,
	strategyService *service.StrategyService,
	strategyMetricService *service.StrategyMetricService,
) Servers {
	var srvs Servers

	srvs = append(srvs, RegisterHTTPService(c, httpSrv,
		authService,
		healthService,
		namespaceService,
		levelService,
		datasourceService,
		strategyService,
		strategyMetricService,
	)...)
	srvs = append(srvs, RegisterGRPCService(c, grpcSrv, healthService, namespaceService, levelService, datasourceService, strategyService, strategyMetricService)...)
	return srvs
}

// RegisterHTTPService registers only HTTP service.
func RegisterHTTPService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	levelService *service.LevelService,
	datasourceService *service.DatasourceService,
	strategyService *service.StrategyService,
	strategyMetricService *service.StrategyMetricService,
) Servers {
	magicboxapiv1.RegisterHealthHTTPServer(httpSrv, healthService)
	magicboxapiv1.RegisterNamespaceHTTPServer(httpSrv, namespaceService)
	apiv1.RegisterLevelHTTPServer(httpSrv, levelService)
	apiv1.RegisterDatasourceHTTPServer(httpSrv, datasourceService)
	apiv1.RegisterStrategyHTTPServer(httpSrv, strategyService)
	apiv1.RegisterStrategyMetricHTTPServer(httpSrv, strategyMetricService)

	oauth2Handler := oauth.NewOAuth2Handler(c.GetOauth2(), authService.Login)
	if err := oauth2Handler.Handler(httpSrv); err != nil {
		panic(err)
	}
	return Servers{newServer("http", httpSrv)}
}

// RegisterGRPCService registers only gRPC service.
func RegisterGRPCService(
	c *conf.Bootstrap,
	grpcSrv *grpc.Server,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	levelService *service.LevelService,
	datasourceService *service.DatasourceService,
	strategyService *service.StrategyService,
	strategyMetricService *service.StrategyMetricService,
) Servers {
	magicboxapiv1.RegisterHealthServer(grpcSrv, healthService)
	magicboxapiv1.RegisterNamespaceServer(grpcSrv, namespaceService)
	apiv1.RegisterLevelServer(grpcSrv, levelService)
	apiv1.RegisterDatasourceServer(grpcSrv, datasourceService)
	apiv1.RegisterStrategyServer(grpcSrv, strategyService)
	apiv1.RegisterStrategyMetricServer(grpcSrv, strategyMetricService)
	return Servers{newServer("grpc", grpcSrv)}
}

var namespaceAllowList = []string{
	magicboxapiv1.OperationNamespaceCreateNamespace,
	magicboxapiv1.OperationNamespaceUpdateNamespace,
	magicboxapiv1.OperationNamespaceUpdateNamespaceStatus,
	magicboxapiv1.OperationNamespaceDeleteNamespace,
	magicboxapiv1.OperationNamespaceGetNamespace,
	magicboxapiv1.OperationNamespaceListNamespace,
	apiv1.OperationLevelCreateLevel,
	apiv1.OperationLevelUpdateLevel,
	apiv1.OperationLevelUpdateLevelStatus,
	apiv1.OperationLevelDeleteLevel,
	apiv1.OperationLevelGetLevel,
	apiv1.OperationLevelListLevel,
	apiv1.OperationLevelSelectLevel,
	apiv1.OperationDatasourceCreateDatasource,
	apiv1.OperationDatasourceUpdateDatasource,
	apiv1.OperationDatasourceDeleteDatasource,
	apiv1.OperationDatasourceGetDatasource,
	apiv1.OperationDatasourceListDatasource,
	apiv1.OperationStrategyCreateStrategyGroup,
	apiv1.OperationStrategyUpdateStrategyGroup,
	apiv1.OperationStrategyUpdateStrategyGroupStatus,
	apiv1.OperationStrategyDeleteStrategyGroup,
	apiv1.OperationStrategyGetStrategyGroup,
	apiv1.OperationStrategyListStrategyGroup,
	apiv1.OperationStrategySelectStrategyGroup,
	apiv1.OperationStrategyStrategyGroupBindReceivers,
	apiv1.OperationStrategyCreateStrategy,
	apiv1.OperationStrategyUpdateStrategy,
	apiv1.OperationStrategyUpdateStrategyStatus,
	apiv1.OperationStrategyDeleteStrategy,
	apiv1.OperationStrategyGetStrategy,
	apiv1.OperationStrategyListStrategy,
	apiv1.OperationStrategyMetricSaveStrategyMetric,
	apiv1.OperationStrategyMetricGetStrategyMetric,
	apiv1.OperationStrategyMetricSaveStrategyMetricLevel,
	apiv1.OperationStrategyMetricUpdateStrategyMetricLevelStatus,
	apiv1.OperationStrategyMetricDeleteStrategyMetricLevel,
	apiv1.OperationStrategyMetricGetStrategyMetricLevel,
	apiv1.OperationStrategyMetricStrategyMetricBindReceivers,
}

var authAllowList = []string{
	magicboxapiv1.OperationHealthHealthCheck,
	oauth.OperationOAuth2Reports,
	oauth.OperationOAuth2Login,
	oauth.OperationOAuth2Callback,
}
