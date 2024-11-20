package vobj

// StorageType 存储类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=StorageType -linecomment
type StorageType int

const (
	// StorageTypeUnknown 未知
	StorageTypeUnknown StorageType = iota // 未知

	// StorageTypePrometheus Prometheus
	StorageTypePrometheus // Prometheus

	// StorageTypeVictoriametrics VictoriaMetrics
	StorageTypeVictoriametrics // VictoriaMetrics
)

const (
	StorageTypeMQUnknown StorageType = iota + 9
	StorageTypeKafka                 // Kafka
	StorageTypeRabbitMQ              // RabbitMQ
	StorageTypeRocketMQ              // RocketMQ
	StorageTypeMQTT                  // MQTT
)
