package agent

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

// TestDeployInstallsAgent 覆盖成功路径：SSH 探测、下载、上传、安装、写配置、启动、上线。
func TestDeployInstallsAgent(t *testing.T) {
	machine := &bo.Agent{
		UID: "agent-1", NamespaceUID: "namespace-1", Name: "web-1", Status: "ACTIVE",
		SSHHost: "10.0.0.8", SSHPort: 22, SSHUser: "root", HasPassword: true,
	}
	repo := &deployAgentRepo{machine: machine, password: "secret"}
	hub := agenthub.NewHub()
	hub.Put(agenthub.NewSession("agent-1", "namespace-1"))
	registry := NewRegistry(repo, hub, testChecker(perm.AgentWrite, perm.AgentRead, perm.MenuPermission("agent")))
	session := &deploySession{}
	dialer := &deployDialer{session: session}
	packages := &deployPackages{name: "alkaid_v1.0.0_linux_amd64.tar.gz", data: []byte("tarball")}
	deployer := NewDeployer(registry, dialer, packages, &conf.Deploy{AliothGrpcAddr: "alioth:9000"})

	started, err := deployer.Start(tenantContext(), "agent-1")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if started.Status != deployRunning || started.Step != "connect" {
		t.Fatalf("started snapshot = %+v", started)
	}
	run := deployerRun(t, deployer, "agent-1")
	waitDone(t, run.done)

	snapshot := snapshot(run)
	if snapshot.Status != deploySucceeded {
		t.Fatalf("status = %q, error = %q, log = %s", snapshot.Status, snapshot.Error, snapshot.Log)
	}
	for _, step := range snapshot.Steps {
		if step.Status != deployStepSucceeded {
			t.Fatalf("step %s = %s", step.Key, step.Status)
		}
	}
	if !strings.Contains(snapshot.Log, "install.sh --no-start") {
		t.Fatalf("log missing install command: %s", snapshot.Log)
	}
	if repo.rotatedHash == "" || repo.rotatedNamespace != "namespace-1" {
		t.Fatalf("token not rotated: %+v", repo)
	}
	if packages.arch != "amd64" {
		t.Fatalf("package arch = %q", packages.arch)
	}
	if !strings.Contains(string(session.uploaded), "tarball") {
		t.Fatalf("uploaded = %q", session.uploaded)
	}
	env := session.stdinFor("cat > /etc/alioth/alkaid.env")
	if !strings.Contains(env, "ALKAID_ALIOTH_GRPC_ADDR=alioth:9000") {
		t.Fatalf("env = %q", env)
	}
	if !strings.Contains(env, "ALKAID_AGENT_TOKEN=ag-") {
		t.Fatalf("env has no token: %q", env)
	}
	// 令牌只写进目标机器，不能出现在返回给页面的安装日志里。
	token := tokenFromEnv(env)
	if token == "" || strings.Contains(snapshot.Log, token) {
		t.Fatalf("deploy log leaks token: %s", snapshot.Log)
	}
}

// TestDeployConflict 确认同一台机器同时只允许一个安装。
func TestDeployConflict(t *testing.T) {
	machine := &bo.Agent{UID: "agent-1", NamespaceUID: "namespace-1", Status: "ACTIVE", SSHHost: "10.0.0.8", SSHUser: "root"}
	repo := &deployAgentRepo{machine: machine, password: "secret"}
	hub := agenthub.NewHub()
	hub.Put(agenthub.NewSession("agent-1", "namespace-1"))
	registry := NewRegistry(repo, hub, testChecker(perm.AgentWrite, perm.MenuPermission("agent")))
	dialer := &deployDialer{block: make(chan struct{}), started: make(chan struct{})}
	deployer := NewDeployer(registry, dialer, &deployPackages{data: []byte("tarball")}, &conf.Deploy{})
	if _, err := deployer.Start(tenantContext(), "agent-1"); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	select {
	case <-dialer.started:
	case <-time.After(2 * time.Second):
		t.Fatal("dial was not called")
	}
	if _, err := deployer.Start(tenantContext(), "agent-1"); !merr.IsConflict(err) {
		t.Fatalf("second Start error = %v, want conflict", err)
	}
	close(dialer.block)
	run := deployerRun(t, deployer, "agent-1")
	waitDone(t, run.done)
}

// TestDeployConnectFailureKeepsToken 确认 SSH 连不上时不动线上令牌。
func TestDeployConnectFailureKeepsToken(t *testing.T) {
	machine := &bo.Agent{UID: "agent-1", NamespaceUID: "namespace-1", Status: "ACTIVE", SSHHost: "10.0.0.8", SSHUser: "root"}
	repo := &deployAgentRepo{machine: machine, password: "secret"}
	registry := NewRegistry(repo, agenthub.NewHub(), testChecker(perm.AgentWrite, perm.MenuPermission("agent")))
	deployer := NewDeployer(registry, &deployDialer{err: merr.Invalid("ssh connect failed")}, &deployPackages{}, &conf.Deploy{})
	if _, err := deployer.Start(tenantContext(), "agent-1"); err != nil {
		t.Fatalf("Start: %v", err)
	}
	run := deployerRun(t, deployer, "agent-1")
	waitDone(t, run.done)
	snapshot := snapshot(run)
	if snapshot.Status != deployFailed {
		t.Fatalf("status = %q", snapshot.Status)
	}
	if snapshot.Steps[0].Status != deployStepFailed {
		t.Fatalf("connect step = %+v", snapshot.Steps[0])
	}
	if repo.rotatedHash != "" {
		t.Fatal("token rotated before SSH succeeded")
	}
}

// TestDeployRequiresWritePermission 确认没有 agent:write 时不会发起安装。
func TestDeployRequiresWritePermission(t *testing.T) {
	machine := &bo.Agent{UID: "agent-1", NamespaceUID: "namespace-1", Status: "ACTIVE", SSHHost: "10.0.0.8", SSHUser: "root"}
	repo := &deployAgentRepo{machine: machine, password: "secret"}
	registry := NewRegistry(repo, agenthub.NewHub(), testChecker(perm.AgentRead, perm.MenuPermission("agent")))
	deployer := NewDeployer(registry, &deployDialer{}, &deployPackages{}, &conf.Deploy{})
	if _, err := deployer.Start(tenantContext(), "agent-1"); kerrors.Reason(err) != merr.ReasonForbidden {
		t.Fatalf("Start error = %v, want forbidden", err)
	}
}

func waitDone(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("deploy did not finish")
	}
}

func deployerRun(t *testing.T, deployer *Deployer, uid string) *deployRun {
	t.Helper()
	deployer.mu.Lock()
	defer deployer.mu.Unlock()
	run := deployer.deploys[uid]
	if run == nil {
		t.Fatal("deploy run is missing")
	}
	return run
}

func tokenFromEnv(env string) string {
	for _, line := range strings.Split(env, "\n") {
		if strings.HasPrefix(line, "ALKAID_AGENT_TOKEN=") {
			return strings.TrimPrefix(line, "ALKAID_AGENT_TOKEN=")
		}
	}
	return ""
}

type deployAgentRepo struct {
	repository.AgentRepo
	machine          *bo.Agent
	password         string
	rotatedHash      string
	rotatedNamespace string
}

func (repo *deployAgentRepo) GetAgent(_ context.Context, namespaceUID, uid string) (*bo.Agent, error) {
	if repo.machine == nil || repo.machine.UID != uid {
		return nil, merr.NotFound("not found")
	}
	return repo.machine, nil
}

func (repo *deployAgentRepo) GetAgentByUID(_ context.Context, uid string) (*bo.Agent, error) {
	return repo.GetAgent(context.Background(), "", uid)
}

func (repo *deployAgentRepo) GetSSHSecrets(context.Context, string) (string, string, error) {
	return repo.password, "", nil
}

func (repo *deployAgentRepo) SetTokenHash(_ context.Context, namespaceUID, uid, tokenHash, _ string) (*bo.Agent, error) {
	repo.rotatedHash = tokenHash
	repo.rotatedNamespace = namespaceUID
	return repo.machine, nil
}

type deployDialer struct {
	session repository.SSHSession
	err     error
	block   chan struct{}
	started chan struct{}
	once    sync.Once
}

func (dialer *deployDialer) Dial(context.Context, bo.SSHTarget) (repository.SSHSession, error) {
	dialer.once.Do(func() {
		if dialer.started != nil {
			close(dialer.started)
		}
	})
	if dialer.block != nil {
		<-dialer.block
	}
	if dialer.err != nil {
		return nil, dialer.err
	}
	if dialer.session == nil {
		dialer.session = &deploySession{}
	}
	return dialer.session, nil
}

type deployPackages struct {
	name string
	data []byte
	arch string
}

func (packages *deployPackages) Fetch(_ context.Context, arch string) (string, io.ReadCloser, error) {
	packages.arch = arch
	name := packages.name
	if name == "" {
		name = "alkaid_linux_" + arch + ".tar.gz"
	}
	return name, io.NopCloser(bytes.NewReader(packages.data)), nil
}

type deploySession struct {
	mu       sync.Mutex
	commands []deployCommand
	uploaded []byte
}

type deployCommand struct {
	command string
	stdin   []byte
}

func (session *deploySession) Run(_ context.Context, command string, stdin []byte) (string, int, error) {
	session.mu.Lock()
	session.commands = append(session.commands, deployCommand{command: command, stdin: append([]byte(nil), stdin...)})
	session.mu.Unlock()
	switch {
	case command == "uname":
		return "Linux\n", 0, nil
	case command == "uname -m":
		return "x86_64\n", 0, nil
	case command == "id -u":
		return "0\n", 0, nil
	case strings.Contains(command, "install.sh"):
		return "installed\n", 0, nil
	case strings.Contains(command, "mkdir -p"):
		return "alkaid_v1.0.0_linux_amd64\n", 0, nil
	case strings.Contains(command, "systemctl"):
		return "active\n", 0, nil
	default:
		return "", 0, nil
	}
}

func (session *deploySession) Upload(_ context.Context, _ string, content io.Reader) error {
	data, err := io.ReadAll(content)
	if err != nil {
		return err
	}
	session.mu.Lock()
	session.uploaded = data
	session.mu.Unlock()
	return nil
}

func (session *deploySession) Close() error { return nil }

func (session *deploySession) stdinFor(match string) string {
	session.mu.Lock()
	defer session.mu.Unlock()
	for _, command := range session.commands {
		if strings.Contains(command.command, match) {
			return string(command.stdin)
		}
	}
	return ""
}
