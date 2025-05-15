package vobj

// DictType represents the type of a dictionary.
//
//go:generate stringer -type=DictType -linecomment -output=type_dict.string.go
type DictType int8

const (
	DictTypeUnknown    DictType = iota // unknown
	DictTypeAlarmLevel                 // alarmLevel
	DictTypeAlarmPage                  // alarmPage
)
