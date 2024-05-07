package main

import (
	"sync"

	"github.com/aide-family/moon/app/prom_server/internal/conf"
	"github.com/aide-family/moon/app/prom_server/internal/server"
	"github.com/aide-family/moon/pkg/util/hello"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	//增加这段代码
	json.MarshalOptions = protojson.MarshalOptions{
		UseEnumNumbers: true, // 将枚举值作为数字发出，默认为枚举值的字符串
	}
}

var (
	once            sync.Once
	ProviderSetCore = wire.NewSet(
		before,
	)
)

func before() conf.Before {
	return func(bc *conf.Bootstrap) error {
		env := bc.GetEnv()
		once.Do(func() {
			hello.SetName(env.GetName())
			hello.SetVersion(Version)
			hello.SetEnv(env.GetEnv())
			hello.SetMetadata(env.GetMetadata())
			hello.FmtASCIIGenerator()
		})
		return nil
	}
}

func newApp(s *server.Server, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(hello.ID()),
		kratos.Name(hello.Name()),
		kratos.Version(hello.Version()),
		kratos.Metadata(hello.Metadata()),
		kratos.Logger(logger),
		kratos.Server(s.List()...),
	)
}
