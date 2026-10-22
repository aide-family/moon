package data

import (
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/pkg/healthcheck"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewConnectionSettings, NewReadinessChecker)

// NewReadinessChecker 返回 nil：alkaid 不持有可探测的本地依赖，
// 与 alioth 的连接状态由 Connector 自己的重连循环负责。
func NewReadinessChecker() healthcheck.Checker {
	return nil
}

// ConnectionSettings 保存摇光连接玉衡的参数。alkaid 不连业务库。
type ConnectionSettings struct {
	AliothGRPC string
	AgentToken string
	Heartbeat  time.Duration
}

func NewConnectionSettings(c *conf.Alioth) (*ConnectionSettings, func(), error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, nil, merr.Server("alioth grpc address is empty")
	}
	heartbeat := 15 * time.Second
	if c.HeartbeatInterval != nil {
		heartbeat = c.HeartbeatInterval.AsDuration()
	}
	return &ConnectionSettings{
		AliothGRPC: c.Grpc.Addr,
		AgentToken: c.AgentToken,
		Heartbeat:  heartbeat,
	}, func() {}, nil
}
