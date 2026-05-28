package middler

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/transport"
)

// parseAuthorizationValue accepts "Bearer <token>" or a raw token / sk-xxx value.
func parseAuthorizationValue(raw string) (fullAuth string, credential string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}
	if full, cred, ok := ParseBearerAuthorization(raw); ok {
		return full, cred, ok
	}
	credential = raw
	return cnst.HTTPHeaderBearerPrefix + " " + credential, credential, true
}

// authorizationFromContext resolves credentials from Authorization, x-md-global-authorization,
// and Kratos client/server metadata (used by outbound clients and metadata middleware).
func authorizationFromContext(ctx context.Context) (fullAuth string, credential string, ok bool) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if full, cred, ok := authorizationFromHeader(tr.RequestHeader()); ok {
			return full, cred, ok
		}
	}
	if tr, ok := transport.FromClientContext(ctx); ok {
		if full, cred, ok := authorizationFromHeader(tr.RequestHeader()); ok {
			return full, cred, ok
		}
	}
	if md, ok := metadata.FromClientContext(ctx); ok {
		if full, cred, ok := parseAuthorizationValue(md.Get(cnst.MetadataGlobalKeyAuthorization)); ok {
			return full, cred, ok
		}
	}
	if md, ok := metadata.FromServerContext(ctx); ok {
		if full, cred, ok := parseAuthorizationValue(md.Get(cnst.MetadataGlobalKeyAuthorization)); ok {
			return full, cred, ok
		}
	}
	return "", "", false
}

func authorizationFromHeader(header transport.Header) (fullAuth string, credential string, ok bool) {
	if header == nil {
		return "", "", false
	}
	for _, key := range []string{cnst.HTTPHeaderAuthorization, cnst.MetadataGlobalKeyAuthorization} {
		if full, cred, ok := parseAuthorizationValue(header.Get(key)); ok {
			return full, cred, ok
		}
	}
	return "", "", false
}

// ensureRequestAuthorizationHeader copies propagated metadata auth onto Authorization for JWT middleware.
func ensureRequestAuthorizationHeader(ctx context.Context) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return
	}
	if _, _, ok := parseAuthorizationValue(tr.RequestHeader().Get(cnst.HTTPHeaderAuthorization)); ok {
		return
	}
	if full, _, ok := parseAuthorizationValue(tr.RequestHeader().Get(cnst.MetadataGlobalKeyAuthorization)); ok {
		tr.RequestHeader().Set(cnst.HTTPHeaderAuthorization, full)
	}
}
