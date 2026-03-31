package all

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/jade_tree/cmd/run"
)

const cmdAllLong = `Start the Jade Tree service with all services (HTTP, gRPC)`

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "all",
		Short: "Start the Jade Tree service with all services (HTTP, gRPC)",
		Long:  cmdAllLong,
		Run: func(_ *cobra.Command, _ []string) {
			if err := flags.applyToBootstrap(); err != nil {
				klog.Errorw("msg", "apply to bootstrap failed", "error", err)
				return
			}
			run.NewEngine(run.NewEndpoint(WireApp)).Start()
		},
	}
	flags.addFlags(runCmd)
	return runCmd
}
