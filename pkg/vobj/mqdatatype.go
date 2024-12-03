package vobj

// MQDataType MQ数据类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=MQDataType -linecomment
type MQDataType int

const (
	// MQDataTypeUnknown 未知
	MQDataTypeUnknown MQDataType = iota // 未知

	// MQDataTypeString string
	MQDataTypeString // string

	// MQDataTypeNumber number
	MQDataTypeNumber // number

	// MQDataTypeObject object
	MQDataTypeObject // object
)
