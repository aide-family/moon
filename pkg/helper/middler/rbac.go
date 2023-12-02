package middler

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/redis/go-redis/v9"
	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/conn"
	"prometheus-manager/pkg/helper/model"
)

func RbacServer(cache ...*redis.Client) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 1. 解析jwt
			authClaims, ok := GetAuthClaims(ctx)
			if !ok {
				return nil, ErrTokenInvalid
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
			if len(cache) > 0 && cache[0] != nil {
				if err = model.CheckUserRoleExist(ctx, cache[0], authClaims.ID, authClaims.Role); err != nil {
					return nil, perrors.ErrorPermissionDenied("用户角色关系已变化, 请重新登录")
				}
			}

			return handler(ctx, req)
		}
	}
}
