package main

import (
	"fmt"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/moon/cmd/houyi/internal/conf"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/i18n"
	mlog "github.com/aide-family/moon/pkg/log"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/aide-family/moon/pkg/util/load"
)

// Version is the version of the compiled software.
var Version string
var cfgPath string
var rootCmd = &cobra.Command{
	Use:   "moon",
	Short: "CLI for managing Moon monitor houyi Server",
	Long:  `The Moon Server CLI provides a command-line interface for managing and interacting with the Moon monitor houyi Server service.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the moon houyi service from Moon Monitor!")
		run(cfgPath)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "conf", "c", "./cmd/houyi/config", "Path to the configuration files")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(cfgPath string) {
	var bc conf.Bootstrap
	if err := load.Load(cfgPath, &bc); err != nil {
		panic(err)
	}
	hello.SetEnvWithConfig(Version, bc.GetEnvironment(), bc.GetServer())
	i18nConf := bc.GetI18N()
	bundle := i18n.New(i18nConf)
	i18n.RegisterGlobalLocalizer(i18n.NewLocalizer(bundle))

	logger := mlog.New(bc.IsDev(), bc.GetLog())

	app, cleanup, err := wireApp(&bc, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func newApp(c *conf.Bootstrap, srvs server.Servers, logger log.Logger) *kratos.App {
	opts := hello.WithKratosOption(
		c.GetRegistry(),
		kratos.Logger(logger),
		kratos.Server(srvs...),
	)

	return kratos.New(opts...)
}
