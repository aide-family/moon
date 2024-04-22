package fake

import (
	"context"
	"github.com/aide-family/moon/api/cluster/v1beta1"

	"github.com/aide-family/moon/app/kubemoon/internal/server/cluster"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ cluster.Client = &fakeClient{}

const name = "fake-client"

type fakeClient struct {
	name   string
	status cluster.Code
}

func (f fakeClient) RunStatus() cluster.Code {
	return f.status
}

func (f fakeClient) NetStatus() cluster.Code {
	return f.status
}

func (f fakeClient) Ready(ctx context.Context) (int, error) {
	return 200, nil
}

func (f fakeClient) Health(ctx context.Context) (int, error) {
	return 200, nil
}

func (f fakeClient) KubernetesVersion() (string, error) {
	return "v0.0.0", nil
}

func (f fakeClient) APIEnablements() ([]v1beta1.APIEnablement, error) {
	return nil, nil
}

func (f fakeClient) Start(ctx context.Context) error {
	return nil
}

func (f fakeClient) Stop() {
}

func (f fakeClient) Status() cluster.Code {
	return f.status
}

func (f fakeClient) Enable() {
}

func (f fakeClient) Disable() {
}

func (f fakeClient) Ping(ctx context.Context) error {
	return nil
}

func (f fakeClient) Name() string {
	return f.name
}

func (f fakeClient) Client() client.Client {
	return nil
}

func (f fakeClient) Cache() cache.Cache {
	return nil
}

func (f fakeClient) ApiExtensions() clientset.Interface {
	return nil
}

func (f fakeClient) Dynamic() dynamic.Interface {
	return nil
}

func (f fakeClient) RESTMapper() meta.RESTMapper {
	return nil
}

func (f fakeClient) Config() *rest.Config {
	return nil
}

func (f fakeClient) Discovery() discovery.DiscoveryInterface {
	return nil
}
