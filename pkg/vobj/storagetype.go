package vobj

// StorageType 存储类型
//
//go:generate stringer -type=StorageType -linecomment
type StorageType int

const (
	StorageTypeUnknown StorageType = iota // 未知

	StorageTypePrometheus // Prometheus

	StorageTypeVictoriametrics // VictoriaMetrics
)
