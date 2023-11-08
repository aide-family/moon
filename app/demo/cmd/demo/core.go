package main

import (
	"errors"
	"os"
	"sync"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"prometheus-manager/app/demo/internal/conf"
	"prometheus-manager/pkg/hello"
)

var (
	once sync.Once
)

func loadConfig(flagConf *string) (*conf.Bootstrap, error) {
	if flagConf == nil || *flagConf == "" {
		return nil, errors.New("flagConf is empty")
	}
	c := config.New(
		config.WithSource(
			file.NewSource(*flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}

	return before(&bc), nil
}

func before(bc *conf.Bootstrap) *conf.Bootstrap {
	once.Do(func() {
		Name = bc.GetEnv().GetName()
		if Version == "" {
			Version = bc.GetEnv().GetVersion()
		}
		Metadata = bc.GetEnv().GetMetadata()
	})
	hello.FmtASCIIGenerator(Name, Version, bc.GetEnv().GetMetadata())
	return bc
}

// newLogger new a logger.
func newLogger(c *conf.Log) log.Logger {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	return logger
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(Metadata),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	)
}
