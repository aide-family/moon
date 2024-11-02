package vobj

// biteSize
//go:generate go run ../../cmd/server/stringer/cmd.go -type=ByteSize -linecomment

type ByteSize int

const (
	KB ByteSize = 1024
	MB ByteSize = 1024 * KB
	GB ByteSize = 1024 * MB
	TB ByteSize = 1024 * GB
)
