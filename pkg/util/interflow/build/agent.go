package build

import (
	"github.com/aide-family/moon/pkg/servers"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/aide-family/moon/pkg/util/interflow/hook"
	"github.com/aide-family/moon/pkg/util/interflow/mq"
	"github.com/go-kratos/kratos/v2/log"
)

type (
	agentInterflow struct {
		log      log.Logger
		category interflowType

		grpcConf hook.GrpcConfig
		httpConf hook.HttpConfig
		kafkaCli *servers.KafkaMQServer
	}

	AgentInterflowOption func(*agentInterflow)
)

func NewAgentInterflow(opts ...AgentInterflowOption) (interflow.AgentInterflow, error) {
	a := &agentInterflow{}
	for _, opt := range opts {
		opt(a)
	}
	return a.Builder()
}

func (a *agentInterflow) Builder() (interflow.AgentInterflow, error) {
	switch a.category {
	case Grpc:
		return hook.NewAgentHookGrpcInterflow(a.grpcConf, a.log)
	case Kafka:
		return mq.NewAgentKafkaInterflow(a.kafkaCli, a.log)
	default:
		return hook.NewAgentHookHttpInterflow(a.httpConf, a.log)
	}
}
