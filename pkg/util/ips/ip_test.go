package ips_test

import (
	"net"
	"testing"

	"github.com/aide-family/moon/pkg/util/ips"
)

func TestLocalIP(t *testing.T) {
	ip, err := ips.LocalIP()
	if err != nil {
		t.Errorf("LocalIP() error = %v", err)
		return
	}

	// Test that the returned IP is valid
	if ip == "" {
		t.Error("LocalIP() returned empty string")
		return
	}

	// Test that the returned IP is a valid IPv4 address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		t.Errorf("LocalIP() returned invalid IP address: %s", ip)
		return
	}

	if parsedIP.To4() == nil {
		t.Errorf("LocalIP() returned non-IPv4 address: %s", ip)
		return
	}

	// Test that the returned IP is not a loopback address
	if parsedIP.IsLoopback() {
		t.Errorf("LocalIP() returned loopback address: %s", ip)
		return
	}

	t.Logf("LocalIP() returned: %s", ip)
}

func TestLocalIPs(t *testing.T) {
	ips, err := ips.LocalIPs()
	if err != nil {
		t.Errorf("LocalIPs() error = %v", err)
		return
	}

	// Test that we got at least one IP
	if len(ips) == 0 {
		t.Error("LocalIPs() returned empty slice")
		return
	}

	// Test each IP in the slice
	for i, ip := range ips {
		if ip == "" {
			t.Errorf("LocalIPs()[%d] returned empty string", i)
			continue
		}

		// Test that the IP is a valid IPv4 address
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			t.Errorf("LocalIPs()[%d] returned invalid IP address: %s", i, ip)
			continue
		}

		if parsedIP.To4() == nil {
			t.Errorf("LocalIPs()[%d] returned non-IPv4 address: %s", i, ip)
			continue
		}

		// Test that the IP is not a loopback address
		if parsedIP.IsLoopback() {
			t.Errorf("LocalIPs()[%d] returned loopback address: %s", i, ip)
			continue
		}

		t.Logf("LocalIPs()[%d] = %s", i, ip)
	}

	// Test that all IPs are unique
	ipMap := make(map[string]bool)
	for _, ip := range ips {
		if ipMap[ip] {
			t.Errorf("LocalIPs() returned duplicate IP: %s", ip)
		}
		ipMap[ip] = true
	}
}

func TestLocalIP_Consistency(t *testing.T) {
	// Test that LocalIP() returns one of the IPs from LocalIPs()
	allIPs, err := ips.LocalIPs()
	if err != nil {
		t.Errorf("LocalIPs() error = %v", err)
		return
	}

	singleIP, err := ips.LocalIP()
	if err != nil {
		t.Errorf("LocalIP() error = %v", err)
		return
	}

	// Check if the single IP is in the list of all IPs
	found := false
	for _, ip := range allIPs {
		if ip == singleIP {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("LocalIP() returned %s, but it's not in LocalIPs() list: %v", singleIP, allIPs)
	}
}

func TestLocalIPs_NotEmpty(t *testing.T) {
	// This test ensures that we can get at least one local IP
	// This should work on any machine with network connectivity
	ips, err := ips.LocalIPs()
	if err != nil {
		t.Errorf("LocalIPs() error = %v", err)
		return
	}

	if len(ips) == 0 {
		t.Error("LocalIPs() returned empty slice, expected at least one local IP")
		return
	}

	t.Logf("Found %d local IP addresses", len(ips))
}

func TestLocalIP_NotEmpty(t *testing.T) {
	// This test ensures that we can get at least one local IP
	// This should work on any machine with network connectivity
	ip, err := ips.LocalIP()
	if err != nil {
		t.Errorf("LocalIP() error = %v", err)
		return
	}

	if ip == "" {
		t.Error("LocalIP() returned empty string, expected a local IP")
		return
	}

	t.Logf("LocalIP() returned: %s", ip)
}

// Benchmark tests
func BenchmarkLocalIP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ips.LocalIP()
		if err != nil {
			b.Errorf("LocalIP() error = %v", err)
		}
	}
}

func BenchmarkLocalIPs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := ips.LocalIPs()
		if err != nil {
			b.Errorf("LocalIPs() error = %v", err)
		}
	}
}
