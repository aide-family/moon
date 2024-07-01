package vobj

// SendType 字典类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=SendType -linecomment
type SendType int

const (
	SendTypeUnknown SendType = iota // 未知

	SendTypeInhibit // 抑制

	SendTypeAggregate // 聚合
)
