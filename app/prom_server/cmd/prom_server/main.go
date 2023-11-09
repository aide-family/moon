package main

import (
	"flag"
	"os"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// Metadata is the metadata of the compiled software.
	Metadata map[string]string
	// flagConf is the config flag.
	flagConf = flag.String("conf", "../../configs", "config path, eg: -conf config.yaml")

	id, _ = os.Hostname()
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
