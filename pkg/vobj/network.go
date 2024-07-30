package vobj

// Network 网络类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Network -linecomment
type Network int

const (
	// NetworkUnknown 未知
	NetworkUnknown Network = iota // 未知

	// NetworkHTTP http
	NetworkHTTP // http

	// NetworkHTTPS https
	NetworkHTTPS // https

	// NetworkRPC rpc
	NetworkRPC // rpc
)
