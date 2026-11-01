package probe

import (
	"strings"

	"github.com/aide-family/moon/pkg/merr"
)

// 探测驱动。取值与 proto 的 driver 枚举、规则驱动白名单保持一致。
const (
	DriverHTTP = "PROBE_HTTP"
	DriverTCP  = "PROBE_TCP"
	DriverICMP = "PROBE_ICMP"
)

// ValidDriver 判断驱动是否受支持。
func ValidDriver(driver string) bool {
	switch normalizeDriver(driver) {
	case DriverHTTP, DriverTCP, DriverICMP:
		return true
	default:
		return false
	}
}

// NormalizeTarget 按驱动归一化探测目标。HTTP 沿用 NormalizeSpec；TCP 只接受
// host:port；ICMP 只接受主机名或 IP 字面量。未知驱动一律拒绝。
func NormalizeTarget(driver, raw string) (string, error) {
	switch normalizeDriver(driver) {
	case DriverHTTP:
		return NormalizeSpec(raw)
	case DriverTCP:
		return normalizeTCPTarget(raw)
	case DriverICMP:
		return normalizeICMPTarget(raw)
	default:
		return "", merr.Invalid("unsupported probe driver")
	}
}

// normalizeDriver 统一驱动写法，调用方传小写也能识别。
func normalizeDriver(driver string) string {
	return strings.ToUpper(strings.TrimSpace(driver))
}
