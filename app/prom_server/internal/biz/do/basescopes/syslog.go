package basescopes

import (
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const (
	SysLogPreloadKeyUser = "User"
)

const (
	SysLogTableFieldModule   Field = "module"
	SysLogTableFieldModuleId Field = "module_id"
)

// SysLogPreloadUsers 用户
func SysLogPreloadUsers(userIds ...uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) == 0 {
			return db.Preload(SysLogPreloadKeyUser)
		}
		return db.Preload(SysLogPreloadKeyUser, WhereInColumn(BaseFieldID, userIds...))
	}
}

// SysLogWhereModule .
func SysLogWhereModule(moduleName vo.Module, moduleId uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(SysLogTableFieldModule.String(), moduleName).
			Where(SysLogTableFieldModuleId.String(), moduleId)
	}
}
