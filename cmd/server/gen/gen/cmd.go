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
	isBiz      bool
)

func init() {
	flag.StringVar(&datasource, "d", "", "datasource")
	flag.StringVar(&drive, "r", "mysql", "drive")
	flag.StringVar(&outputPath, "o", "./pkg/helper/model/query", "output")
	flag.BoolVar(&isBiz, "b", false, "is biz model")
}

func main() {
	flag.Parse()
	outputPath = "./pkg/palace/model/query"
	if isBiz {
		outputPath = "./pkg/palace/model/bizmodel/bizquery"
	}
	gen.Run(datasource, drive, outputPath, isBiz)
}
