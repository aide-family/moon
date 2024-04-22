package server

import (
	"context"
	"fmt"

	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	clu "github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster/controller"
	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster/set"
	"github.com/go-kratos/kratos/v2/transport"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ transport.Server = (*KubeServer)(nil)

type KubeServer struct {
	c *conf.Bootstrap

	stopCh chan struct{}
}

func NewKubeServer(c *conf.Bootstrap) *KubeServer {
	return &KubeServer{
		c: c,
	}
}

func (k *KubeServer) Start(ctx context.Context) error {
	return k.runCommand(ctx, k.c)
}

func (k *KubeServer) Stop(ctx context.Context) error {
	return nil
}

func (k *KubeServer) runCommand(ctx context.Context, cfg *conf.Bootstrap) error {
	mgr, err := Setup(ctx, cfg)
	if err != nil {
		return err
	}
	return Run(ctx, mgr)
}

func Setup(ctx context.Context, cfg *conf.Bootstrap) (manager.Manager, error) {
	klog.Info("complete options ...")
	errList := cfg.Validate()
	if len(errList) != 0 {
		klog.Errorf("validate options failed with err: %v", errList)
		return nil, errList[0]
	}

	klog.Info("complete config ...")
	config, err := cfg.Complete()
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
