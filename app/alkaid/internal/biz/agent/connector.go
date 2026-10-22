package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/wire"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alkaid/internal/data"
	"github.com/aide-family/Alioth/pkg/grpctls"
	"github.com/aide-family/Alioth/pkg/merr"
)

const maxOutputBytes = 64 * 1024

var ProviderSet = wire.NewSet(NewConnector)

// Connector 作为 kratos Server 跑出站流。
// 控制面是多副本时，Agent 对每个副本各建一条流，任意副本都能直接处理该 Agent 的请求。
type Connector struct {
	addrs     []string
	token     string
	heartbeat time.Duration
	creds     credentials.TransportCredentials
	log       *slog.Logger
	cancel    context.CancelFunc
	done      chan struct{}
	sendMu    sync.Mutex
	shellMu   sync.Mutex
	shell     *liveShell
}

func NewConnector(settings *data.ConnectionSettings, logger *slog.Logger) (*Connector, error) {
	if settings.AgentToken == "" {
		logger.Warn("agent token is empty; outbound stream disabled")
	}
	interval := settings.Heartbeat
	if interval <= 0 {
		interval = 15 * time.Second
	}
	addrs := parseEndpoints(settings.AliothGRPC)
	tlsConfig, err := grpctls.ClientFromEnv("ALKAID_ALIOTH_TLS")
	if err != nil {
		return nil, err
	}
	var creds credentials.TransportCredentials
	if tlsConfig != nil {
		if tlsConfig.ServerName == "" && len(addrs) > 0 {
			if host, _, splitErr := net.SplitHostPort(addrs[0]); splitErr == nil && net.ParseIP(host) == nil {
				tlsConfig.ServerName = host
			}
		}
		creds = credentials.NewTLS(tlsConfig)
	}
	return &Connector{
		addrs: addrs, token: settings.AgentToken, heartbeat: interval, creds: creds,
		log: logger, done: make(chan struct{}),
	}, nil
}

func (c *Connector) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go func() {
		defer close(c.done)
		c.run(runCtx)
	}()
	return nil
}

func (c *Connector) Stop(ctx context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.done == nil {
		return nil
	}
	select {
	case <-c.done:
	case <-ctx.Done():
	}
	return nil
}

// run 维护每个控制面副本一条流。DNS 变化时启动新副本的流并停掉已经不存在的副本。
func (c *Connector) run(ctx context.Context) {
	if c.token == "" || len(c.addrs) == 0 {
		<-ctx.Done()
		return
	}
	active := map[string]context.CancelFunc{}
	resolve := func() {
		desired := map[string]struct{}{}
		for _, raw := range c.addrs {
			for _, addr := range expandEndpoint(raw, net.LookupHost) {
				desired[addr] = struct{}{}
			}
		}
		for addr, cancel := range active {
			if _, ok := desired[addr]; !ok {
				cancel()
				delete(active, addr)
			}
		}
		for addr := range desired {
			if _, ok := active[addr]; ok {
				continue
			}
			streamCtx, cancel := context.WithCancel(ctx)
			active[addr] = cancel
			go c.loop(streamCtx, addr)
		}
	}
	resolve()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			for _, cancel := range active {
				cancel()
			}
			return
		case <-ticker.C:
			resolve()
		}
	}
}

func (c *Connector) loop(ctx context.Context, addr string) {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		started := time.Now()
		err := c.session(ctx, addr)
		if ctx.Err() != nil {
			return
		}
		if time.Since(started) > 30*time.Second {
			backoff = time.Second
		}
		c.log.Warn(fmt.Sprintf("agent stream %s closed: %v; retry in %s", addr, err, backoff))
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

// parseEndpoints 允许配置逗号分隔的多个控制面地址。
func parseEndpoints(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// expandEndpoint 把主机名展开成 DNS 返回的每个地址；直接写 IP 时原样返回。
func expandEndpoint(raw string, lookup func(string) ([]string, error)) []string {
	host, port, err := net.SplitHostPort(raw)
	if err != nil || host == "" {
		return []string{raw}
	}
	if net.ParseIP(host) != nil {
		return []string{raw}
	}
	ips, err := lookup(host)
	if err != nil || len(ips) == 0 {
		return []string{raw}
	}
	out := make([]string, 0, len(ips))
	for _, ip := range ips {
		out = append(out, net.JoinHostPort(ip, port))
	}
	return out
}

func (c *Connector) session(ctx context.Context, addr string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	dialOpts := []grpc.DialOption{}
	if c.creds != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(c.creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return merr.Server("alioth dial failed")
	}
	defer conn.Close()
	client := v1.NewAgentServiceClient(conn)
	streamCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+c.token))
	stream, err := client.Connect(streamCtx)
	if err != nil {
		return err
	}
	hostname, _ := os.Hostname()
	machineID := readMachineID()
	if err := c.send(stream, &v1.AgentFrame{Register: &v1.AgentRegister{
		Hostname: hostname, MachineId: machineID, LocalIp: localIP(), Version: "alkaid-dev",
	}}); err != nil {
		return err
	}
	if err := c.send(stream, c.machineFrame()); err != nil {
		return err
	}
	errCh := make(chan error, 3)
	go func() {
		<-ctx.Done()
		select {
		case errCh <- ctx.Err():
		default:
		}
	}()
	go func() {
		ticker := time.NewTicker(c.heartbeat)
		defer ticker.Stop()
		var seq int64
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				seq++
				if err := c.send(stream, &v1.AgentFrame{Heartbeat: &v1.AgentHeartbeat{Seq: seq}}); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				if err := c.send(stream, c.machineFrame()); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
			}
		}
	}()
	// 慢命令不能阻塞后续帧；并发上限用来保护边缘主机的内存和进程数。
	slots := make(chan struct{}, 8)
	busy := func() bool {
		select {
		case slots <- struct{}{}:
			return false
		default:
			return true
		}
	}
	go func() {
		for {
			frame, err := stream.Recv()
			if err == io.EOF {
				errCh <- nil
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			if frame.GetClose() != nil {
				errCh <- merr.Forbidden(frame.GetClose().GetReason())
				return
			}
			if req := frame.GetExecRequest(); req != nil {
				if busy() {
					_ = c.send(stream, &v1.AgentFrame{ExecResult: &v1.AgentExecResult{RequestUid: req.GetRequestUid(), Error: "agent is busy", ExitCode: 1}})
					continue
				}
				go func() {
					defer func() { <-slots }()
					_ = c.send(stream, &v1.AgentFrame{ExecResult: c.exec(req)})
				}()
			}
			c.handleShell(stream, frame)
		}
	}()
	err = <-errCh
	c.closeShell()
	return err
}

func (c *Connector) send(stream v1.AgentService_ConnectClient, frame *v1.AgentFrame) error {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	return stream.Send(frame)
}

func (c *Connector) machineFrame() *v1.AgentFrame {
	hostname, _ := os.Hostname()
	return &v1.AgentFrame{MachineInfo: &v1.AgentMachineInfo{
		Hostname: hostname, MachineId: readMachineID(), LocalIp: localIP(), Version: "alkaid-dev",
		CollectJson: collectJSON(),
	}}
}

func (c *Connector) exec(req *v1.AgentExecRequest) *v1.AgentExecResult {
	out := &v1.AgentExecResult{RequestUid: req.GetRequestUid()}
	command := osExec.Command("/bin/sh", "-c", req.GetCommand())
	if workdir := strings.TrimSpace(req.GetWorkdir()); workdir != "" {
		resolved, err := expandHomeDir(workdir)
		if err != nil {
			out.ExitCode = 1
			out.Error = err.Error()
			return out
		}
		command.Dir = resolved
	}
	if env := req.GetEnv(); env != "" {
		var vars map[string]string
		if json.Unmarshal([]byte(env), &vars) == nil {
			extraEnv := make([]string, 0, len(vars))
			for key, value := range vars {
				extraEnv = append(extraEnv, key+"="+value)
			}
			command.Env = append(os.Environ(), extraEnv...)
		}
	}
	output, runErr := command.CombinedOutput()
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		out.Truncated = true
	}
	out.Output = string(output)
	if runErr != nil {
		if ee, ok := runErr.(*osExec.ExitError); ok {
			out.ExitCode = int32(ee.ExitCode())
		} else {
			out.ExitCode = 1
			out.Error = runErr.Error()
		}
	}
	return out
}

func expandHomeDir(workdir string) (string, error) {
	if workdir != "~" && !strings.HasPrefix(workdir, "~/") {
		return workdir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	if workdir == "~" {
		return home, nil
	}
	return filepath.Join(home, workdir[2:]), nil
}

func sshDial(host string, port int32, user, password, privateKey string) (*ssh.Client, error) {
	if host == "" {
		host = "127.0.0.1"
	}
	if port <= 0 {
		port = 22
	}
	if user == "" {
		user = "root"
	}
	var auths []ssh.AuthMethod
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, merr.Invalid("invalid ssh private key")
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}
	if password != "" {
		auths = append(auths, ssh.Password(password))
	}
	if len(auths) == 0 {
		return nil, merr.Invalid("ssh credentials missing")
	}
	hostKeyCallback, err := sshHostKeyCallback()
	if err != nil {
		return nil, err
	}
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: hostKeyCallback,
		Timeout:         15 * time.Second,
	}
	addr := net.JoinHostPort(host, itoa(int(port)))
	return ssh.Dial("tcp", addr, cfg)
}

// sshHostKeyCallback 在配置了 ALKAID_SSH_KNOWN_HOSTS 时校验目标主机密钥。
// 未配置时保持旧行为，但部署文档要求生产设置该变量。
func sshHostKeyCallback() (ssh.HostKeyCallback, error) {
	path := strings.TrimSpace(os.Getenv("ALKAID_SSH_KNOWN_HOSTS"))
	if path == "" {
		return ssh.InsecureIgnoreHostKey(), nil
	}
	callback, err := knownhosts.New(path)
	if err != nil {
		return nil, merr.Server("failed to load ssh known_hosts")
	}
	return callback, nil
}

func readMachineID() string {
	raw, err := os.ReadFile("/etc/machine-id")
	if err == nil {
		if value := strings.TrimSpace(string(raw)); value != "" {
			return value
		}
	}
	// 没有 machine-id 的系统（例如 macOS 开发机）用主机名兜底，避免所有 Agent 都上报 "unknown"。
	if hostname, err := os.Hostname(); err == nil && strings.TrimSpace(hostname) != "" {
		return strings.TrimSpace(hostname)
	}
	return "unknown"
}

func localIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if v4 := ip.To4(); v4 != nil {
				return v4.String()
			}
		}
	}
	return ""
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
