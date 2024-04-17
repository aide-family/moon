package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
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

var AideCloudClusterRecheckInterval = 30 * time.Second

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

func Online(ctx context.Context, cli clu.Client) error {
	_, err := Ping(ctx, cli)
	return err
}

func Healthy(ctx context.Context, cli clu.Client) error {
	code, err := Ping(ctx, cli)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("cluster isn't healthy")
	}
	return nil
}

func Ping(ctx context.Context, cli clu.Client) (int, error) {
	code, err := cli.Ready(ctx)
	if err != nil && code == http.StatusNotFound {
		// If the ready check interface is not found, try using the health check interface
		code, err = cli.Health(ctx)
	}
	return code, err
}

func ListNodes(ctx context.Context, cli clu.Client) (*v1.NodeList, error) {
	var err error
	list := &v1.NodeList{}
	if cli.Status() == clu.Started {
		err = cli.Cache().List(ctx, list)
	} else {
		err = cli.Client().List(ctx, list)
	}
	return list, err
}

func NodeSummary(nodes *v1.NodeList) *v1beta1.NodeSummary {
	totalNum := len(nodes.Items)
	readyNum := 0

	for i := range nodes.Items {
		if NodeReady(&nodes.Items[i]) {
			readyNum++
		}
	}

	nodeSummary := &v1beta1.NodeSummary{}
	nodeSummary.TotalNum = int32(totalNum)
	nodeSummary.ReadyNum = int32(readyNum)

	return nodeSummary
}

func NodeReady(node *v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}
