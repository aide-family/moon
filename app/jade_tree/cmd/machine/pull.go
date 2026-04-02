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

func (f *pullFlags) addPersistentFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addPersistentFlags(cmd)
}

func resolvePullEndpoints(args []string, cfg *clientConfig) []string {
	endpoints := append([]string(nil), args...)
	if len(endpoints) == 0 {
		endpoints = append(endpoints, cfg.Endpoints...)
	}
	if len(endpoints) == 0 && strings.TrimSpace(cfg.Endpoint) != "" {
		endpoints = []string{cfg.Endpoint}
	}
	return endpoints
}

func newPullCmd() *cobra.Command {
	flags := &pullFlags{}
	root := &cobra.Command{
		Use:   "pull [endpoint...]",
		Short: "Pull machine information from endpoints",
		Long:  "Print summary per endpoint, or use subcommands cpu|memory|network|disk for details.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			cfg, err := loadClientConfig(flags.ConfigPath)
			if err != nil {
				return err
			}
			endpoints := resolvePullEndpoints(args, cfg)
			if len(endpoints) == 0 {
				return merr.ErrorInvalidArgument("endpoints are required (positional args or configure endpoints/endpoint in ~/.jade_tree/client.yaml)")
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
	flags.addPersistentFlags(root)

	detailKinds := []struct {
		name  string
		short string
	}{
		{"cpu", "Show CPU details for each endpoint"},
		{"memory", "Show memory and swap details for each endpoint"},
		{"network", "Show network details for each endpoint"},
		{"disk", "Show disk and mount details for each endpoint"},
	}
	for _, dk := range detailKinds {
		kind := dk.name
		sub := &cobra.Command{
			Use:   kind + " [endpoint...]",
			Short: dk.short,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := flags.validate(); err != nil {
					return err
				}
				cfg, err := loadClientConfig(flags.ConfigPath)
				if err != nil {
					return err
				}
				endpoints := resolvePullEndpoints(args, cfg)
				if len(endpoints) == 0 {
					return merr.ErrorInvalidArgument("endpoints are required (positional args or configure endpoints/endpoint in ~/.jade_tree/client.yaml)")
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
				items := make([]machineDetailItem, 0, len(endpoints))
				for _, endpoint := range endpoints {
					ep := strings.TrimSpace(endpoint)
					if ep == "" {
						continue
					}
					info, err := client.getMachineInfo(ctx, ep)
					if err != nil {
						return err
					}
					items = append(items, machineDetailItem{Endpoint: ep, Reply: info})
				}
				return renderMachineDetailsMulti(kind, flags.Output, items)
			},
		}
		root.AddCommand(sub)
	}

	return root
}
