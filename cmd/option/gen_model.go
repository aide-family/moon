package option

import (
	"github.com/aide-family/moon/cmd/server/gen"

	"github.com/spf13/cobra"
)

var (
	datasource string
	drive      string
	outputPath string
	modelType  string
)

var genCmd = &cobra.Command{
	Use:     "gen",
	Short:   "gen",
	Long:    `gen`,
	Example: `cmd gen`,
	Run: func(cmd *cobra.Command, args []string) {
		gen.Run(datasource, drive, modelType)
	},
}

func init() {
	genCmd.Flags().StringVarP(&datasource, "datasource", "d", "", "datasource")
	genCmd.Flags().StringVarP(&drive, "drive", "r", "", "drive")
	genCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output")
	genCmd.Flags().StringVarP(&modelType, "modelType", "m", "main", "model type")
}
