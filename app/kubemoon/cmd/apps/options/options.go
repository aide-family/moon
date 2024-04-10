package options

import (
	clusterv1beta1 "github.com/aide-family/moon/api/cluster/v1beta1"
	"github.com/aide-family/moon/app/kubemoon/cmd/apps/config"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(clusterv1beta1.AddToScheme(scheme))
}

// TODO: @sumengzs@gmail.com
type Options struct {
}

func NewOptions() *Options {
	return &Options{}
}

func (s *Options) Flags() (fss cliflag.NamedFlagSets) {
	return fss
}

func (s *Options) Validate() []error {
	var errors []error
	return errors
}

func (s *Options) Complete() (*config.Config, error) {
	return nil, nil
}
