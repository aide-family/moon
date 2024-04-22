package controller

import (
	"time"

	"github.com/aide-family/moon/api/cluster/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Controller) Terminating(c *Context) (*time.Duration, error) {
	var cond *metav1.Condition

	cond = GetOrNewCondition(*c.Status, v1beta1.ClusterCondTerminating)
	r.set.Remove(c.Key.Name)
	// TODO: other validation for terminate cluster
	cond.Status = metav1.ConditionTrue
	cond.Reason = v1beta1.ReasonSuccessful
	SetCondition(c.Status, *cond)

	return nil, nil
}
