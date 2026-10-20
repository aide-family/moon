package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/spf13/cobra"

	agentbiz "github.com/aide-family/Alioth/app/alkaid/internal/biz/agent"
	"github.com/aide-family/Alioth/app/alkaid/internal/conf"
	"github.com/aide-family/Alioth/app/alkaid/internal/data"
	"github.com/aide-family/Alioth/pkg/bootstrap"
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
	bc, err := bootstrap.Load[conf.Bootstrap](confPath, slog.Default())
	if err != nil {
		return err
	}
	logger := log.NewLogger(log.NewHandler(
		log.WithWriter(os.Stdout),
		log.WithFormat(log.FormatJSON),
		log.WithAddSource(true),
	)).With(slog.String("service", "alkaid"))
	app, cleanup, err := wireApp(bc.Server, bc.Alioth, logger)
	if err != nil {
		return err
	}
	defer cleanup()
	return app.Run()
}

func newApp(logger *slog.Logger, hs *http.Server, gs *grpc.Server, connector *agentbiz.Connector, d *data.Data) *kratos.App {
	logger.Info(fmt.Sprintf("alioth grpc endpoint configured: %s", d.AliothGRPC))
	servers := []transport.Server{hs, gs}
	if connector != nil {
		servers = append(servers, connector)
	}
	return kratos.New(
		kratos.Name("alkaid"),
		kratos.Logger(logger),
		kratos.Server(servers...),
		kratos.StopTimeout(10*time.Second),
	)
}
