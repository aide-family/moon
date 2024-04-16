package controller

import (
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func (r *Controller) Running(c *Context) (*time.Duration, error) {
	var err error
	var isEnabled bool
	var isConnected bool
	var cond *metav1.Condition
	defer func() { SetCondition(c.Status, *cond) }()

	cond = GetOrNewCondition(*c.Status, v1beta1.ClusterCondReady)

	// check if the cluster is disabled
	isEnabled = r.checkClusterEnabled(c)
	// check if the cluster is connected
	isConnected, err = r.checkClusterConnection(c)

	if isEnabled && isConnected {
		cond.Status = metav1.ConditionTrue
		cond.Reason = v1beta1.ReasonSuccessful
	} else if !isEnabled {
		cond.Status = metav1.ConditionFalse
		cond.Reason = v1beta1.ReasonDisabled
	} else if !isConnected {
		cond.Status = metav1.ConditionFalse
		cond.Reason = v1beta1.ReasonDisconnected
		cond.Message = err.Error()
	}

	// TODO: sync cluster info to status

	return nil, err
}

func (r *Controller) checkClusterEnabled(c *Context) bool {
	disabledCond := GetOrNewCondition(*c.Status, v1beta1.ClusterCondDisabled)
	defer func() { SetCondition(c.Status, *disabledCond) }()

	if c.Cluster.Spec.Disabled {
		disabledCond.Status = metav1.ConditionTrue
		disabledCond.Reason = v1beta1.ReasonDisabled
		return false
	} else {
		disabledCond.Status = metav1.ConditionFalse
		disabledCond.Reason = v1beta1.ReasonSuccessful
		return true
	}
}

func (r *Controller) checkClusterConnection(c *Context) (bool, error) {
	var cli clu.Client
	var err error
	var cond *metav1.Condition

	cond = GetOrNewCondition(*c.Status, v1beta1.ClusterCondConnection)
	defer func() { SetCondition(c.Status, *cond) }()

	cli = r.set.Client(c.Key.Name)
	if cli == nil {
		cli, err = r.builderFunc(c.Key.Name)
		if err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonBuildFailed
			cond.Message = err.Error()
			return false, err
		}
		if err = cli.Ping(c.Context()); err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonDisconnected
			cond.Message = err.Error()
			return false, err
		}
		if err = r.set.Add(cli); err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonBuildFailed
			cond.Message = err.Error()
			return false, err
		}
	} else {
		if err = cli.Ping(c.Context()); err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonDisconnected
			cond.Message = err.Error()
			return false, err
		}
	}

	cond.Status = metav1.ConditionTrue
	cond.Reason = v1beta1.ReasonSuccessful
	return true, nil
}
