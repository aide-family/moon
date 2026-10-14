package main

import (
	"fmt"
	"os"

	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/app/alkaid/internal/data"
	"github.com/aide-family/Alioth/pkg/bootstrap"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/spf13/cobra"
)

func main() {
	var confPath string
	cmd := &cobra.Command{
		Use:   "alkaid",
		Short: "Alioth edge agent",
		RunE: func(*cobra.Command, []string) error {
			return serve(confPath)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP and gRPC servers",
		RunE: func(*cobra.Command, []string) error {
			return serve(confPath)
		},
	})
	cmd.PersistentFlags().StringVar(&confPath, "conf", "configs/config.yaml", "bootstrap config path")
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(confPath string) error {
	bc, err := loadConfig(confPath)
	if err != nil {
		return err
	}
	logger := log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller, "service", "alkaid")
	app, cleanup, err := wireApp(bc.Server, bc.Alioth, logger)
	if err != nil {
		return err
	}
	defer cleanup()
	return app.Run()
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
	log.NewHelper(logger).Infof("alioth grpc endpoint configured: %s", d.AliothGRPC)
	return kratos.New(
		kratos.Name("alkaid"),
		kratos.Logger(logger),
		kratos.Server(hs, gs),
	)
}
