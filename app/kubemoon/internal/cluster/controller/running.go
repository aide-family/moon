package controller

import (
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	"github.com/go-logr/logr"
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

	if !isEnabled {
		cond.Status = metav1.ConditionFalse
		cond.Reason = v1beta1.ReasonDisabled
		r.set.Remove(c.Key.Name)
		r.l.Info("removed cluster for cluster set", "key", c.Key, "reason", cond.Reason)
	} else if !isConnected {
		cond.Status = metav1.ConditionFalse
		cond.Reason = v1beta1.ReasonOffline
		cond.Message = err.Error()
	} else if isEnabled && isConnected {
		cond.Status = metav1.ConditionTrue
		cond.Reason = v1beta1.ReasonSuccessful
		return SyncClusterStatus(c, r.set.Client(c.Key.Name), r.l)
	}

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
		if err = Online(c.Context(), cli); err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonOffline
			cond.Message = err.Error()
			return false, err
		}
		if err = r.set.Add(cli); err != nil {
			cond.Status = metav1.ConditionFalse
			cond.Reason = v1beta1.ReasonBuildFailed
			cond.Message = err.Error()
			return false, err
		}
		r.l.Info("add cluster for cluster set", "key", c.Key)
	}
	if err = Healthy(c.Context(), cli); err != nil {
		cond.Status = metav1.ConditionFalse
		cond.Reason = v1beta1.ReasonNotHealthy
		cond.Message = err.Error()
		return false, err
	}

	cond.Status = metav1.ConditionTrue
	cond.Reason = v1beta1.ReasonSuccessful
	return true, nil
}

func SyncClusterStatus(c *Context, cli clu.Client, logger logr.Logger) (*time.Duration, error) {
	version, err := cli.KubernetesVersion()
	if err != nil {
		logger.Error(err, "fail to get kubernetes version for cluster", "key", c.Key)
	}
	c.Status.Version = version

	enablements, err := cli.APIEnablements()
	if err != nil {
		logger.Error(err, "fail to get kubernetes api for cluster", "key", c.Key)
	}
	c.Status.APIEnablements = enablements

	nodes, err := ListNodes(c.Context(), cli)
	if err != nil {
		logger.Error(err, "failed to list nodes for cluster", "key", c.Key)
	}
	c.Status.NodeSummary = NodeSummary(nodes)
	return &AideCloudClusterRecheckInterval, nil
}
