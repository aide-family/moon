package main

import (
	"flag"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string

	// flagConf is the config flag.
	flagConf = flag.String("conf", "../../configs", "config path, eg: -conf config.yaml")
)

func main() {
	flag.Parse()

	app, cleanup, err := wireApp(flagConf)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err = app.Run(); err != nil {
		panic(err)
	}
}
