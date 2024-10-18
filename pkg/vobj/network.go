package vobj

import (
	"strings"
)

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

// ToNetwork 获取网络类型
func ToNetwork(s string) Network {
	s = strings.ToLower(s)
	switch s {
	case "http":
		return NetworkHTTP
	case "https":
		return NetworkHTTPS
	default:
		return NetworkRPC
	}
}
