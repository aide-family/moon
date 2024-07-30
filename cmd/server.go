package main

import (
	_ "go.uber.org/automaxprocs"

	"github.com/aide-family/moon/cmd/option"
	"github.com/aide-family/moon/pkg/env"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
)

func main() {
	env.SetVersion(Version)
	option.Execute()

}
