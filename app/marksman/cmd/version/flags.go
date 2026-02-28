package version

import (
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd"
)

type Flags struct {
	*cmd.GlobalFlags
	format string
}

var flags Flags

func (f *Flags) addFlags(c *cobra.Command) {
	c.PersistentFlags().StringVarP(&f.format, "format", "f", "txt", "The format of the version output, supported: [txt, json, yaml]")
}
