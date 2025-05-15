package vobj

// FileType script file type
//
//go:generate stringer -type=FileType -linecomment -output=file_type.string.go
type FileType int8

const (
	FileTypeUnknown FileType = iota
	FileTypeShell
	FileTypeBash
	FileTypePython
	FileTypePython3
)
