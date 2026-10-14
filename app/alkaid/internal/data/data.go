package data

import (
	"fmt"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData)

type Data struct {
	AliothGRPC string
}

func NewData(c *conf.Alioth) (*Data, func(), error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, nil, fmt.Errorf("alioth grpc address is empty")
	}
	return &Data{AliothGRPC: c.Grpc.Addr}, func() {}, nil
}
