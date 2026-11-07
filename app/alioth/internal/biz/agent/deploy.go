package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

var DeployProviderSet = wire.NewSet(NewDeployer)

const (
	deployRunning   = "RUNNING"
	deploySucceeded = "SUCCEEDED"
	deployFailed    = "FAILED"

	deployStepPending   = "PENDING"
	deployStepRunning   = "RUNNING"
	deployStepSucceeded = "SUCCEEDED"
	deployStepFailed    = "FAILED"

	maxDeployLogBytes    = 256 << 10
	deployLogTailKeep    = maxDeployLogBytes / 2
	maxAlkaidPackageSize = 256 << 20

	defaultDeployTimeout       = 30 * time.Second
	deployInstallTimeout       = 5 * time.Minute
	defaultDeployOnlineTimeout = 60 * time.Second
	deployOnlinePoll           = time.Second
)

// deploySteps 是安装过程的固定步骤。键用于前端本地化，标题是服务端兜底文案。
var deploySteps = []struct {
	key   string
	title string
}{
	{"connect", "建立 SSH 连接"},
	{"inspect", "检查目标机器"},
	{"package", "准备摇光安装包"},
	{"upload", "上传并解压安装包"},
	{"install", "安装摇光 Agent"},
	{"configure", "写入服务配置"},
	{"service", "注册并启动 systemd 服务"},
	{"online", "等待 Agent 上报上线"},
}

// Deployer 用机器上保存的 SSH 凭据把摇光安装到目标机器，并把每一步记录下来。
// 安装会话只保存在当前进程内存里；多副本部署时查询请求要落到同一副本。
type Deployer struct {
	registry       *Registry
	dialer         repository.SSHDialer
	packages       repository.AgentPackageSource
	grpcAddr       string
	commandTimeout time.Duration
	onlineTimeout  time.Duration
	mu             sync.Mutex
	deploys        map[string]*deployRun
}

// NewDeployer 组装自动安装用例。
func NewDeployer(registry *Registry, dialer repository.SSHDialer, packages repository.AgentPackageSource, cfg *conf.Deploy) *Deployer {
	commandTimeout := defaultDeployTimeout
	if cfg.GetTimeout() != nil && cfg.GetTimeout().AsDuration() > 0 {
		commandTimeout = cfg.GetTimeout().AsDuration()
	}
	onlineTimeout := defaultDeployOnlineTimeout
	if cfg.GetOnlineTimeout() != nil && cfg.GetOnlineTimeout().AsDuration() > 0 {
		onlineTimeout = cfg.GetOnlineTimeout().AsDuration()
	}
	grpcAddr := strings.TrimSpace(cfg.GetAliothGrpcAddr())
	if grpcAddr == "" {
		grpcAddr = "127.0.0.1:9000"
	}
	return &Deployer{
		registry:       registry,
		dialer:         dialer,
		packages:       packages,
		grpcAddr:       grpcAddr,
		commandTimeout: commandTimeout,
		onlineTimeout:  onlineTimeout,
		deploys:        map[string]*deployRun{},
	}
}

// Start 启动一次安装，返回启动时的快照。同一台机器同时只允许一个安装。
func (deployer *Deployer) Start(ctx context.Context, uid string) (*bo.AgentDeploy, error) {
	ns, err := deployer.registry.access.Allow(ctx, "agent:write")
	if err != nil {
		return nil, err
	}
	machine, err := deployer.registry.agents.GetAgent(ctx, ns, uid)
	if err != nil {
		return nil, deployer.registry.access.Wrap(err)
	}
	if machine.Status != "ACTIVE" {
		return nil, merr.Forbidden("agent is revoked")
	}
	if machine.SSHHost == "" || machine.SSHUser == "" {
		return nil, merr.Invalid("ssh host or user is empty")
	}
	password, privateKey, err := deployer.registry.agents.GetSSHSecrets(ctx, uid)
	if err != nil {
		return nil, merr.Invalid("ssh credentials missing")
	}
	if password == "" && privateKey == "" {
		return nil, merr.Invalid("ssh credentials missing")
	}

	deployer.mu.Lock()
	if current, ok := deployer.deploys[uid]; ok && current.running() {
		deployer.mu.Unlock()
		return nil, merr.Conflict("agent deploy is running")
	}
	run := newDeployRun(uid)
	deployer.deploys[uid] = run
	deployer.mu.Unlock()

	target := bo.SSHTarget{
		Host: machine.SSHHost, Port: machine.SSHPort, User: machine.SSHUser,
		Password: password, PrivateKey: privateKey,
	}
	// 安装不能跟着请求上下文结束，接口返回后继续执行，页面通过 GetDeploy 轮询。
	go deployer.install(context.WithoutCancel(ctx), machine, target, run)
	return snapshot(run), nil
}

// Get 返回最近一次安装的进度。没有安装过时返回 nil。
func (deployer *Deployer) Get(ctx context.Context, uid string) (*bo.AgentDeploy, error) {
	ns, err := deployer.registry.access.Allow(ctx, "agent:read")
	if err != nil {
		return nil, err
	}
	if _, err := deployer.registry.agents.GetAgent(ctx, ns, uid); err != nil {
		return nil, deployer.registry.access.Wrap(err)
	}
	deployer.mu.Lock()
	run := deployer.deploys[uid]
	deployer.mu.Unlock()
	if run == nil {
		return nil, nil
	}
	return snapshot(run), nil
}

func (deployer *Deployer) install(ctx context.Context, machine *bo.Agent, target bo.SSHTarget, run *deployRun) {
	defer run.finish()
	session, err := deployer.stepConnect(ctx, target, run)
	if err != nil {
		run.fail("connect", err)
		return
	}
	defer session.Close()

	info, err := deployer.stepInspect(ctx, session, run)
	if err != nil {
		run.fail("inspect", err)
		return
	}

	name, packageData, err := deployer.stepPackage(ctx, info.arch, run)
	if err != nil {
		run.fail("package", err)
		return
	}

	packageDir, err := deployer.stepUpload(ctx, session, packageData, name, run)
	if err != nil {
		run.fail("upload", err)
		return
	}

	if err := deployer.stepInstall(ctx, session, info.root, target.Password, packageDir, run); err != nil {
		run.fail("install", err)
		return
	}

	if err := deployer.stepConfigure(ctx, session, machine, info.root, target.Password, run); err != nil {
		run.fail("configure", err)
		return
	}

	if err := deployer.stepService(ctx, session, info.root, target.Password, run); err != nil {
		run.fail("service", err)
		return
	}

	if err := deployer.stepOnline(ctx, machine, run); err != nil {
		run.fail("online", err)
		return
	}
	run.succeed()
}

func (deployer *Deployer) stepConnect(ctx context.Context, target bo.SSHTarget, run *deployRun) (repository.SSHSession, error) {
	run.start("connect", fmt.Sprintf("连接 %s@%s:%d", target.User, target.Host, target.Port))
	callCtx, cancel := context.WithTimeout(ctx, deployer.commandTimeout)
	defer cancel()
	session, err := deployer.dialer.Dial(callCtx, target)
	if err != nil {
		return nil, err
	}
	run.report("connect", "SSH 登录成功")
	return session, nil
}

type deployMachineInfo struct {
	arch string
	root bool
}

func (deployer *Deployer) stepInspect(ctx context.Context, session repository.SSHSession, run *deployRun) (deployMachineInfo, error) {
	run.start("inspect", "读取目标机器系统与架构")
	callCtx, cancel := context.WithTimeout(ctx, deployer.commandTimeout)
	defer cancel()
	// 用不带参数的 uname 读内核名；个别精简镜像的 uname -s 返回异常，不带参数更稳。
	osName, err := deployer.mustRaw(callCtx, session, "uname", nil)
	if err != nil {
		return deployMachineInfo{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(osName), "linux") {
		return deployMachineInfo{}, merr.Invalid("target machine is not linux")
	}
	rawArch, err := deployer.mustRaw(callCtx, session, "uname -m", nil)
	if err != nil {
		return deployMachineInfo{}, err
	}
	arch, err := mapPackageArch(rawArch)
	if err != nil {
		return deployMachineInfo{}, err
	}
	uid, err := deployer.mustRaw(callCtx, session, "id -u", nil)
	if err != nil {
		return deployMachineInfo{}, err
	}
	info := deployMachineInfo{arch: arch, root: strings.TrimSpace(uid) == "0"}
	rootText := "不是 root"
	if info.root {
		rootText = "是 root"
	}
	run.report("inspect", fmt.Sprintf("系统 linux，架构 %s，SSH 用户%s", arch, rootText))
	return info, nil
}

func (deployer *Deployer) stepPackage(ctx context.Context, arch string, run *deployRun) (string, []byte, error) {
	run.start("package", "准备 "+arch+" 安装包")
	name, content, err := deployer.packages.Fetch(ctx, arch)
	if err != nil {
		return "", nil, err
	}
	defer content.Close()
	data, err := io.ReadAll(io.LimitReader(content, maxAlkaidPackageSize+1))
	if err != nil {
		return "", nil, merr.Server(fmt.Sprintf("read alkaid package failed: %v", err))
	}
	if len(data) == 0 {
		return "", nil, merr.Server("alkaid package is empty")
	}
	if len(data) > maxAlkaidPackageSize {
		return "", nil, merr.Server("alkaid package is too large")
	}
	run.report("package", fmt.Sprintf("已准备 %s（%d 字节）", name, len(data)))
	return name, data, nil
}

func (deployer *Deployer) stepUpload(ctx context.Context, session repository.SSHSession, data []byte, name string, run *deployRun) (string, error) {
	run.start("upload", "上传并解压 "+name)
	base := fmt.Sprintf("/tmp/alkaid-%d", time.Now().UnixNano())
	packagePath := base + ".tar.gz"
	callCtx, cancel := context.WithTimeout(ctx, deployInstallTimeout)
	defer cancel()
	if err := session.Upload(callCtx, packagePath, bytes.NewReader(data)); err != nil {
		return "", err
	}
	packageDir := base + "-pkg"
	command := fmt.Sprintf("mkdir -p %s && tar -xzf %s -C %s && ls -1 %s", packageDir, packagePath, packageDir, packageDir)
	run.log("$ " + command)
	extractCtx, extractCancel := context.WithTimeout(ctx, deployInstallTimeout)
	defer extractCancel()
	out, err := deployer.mustRaw(extractCtx, session, command, nil)
	if err != nil {
		return "", err
	}
	root := ""
	for _, line := range strings.Split(out, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			root = packageDir + "/" + line
			break
		}
	}
	if root == "" {
		return "", merr.Server("alkaid package has no files")
	}
	run.report("upload", "解压完成："+root)
	return root, nil
}

func (deployer *Deployer) stepInstall(ctx context.Context, session repository.SSHSession, root bool, password, packageDir string, run *deployRun) error {
	run.start("install", "安装摇光二进制与 systemd 单元")
	// 重装时目标上可能还有旧进程占用二进制；先停服务，忽略「未安装 / 无 systemd」的失败。
	stopCtx, stopCancel := context.WithTimeout(ctx, deployer.commandTimeout)
	if _, err := deployer.mustSudo(stopCtx, session, root, password, "systemctl stop alkaid", nil); err == nil {
		run.report("install", "已停止旧的 alkaid 服务")
	}
	stopCancel()
	command := fmt.Sprintf("cd %s && sh ./install.sh --no-start", quoteShell(packageDir))
	run.log("$ " + command)
	callCtx, cancel := context.WithTimeout(ctx, deployInstallTimeout)
	defer cancel()
	out, err := deployer.mustSudo(callCtx, session, root, password, command, nil)
	if err != nil {
		return err
	}
	run.log(out)
	run.report("install", "安装完成")
	return nil
}

func (deployer *Deployer) stepConfigure(ctx context.Context, session repository.SSHSession, machine *bo.Agent, root bool, password string, run *deployRun) error {
	run.start("configure", "写入 alkaid.env，绑定控制面与令牌")
	// 到这一步安装已经成功，再重签令牌；前面的失败不会影响线上 Agent。
	token := newAgentToken()
	if _, err := deployer.registry.agents.SetTokenHash(ctx, machine.NamespaceUID, machine.UID, hashToken(token), "ACTIVE"); err != nil {
		return deployer.registry.access.Wrap(err)
	}
	content := strings.Join([]string{
		"ALKAID_HTTP_ADDR=0.0.0.0:8022",
		"ALKAID_GRPC_ADDR=0.0.0.0:9022",
		"ALKAID_ALIOTH_GRPC_ADDR=" + deployer.grpcAddr,
		"ALKAID_AGENT_TOKEN=" + token,
		"",
	}, "\n")
	command := "umask 077 && cat > /etc/alioth/alkaid.env && chmod 600 /etc/alioth/alkaid.env"
	run.log("$ " + command)
	callCtx, cancel := context.WithTimeout(ctx, deployer.commandTimeout)
	defer cancel()
	if _, err := deployer.mustSudo(callCtx, session, root, password, command, []byte(content)); err != nil {
		return err
	}
	if hint := deployer.controlPlaneHint(machine.SSHHost); hint != "" {
		run.log("    warning: " + hint)
	}
	run.report("configure", "服务配置已写入 /etc/alioth/alkaid.env")
	return nil
}

func (deployer *Deployer) stepService(ctx context.Context, session repository.SSHSession, root bool, password string, run *deployRun) error {
	run.start("service", "启用并重启 alkaid 服务")
	command := "systemctl daemon-reload && systemctl enable alkaid && systemctl restart alkaid && sleep 1 && systemctl is-active alkaid"
	run.log("$ " + command)
	callCtx, cancel := context.WithTimeout(ctx, deployer.commandTimeout)
	defer cancel()
	out, err := deployer.mustSudo(callCtx, session, root, password, command, nil)
	if err != nil {
		return err
	}
	run.log(out)
	if strings.TrimSpace(out) != "active" {
		return merr.Server("alkaid service is not active")
	}
	run.report("service", "服务已启动")
	return nil
}

func (deployer *Deployer) stepOnline(ctx context.Context, machine *bo.Agent, run *deployRun) error {
	run.start("online", "等待 Agent 向控制面注册")
	uid := machine.UID
	deadline := time.NewTimer(deployer.onlineTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(deployOnlinePoll)
	defer ticker.Stop()
	for {
		if deployer.registry.hub.Online(uid) {
			run.report("online", "Agent 已上线")
			return nil
		}
		// 多副本时 Agent 可能注册到别的副本；在线标记也会写进数据库。
		if updated, err := deployer.registry.agents.GetAgentByUID(ctx, uid); err == nil && updated.Online {
			run.report("online", "Agent 已上线")
			return nil
		}
		select {
		case <-ctx.Done():
			return merr.Server("agent did not register before the request was cancelled")
		case <-deadline.C:
			message := fmt.Sprintf("agent did not register within %s", deployer.onlineTimeout)
			if hint := deployer.controlPlaneHint(machine.SSHHost); hint != "" {
				message += ": " + hint
			}
			return merr.Server(message)
		case <-ticker.C:
		}
	}
}

// controlPlaneHint 说明「写进目标机的控制面地址」为什么连不上。
// 目标机不是本机、而配置的地址是 127.0.0.1 时，Agent 只会连回目标机自己，
// 上线必然超时；两种情况都返回空串表示地址本身没问题。
func (deployer *Deployer) controlPlaneHint(targetHost string) string {
	host, _, err := net.SplitHostPort(deployer.grpcAddr)
	if err != nil || host == "" || loopbackHost(targetHost) || !loopbackHost(host) {
		return ""
	}
	return fmt.Sprintf(
		"control plane address %s is a loopback address, so %s cannot reach it; set deploy.alioth_grpc_addr to an address the target machine can reach",
		deployer.grpcAddr, targetHost,
	)
}

// loopbackHost 判断主机是不是只指向本机：localhost、127.0.0.0/8 和 ::1。
func loopbackHost(host string) bool {
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// mustRaw 以 SSH 登录用户直接执行命令，不经过 sudo，用于探测系统信息和 /tmp 操作。
func (deployer *Deployer) mustRaw(ctx context.Context, session repository.SSHSession, command string, stdin []byte) (string, error) {
	output, exitCode, err := session.Run(ctx, command, stdin)
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", merr.Server(fmt.Sprintf("command failed (%d): %s", exitCode, truncateOutput(output)))
	}
	return output, nil
}

// mustSudo 以 root 执行安装命令。SSH 用户本身是 root 时不再套 sudo。
func (deployer *Deployer) mustSudo(ctx context.Context, session repository.SSHSession, root bool, password, command string, stdin []byte) (string, error) {
	if root {
		output, exitCode, err := session.Run(ctx, command, stdin)
		return deployer.result(output, exitCode, err)
	}
	input := make([]byte, 0, len(password)+1+len(stdin))
	input = append(input, password...)
	input = append(input, '\n')
	input = append(input, stdin...)
	output, exitCode, err := session.Run(ctx, "sudo -S -p '' sh -c "+quoteShell(command), input)
	return deployer.result(output, exitCode, err)
}

func (deployer *Deployer) result(output string, exitCode int, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if exitCode != 0 {
		return "", merr.Server(fmt.Sprintf("command failed (%d): %s", exitCode, truncateOutput(output)))
	}
	return output, nil
}

func mapPackageArch(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "x86_64", "amd64":
		return "amd64", nil
	case "aarch64", "arm64":
		return "arm64", nil
	default:
		return "", merr.Invalid("unsupported cpu architecture: " + strings.TrimSpace(raw))
	}
}

func quoteShell(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

// deployErrorMessage 取 Kratos 错误的用户可读 message，避免把 code/reason 拼接串显示到页面。
func deployErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	status := kerrors.FromError(err)
	if status != nil && status.Message != "" {
		return status.Message
	}
	return err.Error()
}

func truncateOutput(output string) string {
	output = strings.TrimSpace(output)
	if len(output) > 512 {
		return output[:512] + "..."
	}
	return output
}

// deployRun 是一次安装的可变状态。日志、步骤和最终状态都从这里取快照。
type deployRun struct {
	uid      string
	agentUID string
	mu       sync.Mutex
	status   string
	step     string
	steps    []bo.AgentDeployStep
	logs     []byte
	logCut   bool
	message  string
	started  time.Time
	finished time.Time
	done     chan struct{}
}

func newDeployRun(agentUID string) *deployRun {
	steps := make([]bo.AgentDeployStep, 0, len(deploySteps))
	for _, step := range deploySteps {
		steps = append(steps, bo.AgentDeployStep{Key: step.key, Title: step.title, Status: deployStepPending})
	}
	return &deployRun{
		uid: id.New(), agentUID: agentUID, status: deployRunning, step: deploySteps[0].key,
		steps: steps, started: time.Now(), done: make(chan struct{}),
	}
}

func (run *deployRun) running() bool {
	run.mu.Lock()
	defer run.mu.Unlock()
	return run.status == deployRunning
}

func (run *deployRun) start(step, line string) {
	run.mu.Lock()
	defer run.mu.Unlock()
	run.step = step
	for index := range run.steps {
		if run.steps[index].Key == step {
			run.steps[index].Status = deployStepRunning
			run.steps[index].Message = line
		}
	}
	run.appendLog("==> " + line + "\n")
}

func (run *deployRun) report(step, line string) {
	run.mu.Lock()
	defer run.mu.Unlock()
	for index := range run.steps {
		if run.steps[index].Key == step {
			run.steps[index].Status = deployStepSucceeded
			run.steps[index].Message = line
		}
	}
	run.appendLog("    " + line + "\n")
}

func (run *deployRun) log(output string) {
	output = strings.TrimRight(output, "\n")
	if output == "" {
		return
	}
	run.mu.Lock()
	defer run.mu.Unlock()
	run.appendLog(output + "\n")
}

func (run *deployRun) fail(step string, err error) {
	run.mu.Lock()
	defer run.mu.Unlock()
	message := deployErrorMessage(err)
	for index := range run.steps {
		if run.steps[index].Key == step {
			run.steps[index].Status = deployStepFailed
			run.steps[index].Message = message
		}
	}
	run.status = deployFailed
	run.message = message
	run.finished = time.Now()
	run.appendLog("!!! " + step + " 失败：" + message + "\n")
}

func (run *deployRun) succeed() {
	run.mu.Lock()
	defer run.mu.Unlock()
	for index := range run.steps {
		run.steps[index].Status = deployStepSucceeded
	}
	run.status = deploySucceeded
	run.finished = time.Now()
	run.appendLog("==> 安装完成，Agent 已上线\n")
}

func (run *deployRun) finish() {
	close(run.done)
}

func (run *deployRun) appendLog(line string) {
	run.logs = append(run.logs, line...)
	if len(run.logs) <= maxDeployLogBytes {
		return
	}
	cut := len(run.logs) - deployLogTailKeep
	for cut < len(run.logs) && !utf8.RuneStart(run.logs[cut]) {
		cut++
	}
	run.logs = run.logs[cut:]
	run.logCut = true
}

func snapshot(run *deployRun) *bo.AgentDeploy {
	run.mu.Lock()
	defer run.mu.Unlock()
	steps := make([]bo.AgentDeployStep, len(run.steps))
	copy(steps, run.steps)
	return &bo.AgentDeploy{
		UID: run.uid, AgentUID: run.agentUID, Status: run.status, Step: run.step,
		Steps: steps, Log: string(run.logs), LogTruncated: run.logCut,
		Error: run.message, StartedAt: run.started, FinishedAt: run.finished,
	}
}
