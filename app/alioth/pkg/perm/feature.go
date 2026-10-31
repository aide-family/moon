package perm

import "strings"

// 功能把菜单和接口按模块打包。功能由代码维护，超级管理员用套餐把功能开通给租户，
// 租户只能基于已开通功能给角色授权。namespace:* 不属于任何功能，始终可用。
const (
	FeatureAlert        = "alert"
	FeatureAlertRule    = "alert_rule"
	FeatureProbe        = "probe"
	FeatureDatasource   = "datasource"
	FeatureNotification = "notification"
	FeatureMember       = "member"
	FeatureRole         = "role"
	FeatureCluster      = "cluster"
	FeatureAgent        = "agent"
	FeatureSSH          = "ssh"
)

// Feature 是一个功能及其包含的租户权限。
// Permissions 的顺序与 Catalog 一致，供租户角色授权页按功能分组展示。
type Feature struct {
	Code        string
	Name        string
	Description string
	Permissions []string
}

// featureOrder 是功能的展示顺序，也是 EffectiveFeatures 的返回顺序。
// 基础套餐按这个顺序取前七个：告警链路与租户日常管理，不含 Kubernetes、机器和 SSH。
var featureOrder = []string{
	FeatureAlert,
	FeatureAlertRule,
	FeatureProbe,
	FeatureDatasource,
	FeatureNotification,
	FeatureMember,
	FeatureRole,
	FeatureCluster,
	FeatureAgent,
	FeatureSSH,
}

type featureInfo struct {
	name        string
	description string
}

var featureInfos = map[string]featureInfo{
	FeatureAlert:        {name: "告警", description: "查看告警事件、处理与干预告警。"},
	FeatureAlertRule:    {name: "告警规则", description: "配置告警规则、规则组与告警级别。"},
	FeatureProbe:        {name: "探测目标", description: "管理探测目标组和探测地址。"},
	FeatureDatasource:   {name: "数据源", description: "接入指标与日志数据源。"},
	FeatureNotification: {name: "通知", description: "配置通知渠道、模板、通知组与订阅。"},
	FeatureMember:       {name: "成员", description: "邀请、移除成员并分配角色。"},
	FeatureRole:         {name: "角色", description: "管理租户角色和角色权限。"},
	FeatureCluster:      {name: "Kubernetes", description: "接入 Kubernetes 集群并使用资源级控制台。"},
	FeatureAgent:        {name: "机器", description: "管理边缘机器和 Agent 状态。"},
	FeatureSSH:          {name: "SSH 与 Web Shell", description: "使用 SSH 命令和 Web Shell。"},
}

// Features 返回全部功能。权限取自 Catalog，新增或调整资源权限会自动映射到功能。
func Features() []Feature {
	items := make([]Feature, 0, len(featureOrder))
	index := make(map[string]int, len(featureOrder))
	for _, code := range featureOrder {
		index[code] = len(items)
		items = append(items, Feature{Code: code, Name: featureInfos[code].name, Description: featureInfos[code].description})
	}
	for _, item := range Catalog() {
		code := featureForResource(item.Resource)
		if code == "" {
			continue
		}
		position, ok := index[code]
		if !ok {
			continue
		}
		items[position].Permissions = append(items[position].Permissions, item.Name)
	}
	return items
}

// featureForResource 把租户权限的资源映射到功能。namespace 不属于任何功能，
// Kubernetes 资源级权限统一归 cluster 功能。
func featureForResource(resource string) string {
	if strings.HasPrefix(resource, "k8s:") {
		return FeatureCluster
	}
	switch resource {
	case "alert":
		return FeatureAlert
	case "alert_rule":
		return FeatureAlertRule
	case "probe":
		return FeatureProbe
	case "datasource":
		return FeatureDatasource
	case "notification":
		return FeatureNotification
	case "member":
		return FeatureMember
	case "role":
		return FeatureRole
	case "cluster":
		return FeatureCluster
	case "agent":
		return FeatureAgent
	case "ssh":
		return FeatureSSH
	default:
		return ""
	}
}

// FeatureOf 返回租户权限所属的模块；namespace:* 不受功能限制，返回空字符串。
// access 用它判断一次权限检查是否还要看租户的功能开通情况。
// 菜单权限 menu:<能力> 返回该能力所属的模块；能力编码本身也返回所属模块。
func FeatureOf(permission string) string {
	if menuCode, ok := MenuOfMenuPermission(permission); ok {
		if menu, exists := MenuOf(menuCode); exists {
			return menu.Group
		}
		return ""
	}
	if item, ok := CapabilityOf(permission); ok {
		return item.Module
	}
	for _, item := range Catalog() {
		if item.Name == permission {
			return featureForResource(item.Resource)
		}
	}
	return ""
}

// IsFeature 判断功能编码是否在目录里。
func IsFeature(code string) bool {
	_, ok := featureInfos[code]
	return ok
}

// BaseFeatures 返回基础套餐的功能：告警、探测、数据源、通知和租户日常管理。
// 新租户默认使用基础套餐，Kubernetes、机器和 SSH 由超级管理员另行开通。
func BaseFeatures() []string {
	return []string{
		FeatureAlert,
		FeatureAlertRule,
		FeatureProbe,
		FeatureDatasource,
		FeatureNotification,
		FeatureMember,
		FeatureRole,
	}
}

// EffectiveFeatures 在套餐功能上叠加租户差异：enabled=true 的加上，false 的减掉。
// 结果按 featureOrder 排序，目录里不存在的旧功能被忽略。
func EffectiveFeatures(packageFeatures []string, overrides map[string]bool) []string {
	enabled := make(map[string]struct{}, len(packageFeatures))
	for _, code := range packageFeatures {
		enabled[code] = struct{}{}
	}
	for code, value := range overrides {
		if value {
			enabled[code] = struct{}{}
			continue
		}
		delete(enabled, code)
	}
	out := make([]string, 0, len(enabled))
	for _, code := range featureOrder {
		if _, ok := enabled[code]; ok {
			out = append(out, code)
		}
	}
	return out
}

// RolePermissionsForFeatures 返回预置角色在指定功能范围内应持有的权限。
// 基础套餐用它给新租户预置角色授仅有的权限。
func RolePermissionsForFeatures(role string, features []string) []string {
	return FilterPermissionsByFeatures(RolePermissions(role), features)
}

// FilterPermissionsByFeatures 只保留功能已开通的权限；namespace:* 不受功能限制，始终保留。
// 角色授权页、当前用户权限和角色权限写入都用它，保证菜单与接口看到同一份范围。
func FilterPermissionsByFeatures(names []string, features []string) []string {
	allowed := make(map[string]struct{}, len(features))
	for _, code := range features {
		allowed[code] = struct{}{}
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		code := FeatureOf(name)
		if code == "" {
			out = append(out, name)
			continue
		}
		if _, ok := allowed[code]; ok {
			out = append(out, name)
		}
	}
	return out
}

// FeatureCodes 按展示顺序返回功能编码，供仓储按固定顺序输出有效功能。
func FeatureCodes() []string {
	out := make([]string, len(featureOrder))
	copy(out, featureOrder)
	return out
}
