// Package all is the all command for the Rabbit service
package all

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/rabbit/cmd"
	"github.com/aide-family/rabbit/cmd/run"
)

const cmdAllLong = `Start the rabbit service with all services (HTTP, gRPC)`

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "all",
		Short: "Start the rabbit service with all services (HTTP, gRPC) and bind Swagger and Metrics",
		Long:  cmdAllLong,
		Annotations: map[string]string{
			"group": cmd.ServiceCommands,
		},
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
