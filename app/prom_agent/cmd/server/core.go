package main

import (
	"sync"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/server"
	"prometheus-manager/pkg/util/hello"
)

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
			hello.SetMetadata(env.GetMetadata())
			hello.SetEnv(env.GetEnv())
		})
		hello.FmtASCIIGenerator()
		return nil
	}
}

func newApp(gs *grpc.Server, hs *http.Server, watch *server.Watch, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(hello.ID()),
		kratos.Name(hello.Name()),
		kratos.Version(hello.Version()),
		kratos.Metadata(hello.Metadata()),
		kratos.Logger(logger),
		kratos.Server(gs, hs, watch),
	)
}
