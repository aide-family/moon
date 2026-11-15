package application

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestParseResourcesSecretAndImages(t *testing.T) {
	manifest := `apiVersion: v1
kind: Secret
metadata:
  name: gateway-secret
stringData:
  password: s3cret
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
spec:
  template:
    spec:
      containers:
        - name: app
          image: harbor.local/library/gateway:1.4.0`
	resources, secrets, errs, _, err := ParseResources([]bo.ApplicationResource{{Item: "r1", Manifest: strings.Split(manifest, "---")[0]}, {Item: "r2", Manifest: strings.Split(manifest, "---")[1]}})
	if err != nil || len(errs) > 0 {
		t.Fatalf("parse failed: err=%v errs=%v", err, errs)
	}
	if len(resources) != 2 {
		t.Fatalf("resources = %d", len(resources))
	}
	if secrets["r1/password"] != "s3cret" {
		t.Fatalf("secret missing: %v", secrets)
	}
	if strings.Contains(resources[0].Manifest, "s3cret") {
		t.Fatal("secret value leaked into stored manifest")
	}
	if len(resources[1].Images) != 1 || resources[1].Images[0] != "harbor.local/library/gateway:1.4.0" {
		t.Fatalf("images = %v", resources[1].Images)
	}
	if len(resources[0].SecretKeys) != 1 || resources[0].SecretKeys[0] != "password" {
		t.Fatalf("secret keys = %v", resources[0].SecretKeys)
	}
}

func TestParseResourcesRejectsBlockedKindAndDuplicate(t *testing.T) {
	blocked := `apiVersion: v1
kind: Namespace
metadata:
  name: other`
	_, _, errs, _, err := ParseResources([]bo.ApplicationResource{{Item: "r1", Manifest: blocked}})
	if err != nil || len(errs) == 0 {
		t.Fatalf("blocked kind should fail: err=%v errs=%v", err, errs)
	}
	duplicate := `apiVersion: v1
kind: ConfigMap
metadata:
  name: same`
	_, _, errs, _, err = ParseResources([]bo.ApplicationResource{{Item: "r1", Manifest: duplicate}, {Item: "r2", Manifest: duplicate}})
	if err != nil || len(errs) == 0 {
		t.Fatalf("duplicate should fail: err=%v errs=%v", err, errs)
	}
}
