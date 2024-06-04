//go:build ignore
// +build ignore

package main

import (
	"flag"

	"github.com/aide-family/moon/cmd/server/demo"
	"github.com/aide-family/moon/pkg/env"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// flagconf is the config flag.
	flagconf string

	// Version is the version of the compiled software.
	Version string
)

func init() {
	flag.StringVar(&flagconf, "c", "../configs", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()
	env.SetVersion(Version)
	demo.Run(flagconf)
}
