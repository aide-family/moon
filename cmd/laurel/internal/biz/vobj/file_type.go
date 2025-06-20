package vobj

import "github.com/aide-family/moon/pkg/plugin/command"

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

// ToFileType returns the file type of the file
func ToFileType(interpreter command.Interpreter) FileType {
	switch interpreter {
	case command.Python:
		return FileTypePython
	case command.Python3:
		return FileTypePython3
	case command.Shell:
		return FileTypeShell
	case command.Bash:
		return FileTypeBash
	}
	return FileTypeUnknown
}
