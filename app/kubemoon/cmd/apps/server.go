package apps

import (
	"context"
	"fmt"
	"github.com/aide-family/moon/app/kubemoon/cmd/apps/options"
	"github.com/spf13/cobra"
	"k8s.io/apiserver/pkg/server"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/term"
	"k8s.io/component-base/version/verflag"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewKubeMoonCommand() *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Use:  "Kube Moon",
		Long: ``,
		RunE: func(cmd *cobra.Command, args []string) error {
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
	errList := opts.Validate()
	if len(errList) != 0 {
		klog.Errorf("validate options failed with err: %v", errList)
		return nil, errList[0]
	}

	conf, err := opts.Complete()
	if err != nil {
		klog.ErrorS(err, "complete config error")
		return nil, err
	}

	klog.Info("build manager ...")
	mgr, err := ctrl.NewManager(conf.KubeConfig, conf.ManagerOptions())
	if err != nil {
		return nil, err
	}

	// TODO: set service

	return mgr, nil
}

func Run(ctx context.Context, mgr manager.Manager) error {
	klog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("unable to run the manager: %v", err)
	}
	return nil
}
