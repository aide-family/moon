package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/app/kubemoon/cmd/apps/options"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	"github.com/aide-family/moon/app/kubemoon/internal/cluster/controller"
	"github.com/aide-family/moon/app/kubemoon/internal/cluster/set"
	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/util/hello"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	once            sync.Once
	ProviderSetCore = wire.NewSet(
		before,
	)
)

func before() conf.Before {
	return func(bc *conf.Bootstrap) error {
		env := bc.GetEnv()
		once.Do(func() {
			hello.SetName(env.GetName())
			hello.SetVersion(Version)
			hello.SetEnv(env.GetEnv())
			hello.SetMetadata(env.GetMetadata())
			hello.FmtASCIIGenerator()
		})
		return nil
	}
}

func newApp(hs *http.Server, logger log.Logger) *kratos.App {
	return kratos.New(
		kratos.ID(hello.ID()),
		kratos.Name(hello.Name()),
		kratos.Version(hello.Version()),
		kratos.Metadata(hello.Metadata()),
		kratos.Logger(logger),
		kratos.Server(hs),
	)
}

func NewKubeMoonCommand() *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:  "Kube Moon",
		Long: ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Info("Kube Moon")
			return runCommand(cmd, opts)
		},
		SilenceUsage: true,
	}

	fs := cmd.Flags()
	namedFlagSets := opts.Flags()
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	usageFmt := "%s\n\nUsage:\n %s \n"
	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cmd.SetHelpFunc(func(command *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), usageFmt, cmd.Long, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStdout(), namedFlagSets, cols)
	})
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		fmt.Fprintf(cmd.OutOrStderr(), usageFmt, cmd.UseLine())
		cliflag.PrintSections(cmd.OutOrStderr(), namedFlagSets, cols)
		return nil
	})
	return cmd

}

func runCommand(cmd *cobra.Command, opts *options.Options) error {
	verflag.PrintAndExitIfRequested()
	cliflag.PrintFlags(cmd.Flags())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		stopCh := server.SetupSignalHandler()
		<-stopCh
		cancel()
	}()
	mgr, err := Setup(ctx, opts)
	if err != nil {
		return err
	}
	return Run(ctx, mgr)
}

func Setup(ctx context.Context, opts *options.Options) (manager.Manager, error) {
	klog.Info("complete options ...")
	errList := opts.Validate()
	if len(errList) != 0 {
		klog.Errorf("validate options failed with err: %v", errList)
		return nil, errList[0]
	}

	klog.Info("complete config ...")
	config, err := opts.Complete()
	if err != nil {
		klog.ErrorS(err, "complete config error")
		return nil, err
	}

	klog.Info("build manager ...")
	mgr, err := ctrl.NewManager(config.KubeConfig, config.ManagerOptions())
	if err != nil {
		return nil, err
	}

	klog.Info("build cluster client manager ...")
	cluSet := set.New(mgr.GetClient())

	klog.Info("add cluster client set to manager...")
	err = mgr.Add(cluSet)
	if err != nil {
		klog.ErrorS(err, "add cluster client set to manager failed")
		return nil, err
	}

	klog.Info("build controller ...")
	err = SetupController(ctx, mgr, cluSet)
	if err != nil {
		klog.ErrorS(err, "setup controller failed")
		return nil, err
	}

	// TODO: set service
	if err = mgr.Add(NewKratos(opts)); err != nil {
		klog.ErrorS(err, "add kratos to manager failed")
	}

	return mgr, nil
}

func SetupController(_ context.Context, mgr manager.Manager, set clu.Set) error {
	klog.Info("build cluster controller ...")
	return controller.Default(mgr.GetClient(), set).SetupWithManager(mgr)
}

func Run(ctx context.Context, mgr manager.Manager) error {
	klog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("unable to run the manager: %v", err)
	}
	return nil
}

var _ manager.Runnable = (*Kratos)(nil)

type Kratos struct {
	opts *options.Options
}

func NewKratos(opts *options.Options) *Kratos {
	return &Kratos{opts: opts}
}

func (k *Kratos) Start(_ context.Context) error {
	go func() {
		defer after.RecoverX()
		app, cleanup, err := wireApp(&k.opts.ConfPath)
		if err != nil {
			panic(err)
		}
		defer cleanup()

		// start and wait for stop signal
		if err = app.Run(); err != nil {
			panic(err)
		}
	}()

	return nil
}
