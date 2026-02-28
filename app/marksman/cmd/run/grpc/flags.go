package grpc

import (
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/marksman/cmd/run"
)

type Flags struct {
	*run.RunFlags
	grpcTimeout time.Duration
}

var flags Flags

func (f *Flags) addFlags(c *cobra.Command) {
	f.RunFlags = run.GetRunFlags()
	c.Flags().StringVar(&f.Server.Grpc.Address, "grpc-address", f.Server.Grpc.Address, `Example: --grpc-address="0.0.0.0:9090", --grpc-address=":9090"`)
	c.Flags().StringVar(&f.Server.Grpc.Network, "grpc-network", f.Server.Grpc.Network, `Example: --grpc-network="tcp"`)
	c.Flags().DurationVar(&f.grpcTimeout, "grpc-timeout", f.Server.Grpc.Timeout.AsDuration(), `Example: --grpc-timeout="10s", --grpc-timeout="1m", --grpc-timeout="1h", --grpc-timeout="1d"`)
}

func (f *Flags) applyToBootstrap() error {
	if err := f.ApplyToBootstrap(); err != nil {
		return err
	}
	if f.grpcTimeout > 0 {
		f.Server.Grpc.Timeout = durationpb.New(f.grpcTimeout)
	}
	return nil
}
