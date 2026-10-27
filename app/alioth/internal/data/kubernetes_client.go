package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/cluster"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
)

const (
	// probeTimeout 是接入或测试集群时解析并探测 API Server 的超时。
	probeTimeout = 10 * time.Second
	// kubeCallTimeout 是单次资源读写的超时。每个请求都用独立的调用上下文，
	// 避免一个慢集群拖住整个请求。
	kubeCallTimeout = 15 * time.Second
	// kubeRequestTimeout 是客户端级请求超时，和单次调用超时保持一致。
	kubeRequestTimeout = 15 * time.Second
)

// clusterClients 缓存按集群 UID 建立的 clientset。
// 更新或删除集群后必须失效，否则后续请求会继续用旧凭据访问旧集群。
type clusterClients struct {
	deps   *Dependencies
	log    *slog.Logger
	mu     sync.RWMutex
	cached map[string]kubernetes.Interface
}

func newClusterClients(deps *Dependencies, logger *slog.Logger) *clusterClients {
	return &clusterClients{deps: deps, log: logger, cached: make(map[string]kubernetes.Interface)}
}

// client 取集群的 clientset，命中缓存时不再读数据库和 kubeconfig。
func (clients *clusterClients) client(ctx context.Context, clusterUID string) (kubernetes.Interface, error) {
	clients.mu.RLock()
	cached, ok := clients.cached[clusterUID]
	clients.mu.RUnlock()
	if ok {
		return cached, nil
	}
	// 读凭据和构建客户端都放在写锁内：否则替换 kubeconfig 的 invalidate 可能
	// 发生在旧客户端写回缓存之后，缓存会继续持有旧凭据。
	clients.mu.Lock()
	defer clients.mu.Unlock()
	if existing, ok := clients.cached[clusterUID]; ok {
		return existing, nil
	}
	kubeconfig, err := clients.kubeconfig(ctx, clusterUID)
	if err != nil {
		return nil, err
	}
	restConfig, _, _, err := parseKubeconfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		// kubeconfig 已经解析通过，构建客户端失败说明配置本身不可用。
		return nil, merr.Invalid("invalid kubeconfig")
	}
	clients.cached[clusterUID] = client
	return client, nil
}

// kubeconfig 读取并解密集群凭据。明文只停留在这个包内。
func (clients *clusterClients) kubeconfig(ctx context.Context, clusterUID string) (string, error) {
	row, err := clients.deps.client.Cluster.Query().
		Where(cluster.UIDEQ(clusterUID), cluster.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", merr.NotFound("not found")
		}
		clients.log.Error(fmt.Sprintf("cluster: %v", err))
		return "", merr.Server("request failed")
	}
	plain, err := clients.deps.box.decrypt(row.KubeconfigCipher)
	if err != nil {
		clients.log.Error(fmt.Sprintf("cluster kubeconfig: %v", err))
		return "", merr.Server("request failed")
	}
	return plain, nil
}

// invalidate 丢掉集群的缓存客户端。替换 kubeconfig 或删除集群后调用。
func (clients *clusterClients) invalidate(clusterUID string) {
	clients.mu.Lock()
	delete(clients.cached, clusterUID)
	clients.mu.Unlock()
}

// parseKubeconfig 解析 kubeconfig，并返回当前 context 指向的集群名。
// 只接受带 current-context 的文本；调用方拿到 rest.Config 后即可访问 API Server。
func parseKubeconfig(kubeconfig string) (*rest.Config, string, string, error) {
	kubeconfig = strings.TrimSpace(kubeconfig)
	if kubeconfig == "" {
		return nil, "", "", merr.Invalid("kubeconfig is empty")
	}
	parsed, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return nil, "", "", merr.Invalid("invalid kubeconfig")
	}
	if len(parsed.Clusters) == 0 || parsed.CurrentContext == "" {
		return nil, "", "", merr.Invalid("invalid kubeconfig")
	}
	kubeContext, ok := parsed.Contexts[parsed.CurrentContext]
	if !ok || kubeContext == nil || kubeContext.Cluster == "" {
		return nil, "", "", merr.Invalid("invalid kubeconfig")
	}
	clusterInfo, ok := parsed.Clusters[kubeContext.Cluster]
	if !ok || clusterInfo == nil {
		return nil, "", "", merr.Invalid("invalid kubeconfig")
	}
	// 只接受内嵌凭据。exec 和 auth-provider 会在 alioth 进程里执行租户提交的
	// 命令；tokenFile 与证书文件路径会读取 alioth 本机文件，tokenFile 的内容
	// 还会被当成 Bearer Token 发往 kubeconfig 指定的地址。这些都跨越了租户边界。
	if authInfo := parsed.AuthInfos[kubeContext.AuthInfo]; authInfo != nil {
		if authInfo.Exec != nil || authInfo.AuthProvider != nil || authInfo.TokenFile != "" ||
			authInfo.ClientCertificate != "" || authInfo.ClientKey != "" {
			return nil, "", "", merr.Invalid("kubeconfig must embed credentials; exec, auth-provider, tokenFile and certificate files are not supported")
		}
	}
	if clusterInfo.ProxyURL != "" {
		return nil, "", "", merr.Invalid("kubeconfig proxy-url is not supported")
	}
	// CA 也要内嵌：文件路径会让 client-go 读取 alioth 本机文件，形成租户可控的
	// 文件探测面（特殊文件还可能阻塞或放大内存）。
	if clusterInfo.CertificateAuthority != "" {
		return nil, "", "", merr.Invalid("kubeconfig must embed the certificate authority; certificate-authority file paths are not supported")
	}
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil || restConfig.Host == "" {
		return nil, "", "", merr.Invalid("invalid kubeconfig")
	}
	// 再清一遍，避免 client-go 后续版本从其它字段回填这些能力。
	restConfig.ExecProvider = nil
	restConfig.AuthProvider = nil
	restConfig.BearerTokenFile = ""
	restConfig.CertFile = ""
	restConfig.KeyFile = ""
	restConfig.CAFile = ""
	restConfig.Proxy = nil
	restConfig.Timeout = kubeRequestTimeout
	return restConfig, kubeContext.Cluster, parsed.CurrentContext, nil
}

// probeKubeconfig 解析并探测一份 kubeconfig，返回 API Server、版本和节点数。
// 节点数在读不到节点时按 0 返回；版本读不到才算探测失败。
func probeKubeconfig(ctx context.Context, logger *slog.Logger, kubeconfig string) (*bo.ClusterProbe, error) {
	restConfig, clusterName, contextName, err := parseKubeconfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	probeConfig := rest.CopyConfig(restConfig)
	probeConfig.Timeout = probeTimeout
	client, err := kubernetes.NewForConfig(probeConfig)
	if err != nil {
		return nil, merr.Invalid("invalid kubeconfig")
	}
	probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	info, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, kubeError(logger, err)
	}
	probe := &bo.ClusterProbe{
		Server:      probeConfig.Host,
		Version:     info.GitVersion,
		ClusterName: clusterName,
		ContextName: contextName,
	}
	nodes, err := client.CoreV1().Nodes().List(probeCtx, metav1.ListOptions{})
	if err != nil {
		// 只读版本也算连通；没有节点读取权限时节点数保持 0。
		if apierrors.IsForbidden(err) {
			return probe, nil
		}
		return nil, kubeError(logger, err)
	}
	probe.NodeCount = int32(len(nodes.Items))
	return probe, nil
}

// kubeError 把 client-go 的错误映射成统一错误。错误文本不带上游响应体，
// 避免把集群地址、凭据或内部细节返回给调用方。
func kubeError(logger *slog.Logger, err error) error {
	if err == nil {
		return nil
	}
	if kerrors.Reason(err) != "" {
		return err
	}
	switch {
	case apierrors.IsNotFound(err):
		return merr.NotFound("not found")
	case apierrors.IsUnauthorized(err):
		// 凭据被集群拒绝不是平台未认证，返回 401 会让浏览器清掉登录态。
		return merr.Invalid("kubeconfig credential was rejected by the cluster")
	case apierrors.IsForbidden(err):
		// 集群侧拒绝不是平台权限问题。返回 403 会让前端跳 /forbidden，
		// 用户就看不到受限 kubeconfig 的真实原因。
		return merr.Client("cluster credential is not allowed to perform this operation")
	case apierrors.IsInvalid(err), apierrors.IsBadRequest(err):
		return merr.Invalid("invalid request")
	case apierrors.IsConflict(err):
		return merr.Conflict("cluster resource conflict")
	case apierrors.IsTooManyRequests(err):
		return merr.Server("cluster is rate limited")
	case apierrors.IsServiceUnavailable(err):
		return merr.Server("cluster is unavailable")
	case isUnreachable(err):
		return merr.Client("cluster is unreachable")
	}
	if logger != nil {
		logger.Error(fmt.Sprintf("kubernetes: %v", err))
	}
	return merr.Server("request failed")
}

// isUnreachable 判断错误来自网络不可达或超时，而不是 API Server 给出的业务错误。
func isUnreachable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	if apierrors.IsTimeout(err) || apierrors.IsServerTimeout(err) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}
	var opErr *net.OpError
	return errors.As(err, &opErr)
}
