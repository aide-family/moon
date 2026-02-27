package connect

import (
	"context"
	"sync"

	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	kuberegistry "github.com/go-kratos/kratos/contrib/registry/kubernetes/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientV3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/dir"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
)

const defaultKubeConfig = "~/.kube/config"

func init() {
	globalRegistry.RegisterReportFactory(config.ReportConfig_KUBERNETES, buildReportFromKubernetes)
	globalRegistry.RegisterReportFactory(config.ReportConfig_ETCD, buildReportFromEtcd)
	globalRegistry.RegisterReportFactory(config.ReportConfig_REPORT_TYPE_UNKNOWN, newDefaultReport)
}

type Report interface {
	registry.Registrar
	registry.Discovery
}

// safeReport wraps a Report to provide thread-safe access to Register and Deregister operations.
// This prevents concurrent map read/write errors when multiple apps share the same registry.
type safeReport struct {
	registrar registry.Registrar
	discovery registry.Discovery
	mu        sync.Mutex
}

func (s *safeReport) Register(ctx context.Context, service *registry.ServiceInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.registrar.Register(ctx, service)
}

func (s *safeReport) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.registrar.Deregister(ctx, service)
}

func (s *safeReport) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.discovery.GetService(ctx, serviceName)
}

func (s *safeReport) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.discovery.Watch(ctx, serviceName)
}

// NewReport creates a new report.
// If the report factory is not registered, it will return an error.
// The report is not closed, you need to call the returned function to close the report.
func NewReport(c *config.ReportConfig) (Report, func() error, error) {
	factory, ok := globalRegistry.GetReportFactory(c.GetReportType())
	if !ok {
		return nil, nil, merr.ErrorInternalServer("report factory not registered")
	}
	report, closeFn, err := factory(c)
	if err != nil {
		return nil, nil, err
	}
	// Wrap the report with thread-safe wrapper to prevent concurrent map access
	safe := &safeReport{
		registrar: report,
		discovery: report,
	}
	return safe, closeFn, nil
}

func buildReportFromKubernetes(c *config.ReportConfig) (Report, func() error, error) {
	kubeConfig := &config.KubernetesOptions{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), kubeConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal kubernetes config failed: %v", err)
		}
	}
	configPath := kubeConfig.GetKubeConfig()
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		if configPath == "" {
			configPath = defaultKubeConfig
		}
		restConfig, err = clientcmd.BuildConfigFromFlags("", dir.ExpandHomeDir(configPath))
		if err != nil {
			return nil, nil, merr.ErrorInternalServer("build kubernetes config failed: %v", err)
		}
	}
	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil, merr.ErrorInternalServer("create kubernetes client set failed: %v", err)
	}

	return kuberegistry.NewRegistry(clientSet, c.GetNamespace()), func() error { return nil }, nil
}

func buildReportFromEtcd(c *config.ReportConfig) (Report, func() error, error) {
	etcdConfig := &config.ETCDOptions{}
	if pointer.IsNotNil(c.GetOptions()) {
		if err := anypb.UnmarshalTo(c.GetOptions(), etcdConfig, proto.UnmarshalOptions{Merge: true}); err != nil {
			return nil, nil, merr.ErrorInternalServer("unmarshal etcd config failed: %v", err)
		}
	}
	client, err := clientV3.New(clientV3.Config{
		Endpoints:   strutil.SplitSkipEmpty(etcdConfig.GetEndpoints(), ","),
		Username:    etcdConfig.GetUsername(),
		Password:    etcdConfig.GetPassword(),
		DialTimeout: etcdConfig.GetDialTimeout().AsDuration(),
	})
	if err != nil {
		return nil, nil, merr.ErrorInternalServer("create etcd client failed: %v", err)
	}
	return etcd.New(client, etcd.Namespace(c.GetNamespace())), client.Close, nil
}

func newDefaultReport(c *config.ReportConfig) (Report, func() error, error) {
	return &defaultReport{}, func() error { return nil }, nil
}

type defaultReport struct{}

// Deregister implements [Report].
func (d *defaultReport) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	return nil
}

// GetService implements [Report].
func (d *defaultReport) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	return nil, nil
}

// Register implements [Report].
func (d *defaultReport) Register(ctx context.Context, service *registry.ServiceInstance) error {
	return nil
}

// Watch implements [Report].
func (d *defaultReport) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return nil, nil
}
