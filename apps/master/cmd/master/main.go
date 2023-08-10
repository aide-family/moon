package main

import (
	"flag"
	"os"
	"sync"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"

	"prometheus-manager/pkg/hello"
	"prometheus-manager/pkg/servers"

	"prometheus-manager/apps/master/internal/conf"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
	once  sync.Once
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(
	env *conf.Env,
	logger log.Logger,
	gs *grpc.Server,
	hs *http.Server,
	ts *servers.Timer,
) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(env.GetName()),
		kratos.Version(env.GetVersion()),
		kratos.Metadata(env.GetMetadata()),
		kratos.Logger(logger),
		kratos.Server(gs, hs, ts),
	)
}

func Init(bc *conf.Bootstrap) *conf.Bootstrap {
	once.Do(func() {
		Name = bc.GetEnv().GetName()
		if Version == "" {
			Version = bc.GetEnv().GetVersion()
		}
	})
	hello.FmtASCIIGenerator(Name, Version, bc.GetEnv().GetMetadata())
	return bc
}

func main() {
	flag.Parse()

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

	conf.Set(Init(&bc))

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	app, cleanup, err := wireApp(&bc, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
