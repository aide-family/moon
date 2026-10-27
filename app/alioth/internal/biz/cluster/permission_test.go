package cluster

import (
	"context"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/pkg/merr"
)

// TestPermissionGates 确认每个入口在没有租户上下文时先失败，不会碰到仓储。
// 这里三个用例对象都不注入仓储：漏掉权限检查会直接因 nil 仓储而 panic。
func TestPermissionGates(t *testing.T) {
	ctx := context.Background()
	registry := &Registry{access: access.Checker{}}
	explorer := &Explorer{access: access.Checker{}}
	operator := &Operator{access: access.Checker{}}
	cases := []struct {
		name string
		run  func() error
	}{
		{name: "create cluster", run: func() error {
			_, err := registry.Create(ctx, "prod", "kubeconfig", "")
			return err
		}},
		{name: "update cluster", run: func() error {
			_, err := registry.Update(ctx, "uid", "prod", "", "")
			return err
		}},
		{name: "delete cluster", run: func() error {
			return registry.Delete(ctx, "uid")
		}},
		{name: "get cluster", run: func() error {
			_, err := registry.Get(ctx, "uid")
			return err
		}},
		{name: "list clusters", run: func() error {
			_, _, _, _, err := registry.List(ctx, 1, 20, "")
			return err
		}},
		{name: "test cluster", run: func() error {
			_, _, err := registry.Test(ctx, "uid")
			return err
		}},
		{name: "probe kubeconfig", run: func() error {
			_, err := registry.Probe(ctx, "kubeconfig")
			return err
		}},
		{name: "list namespaces", run: func() error {
			_, err := explorer.ListNamespaces(ctx, "uid")
			return err
		}},
		{name: "list nodes", run: func() error {
			_, _, err := explorer.ListNodes(ctx, "uid", "")
			return err
		}},
		{name: "list workloads", run: func() error {
			_, _, err := explorer.ListWorkloads(ctx, "uid", "", "", "")
			return err
		}},
		{name: "list pods", run: func() error {
			_, _, _, _, err := explorer.ListPods(ctx, "uid", "", "", "", "", 1, 20)
			return err
		}},
		{name: "list ingresses", run: func() error {
			_, _, _, _, err := explorer.ListIngresses(ctx, "uid", "", "", 1, 20)
			return err
		}},
		{name: "pod logs", run: func() error {
			_, err := explorer.PodLogs(ctx, "uid", "ns", "pod", "", 0, false)
			return err
		}},
		{name: "list services", run: func() error {
			_, _, _, _, err := explorer.ListServices(ctx, "uid", "", "", 1, 20)
			return err
		}},
		{name: "list configmaps", run: func() error {
			_, _, _, _, err := explorer.ListConfigMaps(ctx, "uid", "", "", 1, 20)
			return err
		}},
		{name: "list secrets", run: func() error {
			_, _, _, _, err := explorer.ListSecrets(ctx, "uid", "", "", 1, 20)
			return err
		}},
		{name: "list persistent volumes", run: func() error {
			_, _, _, _, err := explorer.ListPersistentVolumes(ctx, "uid", "", 1, 20)
			return err
		}},
		{name: "list persistent volume claims", run: func() error {
			_, _, _, _, err := explorer.ListPersistentVolumeClaims(ctx, "uid", "", "", 1, 20)
			return err
		}},
		{name: "resource yaml", run: func() error {
			_, err := explorer.ResourceYAML(ctx, "uid", "Pod", "ns", "pod")
			return err
		}},
		{name: "workload detail", run: func() error {
			_, err := explorer.WorkloadDetail(ctx, "uid", "Deployment", "ns", "web")
			return err
		}},
		{name: "ingress detail", run: func() error {
			_, err := explorer.IngressDetail(ctx, "uid", "ns", "web")
			return err
		}},
		{name: "service detail", run: func() error {
			_, err := explorer.ServiceDetail(ctx, "uid", "ns", "web")
			return err
		}},
		{name: "configmap detail", run: func() error {
			_, err := explorer.ConfigMapDetail(ctx, "uid", "ns", "app-config")
			return err
		}},
		{name: "secret detail", run: func() error {
			_, err := explorer.SecretDetail(ctx, "uid", "ns", "db-secret")
			return err
		}},
		{name: "persistent volume claim detail", run: func() error {
			_, err := explorer.PersistentVolumeClaimDetail(ctx, "uid", "ns", "data")
			return err
		}},
		{name: "cordon node", run: func() error {
			return operator.SetNodeUnschedulable(ctx, "uid", "node-a", true)
		}},
		{name: "scale workload", run: func() error {
			_, err := operator.ScaleWorkload(ctx, "uid", "ns", workloadDeployment, "web", 1)
			return err
		}},
		{name: "restart workload", run: func() error {
			_, err := operator.RestartWorkload(ctx, "uid", "ns", workloadDeployment, "web")
			return err
		}},
		{name: "delete pod", run: func() error {
			return operator.DeletePod(ctx, "uid", "ns", "pod")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.run()
			if err == nil {
				t.Fatal("want permission error")
			}
			if reason := kerrors.FromError(err).Reason; reason != merr.ReasonClient {
				t.Fatalf("reason = %q, want %q (err = %v)", reason, merr.ReasonClient, err)
			}
		})
	}
}
