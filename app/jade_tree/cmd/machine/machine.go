package machine

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	machineCmd := &cobra.Command{
		Use:   "machine",
		Short: "Machine information operations",
	}
	machineCmd.AddCommand(newInfoCmd(), newPullCmd(), newPushCmd())
	return machineCmd
}
