package ips

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

// LocalIP returns the local IP address.
func LocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("no local IP address found")
}

// LocalIPs returns the local IP addresses.
func LocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips, nil
}

// GetClientIP returns the real client IP address from HTTP request.
// It checks X-Forwarded-For, X-Real-IP, and RemoteAddr headers in order.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		if isValidIP(xRealIP) {
			return xRealIP
		}
	}

	// Check X-Client-IP header
	if xClientIP := r.Header.Get("X-Client-IP"); xClientIP != "" {
		if isValidIP(xClientIP) {
			return xClientIP
		}
	}

	// Check CF-Connecting-IP header (Cloudflare)
	if cfConnectingIP := r.Header.Get("CF-Connecting-IP"); cfConnectingIP != "" {
		if isValidIP(cfConnectingIP) {
			return cfConnectingIP
		}
	}

	// Fallback to RemoteAddr
	if r.RemoteAddr != "" {
		// RemoteAddr format is "IP:port", extract IP
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil && isValidIP(ip) {
			return ip
		}
		// If SplitHostPort fails, try to use RemoteAddr as is
		if isValidIP(r.RemoteAddr) {
			return r.RemoteAddr
		}
	}

	return ""
}

// isValidIP checks if the given string is a valid IP address
func isValidIP(ip string) bool {
	if ip == "" {
		return false
	}

	// Remove port if present
	if strings.Contains(ip, ":") {
		ip, _, _ = net.SplitHostPort(ip)
	}

	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && !parsedIP.IsLoopback() && !parsedIP.IsPrivate()
}
