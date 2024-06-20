package option

import (
	"github.com/aide-family/moon/cmd/server/gen"

	"github.com/spf13/cobra"
)

var (
	datasource string
	drive      string
	outputPath string
	isBiz      bool
)

var genCmd = &cobra.Command{
	Use:     "gen",
	Short:   "gen",
	Long:    `gen`,
	Example: `cmd gen`,
	Run: func(cmd *cobra.Command, args []string) {
		outputPath = "./pkg/palace/model/query"
		if isBiz {
			outputPath = "./pkg/palace/model/bizmodel/bizquery"
		}
		gen.Run(datasource, drive, outputPath, isBiz)
	},
}

func init() {
	genCmd.Flags().StringVarP(&datasource, "datasource", "d", "", "datasource")
	genCmd.Flags().StringVarP(&drive, "drive", "r", "", "drive")
	genCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output")
	genCmd.Flags().BoolVarP(&isBiz, "isBiz", "b", false, "isBiz")
}
