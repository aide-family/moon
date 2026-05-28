package middler

import (
	"context"
	"testing"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
)

func TestParseAuthorizationValueRawToken(t *testing.T) {
	full, cred, ok := parseAuthorizationValue("eyJhbGciOiJIUzI1NiJ9")
	if !ok || cred != "eyJhbGciOiJIUzI1NiJ9" || full != "Bearer eyJhbGciOiJIUzI1NiJ9" {
		t.Fatalf("parseAuthorizationValue() = (%q, %q, %v)", full, cred, ok)
	}
}

func TestAuthorizationFromClientMetadata(t *testing.T) {
	ctx := metadata.AppendToClientContext(context.Background(),
		cnst.MetadataGlobalKeyAuthorization, FormatServiceKeyAuthorization("sk-test"),
	)
	ctx = contextx.WithAuthMode(ctx, contextx.AuthModeServiceKey)

	full, cred, ok := authorizationFromContext(ctx)
	if !ok || cred != "sk-test" {
		t.Fatalf("authorizationFromContext() = (%q, %q, %v)", full, cred, ok)
	}
}

func TestAuthClientFromClientMetadataOnly(t *testing.T) {
	ctx := metadata.AppendToClientContext(context.Background(),
		cnst.MetadataGlobalKeyAuthorization, FormatServiceKeyAuthorization("sk-test"),
	)
	ctx = contextx.WithAuthMode(ctx, contextx.AuthModeServiceKey)

	clientHeader := headerCarrier{}
	clientCtx := transport.NewClientContext(ctx, &testTransport{
		kind:   string(transport.KindHTTP),
		header: clientHeader,
	})

	mw := AuthClient()
	_, err := mw(func(ctx context.Context, req any) (any, error) {
		return nil, nil
	})(clientCtx, nil)
	if err != nil {
		t.Fatalf("AuthClient() error = %v", err)
	}
	if got := clientHeader[cnst.HTTPHeaderAuthorization]; got != "Bearer sk-test" {
		t.Fatalf("client authorization = %q", got)
	}
}

func TestEnsureRequestAuthorizationHeader(t *testing.T) {
	header := headerCarrier{
		cnst.MetadataGlobalKeyAuthorization: "Bearer sk-abc",
	}
	ctx := transport.NewServerContext(context.Background(), &testTransport{
		kind:   string(transport.KindHTTP),
		header: header,
	})
	ensureRequestAuthorizationHeader(ctx)
	if got := header[cnst.HTTPHeaderAuthorization]; got != "Bearer sk-abc" {
		t.Fatalf("Authorization header = %q", got)
	}
}
