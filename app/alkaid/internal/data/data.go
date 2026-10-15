package data

import (
	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewData)

type Data struct {
	AliothGRPC string
}

func NewData(c *conf.Alioth) (*Data, func(), error) {
	if c == nil || c.Grpc == nil || c.Grpc.Addr == "" {
		return nil, nil, merr.Server("alioth grpc address is empty")
	}
	return &Data{AliothGRPC: c.Grpc.Addr}, func() {}, nil
}
