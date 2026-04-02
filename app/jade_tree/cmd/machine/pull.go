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
}

func (f *pullFlags) addFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addFlags(cmd)
}

func newPullCmd() *cobra.Command {
	flags := &pullFlags{}
	cmd := &cobra.Command{
		Use:   "pull [endpoint...]",
		Short: "Pull machine information from endpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			cfg, err := loadClientConfig(flags.ConfigPath)
			if err != nil {
				return err
			}
			endpoints := args
			if len(endpoints) == 0 {
				endpoints = cfg.Endpoints
			}
			if len(endpoints) == 0 {
				return merr.ErrorInvalidArgument("endpoints are required (use --endpoints or configure endpoints in ~/.jade_tree/client.yaml)")
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
			out := make([]machineRow, 0)
			for _, endpoint := range endpoints {
				ep := strings.TrimSpace(endpoint)
				if ep == "" {
					continue
				}
				info, err := client.getMachineInfo(ctx, ep)
				if err != nil {
					return err
				}
				out = append(out, toMachineRows(ep, []*apiv1.GetMachineInfoReply{info})...)
			}
			return renderRows(flags.Output, out)
		},
	}
	flags.addFlags(cmd)
	return cmd
}
