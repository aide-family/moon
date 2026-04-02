package machine

import (
	"strings"
	"time"

	"github.com/aide-family/magicbox/merr"
	"github.com/spf13/cobra"
)

type machineCommonFlags struct {
	Output     string
	Timeout    time.Duration
	JWT        string
	ConfigPath string
}

func (f *machineCommonFlags) addFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.Output, "output", "o", "table", "output format: table|json|yaml")
	cmd.Flags().DurationVar(&f.Timeout, "timeout", 15*time.Second, "HTTP request timeout")
	cmd.Flags().StringVar(&f.JWT, "jwt", "", "JWT token for API authentication")
	cmd.Flags().StringVar(&f.ConfigPath, "config", defaultClientConfigPath, "client config path")
}

func (f *machineCommonFlags) validate() error {
	switch strings.ToLower(strings.TrimSpace(f.Output)) {
	case "table", "json", "yaml":
		return nil
	default:
		return merr.ErrorInvalidArgument("output must be one of: table, json, yaml")
	}
}
