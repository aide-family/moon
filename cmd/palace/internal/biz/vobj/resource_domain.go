package vobj

// ResourceDomain user status
//
//go:generate stringer -type=ResourceDomain -linecomment -output=resource_domain.string.go
type ResourceDomain int8

const (
	ResourceDomainUnknown ResourceDomain = iota // unknown
)
