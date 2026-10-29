package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/spf13/cobra"

	"github.com/aide-family/moon/app/megrez/internal/biz/eval"
	"github.com/aide-family/moon/app/megrez/internal/biz/probe"
	"github.com/aide-family/moon/app/megrez/internal/conf"
	"github.com/aide-family/moon/pkg/bootstrap"
)

// version 由构建注入：-ldflags "-X main.version=v1.2.3"。
var version = "dev"

func main() {
	var confPath string
	cmd := &cobra.Command{
		Use:   "megrez",
		Short: "Alioth evaluation worker",
		RunE: func(*cobra.Command, []string) error {
			return serve(confPath)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Run the health server",
		RunE: func(*cobra.Command, []string) error {
			return serve(confPath)
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		Run: func(*cobra.Command, []string) {
			fmt.Println(version)
		},
	})
	cmd.PersistentFlags().StringVar(&confPath, "conf", "configs/config.yaml", "bootstrap config path")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(confPath string) error {
	cfg, err := bootstrap.Load[conf.Bootstrap](confPath, slog.Default())
	if err != nil {
		return err
	}
	logger := log.NewLogger(log.NewHandler(
		log.WithWriter(os.Stdout),
		log.WithFormat(log.FormatJSON),
		log.WithAddSource(true),
	)).With(slog.String("service", "megrez"))
	app, cleanup, err := wireApp(cfg.Server, cfg.Alioth, cfg.Redis, logger)
	if err != nil {
		return err
	}
	defer cleanup()
	return app.Run()
}

func newApp(logger *slog.Logger, httpServer *http.Server, worker *eval.Worker, probeWorker *probe.Worker) *kratos.App {
	return kratos.New(
		kratos.Name("megrez"),
		kratos.Logger(logger),
		kratos.Server(httpServer, worker, probeWorker),
		kratos.StopTimeout(10*time.Second),
	)
}
