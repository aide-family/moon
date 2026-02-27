package main

import (
	_ "embed"
	"os"

	"github.com/aide-family/magicbox/log"
	"github.com/aide-family/magicbox/log/stdio"
	"github.com/aide-family/magicbox/merr"
	klog "github.com/go-kratos/kratos/v2/log"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/aide-family/rabbit/cmd"
	"github.com/aide-family/rabbit/cmd/config"
	"github.com/aide-family/rabbit/cmd/run"
	"github.com/aide-family/rabbit/cmd/run/all"
	"github.com/aide-family/rabbit/cmd/run/grpc"
	"github.com/aide-family/rabbit/cmd/run/http"
	"github.com/aide-family/rabbit/cmd/run/job"
	"github.com/aide-family/rabbit/cmd/schema"
	"github.com/aide-family/rabbit/cmd/version"
)

var (
	Name        = "rabbit"
	Version     = "latest"
	BuildTime   = "now"
	Author      = ""
	Email       = ""
	Repo        = "https://github.com/aide-family/rabbit"
	hostname, _ = os.Hostname()
)

//go:embed description.txt
var Description string

//go:embed config/server.yaml
var defaultServerConfig []byte

func init() {
	if err := godotenv.Load(); err != nil {
		panic(merr.ErrorInternalServer("load env failed with error: %v", err).WithCause(err))
	}

	cmd.SetGlobalFlags(
		cmd.WithGlobalFlagsName(Name),
		cmd.WithGlobalFlagsHostname(hostname),
		cmd.WithGlobalFlagsVersion(Version),
		cmd.WithGlobalFlagsBuildTime(BuildTime),
		cmd.WithGlobalFlagsAuthor(Author),
		cmd.WithGlobalFlagsEmail(Email),
		cmd.WithGlobalFlagsREPO(Repo),
		cmd.WithGlobalFlagsDescription(Description),
	)

	logger, err := log.NewLogger(stdio.LoggerDriver())
	if err != nil {
		panic(merr.ErrorInternalServer("new logger failed with error: %v", err).WithCause(err))
	}
	logger = klog.With(logger,
		"ts", klog.DefaultTimestamp,
	)
	filterLogger := klog.NewFilter(logger, klog.FilterLevel(klog.LevelInfo))
	helper := klog.NewHelper(filterLogger)
	klog.SetLogger(helper.Logger())
}

func main() {
	runCmd := run.NewCmd(defaultServerConfig)
	runCmd.AddCommand(grpc.NewCmd(), http.NewCmd(), all.NewCmd(), job.NewCmd())

	children := []*cobra.Command{
		version.NewCmd(),
		schema.NewCmd(),
		runCmd,
		config.NewCmd(defaultServerConfig),
	}
	cmd.Execute(cmd.NewCmd(), children...)
}
