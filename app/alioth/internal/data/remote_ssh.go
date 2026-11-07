package data

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const defaultSSHTimeout = 30 * time.Second

// sshDialer 用 golang.org/x/crypto/ssh 登录目标机器。
// 安装摇光必须从控制面回连目标，这是 Agent 自举阶段唯一需要服务端发起 SSH 的路径。
type sshDialer struct {
	timeout time.Duration
}

// NewSSHDialer 按 deploy.timeout 建立 SSH 连接。
func NewSSHDialer(cfg *conf.Deploy) repository.SSHDialer {
	timeout := defaultSSHTimeout
	if cfg.GetTimeout() != nil && cfg.GetTimeout().AsDuration() > 0 {
		timeout = cfg.GetTimeout().AsDuration()
	}
	return &sshDialer{timeout: timeout}
}

func (d *sshDialer) Dial(ctx context.Context, target bo.SSHTarget) (repository.SSHSession, error) {
	if strings.TrimSpace(target.Host) == "" || strings.TrimSpace(target.User) == "" {
		return nil, merr.Invalid("ssh host or user is empty")
	}
	auth, err := sshAuth(target)
	if err != nil {
		return nil, err
	}
	port := target.Port
	if port <= 0 {
		port = 22
	}
	address := net.JoinHostPort(target.Host, strconv.Itoa(port))
	dialer := net.Dialer{Timeout: d.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, merr.Invalid(fmt.Sprintf("ssh connect failed: %v", err))
	}
	// 机器录入时只保存了主机名、账号和密码，没有 known_hosts。
	// 与 alkaid 的默认行为保持一致：不校验主机密钥，避免首次安装被阻断。
	config := &ssh.ClientConfig{
		User:            target.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         d.timeout,
	}
	clientConn, channels, requests, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		_ = conn.Close()
		return nil, merr.Invalid(fmt.Sprintf("ssh handshake failed: %v", err))
	}
	return &sshSession{client: ssh.NewClient(clientConn, channels, requests)}, nil
}

func sshAuth(target bo.SSHTarget) ([]ssh.AuthMethod, error) {
	methods := make([]ssh.AuthMethod, 0, 2)
	if target.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(target.PrivateKey))
		if err != nil {
			return nil, merr.Invalid(fmt.Sprintf("ssh private key is invalid: %v", err))
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if target.Password != "" {
		methods = append(methods, ssh.Password(target.Password))
	}
	if len(methods) == 0 {
		return nil, merr.Invalid("ssh credentials are empty")
	}
	return methods, nil
}

// sshSession 包装一条 SSH 连接。Run 和 Upload 都由同一连接串行执行。
type sshSession struct {
	client *ssh.Client
}

func (s *sshSession) Run(ctx context.Context, command string, stdin []byte) (string, int, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", -1, merr.Invalid(fmt.Sprintf("ssh session failed: %v", err))
	}
	defer session.Close()
	session.Stdin = bytes.NewReader(stdin)
	// stdout 和 stderr 由不同协程写入，必须分开缓冲，否则并发写同一个 Buffer 会丢输出。
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	type result struct {
		output   string
		exitCode int
		err      error
	}
	done := make(chan result, 1)
	go func() {
		err := session.Run(command)
		code := 0
		if err != nil {
			var exitErr *ssh.ExitError
			if errors.As(err, &exitErr) {
				// 退出码已经能表达命令失败；把 stderr 留给调用方组装错误信息。
				code = exitErr.ExitStatus()
				err = nil
			} else {
				code = -1
			}
		}
		done <- result{output: stdout.String() + stderr.String(), exitCode: code, err: err}
	}()
	select {
	case <-ctx.Done():
		_ = session.Close()
		return "", -1, ctx.Err()
	case got := <-done:
		return got.output, got.exitCode, got.err
	}
}

func (s *sshSession) Upload(ctx context.Context, path string, content io.Reader) error {
	session, err := s.client.NewSession()
	if err != nil {
		return merr.Invalid(fmt.Sprintf("ssh session failed: %v", err))
	}
	defer session.Close()
	session.Stdin = content
	var stderr bytes.Buffer
	session.Stderr = &stderr
	done := make(chan error, 1)
	go func() {
		done <- session.Run("cat > " + shellQuote(path))
	}()
	select {
	case <-ctx.Done():
		_ = session.Close()
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return merr.Invalid(fmt.Sprintf("upload %s failed: %v %s", path, err, strings.TrimSpace(stderr.String())))
		}
		return nil
	}
}

func (s *sshSession) Close() error {
	return s.client.Close()
}

// shellQuote 给路径和单行命令片段加单引号，内部单引号按 POSIX 规则转义。
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
