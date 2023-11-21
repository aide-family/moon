package main

import (
	"sync"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"prometheus-manager/pkg/util/hello"

	"prometheus-manager/app/prom_server/internal/conf"
)

var (
	once            sync.Once
	ProviderSetCore = wire.NewSet(
		before,
	)
)

func before() conf.Before {
	return func(bc *conf.Bootstrap) error {
		once.Do(func() {
			hello.SetName(bc.GetEnv().GetName())
			hello.SetVersion(Version)
			hello.SetMetadata(bc.GetEnv().GetMetadata())
		})
		hello.FmtASCIIGenerator()
		return nil
	}
}

func newApp(gs *grpc.Server, hs *http.Server, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(hello.ID()),
		kratos.Name(hello.Name()),
		kratos.Version(hello.Version()),
		kratos.Metadata(hello.Metadata()),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	)
}
