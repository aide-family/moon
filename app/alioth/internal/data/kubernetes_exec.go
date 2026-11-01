package data

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// podExecShellBuffer 是终端输出的缓冲帧数。浏览器消费慢时只丢新帧，不阻塞 SPDY 读取，
	// 避免一个卡住的页面把远端的 shell 一起卡死。
	podExecShellBuffer = 64
	// defaultPodExecCols 和 defaultPodExecRows 是未指定窗口大小时的默认值。
	defaultPodExecCols = int32(80)
	defaultPodExecRows = int32(24)
	// maxPodExecTerminalSize 与 proto 校验一致，避免绕过 HTTP 的调用要求超大窗口。
	maxPodExecTerminalSize = int32(1000)
)

// ExecPod 通过 pods/exec 子资源建立一条 SPDY 终端会话。
// 远端固定执行 /bin/sh；命令是否存在、凭据是否有 exec 权限由集群侧判定，
// 失败原因会作为会话的关闭原因返回，不把上游错误细节带出数据层。
func (r *kubernetesRepo) ExecPod(ctx context.Context, clusterUID, namespace, name, container string, cols, rows int32) (*bo.KubePodShell, error) {
	namespace, name, container = strings.TrimSpace(namespace), strings.TrimSpace(name), strings.TrimSpace(container)
	if namespace == "" {
		return nil, merr.Invalid("namespace is required")
	}
	if name == "" {
		return nil, merr.Invalid("pod name is required")
	}
	if container == "" {
		return nil, merr.Invalid("container is required")
	}
	restConfig, err := r.clients.restConfig(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	client, err := r.clients.client(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	request := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   []string{"/bin/sh"},
			Stdin:     true,
			Stdout:    true,
			TTY:       true,
		}, scheme.ParameterCodec)
	executor, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, request.URL())
	if err != nil {
		// 到这里说明已保存的 kubeconfig 连 SPDY transport 都建不起来，属于配置不可用。
		if r.log != nil {
			r.log.Error(fmt.Sprintf("pod exec executor: %v", err))
		}
		return nil, merr.Server("pod exec unavailable")
	}
	cols, rows = normalizePodExecSize(cols, rows)
	return newPodExecShell(ctx, r.log, executor, container, cols, rows), nil
}

// newPodExecShell 把 SPDY 的输入输出接到业务对象上：stdin 走管道，stdout/stderr 走有界通道，
// 会话结束时写入一条关闭原因。Stop 可以重复调用，保证 HTTP 处理器和进程退出都能清理。
func newPodExecShell(ctx context.Context, logger *slog.Logger, executor remotecommand.Executor, container string, cols, rows int32) *bo.KubePodShell {
	sessionCtx, cancel := context.WithCancel(ctx)
	stdinReader, stdinWriter := io.Pipe()
	output := newPodExecOutput(podExecShellBuffer)
	sizes := newPodExecSizeQueue(cols, rows)
	closed := make(chan string, 1)
	var once sync.Once
	finish := func(reason string) {
		once.Do(func() {
			cancel()
			sizes.Close()
			// 关闭读端也要做：StreamWithContext 里的 stdin 复制协程会一直等读端。
			_ = stdinWriter.Close()
			_ = stdinReader.Close()
			// 先给原因再关输出通道，处理器收到通道关闭时一定能读到原因。
			closed <- reason
			close(closed)
			output.Close()
		})
	}
	shell := &bo.KubePodShell{
		Container: container,
		Data:      output.data,
		Closed:    closed,
		Write: func(payload []byte) error {
			if _, err := stdinWriter.Write(payload); err != nil {
				return merr.Invalid("terminal session is closed")
			}
			return nil
		},
		Resize: func(nextCols, nextRows int32) error {
			if nextCols <= 0 || nextRows <= 0 {
				return merr.Invalid("terminal size is invalid")
			}
			sizes.Push(nextCols, nextRows)
			return nil
		},
		Stop: func() { finish("closed") },
	}
	go func() {
		err := executor.StreamWithContext(sessionCtx, remotecommand.StreamOptions{
			Stdin:             stdinReader,
			Stdout:            output,
			Stderr:            output,
			Tty:               true,
			TerminalSizeQueue: sizes,
		})
		finish(podExecCloseReason(logger, err, sessionCtx.Err()))
	}()
	return shell
}

// podExecOutput 把 SPDY 的 stdout/stderr 写进有界通道。
// 通道满时丢弃新输出：终端宁可少几帧，也不能让读取端阻塞整个会话。
type podExecOutput struct {
	mu     sync.Mutex
	closed bool
	data   chan []byte
}

func newPodExecOutput(size int) *podExecOutput {
	return &podExecOutput{data: make(chan []byte, size)}
}

func (output *podExecOutput) Write(payload []byte) (int, error) {
	chunk := make([]byte, len(payload))
	copy(chunk, payload)
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.closed {
		return len(payload), nil
	}
	select {
	case output.data <- chunk:
	default:
	}
	return len(payload), nil
}

// Close 关闭输出通道。Write 只持有同一把锁，关闭后不会再写，也不会重复关闭。
func (output *podExecOutput) Close() {
	output.mu.Lock()
	defer output.mu.Unlock()
	if output.closed {
		return
	}
	output.closed = true
	close(output.data)
}

// podExecSizeQueue 把浏览器窗口大小变化转成 remotecommand 的 TerminalSizeQueue。
// 只保留最近一次变化：缩放很快时不需要补发每一帧。
type podExecSizeQueue struct {
	sizes chan remotecommand.TerminalSize
	done  chan struct{}
	once  sync.Once
}

func newPodExecSizeQueue(cols, rows int32) *podExecSizeQueue {
	queue := &podExecSizeQueue{
		sizes: make(chan remotecommand.TerminalSize, 1),
		done:  make(chan struct{}),
	}
	queue.Push(cols, rows)
	return queue
}

// Push 写入最新窗口大小。通道满时用新值替换旧值；关闭后调用也不会阻塞或 panic。
func (queue *podExecSizeQueue) Push(cols, rows int32) {
	size := remotecommand.TerminalSize{Width: uint16(cols), Height: uint16(rows)}
	select {
	case queue.sizes <- size:
		return
	default:
	}
	select {
	case <-queue.sizes:
	default:
	}
	select {
	case queue.sizes <- size:
	default:
	}
}

// Next 满足 remotecommand.TerminalSizeQueue。关闭后返回 nil，让 client-go 的缩放协程退出。
func (queue *podExecSizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-queue.sizes:
		return &size
	case <-queue.done:
		return nil
	}
}

func (queue *podExecSizeQueue) Close() {
	queue.once.Do(func() { close(queue.done) })
}

// normalizePodExecSize 补齐并收敛窗口大小，保证传给远端的是合理值。
func normalizePodExecSize(cols, rows int32) (int32, int32) {
	if cols <= 0 {
		cols = defaultPodExecCols
	}
	if rows <= 0 {
		rows = defaultPodExecRows
	}
	if cols > maxPodExecTerminalSize {
		cols = maxPodExecTerminalSize
	}
	if rows > maxPodExecTerminalSize {
		rows = maxPodExecTerminalSize
	}
	return cols, rows
}

// podExecCloseReason 把会话结束原因转成浏览器可读的文本。
// 不返回上游原始错误，避免把集群地址、凭据或内部细节带到页面。
func podExecCloseReason(logger *slog.Logger, err error, canceled error) string {
	if err == nil {
		return "exited"
	}
	if canceled != nil {
		return "closed"
	}
	switch {
	case apierrors.IsNotFound(err):
		return "pod or container not found"
	case apierrors.IsForbidden(err), apierrors.IsUnauthorized(err):
		return "cluster credential is not allowed to exec into this pod"
	}
	if logger != nil {
		logger.Error(fmt.Sprintf("pod exec: %v", err))
	}
	return "pod exec session ended"
}
