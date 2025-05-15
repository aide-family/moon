package vobj

// StrategyType strategy type
//
//go:generate stringer -type=StrategyType -linecomment -output=type_strategy.string.go
type StrategyType int8

const (
	StrategyTypeUnknown StrategyType = iota // unknown
	StrategyTypeMetric                      // metric
	StrategyTypeEvent                       // event
	StrategyTypeLogs                        // logs
	StrategyTypePort                        // port
	StrategyTypeHTTP                        // http
	StrategyTypePing                        // ping
	StrategyTypeCert                        // cert
)
