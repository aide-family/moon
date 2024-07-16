package vobj

// 策略模板来源
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=TemplateSourceType -linecomment
type TemplateSourceType int

const (
	// TemplateSourceTypeUnknown 未知
	TemplateSourceTypeUnknown TemplateSourceType = iota // 未知

	// TemplateSourceTypeSystem 系统
	TemplateSourceTypeSystem // 系统

	// TemplateSourceTypeTeam 团队
	TemplateSourceTypeTeam // 团队
)
