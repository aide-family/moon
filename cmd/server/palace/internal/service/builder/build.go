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

	IPramsBuilder interface {
		PaginationModuleBuilder() IPaginationModuleBuilder

		AlarmNoticeGroupModuleBuilder() IAlarmNoticeGroupModuleBuilder

		UserModuleBuilder() IUserModuleBuilder

		TeamMemberModuleBuilder() ITeamMemberModuleBuilder

		DatasourceModuleBuilder() IDatasourceModuleBuilder

		MetricDataModuleBuilder() IMetricDataModuleBuilder

		MetricModuleBuilder() IMetricModuleBuilder

		DictModuleBuilder() IDictModuleBuilder

		HookModuleBuilder() IHookModuleBuilder

		MenuModuleBuilder() IMenuModuleBuilder

		RealtimeAlarmModuleBuilder() IRealtimeAlarmModuleBuilder

		ResourceModuleBuilder() IResourceModuleBuilder

		StrategyModuleBuilder() IStrategyModuleBuilder

		SubscriberModuleBuilder() ISubscriberModuleBuilder

		RoleModuleBuilder() IRoleModuleBuilder

		TeamModuleBuilder() ITeamModuleBuilder

		InviteModuleBuilder() InviteModuleBuilder

		AlarmHistoryModuleBuilder() IAlarmHistoryModuleBuilder

		AlarmModuleBuilder() IAlarmModuleBuilder

		YamlModuleBuilder() IFileModuleBuild

		AlarmSendModuleBuilder() IAlarmSendModuleBuilder
	}
)

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
