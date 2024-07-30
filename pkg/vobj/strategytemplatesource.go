package vobj

// StrategyTemplateSource 消息类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=StrategyTemplateSource -linecomment
type StrategyTemplateSource int

const (
	// StrategyTemplateSourceUnknown 未知
	StrategyTemplateSourceUnknown StrategyTemplateSource = iota // 未知

	// StrategyTemplateSourceSystem 系统来源
	StrategyTemplateSourceSystem // 系统来源

	// StrategyTemplateSourceTeam 团队来源
	StrategyTemplateSourceTeam // 团队来源
)
