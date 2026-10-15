package actor

import "context"

type ctxKey int

const (
	userKey ctxKey = iota
	namespaceKey
)

// WithUser 把已认证用户写入上下文。后续用例只从这里读取操作者。
func WithUser(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userKey, uid)
}

// UserID 返回当前用户。缺少令牌或空 uid 时 ok 为 false。
func UserID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(userKey).(string)
	return uid, ok && uid != ""
}

// WithNamespace 把 X-Namespace 写入上下文。
func WithNamespace(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, namespaceKey, uid)
}

// NamespaceID 返回当前租户。未选择租户时 ok 为 false。
func NamespaceID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(namespaceKey).(string)
	return uid, ok && uid != ""
}
