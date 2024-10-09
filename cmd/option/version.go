package option

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/spf13/cobra"
)

var logFlag bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long:  `version info`,
	Example: `cmd version
cmd v`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infow("name", "moon cli", "version", "0.0.1")
		showLog()
	},
}

func init() {
	// --log时候显示日志
	versionCmd.Flags().BoolVarP(&logFlag, "log", "l", false, "show log")
}

func showLog() {
	if logFlag {
		log.Info("TODO 增加日志获取逻辑")
		// TODO 增加更新日志
	}
}
