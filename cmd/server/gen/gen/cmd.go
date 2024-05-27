//go:build ignore
// +build ignore

package main

import (
	"flag"

	"github.com/aide-cloud/moon/cmd/server/gen"
)

var (
	datasource string
	drive      string
	outputPath string
)

func init() {
	flag.StringVar(&datasource, "d", "", "datasource")
	flag.StringVar(&drive, "r", "", "drive")
	flag.StringVar(&outputPath, "o", "", "output")
}

func main() {
	flag.Parse()
	gen.Run(datasource, drive, outputPath)
}
