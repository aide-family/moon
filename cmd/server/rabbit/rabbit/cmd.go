package main

import (
	"flag"

	_ "go.uber.org/automaxprocs"

	"github.com/aide-family/moon/pkg/helper"

	"github.com/aide-family/moon/cmd/server/rabbit"
	"github.com/aide-family/moon/pkg/env"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// flagconf is the config flag.
	flagconf string

	// configType is the config file type.
	configType string

	// Version is the version of the compiled software.
	Version string

	// pprofAddress is the pprof address.
	pprofAddress string
)

func init() {
	flag.StringVar(&flagconf, "c", "../configs", "config path, eg: -c ./configs")
	flag.StringVar(&configType, "config_ext", "yaml", "config file ext name, eg: -config_ext yaml")
	flag.StringVar(&pprofAddress, "pprof_address", "", "pprof address, eg: -pprof_address 0.0.0.0:6060")
}

func main() {
	flag.Parse()
	env.SetVersion(Version)
	helper.Pprof(pprofAddress)
	rabbit.Run(flagconf, configType)
}
