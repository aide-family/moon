//go:build ignore
// +build ignore

package main

import (
	"flag"

	"github.com/aide-family/moon/cmd/server/gen"
)

var (
	datasource string
	drive      string
	outputPath string
	modelType  int
)

func init() {
	flag.StringVar(&datasource, "d", "", "datasource")
	flag.StringVar(&drive, "r", "mysql", "drive")
	flag.StringVar(&outputPath, "o", "./pkg/helper/model/query", "output")
	flag.IntVar(&modelType, "m", 1, "model type")
}

func main() {
	flag.Parse()
	gen.Run(datasource, drive, modelType)
}
