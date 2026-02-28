// Package http is the http command for the Rabbit service
package http

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd"
	"github.com/aide-family/marksman/cmd/run"
)

const cmdHTTPLong = `Start the marksman HTTP service only`

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "http",
		Short: "Start the marksman HTTP service only",
		Long:  cmdHTTPLong,
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
