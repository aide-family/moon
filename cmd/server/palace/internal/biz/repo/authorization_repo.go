package repo

type AuthorizationRepo interface {
	// 检查用户是否被团队禁用
	// 检查用户是否有该资源权限
	// 检查token是否过期
	// 检查token是否被登出
	// 检查用户是否被系统禁用
}
