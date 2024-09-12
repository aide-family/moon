package option

import (
	"github.com/aide-family/moon/cmd/server/demo"
	"github.com/aide-family/moon/cmd/server/houyi"
	"github.com/aide-family/moon/cmd/server/palace"
	"github.com/aide-family/moon/cmd/server/rabbit"

	"github.com/spf13/cobra"
)

var (
	// flagconf is the config flag.
	flagconf string
	// configType is the config file type.
	configType string
	// name is the name of the service.
	name string
)

const (
	servicePalaceName = "palace"
	serviceDemoName   = "demo"
	serviceRabbitName = "rabbit"
	serviceHouYiName  = "houyi"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "server",
	Long:    `运行moon服务`,
	Example: `cmd server`,
	Run: func(cmd *cobra.Command, args []string) {
		switch name {
		case serviceDemoName:
			demo.Run(flagconf)
		case serviceRabbitName:
			rabbit.Run(flagconf, configType)
		case serviceHouYiName:
			houyi.Run(flagconf, configType)
		default:
			palace.Run(flagconf, configType)
		}
	},
}

func init() {
	// conf参数
	serverCmd.Flags().StringVarP(&flagconf, "conf", "c", "./configs", "config path, eg: -conf ./configs")
	serverCmd.Flags().StringVarP(&name, "name", "n", servicePalaceName, "name of the service")
	serverCmd.Flags().StringVarP(&configType, "config_ext", "ct", "yaml", "config file ext name, eg: -config_ext yaml")
}
