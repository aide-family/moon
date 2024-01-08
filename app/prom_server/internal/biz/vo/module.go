package vo

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

// Remark 领域说明
func (d Domain) Remark() string {
	switch d {
	case DomainSystem:
		return "系统领域，包含系统模块、系统配置模块、系统日志模块、系统任务模块、系统通知模块、系统监控模块"
	case DomainMonitor:
		return "监控领域，包含监控模块、监控配置模块、监控日志模块、监控任务模块、监控通知模块"
	case DomainBusiness:
		return "业务领域，包含业务模块、业务配置模块、业务日志模块、业务任务模块、业务通知模块"
	case DomainOther:
		return "其他领域，包含其他模块、其他配置模块、其他日志模块、其他任务模块、其他通知模块"
	default:
		return "未知领域"
	}
}

// Remark 模块说明
func (m Module) Remark() string {
	switch m {
	case ModuleApi:
		return "接口模块，包含权限模块、菜单模块、角色模块、用户模块、字典模块、配置模块、日志模块、任务模块、通知模块"
	case ModuleMenu:
		return "菜单模块，包含菜单模块、角色模块、用户模块、字典模块、配置模块、日志模块、任务模块、通知模块"
	case ModuleRole:
		return "角色模块，包含角色模块、用户模块、字典模块、配置模块、日志模块、任务模块、通知模块"
	case ModuleUser:
		return "用户模块，包含用户模块、字典模块、配置模块、日志模块、任务模块、通知模块"
	case ModuleDict:
		return "字典模块，包含字典模块、配置模块、日志模块、任务模块、通知模块"
	case ModuleConfig:
		return "配置模块，包含配置模块、日志模块、任务模块、通知模块"
	case ModuleLog:
		return "日志模块，包含日志模块、任务模块、通知模块"
	case ModuleJob:
		return "任务模块，包含任务模块、通知模块"
	default:
		return "未知模块"
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
