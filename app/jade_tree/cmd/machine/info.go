package machine

import (
	"context"
	"net/http"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/spf13/cobra"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

type infoFlags struct {
	machineCommonFlags
	Endpoint string
}

func (f *infoFlags) addPersistentFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addPersistentFlags(cmd)
	cmd.PersistentFlags().StringVar(&f.Endpoint, "endpoint", defaultEndpoint, "jade_tree HTTP endpoint")
}

func newInfoCmd() *cobra.Command {
	flags := &infoFlags{}
	root := &cobra.Command{
		Use:   "info",
		Short: "Get local machine information",
		Long:  "Print summary machine info, or use subcommands cpu|memory|network|disk for details.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			cfg, err := loadClientConfig(flags.ConfigPath)
			if err != nil {
				return err
			}
			endpoint := strings.TrimSpace(flags.Endpoint)
			if endpoint == "" {
				endpoint = strings.TrimSpace(cfg.Endpoint)
			}
			if endpoint == "" {
				return merr.ErrorInvalidArgument("endpoint is required (use --endpoint or configure endpoint in ~/.jade_tree/client.yaml)")
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
			info, err := client.getMachineInfo(ctx, endpoint)
			if err != nil {
				return err
			}
			return renderRows(flags.Output, toMachineRows(endpoint, []*apiv1.GetMachineInfoReply{info}))
		},
	}
	flags.addPersistentFlags(root)

	detailKinds := []struct {
		name  string
		short string
	}{
		{"cpu", "Show CPU details (processors, cores, threads)"},
		{"memory", "Show memory and swap details"},
		{"network", "Show network interfaces, IPs, and traffic counters"},
		{"disk", "Show disks and mount points"},
	}
	for _, dk := range detailKinds {
		kind := dk.name
		sub := &cobra.Command{
			Use:   kind,
			Short: dk.short,
			RunE: func(cmd *cobra.Command, _ []string) error {
				if err := flags.validate(); err != nil {
					return err
				}
				cfg, err := loadClientConfig(flags.ConfigPath)
				if err != nil {
					return err
				}
				endpoint := strings.TrimSpace(flags.Endpoint)
				if endpoint == "" {
					endpoint = strings.TrimSpace(cfg.Endpoint)
				}
				if endpoint == "" {
					return merr.ErrorInvalidArgument("endpoint is required (use --endpoint or configure endpoint in ~/.jade_tree/client.yaml)")
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
				info, err := client.getMachineInfo(ctx, endpoint)
				if err != nil {
					return err
				}
				return renderMachineDetail(kind, flags.Output, endpoint, info)
			},
		}
		root.AddCommand(sub)
	}

	return root
}
