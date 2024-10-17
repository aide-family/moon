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
	modelType  string
)

func init() {
	flag.StringVar(&datasource, "d", "", "datasource")
	flag.StringVar(&drive, "r", "mysql", "drive")
	flag.StringVar(&outputPath, "o", "./pkg/helper/model/query", "output")
	flag.StringVar(&modelType, "m", "main", "model type")
}

func main() {
	flag.Parse()
	gen.Run(datasource, drive, modelType)
}
