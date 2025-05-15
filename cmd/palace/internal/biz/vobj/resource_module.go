package vobj

// ResourceModule user status
//
//go:generate stringer -type=ResourceModule -linecomment -output=resource_module.string.go
type ResourceModule int8

const (
	ResourceModuleUnknown ResourceModule = iota // unknown
)
