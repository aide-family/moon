package contextx

import "context"

type authModeKey struct{}

// AuthMode identifies the exclusive authentication method for the request chain.
type AuthMode string

const (
	AuthModeJWT        AuthMode = "jwt"
	AuthModeServiceKey AuthMode = "service_key"
)

func WithAuthMode(ctx context.Context, mode AuthMode) context.Context {
	return context.WithValue(ctx, authModeKey{}, mode)
}

func GetAuthMode(ctx context.Context) AuthMode {
	mode, _ := ctx.Value(authModeKey{}).(AuthMode)
	return mode
}

func IsServiceKeyAuth(ctx context.Context) bool {
	return GetAuthMode(ctx) == AuthModeServiceKey
}

func IsJWTAuth(ctx context.Context) bool {
	return GetAuthMode(ctx) == AuthModeJWT
}
