package vobj

// Topic 消息类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Topic -linecomment
type Topic int

const (
	// TopicUnknown 未知
	TopicUnknown Topic = iota // 未知

	// TopicStrategy 策略
	TopicStrategy // 策略

	// TopicAlert 单条告警
	TopicAlert // 单条告警

	// TopicAlarm 多条告警
	TopicAlarm // 多条告警

	// TopicAlertMsg 告警消息
	TopicAlertMsg // 告警消息

	// TopicMQDatasource MQ 数据源
	TopicMQDatasource // MQ 数据源

	// TopicEventStrategy 事件策略
	TopicEventStrategy // 事件策略
)
