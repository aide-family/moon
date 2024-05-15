package main

import (
	"github.com/aide-cloud/moon/cmd/option"
	"github.com/aide-cloud/moon/pkg/env"
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
