package data

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
	"strconv"
	"strings"
	"syscall"
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
		return nil, merr.Invalid(describeDialFailure(address, err))
	}
	// 机器录入时只保存了主机名、账号和密码，没有 known_hosts。
	// 与 alkaid 的默认行为保持一致：不校验主机密钥，避免首次安装被阻断。
	//
	// AuthCallback 只用来抄一份服务端认可的方法：认证失败时才能说清楚
	// 是「密码被拒绝」还是「目标机压根没开放我们有的登录方式」。返回 nil
	// 表示不改变默认的方法选择顺序。
	var allowed []string
	config := &ssh.ClientConfig{
		User:            target.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         d.timeout,
		AuthCallback: func(callback *ssh.ClientAuthContext) (ssh.AuthMethod, error) {
			allowed = append([]string(nil), callback.AllowedMethods...)
			return nil, nil
		},
	}
	clientConn, channels, requests, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		_ = conn.Close()
		return nil, merr.Invalid(describeHandshakeFailure(target.User, address, target, allowed, err))
	}
	return &sshSession{client: ssh.NewClient(clientConn, channels, requests)}, nil
}

func sshAuth(target bo.SSHTarget) ([]ssh.AuthMethod, error) {
	methods := make([]ssh.AuthMethod, 0, 3)
	if target.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(target.PrivateKey))
		if err != nil {
			return nil, merr.Invalid(fmt.Sprintf("ssh private key is invalid: %v", err))
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}
	if target.Password != "" {
		methods = append(methods, ssh.Password(target.Password))
		// 有的 sshd 关掉了 password、只留 keyboard-interactive（PAM 应答），
		// 用同一份密码回答它的问题；服务端不支持时客户端会自动跳过。
		methods = append(methods, ssh.KeyboardInteractive(func(_, _ string, questions []string, _ []bool) ([]string, error) {
			answers := make([]string, len(questions))
			for index := range answers {
				answers[index] = target.Password
			}
			return answers, nil
		}))
	}
	if len(methods) == 0 {
		return nil, merr.Invalid("ssh credentials are empty")
	}
	return methods, nil
}

// describeDialFailure 把 TCP 层失败翻译成能直接照着排查的一句话。
func describeDialFailure(address string, err error) string {
	var netErr net.Error
	switch {
	case errors.As(err, &netErr) && netErr.Timeout():
		return fmt.Sprintf(
			"ssh connect to %s timed out: the host did not answer on that port; check the network and whether sshd is listening there",
			address,
		)
	case errors.Is(err, syscall.ECONNREFUSED):
		return fmt.Sprintf("ssh connect to %s was refused: nothing is listening on that port", address)
	default:
		return fmt.Sprintf("ssh connect to %s failed: %v", address, err)
	}
}

// describeHandshakeFailure 解释认证为什么失败。
// allowed 是 sshd 交给客户端的方法列表；目标机拒绝凭据时它通常还是完整的，
// 只有目标机自己拒绝时才可能为空。
func describeHandshakeFailure(user, address string, target bo.SSHTarget, allowed []string, err error) string {
	if !strings.Contains(err.Error(), "unable to authenticate") {
		return fmt.Sprintf("ssh handshake with %s failed: %v", address, err)
	}
	offered := strings.Join(allowed, ",")
	hasPassword := target.Password != ""
	hasKey := target.PrivateKey != ""
	var usable []string
	if hasKey && slices.Contains(allowed, "publickey") {
		usable = append(usable, "private key")
	}
	if hasPassword && (slices.Contains(allowed, "password") || slices.Contains(allowed, "keyboard-interactive")) {
		usable = append(usable, "password")
	}
	switch {
	case len(allowed) == 0:
		return fmt.Sprintf(
			"ssh authentication failed for %s@%s: the target rejected the credentials and did not accept any of the saved methods",
			user, address,
		)
	case len(usable) == 0:
		return fmt.Sprintf(
			"ssh authentication failed for %s@%s: the target only allows %s, but the saved credentials are %s; enable a matching login method on the target or save the right credential",
			user, address, offered, savedMethods(target),
		)
	case hasPassword && slices.Contains(allowed, "password"):
		return fmt.Sprintf(
			"ssh authentication failed for %s@%s: the target rejected the password (sshd allows %s); the password is wrong, or the account cannot log in with a password (locked, expired, or password login disabled)",
			user, address, offered,
		)
	default:
		return fmt.Sprintf(
			"ssh authentication failed for %s@%s: the target rejected the credentials (sshd allows %s)",
			user, address, offered,
		)
	}
}

// savedMethods 用人类可读的短句描述这条机器记录里保存的凭据。
func savedMethods(target bo.SSHTarget) string {
	switch {
	case target.Password != "" && target.PrivateKey != "":
		return "a private key and a password"
	case target.PrivateKey != "":
		return "a private key only"
	default:
		return "a password only"
	}
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
