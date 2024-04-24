package conf

import (
	"github.com/aide-family/moon/pkg/util/interflow/hook"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"

	"github.com/aide-family/moon/pkg/helper/plog"
)

// ProviderSetConf is conf providers.
var ProviderSetConf = wire.NewSet(
	wire.FieldsOf(new(*Bootstrap), "Server"),
	wire.FieldsOf(new(*Bootstrap), "Data"),
	wire.FieldsOf(new(*Bootstrap), "Env"),
	wire.FieldsOf(new(*Bootstrap), "Log"),
	wire.FieldsOf(new(*Bootstrap), "WatchProm"),
	wire.FieldsOf(new(*Bootstrap), "Interflow"),
	wire.Bind(new(plog.Config), new(*Log)),
	LoadConfig,
)

type Before func(bc *Bootstrap) error

func LoadConfig(flagConf *string, before Before) (*Bootstrap, error) {
	if flagConf == nil || *flagConf == "" {
		return nil, errors.NotFound("FLAG_CONFIGS", "config path not found")
	}
	c := config.New(
		config.WithSource(
			file.NewSource(*flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}

	return &bc, before(&bc)
}

var _ hook.Config = (*InterflowHook)(nil)
var _ hook.HttpConfig = (*InterflowHookHTTP)(nil)
var _ hook.GrpcConfig = (*InterflowHookGRPC)(nil)
var _ hook.HttpServerConfig = (*InterflowHookHTTPServer)(nil)
var _ hook.GrpcServerConfig = (*InterflowHookGRPCServer)(nil)

type (
	InterflowHook struct {
		Http *InterflowHookHTTP
		Grpc *InterflowHookGRPC
	}

	InterflowHookHTTP struct {
		Agent  *InterflowHookHTTPServer
		Server *InterflowHookHTTPServer
	}

	InterflowHookGRPC struct {
		Agent  *InterflowHookGRPCServer
		Server *InterflowHookGRPCServer
	}

	InterflowHookHTTPServer struct {
		*Interflow_Hook_HTTP_Server
	}

	InterflowHookGRPCServer struct {
		*Interflow_Hook_GRPC_Server
	}
)

func (c *InterflowHook) GetHttp() hook.HttpConfig {
	return c.Http
}

func (c *InterflowHook) GetGrpc() hook.GrpcConfig {
	return c.Grpc
}

func (c *InterflowHookHTTP) GetAgent() hook.HttpServerConfig {
	return c.Agent
}

func (c *InterflowHookHTTP) GetServer() hook.HttpServerConfig {
	return c.Server
}

func (c *InterflowHookGRPC) GetAgent() hook.GrpcServerConfig {
	return c.Agent
}

func (c *InterflowHookGRPC) GetServer() hook.GrpcServerConfig {
	return c.Server
}

func (c *InterflowHookHTTPServer) GetUrl() string {
	return c.Interflow_Hook_HTTP_Server.GetUrl()
}

func (c *InterflowHookGRPCServer) GetEndpoints() []string {
	return c.Interflow_Hook_GRPC_Server.GetEndpoints()
}

// BuilderInterflowHook 构建钩子配置
func BuilderInterflowHook(c *Interflow_Hook) hook.Config {
	if c == nil {
		return nil
	}
	httpConf := c.GetHttp()
	grpcConf := c.GetGrpc()
	hookConf := &InterflowHook{}
	if httpConf != nil {
		hookConf.Http = &InterflowHookHTTP{
			Agent: &InterflowHookHTTPServer{
				Interflow_Hook_HTTP_Server: httpConf.GetAgent(),
			},
			Server: &InterflowHookHTTPServer{
				Interflow_Hook_HTTP_Server: httpConf.GetServer(),
			},
		}
	}
	if grpcConf != nil {
		hookConf.Grpc = &InterflowHookGRPC{
			Agent: &InterflowHookGRPCServer{
				Interflow_Hook_GRPC_Server: grpcConf.GetAgent(),
			},
			Server: &InterflowHookGRPCServer{
				Interflow_Hook_GRPC_Server: grpcConf.GetServer(),
			},
		}
	}
	return hookConf
}
