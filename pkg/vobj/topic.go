package vobj

// Topic 消息类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Topic -linecomment
type Topic int

const (
	TopicUnknown Topic = iota // 未知

	// TODO 定义其他的消息类型
)
