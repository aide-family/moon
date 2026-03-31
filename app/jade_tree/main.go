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

	"github.com/aide-family/jade_tree/cmd"
	"github.com/aide-family/jade_tree/cmd/run"
	"github.com/aide-family/jade_tree/cmd/run/all"
	"github.com/aide-family/jade_tree/cmd/run/grpc"
	"github.com/aide-family/jade_tree/cmd/run/http"
)

var (
	Name        = "jade_tree"
	Version     = "latest"
	BuildTime   = "now"
	Author      = ""
	Email       = ""
	Repo        = "https://github.com/aide-family/jade_tree"
	hostname, _ = os.Hostname()
)

//go:embed config/server.yaml
var defaultServerConfig []byte

func init() {
	if err := godotenv.Load(); err != nil {
		klog.Warnf("load env failed with error: %v", err)
	}
	cmd.SetGlobalFlags(
		cmd.WithGlobalFlagsName(Name),
		cmd.WithGlobalFlagsHostname(hostname),
		cmd.WithGlobalFlagsVersion(Version),
		cmd.WithGlobalFlagsBuildTime(BuildTime),
		cmd.WithGlobalFlagsAuthor(Author),
		cmd.WithGlobalFlagsEmail(Email),
		cmd.WithGlobalFlagsREPO(Repo),
	)

	logger, err := log.NewLogger(stdio.LoggerDriver())
	if err != nil {
		panic(merr.ErrorInternalServer("new logger failed with error: %v", err).WithCause(err))
	}
	logger = klog.With(logger, "ts", klog.DefaultTimestamp)
	filterLogger := klog.NewFilter(logger, klog.FilterLevel(klog.LevelInfo))
	helper := klog.NewHelper(filterLogger)
	klog.SetLogger(helper.Logger())
}

func main() {
	runCmd := run.NewCmd(defaultServerConfig)
	runCmd.AddCommand(grpc.NewCmd(), http.NewCmd(), all.NewCmd())
	children := []*cobra.Command{runCmd}
	cmd.Execute(cmd.NewCmd(), children...)
}
