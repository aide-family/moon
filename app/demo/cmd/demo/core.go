package main

import (
	"sync"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"prometheus-manager/app/demo/internal/conf"
	"prometheus-manager/pkg/hello"
	"prometheus-manager/pkg/plog"
)

var _ plog.ServerEnv = (*core)(nil)

type (
	core struct {
		name     string
		id       string
		version  string
		metadata map[string]string
	}
)

func (l *core) GetId() string {
	return l.id
}

func (l *core) GetName() string {
	return l.name
}

func (l *core) GetVersion() string {
	return l.version
}

func newCore(_ *conf.Bootstrap) plog.ServerEnv {
	return &core{
		name:     Name,
		id:       id,
		version:  Version,
		metadata: Metadata,
	}
}

var (
	once            sync.Once
	ProviderSetCore = wire.NewSet(
		before,
		newCore,
	)
)

func before() conf.Before {
	return func(bc *conf.Bootstrap) error {
		once.Do(func() {
			Name = bc.GetEnv().GetName()
			if Version == "" {
				Version = bc.GetEnv().GetVersion()
			}
			Metadata = bc.GetEnv().GetMetadata()
		})
		hello.FmtASCIIGenerator(Name, Version, bc.GetEnv().GetMetadata())
		return nil
	}
}

func newApp(gs *grpc.Server, hs *http.Server, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(Metadata),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	)
}
