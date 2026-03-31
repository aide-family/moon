package impl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/aide-family/magicbox/merr"
	"golang.org/x/crypto/ssh"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data"
)

func NewSSHRepository(d *data.Data) repository.SSHOperator {
	return &sshRepositoryImpl{
		d: d,
	}
}

type sshRepositoryImpl struct {
	d *data.Data
}

func (s *sshRepositoryImpl) Exec(ctx context.Context, req *bo.SSHExecRequest) (*bo.SSHExecReply, error) {
	if req == nil {
		return nil, merr.ErrorInvalidArgument("ssh exec request is required")
	}
	if req.Host == "" || req.Username == "" || req.Command == "" {
		return nil, merr.ErrorInvalidArgument("host, username and command are required")
	}
	if req.Password == "" && req.PrivateKey == "" {
		return nil, merr.ErrorInvalidArgument("password or private key is required")
	}

	authMethod, err := buildAuthMethod(req)
	if err != nil {
		return nil, err
	}

	port := req.Port
	if port == 0 {
		port = 22
	}
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	clientConfig := &ssh.ClientConfig{
		User:            req.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(req.Host, fmt.Sprintf("%d", port)), clientConfig)
	if err != nil {
		return nil, merr.ErrorInternalServer("dial ssh server failed").WithCause(err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, merr.ErrorInternalServer("create ssh session failed").WithCause(err)
	}
	defer session.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	session.Stdout = stdout
	session.Stderr = stderr

	command := buildCommand(req)
	if err := session.Start(command); err != nil {
		return nil, merr.ErrorInternalServer("start remote command failed").WithCause(err)
	}

	done := make(chan error, 1)
	go func() {
		done <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGKILL)
		return nil, merr.ErrorInternalServer("remote command canceled").WithCause(ctx.Err())
	case err = <-done:
		reply := &bo.SSHExecReply{
			Stdout: stdout.String(),
			Stderr: stderr.String(),
		}
		if err == nil {
			return reply, nil
		}
		var exitErr *ssh.ExitError
		if ok := errors.As(err, &exitErr); ok {
			reply.ExitCode = exitErr.ExitStatus()
			return reply, nil
		}
		return nil, merr.ErrorInternalServer("wait remote command failed").WithCause(err)
	}
}

func buildAuthMethod(req *bo.SSHExecRequest) (ssh.AuthMethod, error) {
	if req.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(req.PrivateKey))
		if err != nil {
			return nil, merr.ErrorInvalidArgument("parse private key failed").WithCause(err)
		}
		return ssh.PublicKeys(signer), nil
	}
	return ssh.Password(req.Password), nil
}

func buildCommand(req *bo.SSHExecRequest) string {
	parts := make([]string, 0, len(req.Env)+2)
	for key, value := range req.Env {
		parts = append(parts, fmt.Sprintf("%s=%s", shellEscape(key), shellEscape(value)))
	}
	if req.WorkDir != "" {
		parts = append(parts, fmt.Sprintf("cd %s", shellEscape(req.WorkDir)))
	}
	parts = append(parts, req.Command)
	return strings.Join(parts, " && ")
}

func shellEscape(raw string) string {
	return "'" + strings.ReplaceAll(raw, "'", `'"'"'`) + "'"
}
