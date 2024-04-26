package build

import (
	"github.com/aide-family/moon/pkg"
	"github.com/aide-family/moon/pkg/servers"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/aide-family/moon/pkg/util/interflow/hook"
	"github.com/go-kratos/kratos/v2/log"
)

// WithServerNetwork 设置hook通信网络
func WithServerNetwork(network interflow.Network) ServerInterflowOption {
	return func(s *serverInterflow) {
		s.network = network
		if s.network.IsGRPC() && s.category == Http {
			s.category = Grpc
		}
	}
}

// WithServerLogger 设置日志
func WithServerLogger(logger log.Logger) ServerInterflowOption {
	return func(s *serverInterflow) {
		if pkg.IsNil(logger) {
			return
		}
		s.log = logger
	}
}

// WithServerKafka 设置kafka
func WithServerKafka(kafkaMQServer *servers.KafkaMQServer) ServerInterflowOption {
	return func(s *serverInterflow) {
		if pkg.IsNil(kafkaMQServer) {
			return
		}
		s.kafka = kafkaMQServer
		if s.network.IsGRPC() && s.category == Http {
			s.category = Kafka
		}
	}
}

// WithAgentHttpNetwork 设置http hook通信网络
func WithAgentHttpNetwork(network hook.HttpConfig) AgentInterflowOption {
	return func(a *agentInterflow) {
		if pkg.IsNil(network) {
			return
		}
		a.httpConf = network
	}
}

// WithAgentGrpcNetwork 设置grpc hook通信网络
func WithAgentGrpcNetwork(network hook.GrpcConfig) AgentInterflowOption {
	return func(a *agentInterflow) {
		if pkg.IsNil(network) {
			return
		}
		a.grpcConf = network
		if a.category == Http {
			a.category = Grpc
		}
	}
}

// WithAgentLogger 设置日志
func WithAgentLogger(logger log.Logger) AgentInterflowOption {
	return func(a *agentInterflow) {
		if pkg.IsNil(logger) {
			return
		}
		a.log = logger
	}
}

// WithAgentKafka 设置kafka
func WithAgentKafka(kafkaMQServer *servers.KafkaMQServer) AgentInterflowOption {
	return func(a *agentInterflow) {
		if pkg.IsNil(kafkaMQServer) {
			return
		}
		a.kafkaCli = kafkaMQServer
		if a.category == Http {
			a.category = Kafka
		}
	}
}
