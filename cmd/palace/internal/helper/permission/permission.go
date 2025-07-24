// Package permission is a permission package for kratos.
package permission

import (
	"context"
)

type userIDContextKey struct{}

// GetUserIDByContext Retrieve the user id from the context.
func GetUserIDByContext(ctx context.Context) (uint32, bool) {
	userID, ok := ctx.Value(userIDContextKey{}).(uint32)
	return userID, ok
}

func GetUserIDByContextWithDefault(ctx context.Context, defaultUserID uint32) uint32 {
	userID, ok := GetUserIDByContext(ctx)
	if !ok {
		return defaultUserID
	}
	return userID
}

// WithUserIDContext Set the user id in the context.
func WithUserIDContext(ctx context.Context, userID uint32) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

type teamIDContextKey struct{}

// GetTeamIDByContext Retrieve the team id from the context.
func GetTeamIDByContext(ctx context.Context) (uint32, bool) {
	teamID, ok := ctx.Value(teamIDContextKey{}).(uint32)
	return teamID, ok
}

// GetTeamIDByContextWithZeroValue Retrieve the team id from the context with a zero value.
func GetTeamIDByContextWithZeroValue(ctx context.Context) uint32 {
	teamID, ok := GetTeamIDByContext(ctx)
	if !ok {
		return 0
	}
	return teamID
}

// WithTeamIDContext Set the team id in the context.
func WithTeamIDContext(ctx context.Context, teamID uint32) context.Context {
	return context.WithValue(ctx, teamIDContextKey{}, teamID)
}

type tokenContextKey struct{}

// GetTokenByContext Retrieve the token from the context.
func GetTokenByContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenContextKey{}).(string)
	return token, ok
}

// WithTokenContext Set the token in the context.
func WithTokenContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

type operationContextKey struct{}

// GetOperationByContext Retrieve the operation from the context.
func GetOperationByContext(ctx context.Context) (string, bool) {
	operation, ok := ctx.Value(operationContextKey{}).(string)
	return operation, ok
}

// GetOperationByContextWithDefault Retrieve the operation from the context with a default value.
func GetOperationByContextWithDefault(ctx context.Context, defaultOperation ...string) string {
	operation, ok := GetOperationByContext(ctx)
	if ok {
		return operation
	}
	if len(defaultOperation) > 0 {
		return defaultOperation[0]
	}
	return ""
}

// WithOperationContext Set the operation in the context.
func WithOperationContext(ctx context.Context, operation string) context.Context {
	return context.WithValue(ctx, operationContextKey{}, operation)
}
