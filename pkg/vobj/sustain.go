package vobj

//var _ driver.Valuer = Sustain(0)

// Sustain 持续类型定义
//
//	m时间内出现n次
//	m时间内最多出现n次
//	m时间内最少出现n次
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Sustain -linecomment
type Sustain int

const (
	SustainUnknown Sustain = iota // 未知

	SustainFor // m时间内出现n次

	SustainMax // m时间内最多出现n次

	SustainMin // m时间内最少出现n次
)
