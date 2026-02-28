package all

import (
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/marksman/cmd/run"
)

type Flags struct {
	*run.RunFlags

	httpTimeout time.Duration
	grpcTimeout time.Duration
}

var flags Flags

func (f *Flags) addFlags(c *cobra.Command) {
	f.RunFlags = run.GetRunFlags()
	c.Flags().StringVar(&f.Server.Http.Address, "http-address", f.Server.Http.Address, `Example: --http-address="0.0.0.0:8080", --http-address=":8080"`)
	c.Flags().StringVar(&f.Server.Http.Network, "http-network", f.Server.Http.Network, `Example: --http-network="tcp"`)
	c.Flags().DurationVar(&f.httpTimeout, "http-timeout", f.Server.Http.Timeout.AsDuration(), `Example: --http-timeout="10s", --http-timeout="1m", --http-timeout="1h", --http-timeout="1d"`)
	c.Flags().StringVar(&f.Server.Grpc.Address, "grpc-address", f.Server.Grpc.Address, `Example: --grpc-address="0.0.0.0:9090", --grpc-address=":9090"`)
	c.Flags().StringVar(&f.Server.Grpc.Network, "grpc-network", f.Server.Grpc.Network, `Example: --grpc-network="tcp"`)
	c.Flags().DurationVar(&f.grpcTimeout, "grpc-timeout", f.Server.Grpc.Timeout.AsDuration(), `Example: --grpc-timeout="10s", --grpc-timeout="1m", --grpc-timeout="1h", --grpc-timeout="1d"`)
}

func (f *Flags) applyToBootstrap() error {
	if err := f.ApplyToBootstrap(); err != nil {
		return err
	}
	if f.httpTimeout > 0 {
		f.Server.Http.Timeout = durationpb.New(f.httpTimeout)
	}
	if f.grpcTimeout > 0 {
		f.Server.Grpc.Timeout = durationpb.New(f.grpcTimeout)
	}
	return nil
}
