package demo

import (
	sLog "github.com/aide-family/moon/pkg/util/log"
	"github.com/aide-family/moon/pkg/util/types"
	_ "go.uber.org/automaxprocs"

	"github.com/aide-family/moon/cmd/server/demo/internal/democonf"
	"github.com/aide-family/moon/cmd/server/demo/internal/server"
	"github.com/aide-family/moon/pkg/env"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
)

func newApp(srv *server.Server, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(env.ID()),
		kratos.Name(env.Name()),
		kratos.Version(env.Version()),
		kratos.Metadata(env.Metadata()),
		kratos.Logger(logger),
		kratos.Server(srv.GetServers()...),
	)
}

func Run(flagconf string) {
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); !types.IsNil(err) {
		panic(err)
	}

	var bc democonf.Bootstrap
	if err := c.Scan(&bc); !types.IsNil(err) {
		panic(err)
	}

	env.SetName(bc.GetServer().GetName())
	env.SetMetadata(bc.GetServer().GetMetadata())
	env.SetEnv(bc.GetEnv())

	logger := sLog.GetLogger()
	log.SetLogger(logger)
	app, cleanup, err := wireApp(&bc, logger)
	if !types.IsNil(err) {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err = app.Run(); !types.IsNil(err) {
		panic(err)
	}
}
