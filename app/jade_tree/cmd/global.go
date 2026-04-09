package cmd

import (
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

type globalFlagsOption func(*GlobalFlags)

type GlobalFlags struct {
	Name      string
	Hostname  string
	Version   string
	BuildTime string
	Author    string
	Email     string
	REPO      string
	LogLevel  string

	mu           sync.RWMutex
	httpEndpoint string
	grpcEndpoint string
}

func WithGlobalFlagsName(name string) globalFlagsOption {
	return func(g *GlobalFlags) { g.Name = name }
}

func WithGlobalFlagsHostname(hostname string) globalFlagsOption {
	return func(g *GlobalFlags) { g.Hostname = hostname }
}

func WithGlobalFlagsVersion(version string) globalFlagsOption {
	return func(g *GlobalFlags) { g.Version = version }
}

func WithGlobalFlagsBuildTime(buildTime string) globalFlagsOption {
	return func(g *GlobalFlags) { g.BuildTime = buildTime }
}

func WithGlobalFlagsAuthor(author string) globalFlagsOption {
	return func(g *GlobalFlags) { g.Author = author }
}

func WithGlobalFlagsEmail(email string) globalFlagsOption {
	return func(g *GlobalFlags) { g.Email = email }
}

func WithGlobalFlagsREPO(repo string) globalFlagsOption {
	return func(g *GlobalFlags) { g.REPO = repo }
}

func (g *GlobalFlags) addFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVar(&g.LogLevel, "log-level", g.LogLevel, "log level")
}

var globalFlags = &GlobalFlags{}

func GetGlobalFlags() *GlobalFlags {
	return globalFlags
}

func SetGlobalFlags(opts ...globalFlagsOption) {
	for _, opt := range opts {
		opt(globalFlags)
	}
}

func SetServerEndpoints(httpEndpoint, grpcEndpoint string) {
	globalFlags.mu.Lock()
	defer globalFlags.mu.Unlock()
	globalFlags.httpEndpoint = strings.TrimSpace(httpEndpoint)
	globalFlags.grpcEndpoint = strings.TrimSpace(grpcEndpoint)
}

func GetServerEndpoints() (string, string) {
	globalFlags.mu.RLock()
	defer globalFlags.mu.RUnlock()
	return globalFlags.httpEndpoint, globalFlags.grpcEndpoint
}
