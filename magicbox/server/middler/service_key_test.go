package middler

import "testing"

func TestIsServiceKeyCredential(t *testing.T) {
	tests := []struct {
		name       string
		credential string
		want       bool
	}{
		{name: "service key", credential: "sk-test-key", want: true},
		{name: "jwt like", credential: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", want: false},
		{name: "empty", credential: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsServiceKeyCredential(tt.credential); got != tt.want {
				t.Fatalf("IsServiceKeyCredential() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatServiceKeyAuthorization(t *testing.T) {
	if got := FormatServiceKeyAuthorization("sk-abc"); got != "Bearer sk-abc" {
		t.Fatalf("FormatServiceKeyAuthorization() = %q", got)
	}
	if got := FormatServiceKeyAuthorization("Bearer sk-abc"); got != "Bearer sk-abc" {
		t.Fatalf("FormatServiceKeyAuthorization() = %q", got)
	}
}

func TestValidateServiceKey(t *testing.T) {
	allowed := []string{"sk-valid"}
	if !ValidateServiceKey("sk-valid", allowed) {
		t.Fatal("expected valid service key")
	}
	if ValidateServiceKey("sk-invalid", allowed) {
		t.Fatal("expected invalid service key")
	}
	if ValidateServiceKey("Bearer sk-valid", allowed) {
		t.Fatal("credential must not include Bearer prefix")
	}
}

func TestParseBearerAuthorization(t *testing.T) {
	full, cred, ok := ParseBearerAuthorization("Bearer sk-abc")
	if !ok || full != "Bearer sk-abc" || cred != "sk-abc" {
		t.Fatalf("ParseBearerAuthorization() = (%q, %q, %v)", full, cred, ok)
	}
	_, _, ok = ParseBearerAuthorization("sk-abc")
	if ok {
		t.Fatal("expected missing scheme to fail")
	}
}
