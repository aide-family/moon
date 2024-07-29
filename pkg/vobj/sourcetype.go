package vobj

// SourceType 来源类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=SourceType -linecomment
type SourceType int

const (
	// SourceTypeUnknown 未知
	SourceTypeUnknown SourceType = iota // 未知

	SourceTypeSystem // 系统来源

	SourceTypeTeam // 团队来源
)

const (
	SourceCodeSystem = "System"
	SourceCodeTeam   = "Team"
)

func GetSourceType(sourceCode string) SourceType {
	switch sourceCode {
	case SourceCodeSystem:
		return SourceTypeSystem
	case SourceCodeTeam:
		return SourceTypeTeam
	default:
		return SourceTypeTeam
	}
}
