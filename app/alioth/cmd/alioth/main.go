package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data"
	"github.com/aide-family/Alioth/pkg/bootstrap"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/spf13/cobra"
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
			bc, err := loadConfig(confPath)
			if err != nil {
				return err
			}
			logger := log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller, "service", "alioth")
			app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
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
			bc, err := loadConfig(confPath)
			if err != nil {
				return err
			}
			if bc.Data == nil || bc.Data.Database == nil || bc.Data.Database.Source == "" {
				return fmt.Errorf("database source is empty")
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return data.Migrate(ctx, bc.Data.Database.Source)
		},
	}
	cmd.Flags().StringVar(&confPath, "conf", "configs/config.yaml", "bootstrap config path")
	return cmd
}

func loadConfig(path string) (*conf.Bootstrap, error) {
	c := config.New(config.WithSource(bootstrap.NewSource(path)))
	defer c.Close()
	if err := c.Load(); err != nil {
		return nil, err
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, d *data.Data) *kratos.App {
	if d == nil {
		panic("database is not connected")
	}
	return kratos.New(
		kratos.Name("alioth"),
		kratos.Logger(logger),
		kratos.Server(hs, gs),
	)
}
