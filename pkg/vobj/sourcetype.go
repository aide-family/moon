package vobj

// SourceType 来源类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=SourceType -linecomment
type SourceType int

const (
	// SourceTypeUnknown 未知
	SourceTypeUnknown SourceType = iota // 未知

	// SourceTypeSystem 系统来源
	SourceTypeSystem // 系统来源

	// SourceTypeTeam 团队来源
	SourceTypeTeam // 团队来源
)

const (
	sourceCodeSystem = "System"
	sourceCodeTeam   = "Team"
)

// GetSourceType 根据来源编码获取来源类型
func GetSourceType(sourceCode string) SourceType {
	switch sourceCode {
	case sourceCodeSystem:
		return SourceTypeSystem
	case sourceCodeTeam:
		return SourceTypeTeam
	default:
		return SourceTypeTeam
	}
}
