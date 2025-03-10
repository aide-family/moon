package main

import (
	"flag"

	_ "go.uber.org/automaxprocs"

	"github.com/aide-family/moon/cmd/server/jadeTree"

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
)

func init() {
	flag.StringVar(&flagconf, "c", "../configs", "config path, eg: -c ./configs")
	flag.StringVar(&configType, "config_ext", "yaml", "config file ext name, eg: -config_ext yaml")
}

func main() {
	flag.Parse()
	env.SetVersion(Version)
	jadeTree.Run(flagconf, configType)
}
