package machine

import (
	"context"
	"net/http"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/spf13/cobra"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

type pushFlags struct {
	machineCommonFlags
	From     string
	PageSize int32
}

func (f *pushFlags) addFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addFlags(cmd)
	cmd.Flags().StringVar(&f.From, "from", defaultEndpoint, "source jade_tree endpoint")
	cmd.Flags().Int32Var(&f.PageSize, "page-size", 100, "page size when listing source machine info from endpoints")
}

func newPushCmd() *cobra.Command {
	flags := &pushFlags{}
	cmd := &cobra.Command{
		Use:   "push [endpoint...]",
		Short: "Push local and stored machine information to endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			cfg, err := loadClientConfig(flags.ConfigPath)
			if err != nil {
				return err
			}
			from := strings.TrimSpace(flags.From)
			if from == "" {
				return merr.ErrorInvalidArgument("from endpoint is required (use --from or configure endpoint in ~/.jade_tree/client.yaml)")
			}
			endpoints := args
			if len(endpoints) == 0 {
				endpoints = []string{cfg.Endpoint}
			}
			if len(endpoints) == 0 {
				return merr.ErrorInvalidArgument("endpoints are required (use --endpoints or configure endpoints in ~/.jade_tree/client.yaml)")
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
			machines, err := fetchAllMachines(ctx, client, from, flags.PageSize)
			if err != nil {
				return err
			}
			payload := &apiv1.ReportMachineInfosRequest{Machines: machines}

			for _, endpoint := range endpoints {
				ep := strings.TrimSpace(endpoint)
				if ep == "" {
					continue
				}
				if err := client.reportMachineInfos(ctx, ep, payload); err != nil {
					return err
				}
			}
			return renderRows(flags.Output, toMachineRows(from, machines))
		},
	}
	flags.addFlags(cmd)
	return cmd
}
