package client

import (
	"context"
	"fmt"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	clu "github.com/aide-family/moon/app/kubemoon/internal/cluster"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
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
	status     clu.Code
	scheme     *runtime.Scheme
	client     client.Client
	cache      cache.Cache
	mapper     meta.RESTMapper
	dynamic    dynamic.Interface
	discovery  discovery.DiscoveryInterface
	extensions clientset.Interface
	config     *rest.Config
}

// New returns a new clientx or error
// default status code is Stopped
func New(name string, config *rest.Config, scheme *runtime.Scheme, options ...clu.InitOptions) (clu.Client, error) {
	var err error
	cli := new(clientx)
	cli.name = name
	// TODO: new rest mapper
	cli.mapper = meta.NewDefaultRESTMapper(scheme.PreferredVersionAllGroups())

	if cli.client, err = client.New(config, client.Options{Scheme: scheme}); err != nil {
		return nil, fmt.Errorf("failed to create runtime client: %s", err)
	}

	if cli.cache, err = cache.New(config, cache.Options{Scheme: scheme}); err != nil {
		return nil, fmt.Errorf("failed to create runtime cache: %s", err)
	}

	if cli.dynamic, err = dynamic.NewForConfig(config); err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %s", err)
	}

	if cli.extensions, err = clientset.NewForConfig(config); err != nil {
		return nil, fmt.Errorf("failed to create api-extensions client: %s", err)
	}
	cli.extensions.ApiextensionsV1beta1().CustomResourceDefinitions()

	if cli.discovery, err = discovery.NewDiscoveryClientForConfig(config); err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %s", err)
	}

	cli.status = clu.Stopped
	cli.scheme = scheme

	for _, option := range options {
		if err = option(cli); err != nil {
			return nil, fmt.Errorf("failed to initialize options: %s", err)
		}
	}

	return cli, nil
}

func (c *clientx) Start(ctx context.Context) error {
	switch {
	case c.status == clu.Disabled:
		klog.Infof("%s,needs to be enabled first", c)
	case c.status >= clu.Started:
		klog.Infof("%s,already start", c)
	case c.status == clu.Stopped:
		c.ctx, c.cancelFunc = context.WithCancel(ctx)
		go func() {
			_ = c.cache.Start(c.ctx)
		}()
		c.status = clu.Started
		klog.Infof("%s", c)
	default:
		return fmt.Errorf("%s", c)
	}
	return nil
}

func (c *clientx) Stop() {
	if c.status == clu.Disabled {
		klog.Infof("%s,no need stop", c)
	}
	if c.status > clu.Stopped {
		c.cancelFunc()
		c.status = clu.Stopped
	}
	klog.Infof("%s", c)
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

func (c *clientx) Status() clu.Code {
	return c.status
}

func (c *clientx) Enable() {
	if c.status == clu.Disabled {
		c.status = clu.Stopped
	}
}

func (c *clientx) Disable() {
	if c.status >= clu.Started {
		c.Stop()
	}
	c.status = clu.Disabled
}

func (c *clientx) Client() client.Client {
	return c.client
}

func (c *clientx) Cache() cache.Cache {
	return c.cache
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
	return fmt.Sprintf("clientx %s is %s", c.name, c.status)
}
