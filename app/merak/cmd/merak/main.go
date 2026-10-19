package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/http"
	"github.com/spf13/cobra"

	"github.com/aide-family/Alioth/app/merak/internal/biz/notify"
	"github.com/aide-family/Alioth/app/merak/internal/conf"
	"github.com/aide-family/Alioth/app/merak/internal/data"
	"github.com/aide-family/Alioth/pkg/bootstrap"
	"github.com/aide-family/Alioth/pkg/merr"
)

func main() {
	var confPath string
	cmd := &cobra.Command{
		Use:   "merak",
		Short: "Alioth notification worker",
		RunE: func(*cobra.Command, []string) error {
			return serve(confPath)
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Run the health server and notification worker",
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
	logger := log.NewLogger(log.NewHandler(
		log.WithWriter(os.Stdout),
		log.WithFormat(log.FormatJSON),
		log.WithAddSource(true),
	)).With(slog.String("service", "merak"))
	app, cleanup, err := wireApp(bc.Server, bc.Alioth, bc.Redis, bc.Notify, logger)
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
		slog.Default().Error(fmt.Sprintf("config: %v", err))
		return nil, merr.Server("config load failed")
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		slog.Default().Error(fmt.Sprintf("config: %v", err))
		return nil, merr.Server("config load failed")
	}
	return &bc, nil
}

func newApp(logger *slog.Logger, hs *http.Server, worker *notify.Worker, d *data.Data) *kratos.App {
	if d == nil {
		panic("data layer is missing")
	}
	return kratos.New(
		kratos.Name("merak"),
		kratos.Logger(logger),
		kratos.Server(hs, worker),
		kratos.StopTimeout(10*time.Second),
	)
}
