package data

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/probetargetgroup"
)

// TestProbeGroupDriver 目标组驱动落库不再写死 HTTP：三种探测驱动都按传入值
// 保存，空值兜底 HTTP 以兼容只传目标列表的直接仓储调用与旧数据。
func TestProbeGroupDriver(t *testing.T) {
	cases := []struct {
		driver string
		want   probetargetgroup.Driver
	}{
		{"", probetargetgroup.DriverPROBE_HTTP},
		{"PROBE_HTTP", probetargetgroup.DriverPROBE_HTTP},
		{"PROBE_TCP", probetargetgroup.DriverPROBE_TCP},
		{"PROBE_ICMP", probetargetgroup.DriverPROBE_ICMP},
	}
	for _, tc := range cases {
		if got := probeGroupDriver(tc.driver); got != tc.want {
			t.Fatalf("probeGroupDriver(%q) = %q, want %q", tc.driver, got, tc.want)
		}
	}
}
