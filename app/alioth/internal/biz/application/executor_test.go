package application

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestApplyReplicaOverrides(t *testing.T) {
	resources := []bo.ApplicationResource{
		{Item: "api", Kind: "Deployment", Manifest: "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: api\nspec:\n  replicas: 1\n"},
		{Item: "worker", Kind: "StatefulSet", Manifest: "apiVersion: apps/v1\nkind: StatefulSet\nmetadata:\n  name: worker\nspec:\n  replicas: 1\n"},
		{Item: "cm", Kind: "ConfigMap", Manifest: "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\n"},
	}
	got := applyReplicaOverrides(resources, bo.ApplicationDeployValues{Replicas: map[string]int32{"api": 3}})
	if !strings.Contains(got[0].Manifest, "replicas: 3") {
		t.Fatalf("deployment replicas not overridden: %s", got[0].Manifest)
	}
	if !strings.Contains(got[1].Manifest, "replicas: 1") {
		t.Fatalf("unselected statefulset should remain: %s", got[1].Manifest)
	}
	if strings.Contains(got[2].Manifest, "replicas") {
		t.Fatalf("configmap should not be changed: %s", got[2].Manifest)
	}
}

func TestValidateDeployValues(t *testing.T) {
	version := &bo.ApplicationVersion{Resources: []bo.ApplicationResource{{Item: "api", Kind: "Deployment"}, {Item: "cm", Kind: "ConfigMap"}}}
	if err := validateDeployValues(bo.ApplicationDeployValues{Replicas: map[string]int32{"api": 2}}, version); err != nil {
		t.Fatalf("valid replica override rejected: %v", err)
	}
	if err := validateDeployValues(bo.ApplicationDeployValues{Replicas: map[string]int32{"cm": 2}}, version); err == nil {
		t.Fatal("configmap replica override should be rejected")
	}
	if err := validateDeployValues(bo.ApplicationDeployValues{Replicas: map[string]int32{"api": 101}}, version); err == nil {
		t.Fatal("out-of-range replica override should be rejected")
	}
}
