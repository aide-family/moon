package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/pkg/sftp"
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

// maxSFTPBytes 是单次 SFTP 传输上限，与 grpc-go 默认 4 MiB 消息上限一致。
// 超过时明确失败，不做截断，避免把半个文件当成成功结果返回。
const maxSFTPBytes = 4 << 20

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

func NewConnector(d *data.Data, logger *slog.Logger) (*Connector, error) {
	if d.AgentToken == "" {
		logger.Warn("agent token is empty; outbound stream disabled")
	}
	interval := d.Heartbeat
	if interval <= 0 {
		interval = 15 * time.Second
	}
	addrs := parseEndpoints(d.AliothGRPC)
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
		addrs: addrs, token: d.AgentToken, heartbeat: interval, creds: creds,
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
			if req := frame.GetSftpRequest(); req != nil {
				if busy() {
					_ = c.send(stream, &v1.AgentFrame{SftpResult: &v1.AgentSFTPResult{RequestUid: req.GetRequestUid(), Error: "agent is busy"}})
					continue
				}
				go func() {
					defer func() { <-slots }()
					_ = c.send(stream, &v1.AgentFrame{SftpResult: c.sftp(req)})
				}()
			}
			if req := frame.GetK8SRequest(); req != nil {
				if busy() {
					_ = c.send(stream, &v1.AgentFrame{K8SResult: &v1.AgentK8SResult{RequestUid: req.GetRequestUid(), Error: "agent is busy"}})
					continue
				}
				go func() {
					defer func() { <-slots }()
					_ = c.send(stream, &v1.AgentFrame{K8SResult: c.k8s(req)})
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
	client, err := sshDial(req.GetSshHost(), req.GetSshPort(), req.GetSshUser(), req.GetSshPassword(), req.GetSshPrivateKey())
	if err != nil {
		out.Error = err.Error()
		out.ExitCode = 1
		return out
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		out.Error = err.Error()
		out.ExitCode = 1
		return out
	}
	defer session.Close()
	if env := req.GetEnv(); env != "" {
		var vars map[string]string
		if json.Unmarshal([]byte(env), &vars) == nil {
			for k, v := range vars {
				_ = session.Setenv(k, v)
			}
		}
	}
	command := req.GetCommand()
	if wd := strings.TrimSpace(req.GetWorkdir()); wd != "" {
		command = "cd " + shellQuote(wd) + " && " + command
	}
	var buf bytes.Buffer
	session.Stdout = &buf
	session.Stderr = &buf
	runErr := session.Run(command)
	output := buf.String()
	truncated := false
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		truncated = true
	}
	out.Output = output
	out.Truncated = truncated
	if runErr != nil {
		if ee, ok := runErr.(*ssh.ExitError); ok {
			out.ExitCode = int32(ee.ExitStatus())
		} else {
			out.ExitCode = 1
			out.Error = runErr.Error()
		}
	}
	return out
}

func (c *Connector) sftp(req *v1.AgentSFTPRequest) *v1.AgentSFTPResult {
	out := &v1.AgentSFTPResult{RequestUid: req.GetRequestUid()}
	client, err := sshDial(req.GetSshHost(), req.GetSshPort(), req.GetSshUser(), req.GetSshPassword(), req.GetSshPrivateKey())
	if err != nil {
		out.Error = err.Error()
		return out
	}
	defer client.Close()
	fs, err := sftp.NewClient(client)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	defer fs.Close()
	target, err := resolveSFTPPath(fs, req.GetPath())
	if err != nil {
		out.Error = err.Error()
		return out
	}
	switch strings.ToUpper(req.GetAction()) {
	case "LIST":
		infos, err := fs.ReadDir(target)
		if err != nil {
			out.Error = err.Error()
			return out
		}
		for _, info := range infos {
			name := info.Name()
			out.Entries = append(out.Entries, &v1.AgentSFTPEntry{
				Name: name, Path: strings.TrimSuffix(target, "/") + "/" + name,
				IsDir: info.IsDir(), Size: info.Size(), Mode: info.Mode().String(), ModTime: info.ModTime().UTC().Format(time.RFC3339),
			})
		}
		out.Ok = true
	case "DOWNLOAD":
		f, err := fs.Open(target)
		if err != nil {
			out.Error = err.Error()
			return out
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, maxSFTPBytes+1))
		if err != nil {
			out.Error = err.Error()
			return out
		}
		if len(data) > maxSFTPBytes {
			out.Error = "file too large"
			return out
		}
		out.Content = data
		out.Ok = true
	case "UPLOAD":
		if len(req.GetContent()) > maxSFTPBytes {
			out.Error = "file too large"
			return out
		}
		f, err := fs.Create(target)
		if err != nil {
			out.Error = err.Error()
			return out
		}
		if _, err := f.Write(req.GetContent()); err != nil {
			_ = f.Close()
			out.Error = err.Error()
			return out
		}
		_ = f.Close()
		out.Ok = true
	case "MKDIR":
		if err := fs.MkdirAll(target); err != nil {
			out.Error = err.Error()
			return out
		}
		out.Ok = true
	case "DELETE":
		if err := fs.RemoveAll(target); err != nil {
			out.Error = err.Error()
			return out
		}
		out.Ok = true
	default:
		out.Error = "unknown sftp action"
	}
	return out
}

const (
	maxK8sRequestBytes  = 4 << 20
	maxK8sResponseBytes = 4 << 20
)

var k8sMethods = map[string]struct{}{
	http.MethodGet:    {},
	http.MethodPost:   {},
	http.MethodPut:    {},
	http.MethodPatch:  {},
	http.MethodDelete: {},
	http.MethodHead:   {},
}

func (c *Connector) k8s(req *v1.AgentK8SRequest) *v1.AgentK8SResult {
	out := &v1.AgentK8SResult{RequestUid: req.GetRequestUid()}
	method := strings.ToUpper(strings.TrimSpace(req.GetMethod()))
	if _, ok := k8sMethods[method]; !ok {
		out.Error = "unsupported method"
		return out
	}
	conn, err := k8sConnectionFromEnv()
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if conn.base.Scheme != "https" && conn.base.Scheme != "http" {
		out.Error = "invalid kubernetes api address"
		return out
	}
	target, err := k8sTarget(conn.base, req.GetPath(), req.GetQuery())
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if len(req.GetBody()) > maxK8sRequestBytes {
		out.Error = "request body too large"
		return out
	}
	httpReq, err := http.NewRequest(method, target.String(), bytes.NewReader(req.GetBody()))
	if err != nil {
		out.Error = err.Error()
		return out
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("User-Agent", "alkaid")
	if conn.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+conn.token)
	} else if conn.username != "" {
		httpReq.SetBasicAuth(conn.username, conn.password)
	}
	if hj := req.GetHeadersJson(); hj != "" {
		var headers map[string]string
		if json.Unmarshal([]byte(hj), &headers) == nil {
			for k, v := range headers {
				if k8sAllowedHeader(k) {
					httpReq.Header.Set(k, v)
				}
			}
		}
	}
	resp, err := conn.client.Do(httpReq)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxK8sResponseBytes+1))
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if len(body) > maxK8sResponseBytes {
		out.Error = "response body too large"
		return out
	}
	out.StatusCode = int32(resp.StatusCode)
	out.Body = body
	hdr, _ := json.Marshal(resp.Header)
	out.HeadersJson = string(hdr)
	return out
}

// k8sAPIBase 返回 API Server 地址，默认使用集群内 ServiceAccount 入口。
func k8sAPIBase() string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("ALKAID_K8S_API")), "/")
	if base == "" {
		return "https://kubernetes.default.svc"
	}
	return base
}

// k8sTarget 只接受以单个 / 开头的相对路径，避免调用方把请求改到别的主机。
func k8sTarget(base *url.URL, rawPath, rawQuery string) (*url.URL, error) {
	rawPath = strings.TrimSpace(rawPath)
	if !strings.HasPrefix(rawPath, "/") || strings.HasPrefix(rawPath, "//") || strings.Contains(rawPath, "\\") {
		return nil, merr.Invalid("invalid path")
	}
	ref, err := url.Parse(rawPath)
	if err != nil || ref.IsAbs() || ref.Host != "" || ref.User != nil {
		return nil, merr.Invalid("invalid path")
	}
	for _, segment := range strings.Split(ref.Path, "/") {
		if segment == ".." {
			return nil, merr.Invalid("invalid path")
		}
	}
	target := *base
	target.Path = strings.TrimRight(base.Path, "/") + ref.Path
	target.RawPath = ""
	target.RawQuery = rawQuery
	return &target, nil
}

// k8sAllowedHeader 只放行内容协商头，Authorization、Host、Impersonate-* 一律由 Agent 自己决定。
func k8sAllowedHeader(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "accept", "content-type", "accept-encoding":
		return true
	default:
		return false
	}
}

// k8sTLSConfig 默认校验证书；只有显式设置 ALKAID_K8S_INSECURE=true 才跳过。
func k8sTLSConfig(serverName string) *tls.Config {
	cfg := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: serverName}
	if strings.EqualFold(os.Getenv("ALKAID_K8S_INSECURE"), "true") {
		cfg.InsecureSkipVerify = true
		return cfg
	}
	for _, path := range []string{
		os.Getenv("ALKAID_K8S_CA"),
		"/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
	} {
		if path == "" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		pool := x509.NewCertPool()
		if pool.AppendCertsFromPEM(raw) {
			cfg.RootCAs = pool
			break
		}
	}
	return cfg
}

// resolveSFTPPath 在配置了 ALKAID_SFTP_ROOT 时把操作限制在根目录内。
// 未配置时保持原行为，路径由 SSH 账户权限决定。
func resolveSFTPPath(fs *sftp.Client, raw string) (string, error) {
	root := strings.TrimSpace(os.Getenv("ALKAID_SFTP_ROOT"))
	if root == "" {
		return raw, nil
	}
	cleanRoot := path.Clean("/" + strings.TrimPrefix(root, "/"))
	realRoot, err := fs.RealPath(cleanRoot)
	if err != nil {
		realRoot = cleanRoot
	}
	target := raw
	if !strings.HasPrefix(target, "/") {
		target = cleanRoot + "/" + target
	}
	target = path.Clean(target)
	resolved := target
	if real, err := fs.RealPath(target); err == nil {
		resolved = real
	} else if real, err := fs.RealPath(path.Dir(target)); err == nil {
		resolved = path.Join(real, path.Base(target))
	}
	if resolved != realRoot && !strings.HasPrefix(resolved, realRoot+"/") {
		return "", merr.Invalid("path is outside sftp root")
	}
	return target, nil
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
