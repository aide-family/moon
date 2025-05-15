package vobj

// ServerType server type
//
//go:generate stringer -type=ServerType -linecomment -output=type_server.string.go
type ServerType int8

const (
	ServerTypeUnknown ServerType = iota
	ServerTypePalace
	ServerTypeHouyi
	ServerTypeRabbit
	ServerTypeLaurel
)
