package cluster

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/pkg/merr"
)

func TestClusterInput(t *testing.T) {
	in, err := clusterInput("ns-1", " uid-1 ", "  prod  ", "  primary  ")
	if err != nil {
		t.Fatalf("clusterInput: %v", err)
	}
	if in.UID != "uid-1" || in.NamespaceUID != "ns-1" || in.Name != "prod" || in.Remark != "primary" {
		t.Fatalf("input = %+v", in)
	}
	if _, err := clusterInput("ns-1", "", "   ", ""); !merr.IsInvalid(err) {
		t.Fatalf("empty name err = %v, want invalid", err)
	}
	if _, err := clusterInput("ns-1", "", strings.Repeat("名", maxClusterName), ""); err != nil {
		t.Fatalf("max length name: %v", err)
	}
	if _, err := clusterInput("ns-1", "", strings.Repeat("名", maxClusterName+1), ""); !merr.IsInvalid(err) {
		t.Fatalf("too long name err = %v, want invalid", err)
	}
}

func TestValidWorkloadKind(t *testing.T) {
	for _, kind := range []string{workloadDeployment, workloadStatefulSet, workloadDaemonSet} {
		if !validWorkloadKind(kind) {
			t.Fatalf("%s should be valid", kind)
		}
	}
	for _, kind := range []string{"", "Job", "CronJob", "deployment"} {
		if validWorkloadKind(kind) {
			t.Fatalf("%q should be invalid", kind)
		}
	}
}
