package server

import (
	"context"
	"fmt"
	"github.com/aide-family/moon/app/kubemoon/internal/conf"
	"github.com/aide-family/moon/app/kubemoon/internal/data"
	clu "github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster/controller"
	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster/set"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ transport.Server = (*KubeServer)(nil)

type KubeServer struct {
	c    *conf.Bootstrap
	data *data.Data
	mgr  manager.Manager
	set  clu.Set

	stopCh chan struct{}
}

func (k *KubeServer) Start(ctx context.Context) error {
	return Run(ctx, k.mgr)
}

func (k *KubeServer) Stop(ctx context.Context) error {
	// TODO 这里需要做一些清理工作
	return nil
}

func (k *KubeServer) GetManager() manager.Manager {
	return k.mgr
}

func (k *KubeServer) GetCluSet() clu.Set {
	return k.set
}

func NewKubeServer(cfg *conf.Bootstrap, data *data.Data) (*KubeServer, error) {
	log.Info("complete options ...")
	errList := cfg.Validate()
	if len(errList) != 0 {
		log.Errorf("validate options failed with err: %v", errList)
		return nil, errList[0]
	}

	log.Info("complete config ...")
	config, err := cfg.Complete()
	if err != nil {
		log.Error(err, "complete config error")
		return nil, err
	}

	log.Info("build manager ...")
	mgr, err := ctrl.NewManager(config.KubeConfig, config.ManagerOptions())
	if err != nil {
		return nil, err
	}

	log.Info("build cluster client manager ...")
	cluSet := set.New(mgr.GetClient())

	log.Info("add cluster client set to manager...")
	err = mgr.Add(cluSet)
	if err != nil {
		log.Error(err, "add cluster client set to manager failed")
		return nil, err
	}

	log.Info("build controller ...")
	err = SetupController(mgr, cluSet)
	if err != nil {
		log.Error(err, "setup controller failed")
		return nil, err
	}

	return &KubeServer{
		c:    cfg,
		data: data,
		mgr:  mgr,
		set:  cluSet,
	}, nil
}

func SetupController(mgr manager.Manager, set clu.Set) error {
	log.Info("build cluster controller ...")
	return controller.Default(mgr.GetClient(), set).SetupWithManager(mgr)
}

func Run(ctx context.Context, mgr manager.Manager) error {
	log.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("unable to run the manager: %v", err)
	}
	return nil
}
