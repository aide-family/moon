package moon

import (
	sLog "github.com/aide-cloud/moon/pkg/log"
	"github.com/go-kratos/kratos/v2/log"
	_ "go.uber.org/automaxprocs"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-cloud/moon/cmd/moon/internal/conf"
	"github.com/aide-cloud/moon/pkg/env"
)

func newApp(bc *conf.Server, gs *grpc.Server, hs *http.Server, logger log.Logger) *kratos.App {
	env.SetName(bc.GetName())
	env.SetMetadata(bc.GetMetadata())
	return kratos.New(
		kratos.ID(env.ID()),
		kratos.Name(env.Name()),
		kratos.Version(env.Version()),
		kratos.Metadata(env.Metadata()),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func Run(flagconf string) {
	logger := sLog.GetLogger()
	log.SetLogger(logger)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err = app.Run(); err != nil {
		panic(err)
	}
}
