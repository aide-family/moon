// Package cmd provides CLI commands for Jade Tree.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

type flagRoot interface {
	PersistentFlags() interface {
		StringVar(*string, string, string, string)
	}
}

func NewCmd() *cobra.Command {
	rootCmd := &cobra.Command{Use: "jade_tree", Short: "Moon platform - jade_tree service", Run: func(cmd *cobra.Command, args []string) { cmd.Help() }}
	globalFlags.addFlags(rootCmd)
	return rootCmd
}

func Execute(cmd *cobra.Command, children ...*cobra.Command) {
	cmd.AddCommand(children...)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
