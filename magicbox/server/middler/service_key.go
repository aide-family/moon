package middler

import (
	"context"
	"crypto/subtle"
	"strings"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func IsServiceKeyCredential(credential string) bool {
	return strings.HasPrefix(strings.TrimSpace(credential), cnst.HTTPHeaderServiceKeyPrefix)
}

func FormatServiceKeyAuthorization(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(key), strings.ToLower(cnst.HTTPHeaderBearerPrefix)+" ") {
		return key
	}
	return cnst.HTTPHeaderBearerPrefix + " " + key
}

func ValidateServiceKey(provided string, allowedKeys []string) bool {
	provided = strings.TrimSpace(provided)
	if provided == "" || !IsServiceKeyCredential(provided) {
		return false
	}
	for _, allowed := range allowedKeys {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(allowed)) == 1 {
			return true
		}
	}
	return false
}

func ParseBearerAuthorization(authHeader string) (fullAuth string, credential string, ok bool) {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return "", "", false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], cnst.HTTPHeaderBearerPrefix) {
		return "", "", false
	}
	credential = strings.TrimSpace(parts[1])
	if credential == "" {
		return "", "", false
	}
	return authHeader, credential, true
}

func ServiceKeyServe(allowedKeys []string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			ensureRequestAuthorizationHeader(ctx)
			_, credential, ok := authorizationFromContext(ctx)
			if !ok {
				return nil, merr.ErrorUnauthorized("service key is required")
			}
			if !IsServiceKeyCredential(credential) {
				return nil, merr.ErrorUnauthorized("service key is invalid")
			}
			if len(allowedKeys) == 0 {
				return nil, merr.ErrorUnauthorized("service key auth is not configured")
			}
			if !ValidateServiceKey(credential, allowedKeys) {
				return nil, merr.ErrorUnauthorized("service key is invalid")
			}
			ctx = contextx.WithAuthMode(ctx, contextx.AuthModeServiceKey)
			return handler(ctx, req)
		}
	}
}

func BindServiceKeyToken() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if !contextx.IsServiceKeyAuth(ctx) {
				return handler(ctx, req)
			}
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorUnauthorized("wrong context for middleware")
			}
			fullAuth, _, ok := authorizationFromContext(ctx)
			if !ok || strutil.IsEmpty(fullAuth) {
				return nil, merr.ErrorUnauthorized("service key is required")
			}
			tr.RequestHeader().Set(cnst.HTTPHeaderAuthorization, fullAuth)
			tr.RequestHeader().Set(cnst.MetadataGlobalKeyAuthorization, fullAuth)
			return handler(ctx, req)
		}
	}
}
