// Package contextx provides contextx for the application.
package contextx

import (
	"context"

	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
)

type (
	namespaceKey struct{}
	userUIDKey   struct{}
	usernameKey  struct{}
)

func WithNamespace(ctx context.Context, namespace snowflake.ID) context.Context {
	klog.Debugw("msg", "with namespace uid", "namespaceUID", namespace)
	return context.WithValue(ctx, namespaceKey{}, namespace)
}

func GetNamespace(ctx context.Context) snowflake.ID {
	return ctx.Value(namespaceKey{}).(snowflake.ID)
}

func WithUserUID(ctx context.Context, userUID snowflake.ID) context.Context {
	klog.Debugw("msg", "with user uid", "userUID", userUID)
	return context.WithValue(ctx, userUIDKey{}, userUID)
}

func GetUserUID(ctx context.Context) snowflake.ID {
	return ctx.Value(userUIDKey{}).(snowflake.ID)
}

func WithUsername(ctx context.Context, username string) context.Context {
	klog.Debugw("msg", "with username", "username", username)
	return context.WithValue(ctx, usernameKey{}, username)
}

func GetUsername(ctx context.Context) string {
	return ctx.Value(usernameKey{}).(string)
}
