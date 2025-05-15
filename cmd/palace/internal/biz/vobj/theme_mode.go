package vobj

// ThemeMode is the theme mode of the palace.
//
//go:generate stringer -type=ThemeMode -linecomment -output=theme_mode.string.go
type ThemeMode int8

const (
	ThemeModeUnknown ThemeMode = iota // unknown
	ThemeModeLight                    // light
	ThemeModeDark                     // dark
	ThemeModeSystem                   // system
)
