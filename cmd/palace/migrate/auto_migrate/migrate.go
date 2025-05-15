package main

import (
	"fmt"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/spf13/cobra"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/conf"
	"github.com/aide-family/moon/pkg/plugin/gorm"
	"github.com/aide-family/moon/pkg/util/load"
)

var cfgPath string
var rootCmd = &cobra.Command{
	Use:   "moon",
	Short: "CLI for managing Moon monitor palace Server",
	Long:  `The Moon Server CLI provides a command-line interface for managing and interacting with the Moon monitor palace Server service.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the moon palace service from Moon Monitor!")
		run(cfgPath)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "conf", "c", "./cmd/palace/config", "Path to the configuration files")
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

	mainDB, err := gorm.NewDB(bc.GetData().GetMain())
	if err != nil {
		panic(err)
	}

	tableModels := append(system.Models(), team.Models()...)
	if err := mainDB.GetDB().AutoMigrate(tableModels...); err != nil {
		panic(err)
	}
}
