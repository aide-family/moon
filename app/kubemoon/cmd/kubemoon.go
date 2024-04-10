package main

import (
	"github.com/aide-family/moon/app/kubemoon/cmd/apps"
	"log"
)

func main() {
	command := apps.NewKubeMoonCommand()

	if err := command.Execute(); err != nil {
		log.Fatalln(err)
	}
}
