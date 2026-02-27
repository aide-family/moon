package config

import (
	"github.com/spf13/cobra"

	"github.com/aide-family/rabbit/cmd"
)

var output = "deploy/configs/"

func NewCmd(defaultServerConfig []byte) *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Generate config from default config",
		Long:  "Generate a complete config file from default config",
		Annotations: map[string]string{
			"group": cmd.ConfigCommands,
		},
	}
	configCmd.PersistentFlags().StringVarP(&output, "output", "o", output, "Output config file path")
	configCmd.AddCommand(newServerCmd(defaultServerConfig))
	return configCmd
}
