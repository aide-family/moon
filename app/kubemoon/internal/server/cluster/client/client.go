package client

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	klog "github.com/go-kratos/kratos/v2/log"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ clu.Client = &clientx{}

const (
	ReadyPath  = "/readyz"
	HealthPath = "/healthz"
)

type clientx struct {
	name       string
	ctx        context.Context
	cancelFunc context.CancelFunc
	netStatus  clu.Code
	runStatus  clu.Code
	scheme     *runtime.Scheme
	client     client.Client
	cache      cache.Cache
	mapper     meta.RESTMapper
	dynamic    dynamic.Interface
	discovery  discovery.DiscoveryInterface
	kubernetes kubernetes.Interface
	extensions clientset.Interface
	config     *rest.Config
}

// New returns a new clientx or error
// default status code is Stopped
func New(name string, config *rest.Config, scheme *runtime.Scheme, disable bool, options ...clu.InitOptions) (clu.Client, error) {
	var err error
	cli := new(clientx)
	copyConfig := *config
	copyConfig.Timeout = 0
	copyConfig.UserAgent = "kube-moon"
	cWithProtobuf := rest.CopyConfig(&copyConfig)
	cWithProtobuf.ContentType = runtime.ContentTypeProtobuf

	cli.name = name
	cli.scheme = scheme

	// TODO: new rest mapper
	cli.mapper = meta.NewDefaultRESTMapper(scheme.PreferredVersionAllGroups())

	if cli.client, err = client.New(cWithProtobuf, client.Options{Scheme: scheme}); err != nil {
		return nil, fmt.Errorf("failed to create runtime client: %s", err)
	}

	if cli.cache, err = cache.New(cWithProtobuf, cache.Options{Scheme: scheme}); err != nil {
		return nil, fmt.Errorf("failed to create runtime cache: %s", err)
	}

	if cli.dynamic, err = dynamic.NewForConfig(&copyConfig); err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %s", err)
	}

	if cli.extensions, err = clientset.NewForConfig(cWithProtobuf); err != nil {
		return nil, fmt.Errorf("failed to create api-extensions client: %s", err)
	}

	if cli.kubernetes, err = kubernetes.NewForConfig(cWithProtobuf); err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %s", err)
	}

	if cli.discovery, err = discovery.NewDiscoveryClientForConfig(cWithProtobuf); err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %s", err)
	}

	cli.runStatus = clu.Stopped
	if disable {
		cli.runStatus = clu.Disabled
	}

	for _, option := range options {
		if err = option(cli); err != nil {
			return nil, fmt.Errorf("failed to initialize options: %s", err)
		}
	}

	return cli, nil
}

func (c *clientx) Start(ctx context.Context) error {
	c.syncNetStatus(ctx)

	switch {
	case c.runStatus == clu.Disabled:
		klog.Infof("%s,needs to be enabled first", c)
	case c.runStatus >= clu.Running:
		klog.Infof("%s,already start", c)
	case c.runStatus == clu.Stopped:
		if c.netStatus == clu.Ready {
			c.ctx, c.cancelFunc = context.WithCancel(ctx)
			go func() {
				_ = c.cache.Start(c.ctx)
			}()
			c.runStatus = clu.Running
		}
		klog.Infof("%s", c)
	default:
		return fmt.Errorf("%s", c)
	}
	return nil
}

func (c *clientx) Stop() {
	if c.runStatus == clu.Running {
		c.cancelFunc()
		c.runStatus = clu.Stopped
		klog.Infof("%s", c)
	} else {
		klog.Infof("%s,no need stop", c)
	}
}

func (c *clientx) syncNetStatus(ctx context.Context) {
	code, err := c.Ready(ctx)
	if err != nil && code == http.StatusNotFound {
		code, err = c.Health(ctx)
	}
	switch {
	case err != nil:
		c.netStatus = clu.Offline
	case code != http.StatusOK:
		c.netStatus = clu.Unhealthy
	default:
		c.netStatus = clu.Ready
	}
}

func (c *clientx) Ready(ctx context.Context) (int, error) {
	var code int
	resp := c.discovery.RESTClient().Get().AbsPath(ReadyPath).Do(context.TODO()).StatusCode(&code)
	return code, resp.Error()
}

func (c *clientx) Health(ctx context.Context) (int, error) {
	var code int
	resp := c.discovery.RESTClient().Get().AbsPath(HealthPath).Do(context.TODO()).StatusCode(&code)
	return code, resp.Error()
}

func (c *clientx) KubernetesVersion() (string, error) {
	clusterVersion, err := c.discovery.ServerVersion()
	if err != nil {
		return "", err
	}
	return clusterVersion.GitVersion, nil
}

func (c *clientx) APIEnablements() ([]v1beta1.APIEnablement, error) {
	_, apiResourceList, err := c.discovery.ServerGroupsAndResources()
	if len(apiResourceList) == 0 {
		return nil, err
	}

	var apiEnablements []v1beta1.APIEnablement
	for _, list := range apiResourceList {
		var apiResources []v1beta1.APIResource
		for _, resource := range list.APIResources {
			if strings.Contains(resource.Name, "/") {
				continue
			}
			apiResource := v1beta1.APIResource{
				Name: resource.Name,
				Kind: resource.Kind,
			}
			apiResources = append(apiResources, apiResource)
		}
		sort.SliceStable(apiResources, func(i, j int) bool {
			return apiResources[i].Name < apiResources[j].Name
		})
		apiEnablements = append(apiEnablements, v1beta1.APIEnablement{GroupVersion: list.GroupVersion, Resources: apiResources})
	}
	sort.SliceStable(apiEnablements, func(i, j int) bool {
		return apiEnablements[i].GroupVersion < apiEnablements[j].GroupVersion
	})
	return apiEnablements, err
}

func (c *clientx) Name() string {
	return c.name
}

func (c *clientx) RunStatus() clu.Code {
	return c.runStatus
}

func (c *clientx) NetStatus() clu.Code {
	return c.netStatus
}

func (c *clientx) Client() client.Client {
	return c.client
}

func (c *clientx) Cache() cache.Cache {
	return c.cache
}

func (c *clientx) Kubernetes() kubernetes.Interface {
	return c.kubernetes
}

func (c *clientx) ApiExtensions() clientset.Interface {
	return c.extensions
}

func (c *clientx) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *clientx) RESTMapper() meta.RESTMapper {
	return c.mapper
}

func (c *clientx) Config() *rest.Config {
	return c.config
}

func (c *clientx) Discovery() discovery.DiscoveryInterface {
	return c.discovery
}

func (c *clientx) String() string {
	return fmt.Sprintf("clientx %s is %s %s", c.name, c.netStatus, c.runStatus)
}
