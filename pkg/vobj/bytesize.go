package vobj

// biteSize
//go:generate go run ../../cmd/server/stringer/cmd.go -type=ByteSize -linecomment

// ByteSize 字节大小
type ByteSize int

const (
	// KB KB
	KB ByteSize = 1024
	// MB MB
	MB ByteSize = 1024 * KB
	// GB GB
	GB ByteSize = 1024 * MB
	// TB TB
	TB ByteSize = 1024 * GB
)
