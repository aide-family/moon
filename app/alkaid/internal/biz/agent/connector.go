package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alkaid/internal/data"
	"github.com/aide-family/Alioth/pkg/merr"
)

const maxOutputBytes = 64 * 1024

var ProviderSet = wire.NewSet(NewConnector)

// Connector 作为 kratos Server 跑出站流。
type Connector struct {
	addr      string
	token     string
	heartbeat time.Duration
	log       *log.Helper
	cancel    context.CancelFunc
	done      chan struct{}
}

func NewConnector(d *data.Data, logger log.Logger) (*Connector, error) {
	if d.AgentToken == "" {
		log.NewHelper(logger).Warn("agent token is empty; outbound stream disabled")
	}
	interval := d.Heartbeat
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Connector{
		addr: d.AliothGRPC, token: d.AgentToken, heartbeat: interval,
		log: log.NewHelper(logger), done: make(chan struct{}),
	}, nil
}

func (c *Connector) Start(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	go func() {
		defer close(c.done)
		c.loop(runCtx)
	}()
	return nil
}

func (c *Connector) Stop(context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	<-c.done
	return nil
}

func (c *Connector) loop(ctx context.Context) {
	if c.token == "" {
		<-ctx.Done()
		return
	}
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		err := c.session(ctx)
		if ctx.Err() != nil {
			return
		}
		c.log.Warnf("agent stream closed: %v; retry in %s", err, backoff)
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

func (c *Connector) session(ctx context.Context) error {
	conn, err := grpc.NewClient(c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	if err := stream.Send(&v1.AgentFrame{Register: &v1.AgentRegister{
		Hostname: hostname, MachineId: machineID, LocalIp: localIP(), Version: "alkaid-dev",
	}}); err != nil {
		return err
	}
	errCh := make(chan error, 2)
	go func() {
		ticker := time.NewTicker(c.heartbeat)
		defer ticker.Stop()
		var seq int64
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-ticker.C:
				seq++
				if err := stream.Send(&v1.AgentFrame{Heartbeat: &v1.AgentHeartbeat{Seq: seq}}); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
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
				result := c.exec(req)
				if err := stream.Send(&v1.AgentFrame{ExecResult: result}); err != nil {
					errCh <- err
					return
				}
			}
			if req := frame.GetSftpRequest(); req != nil {
				result := c.sftp(req)
				if err := stream.Send(&v1.AgentFrame{SftpResult: result}); err != nil {
					errCh <- err
					return
				}
			}
			if req := frame.GetK8SRequest(); req != nil {
				result := c.k8s(req)
				if err := stream.Send(&v1.AgentFrame{K8SResult: result}); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
	return <-errCh
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
	switch strings.ToUpper(req.GetAction()) {
	case "LIST":
		infos, err := fs.ReadDir(req.GetPath())
		if err != nil {
			out.Error = err.Error()
			return out
		}
		for _, info := range infos {
			name := info.Name()
			out.Entries = append(out.Entries, &v1.AgentSFTPEntry{
				Name: name, Path: strings.TrimSuffix(req.GetPath(), "/") + "/" + name,
				IsDir: info.IsDir(), Size: info.Size(), Mode: info.Mode().String(), ModTime: info.ModTime().UTC().Format(time.RFC3339),
			})
		}
		out.Ok = true
	case "DOWNLOAD":
		f, err := fs.Open(req.GetPath())
		if err != nil {
			out.Error = err.Error()
			return out
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, maxOutputBytes+1))
		if err != nil {
			out.Error = err.Error()
			return out
		}
		if len(data) > maxOutputBytes {
			data = data[:maxOutputBytes]
		}
		out.Content = data
		out.Ok = true
	case "UPLOAD":
		f, err := fs.Create(req.GetPath())
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
		if err := fs.MkdirAll(req.GetPath()); err != nil {
			out.Error = err.Error()
			return out
		}
		out.Ok = true
	case "DELETE":
		if err := fs.RemoveAll(req.GetPath()); err != nil {
			out.Error = err.Error()
			return out
		}
		out.Ok = true
	default:
		out.Error = "unknown sftp action"
	}
	return out
}

func (c *Connector) k8s(req *v1.AgentK8SRequest) *v1.AgentK8SResult {
	out := &v1.AgentK8SResult{RequestUid: req.GetRequestUid()}
	base := strings.TrimRight(os.Getenv("ALKAID_K8S_API"), "/")
	if base == "" {
		base = "https://kubernetes.default.svc"
	}
	token := os.Getenv("ALKAID_K8S_TOKEN")
	if token == "" {
		raw, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err == nil {
			token = strings.TrimSpace(string(raw))
		}
	}
	u := base + req.GetPath()
	if q := req.GetQuery(); q != "" {
		if strings.Contains(u, "?") {
			u += "&" + q
		} else {
			u += "?" + q
		}
	}
	httpReq, err := http.NewRequest(req.GetMethod(), u, bytes.NewReader(req.GetBody()))
	if err != nil {
		out.Error = err.Error()
		return out
	}
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	if hj := req.GetHeadersJson(); hj != "" {
		var headers map[string]string
		if json.Unmarshal([]byte(hj), &headers) == nil {
			for k, v := range headers {
				httpReq.Header.Set(k, v)
			}
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Do(httpReq)
	if err != nil {
		out.Error = err.Error()
		return out
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxOutputBytes))
	out.StatusCode = int32(resp.StatusCode)
	out.Body = body
	hdr, _ := json.Marshal(resp.Header)
	out.HeadersJson = string(hdr)
	return out
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
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}
	addr := net.JoinHostPort(host, itoa(int(port)))
	return ssh.Dial("tcp", addr, cfg)
}

func readMachineID() string {
	raw, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(raw))
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
