package model

// Models 注册 model 下全部模型
func Models() []any {
	return []any{
		&SysAPI{},
		&SysMenu{},
		&SysTeam{},
		&SysTeamEmail{},
		&SysUser{},
		&SysDict{},
		&StrategyTemplate{},
		&StrategyLevelTemplate{},
		&StrategyTemplateCategories{},
		&SysOAuthUser{},
		&SysUserMessage{},
	}
}
