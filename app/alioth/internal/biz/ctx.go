package biz

import "context"

type ctxKey int

const (
	userKey ctxKey = iota
	namespaceKey
)

func WithUser(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userKey, uid)
}

func UserID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userKey).(string)
	return uid, ok && uid != ""
}

func WithNamespace(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, namespaceKey, uid)
}

func NamespaceID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(namespaceKey).(string)
	return uid, ok && uid != ""
}
