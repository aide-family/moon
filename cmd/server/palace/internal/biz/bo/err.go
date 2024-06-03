package bo

import (
	"context"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/go-kratos/kratos/v2/errors"
)

func UnLoginErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nUnLoginErr(ctx).WithMetadata(map[string]string{
		"redirect": "/login",
	})
}

// PasswordErr 密码错误
func PasswordErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nPasswordErr(ctx).WithMetadata(map[string]string{
		"password": "密码错误",
	})
}

// SystemErr 系统错误
func SystemErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nSystemErr(ctx)
}

// NoPermissionErr 没有权限
func NoPermissionErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nNoPermissionErr(ctx)
}

// UserNotFoundErr 用户不存在
func UserNotFoundErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nUserNotFoundErr(ctx)
}

// PasswordSameErr 新旧密码不能相同
func PasswordSameErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nPasswordSameErr(ctx)
}

// AdminUserDeleteErr 不允许删除管理员
func AdminUserDeleteErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nAdminUserDeleteErr(ctx)
}

// ResourceNotFoundErr 资源不存在
func ResourceNotFoundErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nResourceNotFoundErr(ctx)
}

// TeamNotFoundErr 团队不存在
func TeamNotFoundErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nTeamNotFoundErr(ctx)
}

// TeamLeaderErr 你不是团队的负责人
func TeamLeaderErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nTeamLeaderErr(ctx)
}

// TeamLeaderRepeatErr 你已经是团队负责人了
func TeamLeaderRepeatErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nTeamLeaderRepeatErr(ctx)
}

// DeleteSelfErr 不允许删除自己
func DeleteSelfErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nDeleteSelfErr(ctx)
}

// TeamRoleNotFoundErr 团队角色不存在
func TeamRoleNotFoundErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nTeamRoleNotFoundErr(ctx)
}

// TeamNameExistErr 团队名称已存在
func TeamNameExistErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nTeamNameExistErr(ctx)
}

// DatasourceNotFoundErr 数据源不存在
func DatasourceNotFoundErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nDatasourceNotFoundErr(ctx)
}

// LockFailedErr 加锁失败
func LockFailedErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nLockFailedErr(ctx)
}

// RetryLaterErr 数据源同步中，请稍后重试
func RetryLaterErr(ctx context.Context) *errors.Error {
	return merr.ErrorI18nRetryLaterErr(ctx)
}
