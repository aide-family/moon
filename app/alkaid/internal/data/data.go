package data

import (
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/pkg/healthcheck"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewData, NewReadiness)

// NewReadiness 返回 nil：alkaid 不持有可探测的本地依赖，
// 与 alioth 的连接状态由 Connector 自己的重连循环负责。
func NewReadiness() healthcheck.Checker {
	return nil
}

// Data 保存出站连接配置。alkaid 不连业务库。
type Data struct {
	AliothGRPC string
	AgentToken string
	Heartbeat  time.Duration
}

func NewData(c *conf.Alioth) (*Data, func(), error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, nil, merr.Server("alioth grpc address is empty")
	}
	heartbeat := 15 * time.Second
	if c.HeartbeatInterval != nil {
		heartbeat = c.HeartbeatInterval.AsDuration()
	}
	return &Data{
		AliothGRPC: c.Grpc.Addr,
		AgentToken: c.AgentToken,
		Heartbeat:  heartbeat,
	}, func() {}, nil
}
