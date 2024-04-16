package controller

import (
	"context"
	"encoding/json"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	"github.com/go-logr/logr"
	"k8s.io/client-go/util/retry"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ( // phase
	Unknown = iota
	ClusterPhaseInitial
	ClusterPhaseRunning
	ClusterPhaseTerminating
)

const (
	AideCloudClusterFinalizer = "aide.cloud.cn/cluster"
)

func PhaseCode(phase v1beta1.ClusterPhase) float64 {
	switch phase {
	case v1beta1.ClusterPhaseInitial:
		return ClusterPhaseInitial
	case v1beta1.ClusterPhaseRunning:
		return ClusterPhaseRunning
	case v1beta1.ClusterPhaseTerminating:
		return ClusterPhaseTerminating
	}
	return Unknown
}

func UpdateClusterStatusInternal(ctx *Context, l logr.Logger, client client.Client) error {
	if reflect.DeepEqual(ctx.Origin.Status, *ctx.Status) {
		return nil
	}
	clone := ctx.Origin.DeepCopy()
	if err := retry.OnError(retry.DefaultBackoff, func(err error) bool { return true }, func() error {
		if err := client.Get(ctx.Context(), ctx.Key, clone); err != nil {
			l.Error(err, "error getting updated cluster from client", "key", ctx.Key)
			return err
		}
		clone.Status = *ctx.Status
		return client.Status().Update(context.TODO(), clone)
	}); err != nil {
		l.Error(err, "update cluster status failed", "key", ctx.Key)
		return err
	}
	l.V(4).Info("cluster updated status", "key", ctx.Key, "form", DumpJSON(ctx.Origin.Status), "to", DumpJSON(*ctx.Status))
	ctx.Origin.Status = *ctx.Status
	return nil
}

func DumpJSON(o interface{}) string {
	by, _ := json.Marshal(o)
	return string(by)
}
