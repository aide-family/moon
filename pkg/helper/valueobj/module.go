package valueobj

type Domain int32
type Module int32

const (
	_ Domain = iota
	// DomainSystem 系统领域
	DomainSystem
	// DomainMonitor 监控领域
	DomainMonitor
	// DomainBusiness 业务领域
	DomainBusiness
	// DomainOther 其他领域
	DomainOther
)

const (
	_ Module = iota
	// ModuleApi 接口模块
	ModuleApi
	// ModuleMenu 菜单模块
	ModuleMenu
	// ModuleRole 角色模块
	ModuleRole
	// ModuleUser 用户模块
	ModuleUser
	// ModuleDict 字典模块
	ModuleDict
	// ModuleConfig 配置模块
	ModuleConfig
	// ModuleLog 日志模块
	ModuleLog
	// ModuleJob 任务模块
	ModuleJob
	// ModuleNotify 通知模块
	ModuleNotify
	// ModuleSystem 系统模块
	ModuleSystem
	// ModuleMonitor 监控模块
	ModuleMonitor
	// ModuleBusiness 业务模块
	ModuleBusiness
	// ModuleOther 其他模块
	ModuleOther
)

// String 获取模块名称
func (m Module) String() string {
	switch m {
	case ModuleApi:
		return "权限模块"
	case ModuleMenu:
		return "菜单模块"
	case ModuleRole:
		return "角色模块"
	case ModuleUser:
		return "用户模块"
	case ModuleDict:
		return "字典模块"
	case ModuleConfig:
		return "配置模块"
	case ModuleLog:
		return "日志模块"
	case ModuleJob:
		return "任务模块"
	case ModuleNotify:
		return "通知模块"
	case ModuleSystem:
		return "系统模块"
	case ModuleMonitor:
		return "监控模块"
	case ModuleBusiness:
		return "业务模块"
	case ModuleOther:
		return "其他模块"
	default:
		return "未知模块"
	}
}

// String 获取领域名称
func (d Domain) String() string {
	switch d {
	case DomainSystem:
		return "系统领域"
	case DomainMonitor:
		return "监控领域"
	case DomainBusiness:
		return "业务领域"
	case DomainOther:
		return "其他领域"
	default:
		return "未知领域"
	}
}

// Value 获取模块值
func (m Module) Value() int32 {
	return int32(m)
}

// Value 获取领域值
func (d Domain) Value() int32 {
	return int32(d)
}
