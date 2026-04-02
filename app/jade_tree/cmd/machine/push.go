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
	cmd.Flags().Int32Var(&f.PageSize, "page-size", 100, "page size for source machine info pull")
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
				endpoints = cfg.Endpoints
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

func fetchAllMachines(ctx context.Context, client *apiClient, from string, pageSize int32) ([]*apiv1.GetMachineInfoReply, error) {
	var page int32 = 1
	merged := make(map[string]*apiv1.GetMachineInfoReply)
	for {
		reply, err := client.listMachineInfos(ctx, from, page, pageSize)
		if err != nil {
			return nil, err
		}
		for _, item := range reply.GetMachines() {
			if item == nil {
				continue
			}
			key := item.GetHost().GetMachineUuid()
			if key == "" {
				key = item.GetHost().GetHostName()
			}
			if key == "" {
				continue
			}
			merged[key] = item
		}
		if len(reply.GetMachines()) == 0 || int32(len(reply.GetMachines())) < pageSize {
			break
		}
		page++
	}
	localInfo, err := client.getMachineInfo(ctx, from)
	if err == nil && localInfo != nil {
		key := localInfo.GetHost().GetMachineUuid()
		if key == "" {
			key = localInfo.GetHost().GetHostName()
		}
		if key != "" {
			merged[key] = localInfo
		}
	}

	out := make([]*apiv1.GetMachineInfoReply, 0, len(merged))
	for _, item := range merged {
		out = append(out, item)
	}
	return out, nil
}
