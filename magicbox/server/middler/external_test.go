package middler

import (
	"context"
	"testing"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/metadata"
)

func TestExternalClientContextPropagatesNamespaceFromContext(t *testing.T) {
	namespaceUID := snowflake.ID(2029537653437005824)
	ctx := contextx.WithNamespace(context.Background(), namespaceUID)
	ctx = contextx.WithAuthMode(ctx, contextx.AuthModeServiceKey)

	outCtx := ExternalClientContext(ctx, &config.ExternalDomainConfig{
		ServiceKey: "sk-rabbit-dev",
		Namespace:  "999",
	})

	md, ok := metadata.FromClientContext(outCtx)
	if !ok {
		t.Fatal("expected client metadata")
	}
	if got := md.Get(cnst.MetadataGlobalKeyNamespace); got != "2029537653437005824" {
		t.Fatalf("namespace metadata = %q, want context namespace", got)
	}
	if got := md.Get(cnst.MetadataGlobalKeyAuthorization); got != "Bearer sk-rabbit-dev" {
		t.Fatalf("authorization metadata = %q", got)
	}
}
