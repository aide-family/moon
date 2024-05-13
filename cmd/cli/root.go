package cli

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn("Not Command")
	},
}

func Execute() {
	rootCmd.AddCommand(serverCmd, versionCmd, genCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
