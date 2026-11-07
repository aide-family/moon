package agent

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/conf"
)

// 控制面地址是本机回环、目标机是另一台机器时，Agent 不可能上线：必须提前说清楚。
func TestControlPlaneHintRejectsLoopbackAddress(t *testing.T) {
	deployer := NewDeployer(nil, nil, nil, &conf.Deploy{AliothGrpcAddr: "127.0.0.1:9000"})
	hint := deployer.controlPlaneHint("192.168.10.168")
	for _, want := range []string{"127.0.0.1:9000", "192.168.10.168", "deploy.alioth_grpc_addr"} {
		if !strings.Contains(hint, want) {
			t.Fatalf("hint %q does not contain %q", hint, want)
		}
	}
}

// 目标机就是本机（例如 Docker 宿主）时，回环地址是正确的，不能报错。
func TestControlPlaneHintAllowsLoopbackTarget(t *testing.T) {
	deployer := NewDeployer(nil, nil, nil, &conf.Deploy{AliothGrpcAddr: "127.0.0.1:9000"})
	if hint := deployer.controlPlaneHint("127.0.0.1"); hint != "" {
		t.Fatalf("hint = %q, want empty", hint)
	}
	if hint := deployer.controlPlaneHint("localhost"); hint != "" {
		t.Fatalf("hint = %q, want empty", hint)
	}
}

// 控制面地址本身可被目标机访问时不提示。
func TestControlPlaneHintAllowsReachableAddress(t *testing.T) {
	deployer := NewDeployer(nil, nil, nil, &conf.Deploy{AliothGrpcAddr: "192.168.10.68:9000"})
	if hint := deployer.controlPlaneHint("192.168.10.168"); hint != "" {
		t.Fatalf("hint = %q, want empty", hint)
	}
}

func TestLoopbackHost(t *testing.T) {
	cases := map[string]bool{
		"127.0.0.1":   true,
		"127.1.2.3":   true,
		"::1":         true,
		"localhost":   true,
		"LOCALHOST":   true,
		"192.168.1.1": false,
		"alioth":      false,
	}
	for host, want := range cases {
		if got := loopbackHost(host); got != want {
			t.Fatalf("loopbackHost(%q) = %v, want %v", host, got, want)
		}
	}
}
