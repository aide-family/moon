package controller

import (
	"context"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

type Context struct {
	ctx   context.Context
	Key   types.NamespacedName
	Phase v1beta1.ClusterPhase

	Origin  *v1beta1.Cluster
	Cluster *v1beta1.Cluster
	Status  *v1beta1.ClusterStatus

	index    int
	handlers []HandlerFunc
}

func newContext(ctx context.Context, key types.NamespacedName, origin *v1beta1.Cluster) *Context {
	phase := origin.Status.Phase
	if origin.DeletionTimestamp != nil {
		phase = v1beta1.ClusterPhaseTerminating
	}
	return &Context{
		ctx:      ctx,
		Key:      key,
		Phase:    phase,
		Origin:   origin,
		Cluster:  origin.DeepCopy(),
		Status:   origin.Status.DeepCopy(),
		index:    -1,
		handlers: make([]HandlerFunc, 0),
	}
}

func (c *Context) Context() context.Context {
	return c.ctx
}

func (c *Context) Next() (*time.Duration, error) {
	c.index++
	l := len(c.handlers)
	for ; c.index < l; c.index++ {
		after, err := c.handlers[c.index](c)
		if after != nil || err != nil {
			return after, err
		}
	}
	return nil, nil
}
