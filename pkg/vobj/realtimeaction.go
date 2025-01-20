package vobj

// RealTimeAction 实时告警操作类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=RealTimeAction -linecomment
type RealTimeAction int

const (
	// RealTimeActionUnknown 未知
	RealTimeActionUnknown RealTimeAction = iota // 未知
	// RealTimeActionMark 标记
	RealTimeActionMark // 标记
	// RealTimeActionDelete 删除
	RealTimeActionDelete // 删除
	// RealTimeActionSuppress 抑制
	RealTimeActionSuppress // 抑制
	// RealTimeActionUpgrade 升级
	RealTimeActionUpgrade // 升级
)

// EnUSString 英文字符串
func (a RealTimeAction) EnUSString() string {
	switch a {
	case RealTimeActionMark:
		return "mark"
	case RealTimeActionDelete:
		return "delete"
	case RealTimeActionSuppress:
		return "suppress"
	case RealTimeActionUpgrade:
		return "upgrade"
	}
	return "other"
}
