package impl

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"golang.org/x/crypto/ssh"
)

func TestSSHRepositoryExecInvalidRequest(t *testing.T) {
	repo := &sshRepositoryImpl{}

	_, err := repo.Exec(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}

	_, err = repo.Exec(context.Background(), &bo.SSHExecRequest{})
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestSSHRepositoryExecSuccess(t *testing.T) {
	addr, cleanup := startTestSSHServer(t)
	defer cleanup()

	host, port := splitHostPort(t, addr)
	repo := &sshRepositoryImpl{}

	reply, err := repo.Exec(context.Background(), &bo.SSHExecRequest{
		Host:     host,
		Port:     port,
		Username: "tester",
		Password: "secret",
		Command:  "echo hello",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if reply.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", reply.ExitCode)
	}
	if strings.TrimSpace(reply.Stdout) != "ok" {
		t.Fatalf("expected stdout 'ok', got %q", reply.Stdout)
	}
}

func TestSSHRepositoryExecExitCode(t *testing.T) {
	addr, cleanup := startTestSSHServer(t)
	defer cleanup()

	host, port := splitHostPort(t, addr)
	repo := &sshRepositoryImpl{}

	reply, err := repo.Exec(context.Background(), &bo.SSHExecRequest{
		Host:     host,
		Port:     port,
		Username: "tester",
		Password: "secret",
		Command:  "fail now",
	})
	if err != nil {
		t.Fatalf("expected nil error for exit status, got %v", err)
	}
	if reply.ExitCode != 7 {
		t.Fatalf("expected exit code 7, got %d", reply.ExitCode)
	}
	if !strings.Contains(reply.Stderr, "failed") {
		t.Fatalf("expected stderr to contain failed, got %q", reply.Stderr)
	}
}

func startTestSSHServer(t *testing.T) (string, func()) {
	t.Helper()

	signer, err := newSigner()
	if err != nil {
		t.Fatalf("new signer failed: %v", err)
	}

	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "tester" && string(pass) == "secret" {
				return nil, nil
			}
			return nil, errors.New("auth failed")
		},
	}
	cfg.AddHostKey(signer)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleConn(conn, cfg)
		}
	}()

	cleanup := func() {
		_ = ln.Close()
		<-done
	}
	return ln.Addr().String(), cleanup
}

func handleConn(rawConn net.Conn, cfg *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(rawConn, cfg)
	if err != nil {
		_ = rawConn.Close()
		return
	}
	defer sshConn.Close()
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go handleSession(channel, requests)
	}
}

func handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		if req.Type != "exec" {
			_ = req.Reply(false, nil)
			continue
		}
		var payload struct {
			Command string
		}
		if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
			_ = req.Reply(false, nil)
			return
		}
		_ = req.Reply(true, nil)
		if strings.Contains(payload.Command, "fail") {
			_, _ = io.WriteString(channel.Stderr(), "failed\n")
			_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: 7}))
			return
		}
		_, _ = io.WriteString(channel, "ok\n")
		_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{Status: 0}))
		return
	}
}

func newSigner() (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return ssh.NewSignerFromKey(key)
}

func splitHostPort(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split host port failed: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port failed: %v", err)
	}
	return host, port
}
