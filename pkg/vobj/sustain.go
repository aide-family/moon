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
	// SustainUnknown 未知
	SustainUnknown Sustain = iota // 未知

	// SustainFor m时间内出现n次
	SustainFor // m时间内出现n次

	// SustainMax m时间内最多出现n次
	SustainMax // m时间内最多出现n次

	// SustainMin m时间内最少出现n次
	SustainMin // m时间内最少出现n次
)

// Judge 判断是否符合条件
func (s Sustain) Judge(condition Condition, count uint32, threshold float64) func(values []float64) bool {
	total := uint32(0)
	switch s {
	case SustainFor:
		return func(values []float64) bool {
			for _, v := range values {
				if condition.Judge(threshold, v) {
					total++
				} else {
					total = 0
				}
			}
			return total >= count
		}
	case SustainMax:
		return func(values []float64) bool {
			for _, v := range values {
				if condition.Judge(threshold, v) {
					total++
				}
			}
			return total <= count
		}
	case SustainMin:
		return func(values []float64) bool {
			for _, v := range values {
				if condition.Judge(threshold, v) {
					total++
				}
			}
			return total >= count
		}
	default:
		return func(values []float64) bool {
			return false
		}
	}
}
