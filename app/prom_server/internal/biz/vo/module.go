package vo

type Domain int32
type Module int32

const (
	// DomainOther 其他领域
	DomainOther Domain = iota
	// DomainSystem 系统领域
	DomainSystem
	// DomainMonitor 监控领域
	DomainMonitor
	// DomainBusiness 业务领域
	DomainBusiness
)

const (
	// ModuleOther 其他模块
	ModuleOther Module = iota
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
	// ModuleRealtimeAlarm 实时告警模块
	ModuleRealtimeAlarm
	// ModuleAlarmHistory 告警历史模块
	ModuleAlarmHistory
	// ModuleStrategyGroup 策略组模块
	ModuleStrategyGroup
	// ModuleStrategy 策略模块
	ModuleStrategy
	// ModuleDatasource 数据源模块
	ModuleDatasource
	// ModuleAlarmPage 告警页面模块
	ModuleAlarmPage
	// ModuleAlarmNotifyGroup 告警通知组模块
	ModuleAlarmNotifyGroup
	// ModuleAlarmNotifyHook 告警通知机器人模块
	ModuleAlarmNotifyHook
	// ModuleDashboardChart 仪表盘图表模块
	ModuleDashboardChart
	// ModuleDashboard 仪表盘模块
	ModuleDashboard
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
	case ModuleRealtimeAlarm:
		return "实时告警模块"
	case ModuleAlarmHistory:
		return "告警历史模块"
	case ModuleStrategyGroup:
		return "策略组模块"
	case ModuleStrategy:
		return "策略模块"
	case ModuleDatasource:
		return "数据源模块"
	case ModuleAlarmPage:
		return "告警页面模块"
	case ModuleAlarmNotifyGroup:
		return "告警通知组模块"
	case ModuleAlarmNotifyHook:
		return "告警通知机器人模块"
	case ModuleDashboardChart:
		return "仪表板图表模块"
	case ModuleDashboard:
		return "仪表板模块"
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
