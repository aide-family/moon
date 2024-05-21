package bo

import (
	"github.com/aide-cloud/moon/api/merr"
)

var (
	UnLoginErr = merr.ErrorRedirect("请先登录").WithMetadata(map[string]string{
		"redirect": "/login",
	})
	// PasswordErr 密码错误
	PasswordErr = merr.ErrorAlert("密码错误").WithMetadata(map[string]string{
		"password": "密码错误",
	})
	// SystemErr 系统错误
	SystemErr = merr.ErrorNotification("系统错误")
	// NoPermissionErr 没有权限
	NoPermissionErr = merr.ErrorModal("没有权限")
	// UserNotFoundErr 用户不存在
	UserNotFoundErr = merr.ErrorAlert("用户不存在")
	// PasswordSameErr 密码不一致
	PasswordSameErr = merr.ErrorAlert("新旧密码一致")
	// AdminUserDeleteErr 不允许删除管理员
	AdminUserDeleteErr = merr.ErrorAlert("不允许删除管理员")
	// ResourceNotFoundErr 资源不存在
	ResourceNotFoundErr = merr.ErrorAlert("资源不存在")
)
