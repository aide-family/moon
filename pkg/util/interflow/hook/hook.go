package hook

import (
	"errors"

	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/go-kratos/kratos/v2/log"
)

type (
	HttpServerConfig interface {
		GetUrl() string
	}

	GrpcServerConfig interface {
		GetEndpoints() []string
	}

	HttpConfig interface {
		GetAgent() HttpServerConfig
		GetServer() HttpServerConfig
	}

	GrpcConfig interface {
		GetAgent() GrpcServerConfig
		GetServer() GrpcServerConfig
	}

	Config interface {
		GetHttp() HttpConfig
		GetGrpc() GrpcConfig
	}
)

func NewAgent(c Config, logger log.Logger) (interflow.AgentInterflow, error) {
	switch {
	case c.GetHttp() != nil:
		return NewHookHttpInterflow(c.GetHttp(), logger), nil
	case c.GetGrpc() != nil:
		return NewHookGrpcInterflow(c.GetGrpc(), logger), nil
	default:
		return nil, errors.New("no config found")
	}
}

func NewServer(network Network, logger log.Logger) (interflow.ServerInterflow, error) {
	return NewServerHookHttpInterflow(network, logger), nil
}
