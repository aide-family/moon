package option

import (
	"github.com/aide-cloud/moon/cmd/server/demo"
	"github.com/aide-cloud/moon/cmd/server/palace"
	"github.com/spf13/cobra"
)

var (
	// flagconf is the config flag.
	flagconf string
	// name is the name of the service.
	name string
)

const (
	ServicePalaceName = "palace"
	ServiceDemoName   = "demo"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "server",
	Long:    `运行moon服务`,
	Example: `cmd server`,
	Run: func(cmd *cobra.Command, args []string) {
		switch name {
		case ServiceDemoName:
			demo.Run(flagconf)
		default:
			palace.Run(flagconf)
		}
	},
}

func init() {
	// conf参数
	serverCmd.Flags().StringVarP(&flagconf, "conf", "c", "./configs", "config path, eg: -conf config.yaml")
	serverCmd.Flags().StringVarP(&name, "name", "n", ServicePalaceName, "name of the service")
}
