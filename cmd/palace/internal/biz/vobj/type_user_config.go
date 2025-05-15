package vobj

// UserConfigType user config type
//
//go:generate stringer -type=UserConfigType -linecomment -output=type_user_config.string.go
type UserConfigType int8

const (
	UserConfigTypeUnknown   UserConfigType = iota // unknown
	UserConfigTypeNotice                          // notice
	UserConfigTypeTheme                           // theme
	UserConfigTypeDashboard                       // dashboard
	UserConfigTypeLayout                          // layout
	UserConfigTypeLanguage                        // language
	UserConfigTypeTimeZone                        // time-zone
	UserConfigTypeTable                           // table
)
