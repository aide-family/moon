package builder

import (
	"context"
)

// NewParamsBuild 创建参数构造器
func NewParamsBuild(ctx context.Context) IPramsBuilder {
	return &paramsBuilder{ctx: ctx}
}

type (
	paramsBuilder struct {
		ctx context.Context
	}

	// IPramsBuilder 参数构造器
	IPramsBuilder interface {
		// PaginationModuleBuilder 分页模块构造器
		PaginationModuleBuilder() IPaginationModuleBuilder

		// AlarmNoticeGroupModuleBuilder 告警组模块构造器
		AlarmNoticeGroupModuleBuilder() IAlarmNoticeGroupModuleBuilder

		// UserModuleBuilder 用户模块构造器
		UserModuleBuilder() IUserModuleBuilder

		// TeamMemberModuleBuilder 团队成员模块构造器
		TeamMemberModuleBuilder() ITeamMemberModuleBuilder

		// DatasourceModuleBuilder 数据源模块构造器
		DatasourceModuleBuilder() IDatasourceModuleBuilder

		// MetricDataModuleBuilder 指标数据模块构造器
		MetricDataModuleBuilder() IMetricDataModuleBuilder

		// MetricModuleBuilder 指标模块构造器
		MetricModuleBuilder() IMetricModuleBuilder

		// DictModuleBuilder 字典模块构造器
		DictModuleBuilder() IDictModuleBuilder

		// HookModuleBuilder 钩子模块构造器
		HookModuleBuilder() IHookModuleBuilder

		MenuModuleBuilder() IMenuModuleBuilder

		// RealtimeAlarmModuleBuilder 实时告警模块构造器
		RealtimeAlarmModuleBuilder() IRealtimeAlarmModuleBuilder

		// ResourceModuleBuilder 资源模块构造器
		ResourceModuleBuilder() IResourceModuleBuilder

		// StrategyModuleBuilder 策略模块构造器
		StrategyModuleBuilder() IStrategyModuleBuilder

		// SubscriberModuleBuilder 订阅者模块构造器
		SubscriberModuleBuilder() ISubscriberModuleBuilder

		// RoleModuleBuilder 角色模块构造器
		RoleModuleBuilder() IRoleModuleBuilder

		// TeamModuleBuilder 团队模块构造器
		TeamModuleBuilder() ITeamModuleBuilder

		// InviteModuleBuilder 邀请模块构造器
		InviteModuleBuilder() InviteModuleBuilder

		// AlarmHistoryModuleBuilder 告警历史模块构造器
		AlarmHistoryModuleBuilder() IAlarmHistoryModuleBuilder

		// AlarmModuleBuilder 告警模块构造器
		AlarmModuleBuilder() IAlarmModuleBuilder

		// YamlModuleBuilder 文件模块构造器
		YamlModuleBuilder() IFileModuleBuild

		// AlarmSendModuleBuilder 告警发送模块构造器
		AlarmSendModuleBuilder() IAlarmSendModuleBuilder

		// TimeEngineRuleModuleBuilder 时间引擎规则模块构造器
		TimeEngineRuleModuleBuilder() ITimeEngineRuleModuleBuilder

		// TimeEngineModuleBuilder 时间引擎模块构造器
		TimeEngineModuleBuilder() ITimeEngineModuleBuilder

		// OauthModuleBuilder oauth模块构造器
		OauthModuleBuilder() IOauthModuleBuilder
	}
)

// TimeEngineModuleBuilder implements IPramsBuilder.
func (p *paramsBuilder) TimeEngineModuleBuilder() ITimeEngineModuleBuilder {
	return &timeEngineModuleBuilderImpl{ctx: p.ctx}
}

// TimeEngineRuleModuleBuilder implements IPramsBuilder.
func (p *paramsBuilder) TimeEngineRuleModuleBuilder() ITimeEngineRuleModuleBuilder {
	return &timeEngineRuleModuleBuilderImpl{ctx: p.ctx}
}

// OauthModuleBuilder implements IPramsBuilder.
func (p *paramsBuilder) OauthModuleBuilder() IOauthModuleBuilder {
	return NewOauthModuleBuilder(p.ctx)
}

func (p *paramsBuilder) AlarmSendModuleBuilder() IAlarmSendModuleBuilder {
	return &alarmSendModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) YamlModuleBuilder() IFileModuleBuild {
	return &fileModuleBuild{ctx: p.ctx}
}

func (p *paramsBuilder) AlarmModuleBuilder() IAlarmModuleBuilder {
	return &alarmModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) AlarmHistoryModuleBuilder() IAlarmHistoryModuleBuilder {
	return &alarmHistoryModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) InviteModuleBuilder() InviteModuleBuilder {
	return &inviteModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) TeamModuleBuilder() ITeamModuleBuilder {
	return &teamModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) TeaModuleBuilder() ITeamModuleBuilder {
	return &teamModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) RoleModuleBuilder() IRoleModuleBuilder {
	return &roleModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) SubscriberModuleBuilder() ISubscriberModuleBuilder {
	return &subscriberModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) StrategyModuleBuilder() IStrategyModuleBuilder {
	return &strategyModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) ResourceModuleBuilder() IResourceModuleBuilder {
	return &resourceModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) RealtimeAlarmModuleBuilder() IRealtimeAlarmModuleBuilder {
	return &realtimeAlarmModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) MenuModuleBuilder() IMenuModuleBuilder {
	return &menuModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) HookModuleBuilder() IHookModuleBuilder {
	return &hookModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) DictModuleBuilder() IDictModuleBuilder {
	return &dictModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) MetricModuleBuilder() IMetricModuleBuilder {
	return &metricModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) MetricDataModuleBuilder() IMetricDataModuleBuilder {
	return &metricDataModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) DatasourceModuleBuilder() IDatasourceModuleBuilder {
	return &datasourceModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) TeamMemberModuleBuilder() ITeamMemberModuleBuilder {
	return &teamMemberModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) UserModuleBuilder() IUserModuleBuilder {
	return &userModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) PaginationModuleBuilder() IPaginationModuleBuilder {
	return &paginationModuleBuilder{ctx: p.ctx}
}

func (p *paramsBuilder) AlarmNoticeGroupModuleBuilder() IAlarmNoticeGroupModuleBuilder {
	return &alarmNoticeGroupModuleBuilder{ctx: p.ctx}
}
