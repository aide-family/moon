package k8sproxy

import "testing"

func TestValidK8SPath(t *testing.T) {
	good := []string{
		"/api/v1/pods",
		"/api/v1/namespaces/default/pods",
		"/apis/apps/v1/deployments",
	}
	for _, path := range good {
		if !validK8sPath(path) {
			t.Fatalf("validK8sPath(%q) = false", path)
		}
	}
	bad := []string{
		"",
		"api/v1/pods",
		"//evil.example/x",
		"@evil.example/x",
		"/api/../secrets",
		`/api\v1`,
	}
	for _, path := range bad {
		if validK8sPath(path) {
			t.Fatalf("validK8sPath(%q) = true", path)
		}
	}
}

func TestK8SMethodAllowed(t *testing.T) {
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"} {
		if !k8sMethodAllowed(method) {
			t.Fatalf("method %s should be allowed", method)
		}
	}
	for _, method := range []string{"", "CONNECT", "TRACE", "OPTIONS"} {
		if k8sMethodAllowed(method) {
			t.Fatalf("method %s should be rejected", method)
		}
	}
}
