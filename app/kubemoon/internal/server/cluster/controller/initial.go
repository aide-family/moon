package controller

import (
	"time"

	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster/builder"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func (r *Controller) Initial(c *Context) (*time.Duration, error) {
	var cli clu.Client
	var cond *metav1.Condition
	var err error

	defer func() { SetCondition(c.Status, *cond) }()

	if len(c.Phase) == 0 {
		c.Phase = v1beta1.ClusterPhaseInitial
	}

	cond = GetOrNewCondition(*c.Status, v1beta1.ClusterCondInitial)
	if cond.Status == metav1.ConditionTrue {
		c.Status.Phase = v1beta1.ClusterPhaseRunning
	} else {
		cli = r.set.Client(c.Key.Name)
		if cli == nil {
			cli, err = r.builderFunc(c.Key.Name)
			if err != nil {
				cond.Status = metav1.ConditionFalse
				cond.Reason = v1beta1.ReasonBuildFailed
				cond.Message = err.Error()
				return nil, err
			}
		}
		if err = Online(c.Context(), cli); err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonOffline
			cond.Message = err.Error()
			return nil, err
		}
		cond.Status = metav1.ConditionTrue
		cond.Reason = v1beta1.ReasonSuccessful
	}
	return nil, nil
}

func (r *Controller) builder(name string) (clu.Client, error) {
	sc := runtime.NewScheme()
	_ = scheme.AddToScheme(sc)
	return builder.By(r.confGetter).Named(name).WithScheme(sc).Complete()
}
