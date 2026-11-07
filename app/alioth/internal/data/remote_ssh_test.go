package data

import (
	"errors"
	"net"
	"strings"
	"syscall"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

const authFailure = "ssh: unable to authenticate, attempted methods [none password], no supported methods remain"

// 目标机开放密码登录却拒绝密码：必须说清楚是密码被拒，而不是只说认证失败。
func TestDescribeHandshakeFailureExplainsRejectedPassword(t *testing.T) {
	target := bo.SSHTarget{Password: "secret"}
	message := describeHandshakeFailure(
		"root", "192.168.10.168:22", target, []string{"publickey", "password"}, errors.New(authFailure),
	)
	for _, want := range []string{"root@192.168.10.168:22", "rejected the password", "publickey,password"} {
		if !strings.Contains(message, want) {
			t.Fatalf("message %q does not contain %q", message, want)
		}
	}
}

// 目标机只收公钥、记录里却只有密码：提示要换登录方式，而不是让用户反复试密码。
func TestDescribeHandshakeFailureNamesMissingMethod(t *testing.T) {
	target := bo.SSHTarget{Password: "secret"}
	message := describeHandshakeFailure(
		"root", "10.0.0.8:22", target, []string{"publickey"}, errors.New(authFailure),
	)
	for _, want := range []string{"only allows publickey", "a password only", "10.0.0.8:22"} {
		if !strings.Contains(message, want) {
			t.Fatalf("message %q does not contain %q", message, want)
		}
	}
}

// 记录里同时有私钥和密码、目标机只认密码时，失败原因仍然是密码被拒。
func TestDescribeHandshakeFailureBlamesPasswordWhenKeyUnsupported(t *testing.T) {
	target := bo.SSHTarget{Password: "secret", PrivateKey: "key"}
	message := describeHandshakeFailure(
		"root", "10.0.0.8:22", target, []string{"password"}, errors.New(authFailure),
	)
	if !strings.Contains(message, "rejected the password") {
		t.Fatalf("message = %q", message)
	}
	if !strings.Contains(message, "sshd allows password") {
		t.Fatalf("message = %q", message)
	}
}

func TestSavedMethods(t *testing.T) {
	cases := map[string]bo.SSHTarget{
		"a private key and a password": {Password: "secret", PrivateKey: "key"},
		"a private key only":           {PrivateKey: "key"},
		"a password only":              {Password: "secret"},
	}
	for want, target := range cases {
		if got := savedMethods(target); got != want {
			t.Fatalf("savedMethods(%+v) = %q, want %q", target, got, want)
		}
	}
}

// 算法协商失败等不是认证问题的错误原样带上，不被改写成「密码错误」。
func TestDescribeHandshakeFailureKeepsOtherErrors(t *testing.T) {
	message := describeHandshakeFailure(
		"root", "10.0.0.8:22", bo.SSHTarget{Password: "secret"}, nil, errors.New("ssh: no common algorithm"),
	)
	if !strings.Contains(message, "ssh handshake with 10.0.0.8:22 failed") {
		t.Fatalf("message = %q", message)
	}
	if strings.Contains(message, "rejected the password") {
		t.Fatalf("message %q blames the password", message)
	}
}

func TestDescribeDialFailure(t *testing.T) {
	timeout := describeDialFailure("10.0.0.8:22", &net.DNSError{Err: "i/o timeout", IsTimeout: true})
	if !strings.Contains(timeout, "timed out") {
		t.Fatalf("timeout message = %q", timeout)
	}
	refused := describeDialFailure("10.0.0.8:22", &net.OpError{Op: "dial", Net: "tcp", Err: syscall.ECONNREFUSED})
	if !strings.Contains(refused, "was refused") {
		t.Fatalf("refused message = %q", refused)
	}
}
