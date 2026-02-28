package config

import (
	"os"
	"path/filepath"

	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd"
)

func newServerCmd(defaultServerConfig []byte) *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Generate server config",
		Long:  "Generate a complete server config file from GORM models so each upgrade can be versioned.",
		Annotations: map[string]string{
			"group": cmd.ConfigCommands,
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := os.MkdirAll(output, 0o755); err != nil {
				return err
			}
			filename := filepath.Join(output, "server.yaml")
			if err := os.WriteFile(filename, defaultServerConfig, 0o644); err != nil {
				return err
			}
			klog.Infof("server config generated successfully: %s", filename)
			return nil
		},
	}
	return serverCmd
}
