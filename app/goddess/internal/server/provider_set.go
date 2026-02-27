// Package server is a server package for kratos.
package server

import (
	"embed"
	nethttp "net/http"
	"strings"

	"buf.build/go/protoyaml"
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

	"github.com/aide-family/goddess/internal/conf"
	"github.com/aide-family/goddess/internal/service"
	magicboxv1 "github.com/aide-family/magicbox/api/v1"
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
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
) Servers {
	var srvs Servers

	srvs = append(srvs, RegisterHTTPService(c, httpSrv,
		authService,
		healthService,
		namespaceService,
		userService,
		memberService,
		selfService,
	)...)
	srvs = append(srvs, RegisterGRPCService(c, grpcSrv,
		healthService,
		namespaceService,
		userService,
		memberService,
		selfService,
	)...)
	return srvs
}

// RegisterHTTPService registers only HTTP service.
func RegisterHTTPService(
	c *conf.Bootstrap,
	httpSrv *http.Server,
	authService *service.AuthService,
	healthService *service.HealthService,
	namespaceService *service.NamespaceService,
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
) Servers {
	magicboxv1.RegisterHealthHTTPServer(httpSrv, healthService)
	magicboxv1.RegisterNamespaceHTTPServer(httpSrv, namespaceService)
	magicboxv1.RegisterUserHTTPServer(httpSrv, userService)
	magicboxv1.RegisterMemberHTTPServer(httpSrv, memberService)
	magicboxv1.RegisterSelfHTTPServer(httpSrv, selfService)

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
	userService *service.UserService,
	memberService *service.MemberService,
	selfService *service.SelfService,
) Servers {
	magicboxv1.RegisterHealthServer(grpcSrv, healthService)
	magicboxv1.RegisterNamespaceServer(grpcSrv, namespaceService)
	magicboxv1.RegisterUserServer(grpcSrv, userService)
	magicboxv1.RegisterMemberServer(grpcSrv, memberService)
	magicboxv1.RegisterSelfServer(grpcSrv, selfService)
	return Servers{newServer("grpc", grpcSrv)}
}

var namespaceAllowList = []string{
	magicboxv1.OperationNamespaceCreateNamespace,
	magicboxv1.OperationNamespaceUpdateNamespace,
	magicboxv1.OperationNamespaceUpdateNamespaceStatus,
	magicboxv1.OperationNamespaceDeleteNamespace,
	magicboxv1.OperationNamespaceGetNamespace,
	magicboxv1.OperationNamespaceListNamespace,
	magicboxv1.OperationNamespaceSelectNamespace,
	magicboxv1.OperationHealthHealthCheck,
	magicboxv1.OperationSelfInfo,
	magicboxv1.OperationSelfNamespaces,
	magicboxv1.OperationSelfChangeEmail,
	magicboxv1.OperationSelfChangeAvatar,
	magicboxv1.OperationSelfRefreshToken,
	magicboxv1.OperationUserGetUser,
	magicboxv1.OperationUserListUser,
	magicboxv1.OperationUserSelectUser,
	magicboxv1.OperationUserBanUser,
	magicboxv1.OperationUserPermitUser,
}

var authAllowList = []string{
	magicboxv1.OperationHealthHealthCheck,
	oauth.OperationOAuth2Reports,
	oauth.OperationOAuth2Login,
	oauth.OperationOAuth2Callback,
}
