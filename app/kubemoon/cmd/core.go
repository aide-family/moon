package main

import (
	"sync"

	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/aide-family/moon/app/kubemoon/internal/server"
	"github.com/aide-family/moon/pkg/util/hello"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
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
			hello.SetEnv(env.GetEnv())
			hello.SetMetadata(env.GetMetadata())
			hello.FmtASCIIGenerator()
		})
		return nil
	}
}

func newApp(hs *http.Server, ks *server.KubeServer, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(hello.ID()),
		kratos.Name(hello.Name()),
		kratos.Version(hello.Version()),
		kratos.Metadata(hello.Metadata()),
		kratos.Logger(logger),
		kratos.Server(hs, ks),
	)
}
