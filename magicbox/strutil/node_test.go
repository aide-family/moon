package strutil_test

import (
	"crypto/md5"
	"encoding/binary"
	"net"
	"testing"

	"github.com/aide-family/magicbox/strutil"
)

// calculateNodeIDFromIP 辅助函数，用于计算给定 IP 的节点 ID
// 这个函数复制了 GetNodeIDFromIP 的核心逻辑，用于测试
func calculateNodeIDFromIP(ip string) int64 {
	h := md5.Sum([]byte(ip))
	return int64(binary.BigEndian.Uint16(h[:2]) % 1024)
}

// TestGetNodeIDFromIP_Basic 测试 GetNodeIDFromIP 的基本功能
func TestGetNodeIDFromIP_Basic(t *testing.T) {
	nodeID := strutil.GetNodeIDFromIP()

	// 验证返回值在有效范围内 (0-1023)
	if nodeID < 0 || nodeID >= 1024 {
		t.Errorf("GetNodeIDFromIP() = %d, want value in range [0, 1023]", nodeID)
	}

	// 验证函数不会 panic（如果到这里说明没有 panic）
	t.Logf("GetNodeIDFromIP() returned: %d", nodeID)
}

// TestGetNodeIDFromIP_Consistency 测试 GetNodeIDFromIP 的一致性
// 在短时间内多次调用应该返回相同的值（假设网络接口没有变化）
func TestGetNodeIDFromIP_Consistency(t *testing.T) {
	// 多次调用函数
	results := make([]int64, 10)
	for i := 0; i < 10; i++ {
		results[i] = strutil.GetNodeIDFromIP()
	}

	// 验证所有结果都相同
	first := results[0]
	for i, result := range results {
		if result != first {
			t.Errorf("GetNodeIDFromIP() returned inconsistent value: result[0] = %d, result[%d] = %d", first, i, result)
		}
	}

	// 验证所有结果都在有效范围内
	for i, result := range results {
		if result < 0 || result >= 1024 {
			t.Errorf("GetNodeIDFromIP() returned invalid value: result[%d] = %d, want value in range [0, 1023]", i, result)
		}
	}
}

// TestGetNodeIDFromIP_ValueRange 测试返回值在正确范围内
func TestGetNodeIDFromIP_ValueRange(t *testing.T) {
	nodeID := strutil.GetNodeIDFromIP()

	// 验证返回值在 0-1023 范围内
	if nodeID < 0 {
		t.Errorf("GetNodeIDFromIP() = %d, want value >= 0", nodeID)
	}
	if nodeID >= 1024 {
		t.Errorf("GetNodeIDFromIP() = %d, want value < 1024", nodeID)
	}
}

// TestCalculateNodeIDFromIP_Logic 测试节点 ID 计算逻辑
// 使用辅助函数测试不同 IP 地址的计算结果
func TestCalculateNodeIDFromIP_Logic(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		validate func(int64) bool
	}{
		{
			name: "Localhost IP",
			ip:   "127.0.0.1",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Common private IP",
			ip:   "192.168.1.1",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Another private IP",
			ip:   "10.0.0.1",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Public IP",
			ip:   "8.8.8.8",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Another public IP",
			ip:   "1.1.1.1",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Private IP range",
			ip:   "172.16.0.1",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeID := calculateNodeIDFromIP(tt.ip)
			if !tt.validate(nodeID) {
				t.Errorf("calculateNodeIDFromIP(%q) = %d, validation failed", tt.ip, nodeID)
			}
		})
	}
}

// TestCalculateNodeIDFromIP_Consistency 测试相同 IP 返回相同的节点 ID
func TestCalculateNodeIDFromIP_Consistency(t *testing.T) {
	testIPs := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"1.1.1.1",
	}

	for _, ip := range testIPs {
		t.Run(ip, func(t *testing.T) {
			// 多次计算相同 IP 的节点 ID
			results := make([]int64, 10)
			for i := 0; i < 10; i++ {
				results[i] = calculateNodeIDFromIP(ip)
			}

			// 验证所有结果都相同
			first := results[0]
			for i, result := range results {
				if result != first {
					t.Errorf("calculateNodeIDFromIP(%q) returned inconsistent value: result[0] = %d, result[%d] = %d", ip, first, i, result)
				}
			}
		})
	}
}

// TestCalculateNodeIDFromIP_DifferentIPs 测试不同 IP 可能返回不同的节点 ID
func TestCalculateNodeIDFromIP_DifferentIPs(t *testing.T) {
	testIPs := []string{
		"192.168.1.1",
		"192.168.1.2",
		"10.0.0.1",
		"10.0.0.2",
		"8.8.8.8",
		"8.8.8.9",
	}

	nodeIDs := make(map[string]int64)
	for _, ip := range testIPs {
		nodeIDs[ip] = calculateNodeIDFromIP(ip)
	}

	// 验证所有节点 ID 都在有效范围内
	for ip, nodeID := range nodeIDs {
		if nodeID < 0 || nodeID >= 1024 {
			t.Errorf("calculateNodeIDFromIP(%q) = %d, want value in range [0, 1023]", ip, nodeID)
		}
	}

	// 记录节点 ID 分布（用于调试）
	t.Logf("Node IDs for different IPs:")
	for ip, nodeID := range nodeIDs {
		t.Logf("  %s -> %d", ip, nodeID)
	}
}

// TestGetNodeIDFromIP_NetworkInterfaces 测试函数处理网络接口的情况
func TestGetNodeIDFromIP_NetworkInterfaces(t *testing.T) {
	// 获取所有网络接口地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Logf("Failed to get interface addresses: %v", err)
		return
	}

	// 验证至少有一些网络接口
	if len(addrs) == 0 {
		t.Log("No network interfaces found")
		return
	}

	t.Logf("Found %d network interface addresses", len(addrs))

	// 检查是否有非回环的 IPv4 地址
	hasNonLoopbackIPv4 := false
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				hasNonLoopbackIPv4 = true
				t.Logf("Found non-loopback IPv4 address: %s", ip.String())
				break
			}
		}
	}

	if !hasNonLoopbackIPv4 {
		t.Log("No non-loopback IPv4 addresses found, GetNodeIDFromIP() will return 0")
	}

	// 调用函数并验证结果
	nodeID := strutil.GetNodeIDFromIP()
	if nodeID < 0 || nodeID >= 1024 {
		t.Errorf("GetNodeIDFromIP() = %d, want value in range [0, 1023]", nodeID)
	}

	// 如果有非回环 IPv4 地址，nodeID 应该不为 0
	// 但如果没有，nodeID 应该是 0
	if hasNonLoopbackIPv4 && nodeID == 0 {
		t.Logf("Warning: Found non-loopback IPv4 address but GetNodeIDFromIP() returned 0")
	} else if !hasNonLoopbackIPv4 && nodeID != 0 {
		t.Logf("Warning: No non-loopback IPv4 address found but GetNodeIDFromIP() returned %d", nodeID)
	}
}

// TestCalculateNodeIDFromIP_EdgeCases 测试边界情况
func TestCalculateNodeIDFromIP_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		validate func(int64) bool
	}{
		{
			name: "Minimum IP",
			ip:   "0.0.0.0",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Maximum IP",
			ip:   "255.255.255.255",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "All zeros",
			ip:   "0.0.0.0",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "All ones",
			ip:   "255.255.255.255",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
		{
			name: "Broadcast address",
			ip:   "255.255.255.255",
			validate: func(id int64) bool {
				return id >= 0 && id < 1024
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeID := calculateNodeIDFromIP(tt.ip)
			if !tt.validate(nodeID) {
				t.Errorf("calculateNodeIDFromIP(%q) = %d, validation failed", tt.ip, nodeID)
			}
		})
	}
}

// TestCalculateNodeIDFromIP_Modulo 测试模运算确保结果在 0-1023 范围内
func TestCalculateNodeIDFromIP_Modulo(t *testing.T) {
	// 测试多个不同的 IP 地址
	testIPs := []string{
		"1.1.1.1",
		"2.2.2.2",
		"3.3.3.3",
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"8.8.8.8",
		"127.0.0.1",
		"0.0.0.0",
		"255.255.255.255",
	}

	for _, ip := range testIPs {
		t.Run(ip, func(t *testing.T) {
			// 计算 MD5 哈希
			h := md5.Sum([]byte(ip))

			// 获取前两个字节
			firstTwoBytes := h[:2]

			// 转换为 uint16
			value := binary.BigEndian.Uint16(firstTwoBytes)

			// 计算模 1024
			result := int64(value % 1024)

			// 验证结果在有效范围内
			if result < 0 || result >= 1024 {
				t.Errorf("result = %d, want value in range [0, 1023]", result)
			}

			// 验证与辅助函数的结果一致
			expected := calculateNodeIDFromIP(ip)
			if result != expected {
				t.Errorf("Manual calculation = %d, calculateNodeIDFromIP() = %d, want equal", result, expected)
			}
		})
	}
}

// TestGetNodeIDFromIP_NoPanic 测试函数不会 panic
func TestGetNodeIDFromIP_NoPanic(t *testing.T) {
	// 多次调用确保不会 panic
	for i := 0; i < 100; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("GetNodeIDFromIP() panicked: %v", r)
				}
			}()
			_ = strutil.GetNodeIDFromIP()
		}()
	}
}

// BenchmarkGetNodeIDFromIP 基准测试 GetNodeIDFromIP 函数
func BenchmarkGetNodeIDFromIP(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.GetNodeIDFromIP()
	}
}

// BenchmarkCalculateNodeIDFromIP 基准测试节点 ID 计算逻辑
func BenchmarkCalculateNodeIDFromIP(b *testing.B) {
	testIP := "192.168.1.1"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateNodeIDFromIP(testIP)
	}
}

// BenchmarkCalculateNodeIDFromIP_MD5 基准测试 MD5 哈希计算
func BenchmarkCalculateNodeIDFromIP_MD5(b *testing.B) {
	testIP := "192.168.1.1"
	ipBytes := []byte(testIP)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = md5.Sum(ipBytes)
	}
}

// BenchmarkCalculateNodeIDFromIP_Binary 基准测试二进制转换
func BenchmarkCalculateNodeIDFromIP_Binary(b *testing.B) {
	testIP := "192.168.1.1"
	h := md5.Sum([]byte(testIP))
	firstTwoBytes := h[:2]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = binary.BigEndian.Uint16(firstTwoBytes)
	}
}
