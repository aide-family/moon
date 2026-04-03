package machine

import (
	"context"
	"net/http"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/spf13/cobra"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

type getFlags struct {
	machineCommonFlags
	Endpoint string
}

func (f *getFlags) addPersistentFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addPersistentFlags(cmd)
	cmd.PersistentFlags().StringVar(&f.Endpoint, "endpoint", defaultEndpoint, "default jade_tree HTTP endpoint when no args and no endpoints in config")
}

func newGetCmd() *cobra.Command {
	flags := &getFlags{}
	root := &cobra.Command{
		Use:   "get [endpoint...]",
		Short: "Get machine information from endpoints (GET /v1/machine-info)",
		Long:  "Print one summary row per machine returned from each endpoint's local probe. Positional args or config endpoints override --endpoint. Subcommands cpu|memory|network|disk|sys print details per endpoint.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			cfg, err := loadClientConfig(flags.ConfigPath)
			if err != nil {
				return err
			}
			endpoints := resolveMachineEndpoints(args, cfg, flags.Endpoint)
			if len(endpoints) == 0 {
				return merr.ErrorInvalidArgument("endpoints are required (positional args, configure endpoints/endpoint in ~/.jade_tree/client.yaml, or use --endpoint)")
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
			out := make([]machineRow, 0, len(endpoints))
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
		name    string
		aliases []string
		short   string
	}{
		{DetailKindCPU, []string{"c"}, "Show CPU details (processors, cores, threads)"},
		{DetailKindMemory, []string{"m", "mem"}, "Show memory and swap details"},
		{DetailKindNetwork, []string{"n", "net"}, "Show network interfaces, IPs, and traffic counters"},
		{DetailKindDisk, []string{"d"}, "Show disks and mount points"},
		{DetailKindSystem, []string{"s", "sys"}, "Show OS, architecture, version, and kernel"},
	}
	for _, dk := range detailKinds {
		kind := dk.name
		sub := &cobra.Command{
			Use:     kind + " [endpoint...]",
			Aliases: dk.aliases,
			Short:   dk.short,
			RunE: func(cmd *cobra.Command, args []string) error {
				if err := flags.validate(); err != nil {
					return err
				}
				cfg, err := loadClientConfig(flags.ConfigPath)
				if err != nil {
					return err
				}
				endpoints := resolveMachineEndpoints(args, cfg, flags.Endpoint)
				if len(endpoints) == 0 {
					return merr.ErrorInvalidArgument("endpoints are required (positional args, configure endpoints/endpoint in ~/.jade_tree/client.yaml, or use --endpoint)")
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
