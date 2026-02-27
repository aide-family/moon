package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var hostname, _ = os.Hostname()

type GlobalFlags struct {
	Name        string `json:"name" yaml:"name"`
	Author      string `json:"author" yaml:"author"`
	Email       string `json:"email" yaml:"email"`
	Repo        string `json:"repo" yaml:"repo"`
	Description string `json:"description" yaml:"description"`
	Version     string `json:"version" yaml:"version"`
	Built       string `json:"built" yaml:"built"`

	Hostname  string `json:"-" yaml:"-"`
	LogFormat string `json:"-" yaml:"-"`
	LogLevel  string `json:"-" yaml:"-"`
}

func (g *GlobalFlags) addFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&g.LogFormat, "log-format", "TEXT", "The format of the log")
	cmd.PersistentFlags().StringVar(&g.LogLevel, "log-level", "DEBUG", "The level of the log")
}

type GlobalOption func(*GlobalFlags)

var globalFlags GlobalFlags

func GetGlobalFlags() *GlobalFlags {
	return &globalFlags
}

func SetGlobalFlags(opts ...GlobalOption) {
	for _, opt := range opts {
		opt(&globalFlags)
	}
}

func WithGlobalFlagsVersion(version string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Version = version
	}
}

func WithGlobalFlagsBuildTime(buildTime string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Built = buildTime
	}
}

func WithGlobalFlagsEmail(email string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Email = email
	}
}

func WithGlobalFlagsAuthor(author string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Author = author
	}
}

func WithGlobalFlagsDescription(description string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Description = description
	}
}

func WithGlobalFlagsREPO(repo string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Repo = repo
	}
}

func WithGlobalFlagsName(name string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Name = name
	}
}

func WithGlobalFlagsHostname(hostname string) GlobalOption {
	return func(g *GlobalFlags) {
		g.Hostname = hostname
	}
}
