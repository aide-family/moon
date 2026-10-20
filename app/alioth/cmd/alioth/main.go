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

func main() {
	root := &cobra.Command{Use: "alioth", Short: "Alioth control plane"}
	root.AddCommand(newServeCmd(), newMigrateCmd())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newServeCmd() *cobra.Command {
	var confPath string
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP and gRPC servers",
		RunE: func(*cobra.Command, []string) error {
			bc, err := bootstrap.Load[conf.Bootstrap](confPath, slog.Default())
			if err != nil {
				return err
			}
			logger := log.NewLogger(log.NewHandler(
				log.WithWriter(os.Stdout),
				log.WithFormat(log.FormatJSON),
				log.WithAddSource(true),
			)).With(slog.String("service", "alioth"))
			app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, logger)
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
			bc, err := bootstrap.Load[conf.Bootstrap](confPath, slog.Default())
			if err != nil {
				return err
			}
			if bc.Data == nil || bc.Data.Database == nil || bc.Data.Database.Source == "" {
				return merr.Server("database source is empty")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return data.Migrate(ctx, bc.Data.Database.Source)
		},
	}
	cmd.Flags().StringVar(&confPath, "conf", "configs/config.yaml", "bootstrap config path")
	return cmd
}

func newApp(logger *slog.Logger, hs *http.Server, gs *grpc.Server, sched *server.Scheduler, hub *agenthub.Hub, d *data.Data) *kratos.App {
	if d == nil {
		panic("database is not connected")
	}
	return kratos.New(
		kratos.Name("alioth"),
		kratos.Logger(logger),
		kratos.Server(hs, gs, sched),
		kratos.StopTimeout(10*time.Second),
		kratos.BeforeStop(func(_ context.Context) error {
			hub.CloseAll()
			return nil
		}),
	)
}
