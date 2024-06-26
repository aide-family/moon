package vobj

// Network 网络类型
//
//go:generate go run ../../cmd/server/stringer/cmd.go -type=Network -linecomment
type Network int

const (
	NetworkUnknown Network = iota // 未知

	NetworkHttp // http

	NetworkHttps // https

	NetworkRpc // rpc
)
