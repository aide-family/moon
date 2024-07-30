package bizmodel

// Models 注册biz model下全部模型
func Models() []any {
	return []any{
		&CasbinRule{},
		&Datasource{},
		&MetricLabelValue{},
		&DatasourceMetric{},
		&MetricLabel{},
		&SysTeamAPI{},
		&SysTeamMemberRole{},
		&SysTeamMember{},
		&SysTeamRoleAPI{},
		&SysTeamRole{},
		&SysTeamMenu{},
		&SysDict{},
		&Strategy{},
		&StrategyLevel{},
		&StrategyTemplate{},
		&StrategyLevelTemplate{},
		&SendStrategy{},
		&StrategyGroup{},
	}
}

// AlarmModels 注册biz model下告警相关模型
func AlarmModels() []any {
	return []any{
		&RealtimeAlarm{},
	}
}
