package data

import (
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewData)

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
