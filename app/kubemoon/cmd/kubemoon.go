package main

import (
	"log"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
)

func main() {
	command := NewKubeMoonCommand()
	if err := command.Execute(); err != nil {
		log.Fatalln(err)
	}
}
