// Package grpc is the grpc command for the Rabbit service
package grpc

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd"
	"github.com/aide-family/marksman/cmd/run"
)

const cmdGRPCLong = `Start the marksman gRPC service only`

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "grpc",
		Short: "Start the marksman gRPC service only",
		Long:  cmdGRPCLong,
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
