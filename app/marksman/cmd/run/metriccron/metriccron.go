// Package metriccron is the metric-cron command for the marksman service
package metriccron

import (
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd"
	"github.com/aide-family/marksman/cmd/run"
)

const cmdMetricCronLong = `Start the marksman metric cron service only`

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "metriccron",
		Short: "Start the marksman metric cron service only",
		Long:  cmdMetricCronLong,
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
