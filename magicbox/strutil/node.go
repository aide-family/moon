package strutil

import (
	"crypto/md5"
	"encoding/binary"
	"net"
)

func GetNodeIDFromIP() int64 {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				h := md5.Sum([]byte(ip.String()))
				return int64(binary.BigEndian.Uint16(h[:2]) % 1024)
			}
		}
	}
	return 0
}
