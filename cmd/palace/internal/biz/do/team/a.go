package team

func Models() []any {
	return []any{
		&SmsConfig{},
		&EmailConfig{},
		&Dashboard{},
		&DashboardChart{},
		&DatasourceMetricMetadata{},
		&DatasourceMetric{},
		&Dict{},
		&OperateLog{},
		&NoticeGroup{},
		&NoticeHook{},
		&NoticeMember{},
		&Strategy{},
		&StrategyGroup{},
		&StrategyMetric{},
		&StrategySubscriber{},
		&StrategyMetricRule{},
		&StrategyMetricRuleLabelNotice{},
		&TimeEngine{},
		&TimeEngineRule{},
	}
}
