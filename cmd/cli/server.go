package cli

import (
	"github.com/aide-cloud/moon/cmd/moon"
	"github.com/spf13/cobra"
)

// flagconf is the config flag.
var flagconf string

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "server",
	Long:    `运行moon服务`,
	Example: `cmd server`,
	Run: func(cmd *cobra.Command, args []string) {
		moon.Run(flagconf)
	},
}

func init() {
	// conf参数
	serverCmd.Flags().StringVarP(&flagconf, "conf", "c", "./configs", "config path, eg: -conf config.yaml")
}
