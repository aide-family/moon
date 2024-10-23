package vobj

// StrategyType 策略类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=StrategyType -linecomment
type StrategyType int

const (
	// StrategyTypeUnknown 未知
	StrategyTypeUnknown StrategyType = iota // unknown

	// StrategyTypeMetric 指标策略
	StrategyTypeMetric // metric

	// StrategyTypeDomainCertificate 域名证书策略
	StrategyTypeDomainCertificate // domain_certificate

	// StrategyTypeDomainPort 域名端口策略
	StrategyTypeDomainPort // domain_port

	// StrategyTypePing 网络连通性策略
	StrategyTypePing // ping
)
