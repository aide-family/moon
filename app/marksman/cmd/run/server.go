// Package run is the run command for the Rabbit service
package run

import (
	"strings"
	"sync"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config/env"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/spf13/cobra"

	"github.com/aide-family/magicbox/hello"
	"github.com/aide-family/magicbox/strutil"
	"github.com/aide-family/marksman/internal/conf"
	"github.com/aide-family/marksman/internal/data"
	"github.com/aide-family/marksman/internal/server"
)

const cmdRunLong = `Run the marksman services`

func NewCmd(defaultServerConfigBytes []byte) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the marksman services",
		Long:  cmdRunLong,
	}
	var bc conf.Bootstrap
	if err := conf.Load(&bc, env.NewSource(), conf.NewBytesSource(defaultServerConfigBytes)); err != nil {
		klog.Errorw("msg", "load config failed", "error", err)
		panic(err)
	}
	runFlags.addFlags(runCmd, &bc)

	return runCmd
}

func NewEndpoint(wireApp WireAppFunc) *endpoint {
	return &endpoint{
		serviceName: strings.Join([]string{runFlags.Name, runFlags.Server.Name}, "."),
		wireAppFunc: wireApp,
	}
}

func NewEngine(endpoints ...*endpoint) *Engine {
	return &Engine{
		endpoints:   endpoints,
		beforeFuncs: []func(...bool){hello.Hello},
		afterFuncs:  []func(){},
	}
}

type WireAppFunc func(serviceName string, bc *conf.Bootstrap, helper *klog.Helper) ([]*kratos.App, func(), error)

type endpoint struct {
	serviceName string
	wireAppFunc WireAppFunc
	apps        []*kratos.App
	cleanup     func()
	helper      *klog.Helper
	err         error
}

type Engine struct {
	endpoints   []*endpoint
	beforeFuncs []func(...bool)
	afterFuncs  []func()
}

func (e *Engine) AddAfterFunc(afterFunc func()) *Engine {
	e.afterFuncs = append(e.afterFuncs, afterFunc)
	return e
}

func (e *Engine) AddBeforeFunc(beforeFunc func(...bool)) *Engine {
	e.beforeFuncs = append(e.beforeFuncs, beforeFunc)
	return e
}

func (e *Engine) init() *Engine {
	serverConf := runFlags.GetServer()
	envOpts := []hello.Option{
		hello.WithVersion(runFlags.Version),
		hello.WithID(runFlags.Hostname),
		hello.WithEnv(runFlags.Environment.String()),
		hello.WithMetadata(serverConf.GetMetadata()),
		hello.WithName(serverConf.GetName()),
	}
	if strings.EqualFold(runFlags.GetUseRandomID(), "true") {
		envOpts = append(envOpts, hello.WithID(strutil.RandomID()))
	}
	hello.SetEnvWithOption(envOpts...)
	return e
}

func (e *Engine) Start() {
	e.init()
	wg := new(sync.WaitGroup)
	for _, endpoint := range e.endpoints {
		endpoint.init()
	}
	for _, beforeFunc := range e.beforeFuncs {
		beforeFunc()
	}
	for _, endpoint := range e.endpoints {
		endpoint.start(wg)
	}
	wg.Wait()
	for _, endpoint := range e.endpoints {
		endpoint.Cleanup()
	}
	for _, afterFunc := range e.afterFuncs {
		afterFunc()
	}
}

func (e *endpoint) init() {
	e.helper = klog.NewHelper(klog.With(klog.GetLogger(),
		"service.name", e.serviceName,
		"service.id", hello.ID(),
		"caller", klog.DefaultCaller,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID()),
	)
	e.apps, e.cleanup, e.err = e.wireAppFunc(e.serviceName, runFlags.Bootstrap, e.helper)
}

func (e *endpoint) start(wg *sync.WaitGroup) {
	if e.err != nil {
		e.helper.Errorw("msg", "endpoint init failed", "error", e.err)
		return
	}
	for _, app := range e.apps {
		appName := app.Name()
		_app := app
		wg.Go(func() {
			if err := _app.Run(); err != nil {
				e.helper.Errorf("app [%s] run failed, error: %v", appName, err)
				return
			}
		})
	}
}

func (e *endpoint) Cleanup() {
	if e.cleanup == nil {
		return
	}
	e.cleanup()
}

func NewApp(serviceName string, d *data.Data, srvs server.Servers, bc *conf.Bootstrap, helper *klog.Helper) ([]*kratos.App, error) {
	apps := make([]*kratos.App, 0, len(srvs))
	if len(srvs) == 0 {
		panic("no servers")
	}

	for _, srv := range srvs {
		opts := []kratos.Option{
			kratos.Name(strings.Join([]string{serviceName, srv.Name()}, ".")),
			kratos.ID(hello.ID()),
			kratos.Version(hello.Version()),
			kratos.Metadata(hello.Metadata()),
			kratos.Logger(helper.Logger()),
			kratos.Server(srv.Instance()),
		}

		if registry := d.Registry(); registry != nil {
			opts = append(opts, kratos.Registrar(registry))
		}

		if srvName := srv.Name(); srvName == "http" {
			instance := srv.Instance()
			httpSrv, ok := instance.(*http.Server)
			if !ok {
				panic("server instance is not a *http.Server")
			}
			server.BindSwagger(httpSrv, bc)
			server.BindMetrics(httpSrv, bc)
		}

		apps = append(apps, kratos.New(opts...))
	}
	return apps, nil
}
