package machine

import (
	"context"
	"net/http"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/spf13/cobra"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

type pullFlags struct {
	machineCommonFlags
	To       string
	PageSize int32
}

func (f *pullFlags) addFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addFlags(cmd)
	cmd.Flags().StringVar(&f.To, "to", defaultEndpoint, "target jade_tree endpoint (pull to)")
	cmd.Flags().Int32Var(&f.PageSize, "page-size", 100, "page size when listing target machine info")
}

func newPullCmd() *cobra.Command {
	flags := &pullFlags{}
	cmd := &cobra.Command{
		Use:   "pull [endpoint...]",
		Short: "Pull machine information from one endpoint and store it into target endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			cfg, err := loadClientConfig(flags.ConfigPath)
			if err != nil {
				return err
			}

			to := strings.TrimSpace(flags.To)
			if to == "" {
				return merr.ErrorInvalidArgument("to endpoint is required (use --to or configure endpoint in ~/.jade_tree/client.yaml)")
			}

			targets := args
			if len(targets) == 0 {
				return merr.ErrorInvalidArgument("target endpoint is required (positional args or configure endpoint in ~/.jade_tree/client.yaml)")
			}

			if flags.PageSize <= 0 {
				return merr.ErrorInvalidArgument("page-size must be greater than 0")
			}

			token := strings.TrimSpace(flags.JWT)
			if token == "" {
				token = strings.TrimSpace(cfg.JWT)
			}

			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			client := newAPIClient(&http.Client{Timeout: flags.Timeout}, token)
			machines := make([]*apiv1.GetMachineInfoReply, 0)
			for _, endpoint := range targets {
				ep := strings.TrimSpace(endpoint)
				if ep == "" {
					continue
				}
				ms, err := fetchAllMachines(ctx, client, ep, flags.PageSize)
				if err != nil {
					return err
				}

				payload := &apiv1.ReportMachineInfosRequest{Machines: ms}
				if err := client.reportMachineInfos(ctx, to, payload); err != nil {
					return err
				}
				machines = append(machines, ms...)
			}

			// Show what we pulled from the source endpoint.
			return renderRows(flags.Output, toMachineRows(to, machines))
		},
	}
	flags.addFlags(cmd)
	return cmd
}
