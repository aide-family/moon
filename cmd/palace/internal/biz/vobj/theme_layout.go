package vobj

// ThemeLayout is the theme layout of the palace.
//
//go:generate stringer -type=ThemeLayout -linecomment -output=theme_layout.string.go
type ThemeLayout int8

const (
	ThemeLayoutUnknown ThemeLayout = iota // unknown
	ThemeLayoutDefault                    // default
)
