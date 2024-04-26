package build

import (
	"github.com/aide-family/moon/pkg/servers"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/aide-family/moon/pkg/util/interflow/hook"
	"github.com/aide-family/moon/pkg/util/interflow/mq"
	"github.com/go-kratos/kratos/v2/log"
)

type (
	interflowType   int
	serverInterflow struct {
		network  interflow.Network
		log      log.Logger
		kafka    *servers.KafkaMQServer
		category interflowType
	}

	ServerInterflowOption func(s *serverInterflow)
)

const (
	Http interflowType = iota
	Grpc
	Kafka
)

func NewServerInterflow(opts ...ServerInterflowOption) (interflow.ServerInterflow, error) {
	s := &serverInterflow{
		log: log.GetLogger(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s.Builder()
}

func (s *serverInterflow) Builder() (interflow.ServerInterflow, error) {
	switch s.category {
	case Grpc:
		return hook.NewServerHookGrpcInterflow(s.network, s.log), nil
	case Kafka:
		return mq.NewServerKafkaInterflow(s.kafka, s.log)
	default:
		return hook.NewServerHookHttpInterflow(s.log), nil
	}
}
