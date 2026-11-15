package registry

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func TestSplitImageRef(t *testing.T) {
	cases := []struct {
		image      string
		repository string
		tag        string
		digest     string
	}{
		{"harbor.local/library/gateway:1.4.0", "harbor.local/library/gateway", "1.4.0", ""},
		{"harbor.local:5000/library/gateway:latest", "harbor.local:5000/library/gateway", "latest", ""},
		{"docker.io/library/nginx@sha256:abc", "docker.io/library/nginx", "", "sha256:abc"},
		{"library/nginx", "library/nginx", "", ""},
	}
	for _, tc := range cases {
		repository, tag, digest := splitImageRef(tc.image)
		if repository != tc.repository || tag != tc.tag || digest != tc.digest {
			t.Fatalf("splitImageRef(%q) = %q, %q, %q", tc.image, repository, tag, digest)
		}
	}
}

func TestValidateImageSyncInput(t *testing.T) {
	cases := []struct {
		name    string
		input   bo.ImageSyncInput
		wantErr bool
	}{
		{name: "connection source", input: bo.ImageSyncInput{SourceKind: "CONNECTION", SourceRegistryUID: "r1", SourceImage: "library/nginx:1.27", TargetRegistryUID: "r2", TargetRepository: "library/nginx", TargetTag: "1.27"}},
		{name: "external source", input: bo.ImageSyncInput{SourceKind: "EXTERNAL", SourceImage: "docker.io/library/nginx:1.27", TargetRegistryUID: "r2", TargetRepository: "library/nginx", TargetTag: "1.27"}},
		{name: "missing source", input: bo.ImageSyncInput{SourceKind: "EXTERNAL", TargetRegistryUID: "r2", TargetRepository: "library/nginx", TargetTag: "1.27"}, wantErr: true},
		{name: "missing target tag", input: bo.ImageSyncInput{SourceKind: "EXTERNAL", SourceImage: "docker.io/library/nginx:1.27", TargetRegistryUID: "r2", TargetRepository: "library/nginx"}, wantErr: true},
		{name: "bad platform", input: bo.ImageSyncInput{SourceKind: "EXTERNAL", SourceImage: "docker.io/library/nginx:1.27", TargetRegistryUID: "r2", TargetRepository: "library/nginx", TargetTag: "1.27", Platforms: []string{"amd64"}}, wantErr: true},
	}
	for _, tc := range cases {
		in := tc.input
		err := validateImageSyncInput(&in)
		if (err != nil) != tc.wantErr {
			t.Fatalf("%s: err = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}

func TestValidateRegistry(t *testing.T) {
	in := bo.Registry{Name: "harbor", Driver: "HARBOR", Endpoint: "https://harbor.local"}
	if err := validateRegistry(&in, "secret"); err != nil {
		t.Fatalf("valid registry rejected: %v", err)
	}
	invalid := []bo.Registry{
		{Name: "", Driver: "HARBOR", Endpoint: "https://harbor.local"},
		{Name: "harbor", Driver: "OTHER", Endpoint: "https://harbor.local"},
		{Name: "harbor", Driver: "HARBOR", Endpoint: "ftp://harbor.local"},
		{Name: "harbor", Driver: "HARBOR", Endpoint: "harbor.local"},
	}
	for _, item := range invalid {
		value := item
		if err := validateRegistry(&value, ""); err == nil {
			t.Fatalf("invalid registry accepted: %+v", item)
		}
	}
}

func TestValidateRegistryConnectionAllowsEmptyName(t *testing.T) {
	in := bo.Registry{Driver: "DOCKER_V2", Endpoint: "http://192.168.10.172:5000/"}
	if err := validateRegistryConnection(&in, ""); err != nil {
		t.Fatalf("probe connection rejected: %v", err)
	}
	if in.Endpoint != "http://192.168.10.172:5000" {
		t.Fatalf("endpoint not normalized: %q", in.Endpoint)
	}
}
