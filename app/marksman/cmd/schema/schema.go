// Package schema is the schema command for generating DDL from table models.
package schema

import (
	"github.com/spf13/cobra"

	"github.com/aide-family/marksman/cmd"
)

func NewCmd() *cobra.Command {
	schemaCmd := &cobra.Command{
		Use:   "schema",
		Short: "Generate schema SQL from table models",
		Long:  "Generate a complete DDL file (CREATE TABLE / CREATE INDEX) from GORM models so each upgrade can be versioned.",
		Annotations: map[string]string{
			"group": cmd.DatabaseCommands,
		},
	}
	schemaCmd.AddCommand(newSQLCmd(), newMigrateCmd())
	return schemaCmd
}
