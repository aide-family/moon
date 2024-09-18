package middleware

import (
	"context"

	"github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// CheckRbacFun 权限校验函数
type CheckRbacFun func(ctx context.Context, operation string) (*authorization.CheckPermissionReply, error)

// Rbac 权限校验中间件
func Rbac(check CheckRbacFun) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			operation, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorNotification("get operation failed")
			}
			// 判断该用户在该资源是否有权限
			permission, err := check(ctx, operation.Operation())
			if err != nil {
				return nil, err
			}
			if !permission.GetHasPermission() {
				return nil, merr.ErrorI18nForbidden(ctx)
			}
			ctx = WithTeamRoleContextKey(ctx, vobj.Role(permission.GetTeamMember().GetRole()))
			return handler(ctx, req)
		}
	}
}
