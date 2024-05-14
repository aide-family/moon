package bo

import (
	"github.com/aide-cloud/moon/api/merr"
)

var UnLoginErr = merr.ErrorRedirect("请先登录").WithMetadata(map[string]string{
	"redirect": "/login",
})

// PasswordErr 密码错误
var PasswordErr = merr.ErrorAlert("密码错误").WithMetadata(map[string]string{
	"password": "密码错误",
})

// SystemErr 系统错误
var SystemErr = merr.ErrorNotification("系统错误")
