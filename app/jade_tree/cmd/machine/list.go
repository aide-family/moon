package machine

import (
	"context"
	"net/http"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/spf13/cobra"
)

type listFlags struct {
	machineCommonFlags
	PageSize int32
	Endpoint string
}

func (f *listFlags) addPersistentFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addPersistentFlags(cmd)
	cmd.PersistentFlags().StringVar(&f.Endpoint, "endpoint", defaultEndpoint, "default jade_tree HTTP endpoint when no args and no endpoints in config")
	cmd.PersistentFlags().Int32Var(&f.PageSize, "page-size", 100, "page size when paging GET /v1/machine-infos per endpoint")
}

func newListCmd() *cobra.Command {
	flags := &listFlags{}
	root := &cobra.Command{
		Use:   "list [endpoint...]",
		Short: "List machines known by each endpoint (GET /v1/machine-infos)",
		Long:  "For each jade_tree API endpoint, fetch all pages of cluster-stored machines. Subcommands cpu|memory|network|disk print the same detail sections for every machine from every endpoint.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.validate(); err != nil {
				return err
			}
			if flags.PageSize <= 0 {
				return merr.ErrorInvalidArgument("page-size must be greater than 0")
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
			out := make([]machineRow, 0)
			for _, endpoint := range endpoints {
				ep := strings.TrimSpace(endpoint)
				if ep == "" {
					continue
				}
				machines, err := fetchAllMachines(ctx, client, ep, flags.PageSize)
				if err != nil {
					return err
				}
				out = append(out, toMachineRows(ep, machines)...)
			}
			return renderRows(flags.Output, out)
		},
	}
	flags.addPersistentFlags(root)

	detailKinds := []struct {
		name  string
		short string
	}{
		{"cpu", "Show CPU details for each known machine"},
		{"memory", "Show memory and swap details for each known machine"},
		{"network", "Show network details for each known machine"},
		{"disk", "Show disk and mount details for each known machine"},
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
				if flags.PageSize <= 0 {
					return merr.ErrorInvalidArgument("page-size must be greater than 0")
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
				items := make([]machineDetailItem, 0)
				for _, endpoint := range endpoints {
					ep := strings.TrimSpace(endpoint)
					if ep == "" {
						continue
					}
					machines, err := fetchAllMachines(ctx, client, ep, flags.PageSize)
					if err != nil {
						return err
					}
					for _, m := range machines {
						if m == nil {
							continue
						}
						items = append(items, machineDetailItem{Endpoint: ep, Reply: m})
					}
				}
				return renderMachineDetailsMulti(kind, flags.Output, items)
			},
		}
		root.AddCommand(sub)
	}

	return root
}
