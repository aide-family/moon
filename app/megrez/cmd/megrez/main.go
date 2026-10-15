package main

import (
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/spf13/cobra"

	"github.com/aide-family/Alioth/app/megrez/internal/conf"
	"github.com/aide-family/Alioth/app/megrez/internal/data"
	"github.com/aide-family/Alioth/pkg/bootstrap"
	"github.com/aide-family/Alioth/pkg/merr"
)

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
	logger := log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller, "service", "megrez")
	app, cleanup, err := wireApp(bc.Server, logger)
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
		log.NewHelper(log.DefaultLogger).Errorf("config: %v", err)
		return nil, merr.Server("config load failed")
	}
	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		log.NewHelper(log.DefaultLogger).Errorf("config: %v", err)
		return nil, merr.Server("config load failed")
	}
	return &bc, nil
}

func newApp(logger log.Logger, hs *http.Server, d *data.Data) *kratos.App {
	if d == nil {
		panic("data layer is missing")
	}
	return kratos.New(
		kratos.Name("megrez"),
		kratos.Logger(logger),
		kratos.Server(hs),
	)
}
