package controller

import (
	"context"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	"github.com/aide-family/moon/app/kubemoon/internal/cluster/config"
	"github.com/aide-family/moon/pkg/util/finalize"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

type HandlerFunc func(*Context) (*time.Duration, error)

type Controller struct {
	client.Client
	set         clu.Set
	confGetter  clu.ConfigGetter
	builderFunc func(name string) (clu.Client, error)
	l           logr.Logger
	middlewares []HandlerFunc
	handler     *handler
}

func (r *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	clu := &v1beta1.Cluster{}
	err := r.Get(context.TODO(), req.NamespacedName, clu)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	r.l.Info("Begin to reconcile cluster", "key", req.NamespacedName)
	c := newContext(ctx, req.NamespacedName, clu)
	c.handlers = r.middlewares
	c.handlers = append(c.handlers, r.handler.getHandler(c.Phase))

	// handle cluster and promote status
	after, err := r.handler.handler(c)
	if err != nil {
		r.l.Error(err, "cluster handler failed.", "key", c.Key)
		return ctrl.Result{}, err
	}

	err = UpdateClusterStatusInternal(c, r.l, r.Client)
	if err != nil {
		r.l.Error(err, "update cluster status failed.", "key", c.Key)
		return ctrl.Result{}, err
	}

	if after != nil {
		r.l.Info("cluster requeue.", "key", c.Key, "after", after.String())
		return ctrl.Result{RequeueAfter: *after}, nil
	}
	return ctrl.Result{}, nil
}

// New is the constructor of Controller
func New(client client.Client, set clu.Set) *Controller {
	controller := &Controller{
		l:          klogr.New().WithName("Controller:CloneJob"),
		set:        set,
		confGetter: config.NewKubeConfig(client),
		Client:     client,
		handler:    newHandler(),
	}
	controller.builderFunc = controller.builder
	return controller
}

func (r *Controller) Use(middlewares ...HandlerFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Controller) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Cluster{}).
		Complete(r)
}

// handle adding and handle finalizer logic, it turns if we should continue to reconcile
func (r *Controller) handleFinalizer(c *Context) (*time.Duration, error) {
	// delete instance crd, remove finalizer
	if !c.Cluster.DeletionTimestamp.IsZero() {
		cond := GetCondition(*c.Status, v1beta1.ClusterCondTerminating)
		if cond != nil && cond.Status == metav1.ConditionTrue {
			// Completed
			if controllerutil.ContainsFinalizer(c.Cluster, AideCloudCloneJobFinalizer) {
				err := finalize.UpdateFinalizer(r.Client, c.Cluster, finalize.RemoveFinalizerOpType, AideCloudCloneJobFinalizer)
				if err != nil {
					if !errors.IsNotFound(err) {
						r.l.Error(err, "remove cluster finalizer failed.", "key", c.Key)
						return nil, err
					}
				}
				r.l.Info("remove cluster finalizer success", "key", c.Key)
			}
			return nil, nil
		}
		return nil, nil
	}

	// create instance crd, add finalizer
	if !controllerutil.ContainsFinalizer(c.Cluster, AideCloudCloneJobFinalizer) {
		err := finalize.UpdateFinalizer(r.Client, c.Cluster, finalize.AddFinalizerOpType, AideCloudCloneJobFinalizer)
		if err != nil {
			r.l.Error(err, "register cluster finalizer failed.", "key", c.Key)
			return nil, err
		}
		r.l.Info("register cluster finalizer success", "key", c.Key)
	}
	return nil, nil
}

// Default use Logger() & Recovery middlewares
func Default(client client.Client, set clu.Set) *Controller {
	c := New(client, set)
	c.Use(Logger(), Recovery(), Metrics(), c.handleFinalizer)
	c.handler.addHandler("", c.Initial)
	c.handler.addHandler(v1beta1.ClusterPhaseInitial, c.Initial)
	c.handler.addHandler(v1beta1.ClusterPhaseRunning, c.Running)
	c.handler.addHandler(v1beta1.ClusterPhaseTerminating, c.Terminating)
	return c
}
