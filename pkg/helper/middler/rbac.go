package middler

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/conn"
)

const AdminRole = "1"

type CheckUserRoleExistFun func(ctx context.Context, userID uint32, roleID string) error
type GetApiIDByPathAndMethodFun func(ctx context.Context, path, method string) (uint64, error)

func RbacServer(checkFun CheckUserRoleExistFun, getApiFun GetApiIDByPathAndMethodFun) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 1. 解析jwt
			authClaims, ok := GetAuthClaims(ctx)
			if !ok {
				return nil, ErrTokenInvalid
			}

			if authClaims.Role == AdminRole {
				return handler(ctx, req)
			}

			path := GetPath(ctx)
			method := GetMethod(ctx)
			// 2. 校验权限
			enforcer := conn.Enforcer()
			has, err := enforcer.Enforce(authClaims.Role, path, method)
			if err != nil {
				return nil, perrors.ErrorUnknown("系统错误")
			}
			if !has {
				return nil, perrors.ErrorPermissionDenied("请联系管理员分配权限")
			}

			// 3. 校验用户是否具备这个角色, 避免角色被删除后, 用户仍然具备这个角色
			if checkFun != nil {
				if err = checkFun(ctx, authClaims.ID, authClaims.Role); err != nil {
					return nil, perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
				}
			}
			if getApiFun != nil {
				if _, err = getApiFun(ctx, path, method); err != nil {
					return nil, err
				}
			}

			return handler(ctx, req)
		}
	}
}
