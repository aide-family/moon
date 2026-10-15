package datasource

import (
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestRedactClearsSecrets(t *testing.T) {
	item := redact(&bo.Datasource{Password: "plain-password", Secret: "plain-secret", SecretSet: true})
	if item.Password != "" || item.Secret != "" || !item.SecretSet {
		t.Fatalf("redacted = %+v", item)
	}
}
