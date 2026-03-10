package metriccron

import (
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd/run"
)

type Flags struct {
	*run.RunFlags
}

var flags Flags

func (f *Flags) addFlags(c *cobra.Command) {
	f.RunFlags = run.GetRunFlags()
}

func (f *Flags) applyToBootstrap() error {
	return f.ApplyToBootstrap()
}
