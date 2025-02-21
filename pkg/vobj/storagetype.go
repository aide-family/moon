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

	StorageTypeElasticsearch // Elasticsearch
	StorageTypeLoki          // Loki

	StorageAliYunSLS // AliYunSLS
)

const (
	// StorageTypeKafka Kafka
	StorageTypeKafka StorageType = iota + 10 // Kafka

	// StorageTypeRabbitMQ RabbitMQ
	StorageTypeRabbitMQ // RabbitMQ

	// StorageTypeRocketMQ RocketMQ
	StorageTypeRocketMQ // RocketMQ

	// StorageTypeMQTT MQTT
	StorageTypeMQTT // MQTT
)
