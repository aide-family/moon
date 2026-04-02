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

func (f *infoFlags) addFlags(cmd *cobra.Command) {
	f.machineCommonFlags.addFlags(cmd)
	cmd.Flags().StringVar(&f.Endpoint, "endpoint", defaultEndpoint, "jade_tree HTTP endpoint")
}

func newInfoCmd() *cobra.Command {
	flags := &infoFlags{}
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get local machine information",
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
	flags.addFlags(cmd)
	return cmd
}
