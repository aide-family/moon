package vobj

// EventDataType Event数据类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=EventDataType -linecomment
type EventDataType int

const (
	// EventDataTypeUnknown 未知
	EventDataTypeUnknown EventDataType = iota // 未知

	// EventDataTypeString string
	EventDataTypeString // string

	// EventDataTypeNumber number
	EventDataTypeNumber // number

	// EventDataTypeObject object
	EventDataTypeObject // object
)
