package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/spf13/cobra"

	"github.com/aide-family/Alioth/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data"
	"github.com/aide-family/Alioth/app/alioth/internal/server"
	"github.com/aide-family/Alioth/pkg/bootstrap"
	"github.com/aide-family/Alioth/pkg/merr"
)

// version 由构建注入：-ldflags "-X main.version=v1.2.3"。
var version = "dev"

func main() {
	root := &cobra.Command{Use: "alioth", Short: "Alioth control plane"}
	root.AddCommand(newServeCmd(), newMigrateCmd(), newVersionCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		Run: func(*cobra.Command, []string) {
			fmt.Println(version)
		},
	}
}

func newServeCmd() *cobra.Command {
	var confPath string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP and gRPC servers",
		RunE: func(*cobra.Command, []string) error {
			cfg, err := bootstrap.Load[conf.Bootstrap](confPath, slog.Default())
			if err != nil {
				return err
			}
			logger := log.NewLogger(log.NewHandler(
				log.WithWriter(os.Stdout),
				log.WithFormat(log.FormatJSON),
				log.WithAddSource(true),
			)).With(slog.String("service", "alioth"))
			app, cleanup, err := wireApp(cfg.Server, cfg.Data, cfg.Auth, logger)
			if err != nil {
				return err
			}
			defer cleanup()
			return app.Run()
		},
	}
	cmd.Flags().StringVar(&confPath, "conf", "configs/config.yaml", "bootstrap config path")
	return cmd
}

func newMigrateCmd() *cobra.Command {
	var confPath string
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply the ent schema to PostgreSQL",
		RunE: func(*cobra.Command, []string) error {
			cfg, err := bootstrap.Load[conf.Bootstrap](confPath, slog.Default())
			if err != nil {
				return err
			}
			if cfg.Data == nil || cfg.Data.Database == nil || cfg.Data.Database.Source == "" {
				return merr.Server("database source is empty")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return data.Migrate(ctx, cfg.Data.Database.Source)
		},
	}
	cmd.Flags().StringVar(&confPath, "conf", "configs/config.yaml", "bootstrap config path")
	return cmd
}

func newApp(logger *slog.Logger, httpServer *http.Server, grpcServer *grpc.Server, scheduler *server.Scheduler, hub *agenthub.Hub, deps *data.Dependencies) *kratos.App {
	if deps == nil {
		panic("database is not connected")
	}
	return kratos.New(
		kratos.Name("alioth"),
		kratos.Logger(logger),
		kratos.Server(httpServer, grpcServer, scheduler),
		kratos.StopTimeout(10*time.Second),
		kratos.BeforeStop(func(_ context.Context) error {
			hub.CloseAll()
			return nil
		}),
	)
}
