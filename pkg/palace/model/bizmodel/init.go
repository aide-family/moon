package bizmodel

// Models 注册biz model下全部模型
func Models() []any {
	return []any{
		&CasbinRule{},
		&Datasource{},
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
		&StrategyGroupCategories{},
		&Dashboard{},
		&DashboardChart{},
		&AlarmNoticeGroup{},
		&AlarmNoticeMember{},
		&DashboardSelf{},
		&AlarmPageSelf{},
		&AlarmHook{},
		&StrategyLabelNotice{},
		&StrategySubscriber{},
		&StrategyDomain{},
		&StrategyHTTP{},
		&StrategyPing{},
		&MqDatasource{},
	}
}
