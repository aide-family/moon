package middler

import (
	"context"
	"testing"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
)

type headerCarrier map[string]string

func (h headerCarrier) Get(key string) string { return h[key] }
func (h headerCarrier) Set(key string, value string) { h[key] = value }
func (h headerCarrier) Add(key string, value string) { h[key] = value }
func (h headerCarrier) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}
func (h headerCarrier) Values(key string) []string {
	if v := h.Get(key); v != "" {
		return []string{v}
	}
	return nil
}

type testTransport struct {
	kind   string
	header headerCarrier
}

func (t *testTransport) Kind() transport.Kind { return transport.Kind(t.kind) }
func (t *testTransport) Endpoint() string     { return "" }
func (t *testTransport) Operation() string    { return "" }
func (t *testTransport) RequestHeader() transport.Header { return t.header }
func (t *testTransport) ReplyHeader() transport.Header   { return t.header }

func TestAuthClientPropagatesExclusiveServiceKey(t *testing.T) {
	serverHeader := headerCarrier{
		cnst.HTTPHeaderAuthorization: "Bearer sk-test",
	}
	serverCtx := transport.NewServerContext(context.Background(), &testTransport{
		kind:   string(transport.KindHTTP),
		header: serverHeader,
	})
	serverCtx = contextx.WithAuthMode(serverCtx, contextx.AuthModeServiceKey)
	serverCtx = metadata.NewServerContext(serverCtx, metadata.Metadata{
		cnst.MetadataGlobalKeyAuthorization: []string{"Bearer sk-test"},
	})

	clientHeader := headerCarrier{}
	clientCtx := transport.NewClientContext(serverCtx, &testTransport{
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

func TestAuthClientPropagatesExclusiveJWT(t *testing.T) {
	serverHeader := headerCarrier{
		cnst.HTTPHeaderAuthorization: "Bearer eyJhbGciOiJIUzI1NiJ9",
	}
	serverCtx := transport.NewServerContext(context.Background(), &testTransport{
		kind:   string(transport.KindHTTP),
		header: serverHeader,
	})
	serverCtx = contextx.WithAuthMode(serverCtx, contextx.AuthModeJWT)

	clientHeader := headerCarrier{}
	clientCtx := transport.NewClientContext(serverCtx, &testTransport{
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
	if got := clientHeader[cnst.HTTPHeaderAuthorization]; got != "Bearer eyJhbGciOiJIUzI1NiJ9" {
		t.Fatalf("client authorization = %q", got)
	}
}
